package observability

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Request metrics
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_requests_total",
			Help: "Total number of MCP requests",
		},
		[]string{"method", "status", "transport"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "transport"},
	)

	requestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_request_size_bytes",
			Help:    "Request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "transport"},
	)

	responseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_response_size_bytes",
			Help:    "Response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "transport"},
	)

	// Tool metrics
	toolCallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_tool_calls_total",
			Help: "Total number of tool calls",
		},
		[]string{"tool", "status"},
	)

	toolDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_tool_duration_seconds",
			Help:    "Tool execution duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"tool"},
	)

	// Backend metrics
	backendInitialized = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mcp_backend_initialized",
			Help: "Backend initialization status (1 = initialized)",
		},
		[]string{"backend"},
	)

	// Streaming metrics (NEW for v0.2.0)
	streamingEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_streaming_events_total",
			Help: "Total number of streaming events emitted",
		},
		[]string{"tool", "event_type"},
	)

	activeStreams = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mcp_active_streams",
			Help: "Number of currently active streaming executions",
		},
	)

	concurrentExecutions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mcp_concurrent_executions",
			Help: "Number of concurrent tool executions",
		},
	)
)

// RecordRequest records a request metric
func RecordRequest(method, status, transport string) {
	requestsTotal.WithLabelValues(method, status, transport).Inc()
}

// RecordRequestDuration records request duration
func RecordRequestDuration(method, transport string, duration time.Duration) {
	requestDuration.WithLabelValues(method, transport).Observe(duration.Seconds())
}

// RecordRequestSize records request size
func RecordRequestSize(method, transport string, size int64) {
	requestSize.WithLabelValues(method, transport).Observe(float64(size))
}

// RecordResponseSize records response size
func RecordResponseSize(method, transport string, size int64) {
	responseSize.WithLabelValues(method, transport).Observe(float64(size))
}

// RecordToolCall records a tool call
func RecordToolCall(tool, status string) {
	toolCallsTotal.WithLabelValues(tool, status).Inc()
}

// RecordToolDuration records tool execution duration
func RecordToolDuration(tool string, duration time.Duration) {
	toolDuration.WithLabelValues(tool).Observe(duration.Seconds())
}

// RecordBackendInitialized records backend initialization
func RecordBackendInitialized(backend string, initialized bool) {
	value := 0.0
	if initialized {
		value = 1.0
	}
	backendInitialized.WithLabelValues(backend).Set(value)
}

// RecordStreamingEvent records a streaming event (NEW for v0.2.0)
func RecordStreamingEvent(tool, eventType string) {
	streamingEventsTotal.WithLabelValues(tool, eventType).Inc()
}

// IncActiveStreams increments active streams counter (NEW for v0.2.0)
func IncActiveStreams() {
	activeStreams.Inc()
}

// DecActiveStreams decrements active streams counter (NEW for v0.2.0)
func DecActiveStreams() {
	activeStreams.Dec()
}

// IncConcurrentExecutions increments concurrent executions (NEW for v0.2.0)
func IncConcurrentExecutions() {
	concurrentExecutions.Inc()
}

// DecConcurrentExecutions decrements concurrent executions (NEW for v0.2.0)
func DecConcurrentExecutions() {
	concurrentExecutions.Dec()
}
