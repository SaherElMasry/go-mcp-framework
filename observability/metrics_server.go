package observability

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsConfig configures metrics server
type MetricsConfig struct {
	Enabled bool
	Address string
	Path    string
}

// MetricsServer serves Prometheus metrics over HTTP
type MetricsServer struct {
	server  *http.Server
	logger  *slog.Logger
	metrics *Metrics
}

// NewMetricsServer creates a new metrics server
func NewMetricsServer(config MetricsConfig, metrics *Metrics, logger *slog.Logger) *MetricsServer {
	if logger == nil {
		logger = slog.Default()
	}

	mux := http.NewServeMux()
	mux.Handle(config.Path, promhttp.Handler())

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK\n")
	})

	mux.HandleFunc("/runtime", func(w http.ResponseWriter, r *http.Request) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"alloc_bytes":%d,"goroutines":%d}`, m.Alloc, runtime.NumGoroutine())
	})

	return &MetricsServer{
		server: &http.Server{
			Addr:    config.Address,
			Handler: mux,
		},
		logger:  logger,
		metrics: metrics,
	}
}

// Run starts the metrics server
func (s *MetricsServer) Run(ctx context.Context) error {
	s.logger.Info("starting metrics server", "address", s.server.Addr)

	errChan := make(chan error, 1)
	go func() {
		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(shutdownCtx)
	case err := <-errChan:
		return err
	}
}
