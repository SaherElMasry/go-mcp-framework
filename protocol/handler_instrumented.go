package protocol

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/SaherElMasry/go-mcp-framework/backend"
	"github.com/SaherElMasry/go-mcp-framework/cache"
	"github.com/SaherElMasry/go-mcp-framework/observability"
)

// InstrumentedHandler wraps a handler with metrics
type InstrumentedHandler struct {
	*Handler
}

// NewInstrumentedHandler creates a new instrumented handler
func NewInstrumentedHandler(backend backend.ServerBackend, logger *slog.Logger) *InstrumentedHandler {
	return &InstrumentedHandler{
		Handler: NewHandler(backend, logger),
	}
}

// === NEW: SetCache forwards to underlying handler ===
func (h *InstrumentedHandler) SetCache(c cache.Cache, keyGen *cache.KeyGenerator, config *cache.Config) {
	h.Handler.SetCache(c, keyGen, config)
}

// Handle processes a request with metrics
func (h *InstrumentedHandler) Handle(ctx context.Context, data []byte, transportType string) ([]byte, error) {
	start := time.Now()

	// Parse request to get method
	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		observability.RecordRequest(req.Method, "error", transportType)
		return h.Handler.Handle(ctx, data, transportType)
	}

	// Handle request
	resp, err := h.Handler.Handle(ctx, data, transportType)

	// Record metrics
	duration := time.Since(start)
	status := "success"
	if err != nil {
		status = "error"
	}

	observability.RecordRequest(req.Method, status, transportType)
	observability.RecordRequestDuration(req.Method, transportType, duration)
	observability.RecordRequestSize(req.Method, transportType, int64(len(data)))
	observability.RecordResponseSize(req.Method, transportType, int64(len(resp)))

	return resp, err
}
