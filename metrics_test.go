package main

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestIncreaseMySQLErrorCounter(t *testing.T) {
	reg := prometheus.NewRegistry()
	LoadMetricEnvs()
	DefineMetrics()
	reg.MustRegister(errorCounter)

	IncreaseErrorCount("08S01")
	IncreaseErrorCount("08S01")
	IncreaseErrorCount("08S02")

	value := testutil.ToFloat64(errorCounter.WithLabelValues("08S01"))
	if value != 2 {
		t.Fatalf("expected 2 for error code 08S01, got %v", value)
	}

	value = testutil.ToFloat64(errorCounter.WithLabelValues("08S02"))
	if value != 1 {
		t.Fatalf("expected 1 for error code 08S02, got %v", value)
	}

}
