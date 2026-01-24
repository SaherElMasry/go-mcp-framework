package framework

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time" // ADD THIS IMPORT

	"github.com/SaherElMasry/go-mcp-framework/auth"
	"github.com/SaherElMasry/go-mcp-framework/backend"
	"github.com/SaherElMasry/go-mcp-framework/cache" // ADD THIS IMPORT
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
	executor   *engine.Executor

	// Observability
	metricsServer *observability.MetricsServer

	authManager *auth.Manager

	// === NEW: Cache support ===
	cache       cache.Cache         // Cache instance
	cacheConfig *cache.Config       // Cache configuration
	keyGen      *cache.KeyGenerator // Key generator
}

// NewServer creates a new MCP server
func NewServer(opts ...Option) *Server {
	// Auto-detect color support
	color.AutoDetect()

	s := &Server{
		config:      DefaultConfig(),
		authManager: auth.NewManager(),
		logger:      slog.Default(),
		// Cache will be initialized in Initialize() if configured
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

	// === NEW: Initialize cache BEFORE backend ===
	if s.cacheConfig != nil && s.cacheConfig.Enabled {
		var err error
		s.cache, err = cache.New(s.cacheConfig)
		if err != nil {
			return fmt.Errorf("failed to create cache: %w", err)
		}

		s.keyGen = cache.NewKeyGenerator()

		s.logger.Info("cache initialized",
			"type", s.cacheConfig.Type,
			"ttl", s.cacheConfig.GetTTLDuration(),
			"max_size", s.cacheConfig.MaxSize)

		// Start background cleanup for memory cache
		if s.cacheConfig.Type == cache.TypeShort {
			go s.startCacheCleanup(ctx)
		}
	} else {
		// No cache configured - use NoOp
		s.cache = cache.NewNoOpCache()
		s.logger.Debug("cache disabled")
	}

	// Initialize backend if not provided
	if s.backend == nil {
		var err error
		s.backend, err = backend.Create(s.config.Backend.Type)
		if err != nil {
			return fmt.Errorf("failed to create backend: %w", err)
		}
	}

	// Set auth manager on backend
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

	// Validate all auth providers
	if s.authManager != nil && len(s.authManager.List()) > 0 {
		if err := s.authManager.ValidateAll(ctx); err != nil {
			s.logger.Warn("auth validation failed", "error", err)
		}
	}

	// Initialize streaming executor
	if s.config.Streaming.Enabled {
		executorConfig := engine.ExecutorConfig{
			BufferSize:    s.config.Streaming.BufferSize,
			Timeout:       s.config.Streaming.Timeout,
			MaxEvents:     s.config.Streaming.MaxEvents,
			MaxConcurrent: s.config.Streaming.MaxConcurrent,
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

	// === NEW: Configure cache in handler ===
	if s.cache != nil && s.keyGen != nil {
		// Type assertion to access SetCache
		if h, ok := handler.(*protocol.InstrumentedHandler); ok {
			h.SetCache(s.cache, s.keyGen, s.cacheConfig)
		} else if h, ok := handler.(*protocol.Handler); ok {
			h.SetCache(s.cache, s.keyGen, s.cacheConfig)
		}
		s.logger.Info("cache configured in protocol handler",
			"enabled", s.cacheConfig.Enabled,
			"type", s.cacheConfig.Type)
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

		s.transport = httpTransport.NewHTTPTransport(
			handler,
			httpConfig,
			s.logger,
			s.backend,
			s.executor,
		)

	case "stdio":
		s.transport = stdioTransport.NewStdioTransport(handler, s.logger)

	default:
		return fmt.Errorf("unknown transport type: %s", s.config.Transport.Type)
	}

	return nil
}

// === NEW: Background cache cleanup ===
func (s *Server) startCacheCleanup(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if mc, ok := s.cache.(*cache.MemoryCache); ok {
				removed := mc.CleanExpired()
				if removed > 0 {
					s.logger.Debug("cleaned expired cache entries",
						"count", removed)
				}
			}
		}
	}
}

// Run starts the server
func (s *Server) Run(ctx context.Context) error {
	// Print colorful startup banner
	if color.IsEnabled() {
		PrintStartupBanner(
			"MCP Server",
			"v0.4.0", // UPDATE VERSION
			"Production-ready MCP framework with caching",
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

	// === NEW: Close cache ===
	if s.cache != nil {
		if err := s.cache.Close(); err != nil {
			s.logger.Error("cache close error", "error", err)
		}
	}

	// Close auth manager
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

// RegisterFunction registers a single streaming function as a tool
func (s *Server) RegisterFunction(name string, handler backend.StreamingHandler) {
	functionBackend := backend.NewFunctionBackend(name, handler)
	s.backend = functionBackend

	s.logger.Info("registered function tool",
		"name", name,
		"style", "v2-simple")
}

// RegisterBackend registers a full backend
func (s *Server) RegisterBackend(b backend.ServerBackend) {
	s.backend = b

	tools := b.ListTools()
	s.logger.Info("registered backend",
		"name", b.Name(),
		"tools", len(tools),
		"style", "v0.2.0-full")
}

// === NEW: Public Getters ===

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

// === NEW: Cache Getters ===

// GetCache returns the cache instance
func (s *Server) GetCache() cache.Cache {
	return s.cache
}

// GetKeyGenerator returns the key generator
func (s *Server) GetKeyGenerator() *cache.KeyGenerator {
	return s.keyGen
}

// GetCacheConfig returns the cache configuration
func (s *Server) GetCacheConfig() *cache.Config {
	return s.cacheConfig
}
