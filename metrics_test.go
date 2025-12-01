package main

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestIncreaseMySQLErrorCounter(t *testing.T) {
	labels := prometheus.Labels{
		"namespace":    "dev",
		"cluster_name": "clusterA",
		"error_code":   "1045",
	}

	// Helper to read the metric value
	getValue := func() float64 {
		return testutil.ToFloat64(MySQLErrorCounter.With(labels))
	}
	IncreaseMySQLErrorCounter(labels)
	if v := getValue(); v != 1 {
		t.Fatalf("expected 1 after first increment, got %v", v)
	}

	IncreaseMySQLErrorCounter(labels)
	if v := getValue(); v != 2 {
		t.Fatalf("expected 2 after second increment, got %v", v)
	}

	IncreaseMySQLErrorCounter(labels)
	if v := getValue(); v != 3 {
		t.Fatalf("expected 3 after second increment, got %v", v)
	}
}

func TestEmptyLabelMySQLErrorCounter(t *testing.T) {
	labels := prometheus.Labels{
		"namespace":    "dev",
		"error_code":   "1045",
		"cluster_name": "",
	}

	// Helper to read the metric value
	getValue := func() float64 {
		return testutil.ToFloat64(MySQLErrorCounter.With(labels))
	}
	IncreaseMySQLErrorCounter(labels)
	if v := getValue(); v != 1 {
		t.Fatalf("expected 1 after second increment, got %v", v)
	}
}
