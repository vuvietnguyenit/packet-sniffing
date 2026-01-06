package mysql

import (
	"encoding/binary"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"sync"

	"time"

	"git.itim.vn/docker/redis-error-sniffer/pkg/metrics"
	"git.itim.vn/docker/redis-error-sniffer/utils"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/spf13/cobra"
)

type mysqlCmdFlags struct {
	UseDNS        bool
	CacheSize     int
	CacheDuration time.Duration
	Nameserver    string
	Iface         string
	Port          int
	ExporterPort  int
	Verbose       bool
}

var cfg mysqlCmdFlags

var MysqlErrorResponseCmd = &cobra.Command{
	Use:   "mysql-error-response",
	Short: "Record error responses are sent to mysql-client and export it as Prometheus exporter base on libpcap",

	Run: func(cmd *cobra.Command, args []string) {
		utils.InitLogger(cfg.Verbose)
		var wg sync.WaitGroup
		wg.Go(func() {
			mysqlSniff()
		})
		if cfg.ExporterPort > 0 {
			wg.Go(func() {
				metrics.RunExporterMetricsServer(cfg.ExporterPort)
			})
		}
		wg.Wait()
	},
}

var alphaNumRegex = regexp.MustCompile(`^[A-Z0-9]+$`)

// isAlphaNumASCIIRegex returns true only if the string matches [A-Z0-9]+
func isValidStr(s string) bool {
	return alphaNumRegex.MatchString(s)
}

func init() {

	MysqlErrorResponseCmd.Flags().BoolVar(&cfg.UseDNS, "use-dns", false, "Resolve IP to domain using reverse DNS lookup")
	MysqlErrorResponseCmd.Flags().IntVar(&cfg.CacheSize, "cache-size", 4096, "DNS cache size")
	MysqlErrorResponseCmd.Flags().DurationVar(&cfg.CacheDuration, "cache-duration", 5*time.Minute, "DNS cache expiration")
	MysqlErrorResponseCmd.Flags().StringVar(&cfg.Nameserver, "nameserver", "", "Custom DNS server (e.g., 8.8.8.8:53)")

	// Packet processing
	MysqlErrorResponseCmd.Flags().StringVar(&cfg.Iface, "iface", "eth0", "Network interface to monitor")
	MysqlErrorResponseCmd.Flags().IntVar(&cfg.Port, "port", 3306, "MySQL port to trace")

	// Exporter
	MysqlErrorResponseCmd.Flags().IntVar(&cfg.ExporterPort, "exporter-port", 2112, "Prometheus exporter port")
	MysqlErrorResponseCmd.Flags().BoolVarP(&cfg.Verbose, "verbose", "v", false, "Enable verbose logging")
}

func mysqlSniff() {

	snaplen := int32(65535)
	promisc := true
	timeout := pcap.BlockForever

	handle, err := pcap.OpenLive(cfg.Iface, snaplen, promisc, timeout)
	if err != nil {
		utils.Slogger.Error("error when open pcap live", "error", err.Error())
		os.Exit(1)
	}
	defer handle.Close()
	// Pre-filter capture packets *from* MySQL server
	filter := fmt.Sprintf("tcp src port %d", cfg.Port)
	if err := handle.SetBPFFilter(filter); err != nil {
		utils.Slogger.Error("Failed to set BPF filter:", "error", err)
	}
	utils.Slogger.Info("listening for MySQL error packets on", "iface", cfg.Iface, "port", cfg.Port)
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	metric := NewMySQLMetric()
	for packet := range packetSource.Packets() {
		mysqlInspect(packet, metric)
	}
}

func mysqlInspect(packet gopacket.Packet, metric *mySQLMetric) {
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
	if tcpLayer == nil || ipv4Layer == nil {
		return
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

	msgIdx := 8 + 5 // from this index to remaining data
	msg := payload[msgIdx:]
	metric.IncreaseErrorCount(strconv.Itoa(int(errCode)), string(stateCodeByteArr))

	utils.Slogger.Info("MySQL ERR packet",
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
