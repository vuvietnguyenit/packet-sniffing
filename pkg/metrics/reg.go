package metrics

import (
	"fmt"
	"net/http"
	"sync"

	"git.itim.vn/docker/redis-error-sniffer/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	defaultLabels prometheus.Labels
	once          sync.Once
)
var Reg = prometheus.NewRegistry()

func NewMetricHandler(port int) http.Handler {
	metrics := promhttp.HandlerFor(
		Reg,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	)
	utils.Slogger.Info("starting metrics server", "port", port)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.Slogger.Debug("scrape request", "from", r.RemoteAddr, "path", r.URL.Path)
		metrics.ServeHTTP(w, r)
	})
}

func RunExporterMetricsServer(port int) {
	http.Handle("/metrics", NewMetricHandler(port))
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		utils.Slogger.Error("failed to start metrics server", "error", err)
	}
}
