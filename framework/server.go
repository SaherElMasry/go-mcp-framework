package framework

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/SaherElMasry/go-mcp-framework/auth"
	"github.com/SaherElMasry/go-mcp-framework/backend"

	"github.com/SaherElMasry/go-mcp-framework/color"
	"github.com/SaherElMasry/go-mcp-framework/engine"
	"github.com/SaherElMasry/go-mcp-framework/observability"
	"github.com/SaherElMasry/go-mcp-framework/protocol"
	"github.com/SaherElMasry/go-mcp-framework/transport"
	httpTransport "github.com/SaherElMasry/go-mcp-framework/transport/http"
	stdioTransport "github.com/SaherElMasry/go-mcp-framework/transport/stdio"
)

// Server is the main MCP server
type Server struct {
	config     *Config
	configFile string
	backend    backend.ServerBackend
	transport  transport.Transport
	logger     *slog.Logger
	executor   *engine.Executor // NEW: Streaming executor

	// Observability
	metricsServer *observability.MetricsServer

	authManager *auth.Manager // === NEW ===
}

// NewServer creates a new MCP server
func NewServer(opts ...Option) *Server {
	// Auto-detect color support
	color.AutoDetect()

	s := &Server{
		config:      DefaultConfig(),
		authManager: auth.NewManager(),
		logger:      slog.Default(),
	}

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Initialize initializes the server
func (s *Server) Initialize(ctx context.Context) error {
	// Load config file if specified
	if s.configFile != "" {
		config, err := LoadConfig(s.configFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		s.config = config
	}

	// Validate configuration
	if err := s.config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Setup logging
	s.logger = observability.SetupLogging(s.config.Logging)

	s.logger.Info("initializing server",
		"backend", s.config.Backend.Type,
		"transport", s.config.Transport.Type)

	// Initialize backend if not provided
	if s.backend == nil {
		var err error
		s.backend, err = backend.Create(s.config.Backend.Type)
		if err != nil {
			return fmt.Errorf("failed to create backend: %w", err)
		}
	}

	// === NEW: Set auth manager on backend ===
	if s.authManager != nil {
		s.backend.SetAuthManager(s.authManager)

		// Log registered auth providers
		providers := s.authManager.List()
		if len(providers) > 0 {
			s.logger.Info("auth providers registered",
				"count", len(providers),
				"providers", providers)
		}
	}

	// Initialize backend
	if err := s.backend.Initialize(ctx, s.config.Backend.Config); err != nil {
		return fmt.Errorf("failed to initialize backend: %w", err)
	}

	// === NEW: Validate all auth providers ===
	if s.authManager != nil && len(s.authManager.List()) > 0 {
		if err := s.authManager.ValidateAll(ctx); err != nil {
			s.logger.Warn("auth validation failed", "error", err)
			// Don't fail startup, just warn
		}
	}

	// NEW: Initialize streaming executor
	if s.config.Streaming.Enabled {
		executorConfig := engine.ExecutorConfig{
			BufferSize:    s.config.Streaming.BufferSize,
			Timeout:       s.config.Streaming.Timeout,
			MaxEvents:     s.config.Streaming.MaxEvents,
			MaxConcurrent: s.config.Streaming.MaxConcurrent, // v2 semaphore
		}
		s.executor = engine.NewExecutor(executorConfig, s.logger)

		s.logger.Info("streaming enabled",
			"buffer_size", executorConfig.BufferSize,
			"timeout", executorConfig.Timeout,
			"max_concurrent", executorConfig.MaxConcurrent)
	}

	// Setup observability
	if s.config.Observability.Enabled {
		s.metricsServer = observability.NewMetricsServer(
			s.config.Observability.MetricsAddress,
			s.logger,
		)

		go func() {
			if err := s.metricsServer.Start(); err != nil {
				s.logger.Error("metrics server failed", "error", err)
			}
		}()
	}

	// Create protocol handler
	var handler transport.Handler
	if s.config.Observability.Enabled {
		handler = protocol.NewInstrumentedHandler(s.backend, s.logger)
	} else {
		handler = protocol.NewHandler(s.backend, s.logger)
	}

	// Setup transport
	switch s.config.Transport.Type {
	case "http":
		httpConfig := httpTransport.HTTPConfig{
			Address:        s.config.Transport.HTTP.Address,
			ReadTimeout:    s.config.Transport.HTTP.ReadTimeout,
			WriteTimeout:   s.config.Transport.HTTP.WriteTimeout,
			MaxRequestSize: s.config.Transport.HTTP.MaxRequestSize,
			AllowedOrigins: s.config.Transport.HTTP.AllowedOrigins,
		}

		// NEW: Pass executor for streaming support
		s.transport = httpTransport.NewHTTPTransport(
			handler,
			httpConfig,
			s.logger,
			s.backend,  // NEW: For SSE streaming
			s.executor, // NEW: For streaming execution
		)

	case "stdio":
		s.transport = stdioTransport.NewStdioTransport(handler, s.logger)

	default:
		return fmt.Errorf("unknown transport type: %s", s.config.Transport.Type)
	}

	return nil
}

// Run starts the server
func (s *Server) Run(ctx context.Context) error {
	// Print colorful startup banner
	if color.IsEnabled() {
		PrintStartupBanner(
			"MCP Server",
			"v0.3.0",
			"Production-ready MCP framework",
		)
	}
	// Initialize
	if err := s.Initialize(ctx); err != nil {
		return err
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		s.logger.Info("shutdown signal received")
		cancel()
	}()

	// Run transport
	s.logger.Info("server starting",
		"transport", s.config.Transport.Type,
		"address", s.getAddress())

	if err := s.transport.Run(ctx); err != nil {
		return fmt.Errorf("transport error: %w", err)
	}

	// Cleanup
	s.logger.Info("server shutting down")

	// === NEW: Close auth manager ===
	if s.authManager != nil {
		if err := s.authManager.Close(); err != nil {
			s.logger.Error("auth manager close error", "error", err)
		}
	}

	if err := s.backend.Close(); err != nil {
		s.logger.Error("backend close error", "error", err)
	}

	if s.metricsServer != nil {
		s.metricsServer.Stop()
	}

	return nil
}

// getAddress returns the server address for logging
func (s *Server) getAddress() string {
	switch s.config.Transport.Type {
	case "http":
		return s.config.Transport.HTTP.Address
	case "stdio":
		return "stdio"
	default:
		return "unknown"
	}
}

// NEW: v2-style simple registration

// RegisterFunction registers a single streaming function as a tool (v2 style)
// This is a simplified API for quick prototyping and simple use cases
func (s *Server) RegisterFunction(name string, handler backend.StreamingHandler) {
	functionBackend := backend.NewFunctionBackend(name, handler)
	s.backend = functionBackend

	s.logger.Info("registered function tool",
		"name", name,
		"style", "v2-simple")
}

// RegisterBackend registers a full backend (v0.2.0 style)
// This is the complete API for production use cases with multiple tools
func (s *Server) RegisterBackend(b backend.ServerBackend) {
	s.backend = b

	tools := b.ListTools()
	s.logger.Info("registered backend",
		"name", b.Name(),
		"tools", len(tools),
		"style", "v0.2.0-full")
}

// === NEW: Public Getters for Auth ===
//
// GetBackend returns the current backend
func (s *Server) GetBackend() backend.ServerBackend {
	return s.backend
}

// GetExecutor returns the streaming executor
func (s *Server) GetExecutor() *engine.Executor {
	return s.executor
}

// GetAuthManager returns the auth manager
func (s *Server) GetAuthManager() *auth.Manager {
	return s.authManager
}

// GetLogger returns the logger
func (s *Server) GetLogger() *slog.Logger {
	return s.logger
}
