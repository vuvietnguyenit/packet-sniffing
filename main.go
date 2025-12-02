package main

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var slogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

func packetProcessing(cfg *Config) {
	res := &Resolver{}
	if cfg.UseDNS {
		enableCache := true
		if cfg.CacheSize <= 0 {
			enableCache = false
		}
		res = NewResolver(cfg.Nameserver, enableCache, cfg.CacheSize, int(cfg.CacheDuration))
	}

	snaplen := int32(65535)
	promisc := true
	timeout := pcap.BlockForever

	handle, err := pcap.OpenLive(cfg.Iface, snaplen, promisc, timeout)
	if err != nil {
		slogger.Error("error when open pcap live", "error", err.Error())
		os.Exit(1)
	}
	defer handle.Close()
	// Pre-filter capture packets *from* MySQL server
	filter := fmt.Sprintf("tcp src port %d", cfg.Port)
	if err := handle.SetBPFFilter(filter); err != nil {
		slogger.Error("Failed to set BPF filter:", "error", err)
	}
	slogger.Info("listening for MySQL error packets on", "iface", cfg.Iface, "port", cfg.Port)
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		inspect(packet, res)
	}
}

var errRe = regexp.MustCompile(`#([A-Za-z0-9]{5})(.*)`)

func inspect(packet gopacket.Packet, res *Resolver) {
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
	if tcpLayer == nil || ipv4Layer == nil {
		return
	}
	dnsResolved := false
	if res == nil {
		dnsResolved = true
	}

	ip := ipv4Layer.(*layers.IPv4)
	tcp := tcpLayer.(*layers.TCP)
	payload := tcp.Payload

	if len(payload) == 0 {
		return
	}

	offset := 0
	for offset+4 <= len(payload) {
		length := int(payload[offset]) | int(payload[offset+1])<<8 | int(payload[offset+2])<<16
		if offset+4+length > len(payload) {
			break
		} // incomplete packet
		if payload[offset+4] == 0xff {
			srcIP := ip.SrcIP.String()
			dstIp := ip.DstIP.String()
			if dnsResolved {
				srcIP = strings.Join(res.ReverseLookup(srcIP), ",")
				dstIp = strings.Join(res.ReverseLookup(dstIp), ",")
			}
			msg := string(payload[offset+4 : offset+4+length])
			slogger.Info("MySQL ERR packet",
				"src_ip", srcIP,
				"src_port", tcp.SrcPort,
				"dst_ip", dstIp,
				"dst_port", tcp.DstPort,
				"length", length,
				"error_message", msg,
			)
			matches := errRe.FindStringSubmatch(msg)
			if matches == nil {
				slogger.Warn("unable to parse MySQL error code from message", "message", msg)
				return
			}
			errorCode := matches[1]
			IncreaseErrorCount(errorCode)

		}
		offset += 4 + length
	}
}

func main() {
	cfg, err := ParseFlags()
	if err != nil {
		slogger.Error("invalid flags", "error", err)
		os.Exit(1)
	}
	if cfg.Verbose {
		slogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	} else {
		slogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	var wg sync.WaitGroup
	wg.Go(func() {
		packetProcessing(cfg)
	})
	if cfg.ExporterPort > 0 {
		wg.Go(func() {
			RunExporterMetricsServer(cfg.ExporterPort)
		})
	}
	wg.Wait()
}
