package observability

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all application metrics
type Metrics struct {
	RequestsTotal    *prometheus.CounterVec
	RequestDuration  *prometheus.HistogramVec
	RequestsInFlight *prometheus.GaugeVec
	RequestSize      *prometheus.HistogramVec
	ResponseSize     *prometheus.HistogramVec
	UptimeSeconds    prometheus.Gauge
	MemoryUsageBytes prometheus.Gauge
	GoroutineCount   prometheus.Gauge
}

// NewMetrics creates and registers all metrics
func NewMetrics(namespace, subsystem string) *Metrics {
	return &Metrics{
		RequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "requests_total",
				Help:      "Total requests by method/status/transport",
			},
			[]string{"method", "status", "transport"},
		),
		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "request_duration_seconds",
				Help:      "Request duration in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "transport"},
		),
		RequestsInFlight: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "requests_in_flight",
				Help:      "Requests currently being processed",
			},
			[]string{"transport"},
		),
		RequestSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "request_size_bytes",
				Help:      "Request size in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "transport"},
		),
		ResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "response_size_bytes",
				Help:      "Response size in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "transport"},
		),
		UptimeSeconds: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "uptime_seconds",
				Help:      "Server uptime in seconds",
			},
		),
		MemoryUsageBytes: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "memory_usage_bytes",
				Help:      "Memory usage in bytes",
			},
		),
		GoroutineCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "goroutines",
				Help:      "Number of goroutines",
			},
		),
	}
}

// RecordRequest records a request with its duration and status
func (m *Metrics) RecordRequest(method, transport, status string, duration time.Duration) {
	m.RequestsTotal.WithLabelValues(method, status, transport).Inc()
	m.RequestDuration.WithLabelValues(method, transport).Observe(duration.Seconds())
}

// RecordRequestSize records request size
func (m *Metrics) RecordRequestSize(method, transport string, size int) {
	m.RequestSize.WithLabelValues(method, transport).Observe(float64(size))
}

// RecordResponseSize records response size
func (m *Metrics) RecordResponseSize(method, transport string, size int) {
	m.ResponseSize.WithLabelValues(method, transport).Observe(float64(size))
}

// UpdateUptime updates the uptime metric
func (m *Metrics) UpdateUptime(startTime time.Time) {
	m.UptimeSeconds.Set(time.Since(startTime).Seconds())
}

// UpdateMemoryUsage updates the memory usage metric
func (m *Metrics) UpdateMemoryUsage(bytes uint64) {
	m.MemoryUsageBytes.Set(float64(bytes))
}

// UpdateGoroutineCount updates the goroutine count metric
func (m *Metrics) UpdateGoroutineCount(count int) {
	m.GoroutineCount.Set(float64(count))
}
