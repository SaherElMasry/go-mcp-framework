package observability

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsServer serves Prometheus metrics
type MetricsServer struct {
	address string
	server  *http.Server
	logger  *slog.Logger
}

// NewMetricsServer creates a new metrics server
func NewMetricsServer(address string, logger *slog.Logger) *MetricsServer {
	if logger == nil {
		logger = slog.Default()
	}

	return &MetricsServer{
		address: address,
		logger:  logger,
	}
}

// Start starts the metrics server
func (m *MetricsServer) Start() error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	m.server = &http.Server{
		Addr:         m.address,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	m.logger.Info("metrics server starting", "address", m.address)

	if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("metrics server error: %w", err)
	}

	return nil
}

// Stop stops the metrics server
func (m *MetricsServer) Stop() error {
	if m.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	m.logger.Info("metrics server stopping")

	if err := m.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("metrics server shutdown error: %w", err)
	}

	return nil
}
