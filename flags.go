// flags.go
package main

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

// Config holds parsed CLI options.
type Config struct {
	UseDNS        bool
	CacheSize     int
	CacheDuration time.Duration
	Nameserver    string
	Iface         string
	Port          int
	ExporterPort  int
	Verbose       bool
}

func ParseFlags() (*Config, error) {
	cfg := &Config{}

	pflag.BoolVar(&cfg.UseDNS, "use-dns", false, "Resolve IP to domain using reverse DNS lookup")
	pflag.IntVar(&cfg.CacheSize, "cache-size", 4096, "DNS cache size")
	pflag.DurationVar(&cfg.CacheDuration, "cache-duration", 5*time.Minute, "DNS cache expiration")
	pflag.StringVar(&cfg.Nameserver, "nameserver", "", "Custom DNS server (e.g., 8.8.8.8:53)")

	// Packet processing flags
	pflag.StringVar(&cfg.Iface, "iface", "eth0", "Network interface to monitor")
	pflag.IntVar(&cfg.Port, "port", 3306, "MySQL port to trace")

	// Exporter flag (required)
	pflag.IntVar(&cfg.ExporterPort, "exporter-port", 2112, "Prometheus exporter port (required)")
	pflag.BoolVarP(&cfg.Verbose, "verbose", "v", false, "Enable verbose logging")

	pflag.Parse()

	// ------------------
	// Validation Section
	// ------------------
	if cfg.ExporterPort == 0 {
		return nil, fmt.Errorf("--exporter-port is required and must be > 0")
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return nil, fmt.Errorf("invalid --port value")
	}
	if cfg.Iface == "" {
		return nil, fmt.Errorf("--iface cannot be empty")
	}

	return cfg, nil
}
