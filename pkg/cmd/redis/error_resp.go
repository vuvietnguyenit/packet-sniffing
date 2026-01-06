package redis

import (
	"fmt"
	"os"
	"regexp"
	"sync"

	"git.itim.vn/docker/packet-sniffer/pkg/metrics"
	"git.itim.vn/docker/packet-sniffer/utils"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/spf13/cobra"
)

var RedisErrorResponseCmd = &cobra.Command{
	Use:   "redis-error-response",
	Short: "Record error reponses on Redis instance to clients base and export to Prometheus metrics base on libpcap",
	Run: func(cmd *cobra.Command, args []string) {
		utils.InitLogger(cfg.Verbose)
		var wg sync.WaitGroup
		wg.Go(func() {
			redisSniff()
		})
		if cfg.ExporterPort > 0 {
			wg.Go(func() {
				metrics.RunExporterMetricsServer(cfg.ExporterPort)

			})
		}
		wg.Wait()

	},
}

type cmdFlags struct {
	Iface        string
	Port         int
	ExporterPort int
	Verbose      bool
}

var cfg cmdFlags

func init() {
	RedisErrorResponseCmd.Flags().StringVarP(&cfg.Iface, "iface", "i", "eth0", "Network interface to monitor")
	RedisErrorResponseCmd.Flags().IntVarP(&cfg.Port, "port", "p", 6379, "Redis port want to sniff")
	RedisErrorResponseCmd.Flags().IntVarP(&cfg.ExporterPort, "exporter-port", "e", 2112, "Prometheus exporter port")
	RedisErrorResponseCmd.Flags().BoolVarP(&cfg.Verbose, "verbose", "v", false, "Enable verbose logging")
}

func redisSniff() {

	snaplen := int32(65535)
	promisc := true
	timeout := pcap.BlockForever

	handle, err := pcap.OpenLive(cfg.Iface, snaplen, promisc, timeout)
	if err != nil {
		utils.Slogger.Error("error when open pcap live", "error", err.Error())
		os.Exit(1)
	}
	defer handle.Close()
	filter := fmt.Sprintf("tcp src port %d", cfg.Port)
	if err := handle.SetBPFFilter(filter); err != nil {
		utils.Slogger.Error("Failed to set BPF filter:", "error", err)
	}
	utils.Slogger.Info("listening packets on", "iface", cfg.Iface, "port", cfg.Port)
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	redisMetric := NewRedisMetric()

	for packet := range packetSource.Packets() {
		redisInspect(packet, redisMetric)
	}
}

var errRe = regexp.MustCompile(`-(\S+) (.+)`)

func redisInspect(pkt gopacket.Packet, metric *redisMetric) {
	tcpLayer := pkt.Layer(layers.LayerTypeTCP)
	ipv4Layer := pkt.Layer(layers.LayerTypeIPv4)
	if tcpLayer == nil || ipv4Layer == nil {
		return
	}
	ip := ipv4Layer.(*layers.IPv4)
	tcp := tcpLayer.(*layers.TCP)
	payload := tcp.Payload
	if len(payload) == 0 {
		return
	}
	srcIP := ip.SrcIP.String()
	dstIp := ip.DstIP.String()

	// just consider packets only has "-" character (0x2d) at first byte
	// and must contains CRLF (\r\n)
	if payload[0] == 0x2d && payload[len(payload)-1] == 0xa && payload[len(payload)-2] == 0xd {
		msg := string(payload)
		matches := errRe.FindStringSubmatch(msg)
		utils.Slogger.Info("MySQL ERR packet",
			"src_ip", srcIP,
			"src_port", tcp.SrcPort,
			"dst_ip", dstIp,
			"dst_port", tcp.DstPort,
			"length", len(payload),
			"error_message", msg,
		)
		if matches == nil {
			utils.Slogger.Warn("unable to parse message", "message", msg)
			return
		}
		errorPrefix := matches[1]
		metric.IncreaseErrorCount(errorPrefix)
	}

}
