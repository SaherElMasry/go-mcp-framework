package framework

import (
	"github.com/SaherElMasry/go-mcp-framework/backend"
)

// Option configures the server
type Option func(*Server)

// WithBackend sets a specific backend instance
func WithBackend(b backend.ServerBackend) Option {
	return func(s *Server) {
		s.backend = b
	}
}

// WithConfigFile sets the config file path
func WithConfigFile(path string) Option {
	return func(s *Server) {
		s.configFile = path
	}
}

// WithConfig sets the complete configuration
func WithConfig(config *Config) Option {
	return func(s *Server) {
		s.config = config
	}
}

// WithBackendType sets the backend type
func WithBackendType(backendType string) Option {
	return func(s *Server) {
		if s.config == nil {
			s.config = &Config{}
		}
		s.config.Backend.Type = backendType
	}
}

// WithTransport sets the transport type
func WithTransport(transport string) Option {
	return func(s *Server) {
		if s.config == nil {
			s.config = &Config{}
		}
		s.config.Transport.Type = transport
	}
}

// WithHTTPAddress sets the HTTP server address
func WithHTTPAddress(addr string) Option {
	return func(s *Server) {
		if s.config == nil {
			s.config = &Config{}
		}
		s.config.Transport.HTTP.Address = addr
	}
}

// WithObservability enables/disables observability
func WithObservability(enabled bool) Option {
	return func(s *Server) {
		if s.config == nil {
			s.config = &Config{}
		}
		s.config.Observability.Enabled = enabled
	}
}

// WithLogLevel sets the log level
func WithLogLevel(level string) Option {
	return func(s *Server) {
		if s.config == nil {
			s.config = &Config{}
		}
		s.config.Logging.Level = level
	}
}

// WithMetricsAddress sets the metrics server address
func WithMetricsAddress(addr string) Option {
	return func(s *Server) {
		if s.config == nil {
			s.config = &Config{}
		}
		s.config.Observability.MetricsAddress = addr
	}
}
