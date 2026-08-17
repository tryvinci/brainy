package observability

import (
	"regexp"
	"strconv"
	"testing"
	"time"
)

func TestLatencyHistogramSumAddsFloats(t *testing.T) {
	m := NewMetrics()
	m.RecordSearch(100*time.Millisecond, false)
	m.RecordSearch(200*time.Millisecond, false)
	got := prometheusSum(t, m.Prometheus(), "brainy_search_latency_seconds_sum")
	want := 0.3
	if diff := got - want; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("search latency sum: got %v want %v (bit-pattern add would not equal 0.3)", got, want)
	}
}

func prometheusSum(t *testing.T, text, metric string) float64 {
	t.Helper()
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(metric) + ` ([0-9.eE+-]+)`)
	m := re.FindStringSubmatch(text)
	if m == nil {
		t.Fatalf("metric %s not found in:\n%s", metric, text)
	}
	v, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
