package framework

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/SaherElMasry/go-mcp-framework/backend"
	"github.com/SaherElMasry/go-mcp-framework/observability"
	"github.com/SaherElMasry/go-mcp-framework/protocol"
	"github.com/SaherElMasry/go-mcp-framework/transport"
	"github.com/SaherElMasry/go-mcp-framework/transport/http"
	"github.com/SaherElMasry/go-mcp-framework/transport/stdio"
)

// Server is the main MCP server
type Server struct {
	backend    backend.ServerBackend
	config     *Config
	configFile string

	logger        *slog.Logger
	metrics       *observability.Metrics
	healthChecker *observability.HealthChecker
	handler       protocol.Handler
	transportImpl transport.Transport

	startTime time.Time
}

// NewServer creates a new server with the given options
func NewServer(opts ...Option) *Server {
	s := &Server{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Run starts the server and blocks until shutdown
func (s *Server) Run(ctx context.Context) error {
	s.startTime = time.Now()

	if err := s.loadConfig(); err != nil {
		return fmt.Errorf("config load failed: %w", err)
	}

	if err := s.setupObservability(); err != nil {
		return fmt.Errorf("observability setup failed: %w", err)
	}

	s.logger.Info("Starting MCP Server",
		"backend", s.config.Backend.Type,
		"transport", s.config.Transport.Type)

	if err := s.initializeBackend(ctx); err != nil {
		return fmt.Errorf("backend init failed: %w", err)
	}
	defer s.backend.Close()

	if err := s.setupHandler(); err != nil {
		return fmt.Errorf("handler setup failed: %w", err)
	}

	if err := s.setupTransport(); err != nil {
		return fmt.Errorf("transport setup failed: %w", err)
	}

	s.setupHealthChecks()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		s.logger.Info("Received shutdown signal", "signal", sig)
		cancel()
	}()

	s.logger.Info("Server ready")

	if err := s.transportImpl.Run(ctx); err != nil {
		if err == context.Canceled {
			s.logger.Info("Server shutdown complete")
			return nil
		}
		return fmt.Errorf("transport error: %w", err)
	}

	return nil
}

func (s *Server) loadConfig() error {
	if s.config != nil {
		return nil
	}
	config, err := LoadConfig(s.configFile)
	if err != nil {
		return err
	}
	s.config = config
	return nil
}

func (s *Server) setupObservability() error {
	// s.logger = observability.NewLogger(observability.LoggingConfig{
	// 	Level:     observability.LogLevel(s.config.Logging.Level),
	// 	Format:    s.config.Logging.Format,
	// 	AddSource: s.config.Logging.AddSource,
	// })
	s.logger = observability.NewLogger(observability.LoggingConfig{
		Level:     s.config.Logging.Level, // ✅ Direct string
		Format:    s.config.Logging.Format,
		AddSource: s.config.Logging.AddSource,
	})
	slog.SetDefault(s.logger)

	if s.config.Observability.Enabled {
		s.metrics = observability.NewMetrics("mcp", "server")

		metricsServer := observability.NewMetricsServer(
			observability.MetricsConfig{
				Enabled: true,
				Address: s.config.Observability.MetricsAddress,
				Path:    "/metrics",
			},
			s.metrics,
			s.logger,
		)

		go func() {
			s.logger.Info("Starting metrics server",
				"address", s.config.Observability.MetricsAddress)
			if err := metricsServer.Run(context.Background()); err != nil {
				s.logger.Error("Metrics server error", "error", err)
			}
		}()

		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				s.metrics.UpdateUptime(s.startTime)
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				s.metrics.UpdateMemoryUsage(m.Alloc)
				s.metrics.UpdateGoroutineCount(runtime.NumGoroutine())
			}
		}()
	}

	return nil
}

func (s *Server) initializeBackend(ctx context.Context) error {
	if s.backend == nil {
		b, err := backend.New(s.config.Backend.Type)
		if err != nil {
			return fmt.Errorf("backend creation failed: %w", err)
		}
		s.backend = b
	}

	if err := s.backend.Initialize(ctx, s.config.Backend.Config); err != nil {
		return fmt.Errorf("backend initialization failed: %w", err)
	}

	s.logger.Info("Backend initialized", "name", s.backend.Name())
	return nil
}

func (s *Server) setupHandler() error {
	if s.config.Observability.Enabled && s.metrics != nil {
		s.handler = protocol.NewInstrumentedHandler(s.backend, s.metrics, s.logger)
	} else {
		s.handler = protocol.NewHandler(s.backend, s.logger)
	}
	return nil
}

// func (s *Server) setupTransport() error {
// 	switch s.config.Transport.Type {
// 	case "stdio":
// 		s.transportImpl = stdio.NewStdioTransport(s.handler, s.logger)
// 	case "http":
// 		s.transportImpl = http.NewHTTPTransport(s.handler, s.config.Transport.HTTP, s.logger)
// 	default:
// 		return fmt.Errorf("unknown transport: %s", s.config.Transport.Type)
// 	}
// 	return nil
// }

func (s *Server) setupTransport() error {
	switch s.config.Transport.Type {
	case "stdio":
		s.transportImpl = stdio.NewStdioTransport(s.handler, s.logger)
	case "http":
		// Convert framework config to transport config
		httpConfig := http.HTTPConfig{
			Address:        s.config.Transport.HTTP.Address,
			ReadTimeout:    s.config.Transport.HTTP.ReadTimeout,
			WriteTimeout:   s.config.Transport.HTTP.WriteTimeout,
			MaxRequestSize: s.config.Transport.HTTP.MaxRequestSize,
			AllowedOrigins: s.config.Transport.HTTP.AllowedOrigins,
		}
		s.transportImpl = http.NewHTTPTransport(s.handler, httpConfig, s.logger)
	default:
		return fmt.Errorf("unknown transport: %s", s.config.Transport.Type)
	}
	return nil
}

func (s *Server) setupHealthChecks() {
	s.healthChecker = observability.NewHealthChecker()
	s.healthChecker.Register("backend", observability.BackendHealthCheck(s.backend))
	s.healthChecker.Register("uptime", observability.UptimeHealthCheck(s.startTime))
}
