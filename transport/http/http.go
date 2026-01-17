package http

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/SaherElMasry/go-mcp-framework/backend"
	"github.com/SaherElMasry/go-mcp-framework/engine"
	"github.com/SaherElMasry/go-mcp-framework/transport"
)

// HTTPConfig configures the HTTP transport
type HTTPConfig struct {
	Address        string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	MaxRequestSize int64
	AllowedOrigins []string
}

// HTTPTransport implements HTTP-based transport
type HTTPTransport struct {
	handler  transport.Handler
	config   HTTPConfig
	logger   *slog.Logger
	server   *http.Server
	backend  backend.ServerBackend // NEW: For SSE streaming
	executor *engine.Executor      // NEW: For streaming execution
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(
	handler transport.Handler,
	config HTTPConfig,
	logger *slog.Logger,
	backend backend.ServerBackend, // NEW
	executor *engine.Executor, // NEW
) *HTTPTransport {
	if logger == nil {
		logger = slog.Default()
	}

	return &HTTPTransport{
		handler:  handler,
		config:   config,
		logger:   logger,
		backend:  backend,
		executor: executor,
	}
}

// Run starts the HTTP server
func (t *HTTPTransport) Run(ctx context.Context) error {
	mux := http.NewServeMux()

	// Regular JSON-RPC endpoint
	mux.HandleFunc("/rpc", t.handleRPC)

	// NEW: SSE streaming endpoint
	if t.executor != nil {
		sseHandler := NewSSEHandler(t.executor, t.backend, t.logger, 5*time.Minute)
		mux.Handle("/stream", sseHandler)
		t.logger.Info("SSE streaming endpoint enabled", "path", "/stream")
	}

	// Health check endpoint
	mux.HandleFunc("/health", t.handleHealth)

	t.server = &http.Server{
		Addr:         t.config.Address,
		Handler:      t.applyCORS(mux),
		ReadTimeout:  t.config.ReadTimeout,
		WriteTimeout: t.config.WriteTimeout,
	}

	// Graceful shutdown
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := t.server.Shutdown(shutdownCtx); err != nil {
			t.logger.Error("shutdown error", "error", err)
		}
	}()

	t.logger.Info("http transport started", "address", t.config.Address)

	if err := t.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server error: %w", err)
	}

	return nil
}

// handleRPC handles regular JSON-RPC requests
func (t *HTTPTransport) handleRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(io.LimitReader(r.Body, t.config.MaxRequestSize))
	if err != nil {
		t.logger.Error("read error", "error", err)
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Handle request
	resp, err := t.handler.Handle(r.Context(), body, "http")
	if err != nil {
		t.logger.Error("handler error", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(resp); err != nil {
		t.logger.Error("write error", "error", err)
	}
}

// handleHealth handles health check requests
func (t *HTTPTransport) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// applyCORS applies CORS headers
func (t *HTTPTransport) applyCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight requests
		if r.Method == http.MethodOptions {
			t.setCORSHeaders(w)
			w.WriteHeader(http.StatusOK)
			return
		}

		t.setCORSHeaders(w)
		next.ServeHTTP(w, r)
	})
}

// setCORSHeaders sets CORS headers
func (t *HTTPTransport) setCORSHeaders(w http.ResponseWriter) {
	if len(t.config.AllowedOrigins) > 0 {
		w.Header().Set("Access-Control-Allow-Origin", t.config.AllowedOrigins[0])
	}
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}
