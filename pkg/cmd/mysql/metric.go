package mysql

import (
	"unicode/utf8"

	"git.itim.vn/docker/packet-sniffer/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	mysqlErrorMetricName = "mysql_error_response_count"
)

var mysqlErrorLabels = []string{"error_code", "state_code"}

type mySQLMetric struct {
	metric *prometheus.CounterVec
}

func NewMySQLMetric() *mySQLMetric {
	m := &mySQLMetric{
		metric: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: mysqlErrorMetricName,
				Help: "Number of Redis errors received",
			},
			mysqlErrorLabels,
		),
	}

	metrics.Reg.MustRegister(m.metric)
	return m
}

func safeLabel(b string) string {
	if utf8.ValidString(b) {
		return string(b)
	}
	return "unknown" // return a fallback value
}

func (r *mySQLMetric) IncreaseErrorCount(errorCode string, stateCode string) {
	labels := map[string]string{
		"error_code": safeLabel(errorCode),
		"state_code": safeLabel(stateCode),
	}
	r.metric.With(labels).Inc()
}
