package observability

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Metrics struct {
	ingestTotal       atomic.Int64
	ingestErrors      atomic.Int64
	searchTotal       atomic.Int64
	searchErrors      atomic.Int64
	extractionTotal   atomic.Int64
	jobQueueDepth     atomic.Int64
	ingestLatencies   latencyHistogram
	searchLatencies   latencyHistogram
}

type latencyHistogram struct {
	buckets []float64
	counts  []atomic.Int64
	sum     atomic.Uint64
	count   atomic.Int64
}

func newLatencyHistogram(buckets []float64) latencyHistogram {
	return latencyHistogram{
		buckets: buckets,
		counts:  make([]atomic.Int64, len(buckets)+1),
	}
}

func (h *latencyHistogram) observe(seconds float64) {
	h.sum.Add(math.Float64bits(seconds))
	h.count.Add(1)
	for i, b := range h.buckets {
		if seconds <= b {
			h.counts[i].Add(1)
			return
		}
	}
	h.counts[len(h.buckets)].Add(1)
}

func NewMetrics() *Metrics {
	buckets := []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0}
	return &Metrics{
		ingestLatencies: newLatencyHistogram(buckets),
		searchLatencies: newLatencyHistogram(buckets),
	}
}

func (m *Metrics) RecordIngest(duration time.Duration, err bool) {
	m.ingestTotal.Add(1)
	if err {
		m.ingestErrors.Add(1)
	}
	m.ingestLatencies.observe(duration.Seconds())
}

func (m *Metrics) RecordSearch(duration time.Duration, err bool) {
	m.searchTotal.Add(1)
	if err {
		m.searchErrors.Add(1)
	}
	m.searchLatencies.observe(duration.Seconds())
}

func (m *Metrics) RecordExtraction() {
	m.extractionTotal.Add(1)
}

func (m *Metrics) SetJobQueueDepth(depth int64) {
	m.jobQueueDepth.Store(depth)
}

func (m *Metrics) Prometheus() string {
	var b strings.Builder
	writeCounter := func(name, help string, value int64) {
		fmt.Fprintf(&b, "# HELP %s %s\n", name, help)
		fmt.Fprintf(&b, "# TYPE %s counter\n", name)
		fmt.Fprintf(&b, "%s %d\n", name, value)
	}
	writeGauge := func(name, help string, value int64) {
		fmt.Fprintf(&b, "# HELP %s %s\n", name, help)
		fmt.Fprintf(&b, "# TYPE %s gauge\n", name)
		fmt.Fprintf(&b, "%s %d\n", name, value)
	}

	writeCounter("brainy_ingest_total", "Total ingest requests.", m.ingestTotal.Load())
	writeCounter("brainy_ingest_errors_total", "Total ingest errors.", m.ingestErrors.Load())
	writeCounter("brainy_search_total", "Total search requests.", m.searchTotal.Load())
	writeCounter("brainy_search_errors_total", "Total search errors.", m.searchErrors.Load())
	writeCounter("brainy_extraction_total", "Total extractions performed.", m.extractionTotal.Load())
	writeGauge("brainy_job_queue_depth", "Current pending job queue depth.", m.jobQueueDepth.Load())

	writeHistogram := func(name, help string, h *latencyHistogram) {
		fmt.Fprintf(&b, "# HELP %s %s\n", name, help)
		fmt.Fprintf(&b, "# TYPE %s histogram\n", name)
		for i, bucket := range h.buckets {
			fmt.Fprintf(&b, "%s_bucket{le=\"%.3f\"} %d\n", name, bucket, h.counts[i].Load())
		}
		fmt.Fprintf(&b, "%s_bucket{le=\"+Inf\"} %d\n", name, h.counts[len(h.buckets)].Load())
		fmt.Fprintf(&b, "%s_sum %f\n", name, math.Float64frombits(h.sum.Load()))
		fmt.Fprintf(&b, "%s_count %d\n", name, h.countLoaded())
	}

	writeHistogram("brainy_ingest_latency_seconds", "Ingest latency distribution.", &m.ingestLatencies)
	writeHistogram("brainy_search_latency_seconds", "Search latency distribution.", &m.searchLatencies)
	return b.String()
}

func (h *latencyHistogram) countLoaded() int64 {
	var total int64
	for i := range h.counts {
		total += h.counts[i].Load()
	}
	return total
}

var globalMetrics = NewMetrics()
var metricsMu sync.Mutex

func Global() *Metrics {
	metricsMu.Lock()
	defer metricsMu.Unlock()
	return globalMetrics
}
