package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/SaherElMasry/go-mcp-framework/backend"
	"github.com/SaherElMasry/go-mcp-framework/engine"
	"github.com/SaherElMasry/go-mcp-framework/protocol"
)

// SSEHandler handles Server-Sent Events streaming requests
type SSEHandler struct {
	executor *engine.Executor
	backend  backend.ServerBackend
	logger   *slog.Logger
	timeout  time.Duration
}

// NewSSEHandler creates a new SSE handler
func NewSSEHandler(
	executor *engine.Executor,
	backend backend.ServerBackend,
	logger *slog.Logger,
	timeout time.Duration,
) *SSEHandler {
	if logger == nil {
		logger = slog.Default()
	}
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	return &SSEHandler{
		executor: executor,
		backend:  backend,
		logger:   logger,
		timeout:  timeout,
	}
}

// ServeHTTP handles SSE streaming requests
// POST /stream?tool=<tool_name> with JSON body containing arguments
func (h *SSEHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Configure properly in production
	w.Header().Set("X-Accel-Buffering", "no")          // Disable nginx buffering

	// Get flusher for streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Parse request body (tool arguments)
	var args map[string]interface{}
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&args); err != nil && err != io.EOF {
			h.sendErrorEvent(w, flusher, "invalid_request", fmt.Sprintf("Failed to parse arguments: %v", err))
			return
		}
	}

	// Get tool name from query parameter
	toolName := r.URL.Query().Get("tool")
	if toolName == "" {
		h.sendErrorEvent(w, flusher, "missing_tool", "Tool name required in query parameter 'tool'")
		return
	}

	// Verify tool exists
	_, ok = h.backend.GetTool(toolName)
	if !ok {
		h.sendErrorEvent(w, flusher, "tool_not_found", fmt.Sprintf("Tool not found: %s", toolName))
		return
	}

	// Check if tool supports streaming
	if !h.backend.IsStreamingTool(toolName) {
		h.sendErrorEvent(w, flusher, "not_streaming", fmt.Sprintf("Tool %s does not support streaming", toolName))
		return
	}

	// Generate request ID
	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())

	h.logger.Info("starting SSE stream",
		"tool", toolName,
		"request_id", requestID,
		"remote_addr", r.RemoteAddr)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	// Create streaming handler that calls the backend
	handler := func(ctx context.Context, args map[string]interface{}, emit engine.Emitter) error {
		return h.backend.CallStreamingTool(ctx, toolName, args, emit)
	}

	// Execute tool and get event stream
	events := h.executor.Execute(ctx, toolName, requestID, args, handler)

	// Stream events as SSE messages
	h.streamEvents(w, flusher, events, requestID)

	h.logger.Info("SSE stream completed",
		"tool", toolName,
		"request_id", requestID)
}

// streamEvents converts engine events to SSE format and sends them
func (h *SSEHandler) streamEvents(
	w http.ResponseWriter,
	flusher http.Flusher,
	events <-chan engine.Event,
	requestID string,
) {
	for evt := range events {
		// Convert event to SSE using the public protocol function
		sseData := protocol.FormatEventAsSSE(evt, requestID)

		// Write SSE message
		if _, err := w.Write([]byte(sseData)); err != nil {
			h.logger.Error("failed to write SSE message",
				"error", err,
				"request_id", requestID)
			return
		}

		// Flush immediately for streaming
		flusher.Flush()

		// Log progress events
		if evt.Type == engine.EventProgress {
			if payload, ok := evt.Data.(engine.ProgressPayload); ok {
				h.logger.Debug("progress",
					"request_id", requestID,
					"percentage", payload.Percentage,
					"message", payload.Message)
			}
		}
	}
}

// sendErrorEvent sends an error event in SSE format
func (h *SSEHandler) sendErrorEvent(w http.ResponseWriter, flusher http.Flusher, code, message string) {
	errorEvt := engine.NewErrorEvent(nil, message, false)
	sseData := protocol.FormatEventAsSSE(errorEvt, code)
	w.Write([]byte(sseData))
	flusher.Flush()
}
