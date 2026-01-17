package framework

import (
	"time"

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

// NEW: Streaming options (v2 features)

// WithStreaming enables/disables streaming
func WithStreaming(enabled bool) Option {
	return func(s *Server) {
		if s.config == nil {
			s.config = &Config{}
		}
		s.config.Streaming.Enabled = enabled
	}
}

// WithStreamingBufferSize sets the event buffer size
func WithStreamingBufferSize(size int) Option {
	return func(s *Server) {
		if s.config == nil {
			s.config = &Config{}
		}
		s.config.Streaming.BufferSize = size
	}
}

// WithStreamingTimeout sets the execution timeout
func WithStreamingTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		if s.config == nil {
			s.config = &Config{}
		}
		s.config.Streaming.Timeout = timeout
	}
}

// WithMaxConcurrent sets maximum concurrent executions (v2 semaphore)
func WithMaxConcurrent(max int) Option {
	return func(s *Server) {
		if s.config == nil {
			s.config = &Config{}
		}
		s.config.Streaming.MaxConcurrent = max
	}
}

// WithMaxEvents sets maximum events per execution
func WithMaxEvents(max int64) Option {
	return func(s *Server) {
		if s.config == nil {
			s.config = &Config{}
		}
		s.config.Streaming.MaxEvents = max
	}
}
