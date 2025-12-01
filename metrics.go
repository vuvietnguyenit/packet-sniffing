package main

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var MySQLErrorCounter = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "mysql_error_response_count",
		Help: "Number of MySQL error responses",
	},
	[]string{
		"namespace",
		"cluster_name", // REQUIRED: MySQL cluster name
		"error_code",
	},
)
var reg = prometheus.NewRegistry()

func IncreaseMySQLErrorCounter(labels prometheus.Labels) {
	MySQLErrorCounter.With(labels).Inc()
}

func NewMetricHandler(port int) http.Handler {
	metrics := promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			EnableOpenMetrics: false,
		},
	)
	slogger.Info("starting metrics server", "port", port)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slogger.Debug("scrape request", "from", r.RemoteAddr, "path", r.URL.Path)
		metrics.ServeHTTP(w, r)
	})
}

func RunExporterMetricsServer(port int) {
	reg.MustRegister(MySQLErrorCounter)
	http.Handle("/metrics", NewMetricHandler(port))
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		slogger.Error("failed to start metrics server", "error", err)
	}
}
