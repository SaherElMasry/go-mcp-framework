package protocol

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/SaherElMasry/go-mcp-framework/backend"
	"github.com/SaherElMasry/go-mcp-framework/observability"
)

// InstrumentedHandler wraps Handler with observability
type InstrumentedHandler struct {
	base    Handler
	metrics *observability.Metrics
	logger  *slog.Logger
}

// NewInstrumentedHandler creates a handler with observability
func NewInstrumentedHandler(
	b backend.ServerBackend,
	metrics *observability.Metrics,
	logger *slog.Logger,
) Handler {
	base := NewHandler(b, logger)

	return &InstrumentedHandler{
		base:    base,
		metrics: metrics,
		logger:  logger,
	}
}

// Handle implements Handler with full observability
func (h *InstrumentedHandler) Handle(ctx context.Context, requestBytes []byte, transport string) ([]byte, error) {
	start := time.Now()

	var req Request
	_ = json.Unmarshal(requestBytes, &req)
	method := req.Method

	if h.metrics != nil {
		h.metrics.RequestsInFlight.WithLabelValues(transport).Inc()
		defer h.metrics.RequestsInFlight.WithLabelValues(transport).Dec()
		h.metrics.RecordRequestSize(method, transport, len(requestBytes))
	}

	h.logger.InfoContext(ctx, "processing request",
		"method", method,
		"transport", transport,
		"size", len(requestBytes))

	response, err := h.base.Handle(ctx, requestBytes, transport)

	duration := time.Since(start)

	status := "success"
	if err != nil {
		status = "error"
		h.logger.ErrorContext(ctx, "request failed", "method", method, "error", err, "duration", duration)
	} else {
		h.logger.InfoContext(ctx, "request completed", "method", method, "duration", duration)
	}

	if h.metrics != nil {
		h.metrics.RecordRequest(method, transport, status, duration)
		h.metrics.RecordResponseSize(method, transport, len(response))
	}

	return response, err
}
