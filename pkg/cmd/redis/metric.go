package redis

import (
	"git.itim.vn/docker/redis-error-sniffer/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	redisErrorMetricName = "redis_error_response_count"
)

var redisErrorLabels = []string{"error_code"}

type redisMetric struct {
	metric *prometheus.CounterVec
}

func NewRedisMetric() *redisMetric {
	m := &redisMetric{
		metric: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: redisErrorMetricName,
				Help: "Number of Redis errors received",
			},
			redisErrorLabels,
		),
	}

	metrics.Reg.MustRegister(m.metric)
	return m
}

func (r *redisMetric) IncreaseErrorCount(errorCode string) {
	r.metric.WithLabelValues(errorCode).Inc()
}
