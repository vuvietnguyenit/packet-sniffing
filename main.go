package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// flags define
var iface string
var port int

var slogger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func main() {

	flag.StringVar(&iface, "i", "eth0", "Network interface to capture packets")
	flag.IntVar(&port, "p", 3306, "MySQL server port")
	flag.Parse()

	snaplen := int32(65535)
	promisc := true
	timeout := pcap.BlockForever

	handle, err := pcap.OpenLive(iface, snaplen, promisc, timeout)
	if err != nil {
		slogger.Error("error when open pcap live", "error", err.Error())
	}
	defer handle.Close()

	// Pre-filter capture packets *from* MySQL server
	filter := fmt.Sprintf("tcp src port %d", port)
	if err := handle.SetBPFFilter(filter); err != nil {
		slogger.Error("Failed to set BPF filter:", "error", err)
	}
	slogger.Info("listening for MySQL error packets on", "iface", iface, "port", port)

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		processPacket(packet)
	}
}

func processPacket(packet gopacket.Packet) {
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
	if tcpLayer == nil || ipv4Layer == nil {
		return
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
			slogger.Info("MySQL ERR packet",
				"src_ip", ip.SrcIP,
				"src_port", tcp.SrcPort,
				"dst_ip", ip.DstIP,
				"dst_port", tcp.DstPort,
				"length", length,
				"error_message", string(payload[offset+4:offset+4+length]),
			)
		}
		offset += 4 + length
	}
}
