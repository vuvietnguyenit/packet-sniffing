package main

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var slogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

var alphaNumRegex = regexp.MustCompile(`^[A-Z0-9]+$`)

// isAlphaNumASCIIRegex returns true only if the string matches [A-Z0-9]+
func isValidStr(s string) bool {
	return alphaNumRegex.MatchString(s)
}

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

	// error code processing
	oxffIdx := 4
	if len(payload) < 14 {
		return
	}
	if payload[oxffIdx] != 0xff { // not found ff byte
		return
	}
	if payload[7] != 0x23 { // a # character
		return
	}

	errCodeByteArr := make([]byte, 2)
	errCodeByteArr[0] = payload[oxffIdx+1]
	errCodeByteArr[1] = payload[oxffIdx+2]

	errCode := binary.LittleEndian.Uint16(errCodeByteArr)

	// state code processing
	statecodeIdx := 8 // start index at 7, but we need skip # character -> 7 + 1
	stateCodeByteArr := make([]byte, 5)
	for i := range stateCodeByteArr {
		stateCodeByteArr[i] = payload[statecodeIdx+i]
	}
	if !isValidStr(string(stateCodeByteArr)) {
		return
	}

	srcIP := ip.SrcIP.String()
	dstIp := ip.DstIP.String()
	if dnsResolved {
		srcIP = strings.Join(res.ReverseLookup(srcIP), ",")
		dstIp = strings.Join(res.ReverseLookup(dstIp), ",")
	}

	msgIdx := 8 + 5 // from this index to remaining data
	msg := payload[msgIdx:]
	IncreaseErrorCount(strconv.Itoa(int(errCode)), string(stateCodeByteArr))

	slogger.Info("MySQL ERR packet",
		"src_ip", srcIP,
		"src_port", tcp.SrcPort,
		"dst_ip", dstIp,
		"dst_port", tcp.DstPort,
		"length", len(payload),
		"state_code", string(stateCodeByteArr),
		"err_code", errCode,
		"error_message", string(msg),
	)
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
