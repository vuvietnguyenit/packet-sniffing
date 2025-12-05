package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	defaultLabels prometheus.Labels
	errorCounter  *prometheus.CounterVec
	once          sync.Once
)

func LoadMetricEnvs() {
	labels := prometheus.Labels{}

	prefix := "MYSQL_ERROR_ECHO_METRIC_"

	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, prefix) {
			parts := strings.SplitN(kv, "=", 2)
			key := strings.TrimPrefix(parts[0], prefix)
			val := parts[1]
			labels[key] = val
		}
	}
	defaultLabels = labels
}

func DefineMetrics() {
	once.Do(func() {
		labelKeys := []string{"error_code", "state_code"}
		if len(defaultLabels) == 0 {
			slogger.Debug("environment labels isn't defined")
		}

		// Add default label keys
		for k := range defaultLabels {
			labelKeys = append(labelKeys, k)
		}
		slogger.Info("metric info", "labels_included", labelKeys)
		errorCounter = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mysql_error_response_count",
				Help: "Number of MySQL errors received",
			},
			labelKeys,
		)
	})
}
func mergeLabels(base prometheus.Labels, extra prometheus.Labels) prometheus.Labels {
	out := prometheus.Labels{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

func IncreaseErrorCount(errorCode string, stateCode string) {
	labels := mergeLabels(defaultLabels, prometheus.Labels{
		"error_code": errorCode,
		"state_code": stateCode,
	})
	errorCounter.With(labels).Inc()
}

var reg = prometheus.NewRegistry()

func NewMetricHandler(port int) http.Handler {
	metrics := promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	)
	slogger.Info("starting metrics server", "port", port)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slogger.Debug("scrape request", "from", r.RemoteAddr, "path", r.URL.Path)
		metrics.ServeHTTP(w, r)
	})
}

func RunExporterMetricsServer(port int) {
	LoadMetricEnvs()
	DefineMetrics()
	reg.MustRegister(errorCounter)

	http.Handle("/metrics", NewMetricHandler(port))
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		slogger.Error("failed to start metrics server", "error", err)
	}
}
