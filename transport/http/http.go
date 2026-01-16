package http

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

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

// HTTPTransport implements MCP over HTTP
type HTTPTransport struct {
	handler transport.Handler
	logger  *slog.Logger
	server  *http.Server
	config  HTTPConfig
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(handler transport.Handler, config HTTPConfig, logger *slog.Logger) *HTTPTransport {
	if logger == nil {
		logger = slog.Default()
	}

	if config.Address == "" {
		config.Address = ":8080"
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 30 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 30 * time.Second
	}
	if config.MaxRequestSize == 0 {
		config.MaxRequestSize = 10 * 1024 * 1024
	}

	t := &HTTPTransport{
		handler: handler,
		logger:  logger,
		config:  config,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/rpc", t.withMiddleware(t.handleRPC))
	mux.HandleFunc("/health", t.handleHealth)

	t.server = &http.Server{
		Addr:         config.Address,
		Handler:      mux,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	return t
}

// Run starts the HTTP server
func (t *HTTPTransport) Run(ctx context.Context) error {
	t.logger.Info("http transport started", "address", t.server.Addr)

	errChan := make(chan error, 1)
	go func() {
		if err := t.server.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		t.logger.Info("http transport shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return t.server.Shutdown(shutdownCtx)
	case err := <-errChan:
		return err
	}
}

func (t *HTTPTransport) handleRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.logger.Error("failed to read body", "error", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	t.logger.Debug("received HTTP request", "size", len(body))

	response, err := t.handler.Handle(r.Context(), body, "http")
	if err != nil {
		t.logger.Error("handler error", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (t *HTTPTransport) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (t *HTTPTransport) withMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(t.config.AllowedOrigins) > 0 {
			origin := r.Header.Get("Origin")
			for _, allowed := range t.config.AllowedOrigins {
				if origin == allowed || allowed == "*" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
					break
				}
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		r.Body = http.MaxBytesReader(w, r.Body, t.config.MaxRequestSize)

		start := time.Now()
		handler(w, r)

		t.logger.Info("http request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start))
	}
}
