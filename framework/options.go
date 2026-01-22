package framework

import (
	"context"
	"os"
	"time"

	"github.com/SaherElMasry/go-mcp-framework/auth"
	"github.com/SaherElMasry/go-mcp-framework/backend"
)

// Option configures the server
type Option func(*Server)

// ============================================================
// Basic Options
// ============================================================

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

// ============================================================
// Transport Options
// ============================================================

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

// ============================================================
// Observability Options
// ============================================================

// WithObservability enables/disables observability
func WithObservability(enabled bool) Option {
	return func(s *Server) {
		if s.config == nil {
			s.config = DefaultConfig()
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

// ============================================================
// Streaming Options
// ============================================================

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

// WithMaxConcurrent sets maximum concurrent executions
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

// ============================================================
// AUTH OPTIONS
// ============================================================

// WithAuth configures basic authentication
func WithAuth(authType string, config interface{}) Option {
	return func(s *Server) {
		var provider auth.AuthProvider

		switch authType {
		case "api-key":
			if cfg, ok := config.(auth.APIKeyConfig); ok {
				provider = auth.NewAPIKeyProvider("default", cfg)
			} else {
				s.logger.Error("invalid config type for api-key provider")
				return
			}

		case "database":
			if _, ok := config.(auth.DatabaseConfig); ok {
				provider = auth.NewDatabaseProvider("default")
				// Additional database setup would go here
			} else {
				s.logger.Error("invalid config type for database provider")
				return
			}

		default:
			s.logger.Error("unsupported auth type", "type", authType)
			return
		}

		// Register provider
		if err := s.authManager.Register("default", provider); err != nil {
			s.logger.Error("failed to register auth provider", "error", err)
			return
		}

		// Set provider on backend if it exists
		if s.backend != nil {
			s.backend.SetAuthProvider(provider)
		}
	}
}

// WithAuthProvider directly sets an auth provider
func WithAuthProvider(name string, provider auth.AuthProvider) Option {
	return func(s *Server) {
		if err := s.authManager.Register(name, provider); err != nil {
			s.logger.Error("failed to register auth provider",
				"name", name,
				"error", err)
			return
		}

		// Set as default provider on backend if name is "default"
		if s.backend != nil && name == "default" {
			s.backend.SetAuthProvider(provider)
		}
	}
}

// WithAuthResource registers a resource with an auth provider
func WithAuthResource(providerName string, resource auth.ResourceConfig) Option {
	return func(s *Server) {
		provider, err := s.authManager.Get(providerName)
		if err != nil {
			s.logger.Error("auth provider not found",
				"name", providerName,
				"error", err)
			return
		}

		// Check if provider supports RegisterResource
		if apiProvider, ok := provider.(*auth.APIKeyProvider); ok {
			apiProvider.RegisterResource(resource)
		} else if dbProvider, ok := provider.(*auth.DatabaseProvider); ok {
			dbProvider.RegisterResource(resource)
		} else if oauth2Provider, ok := provider.(*auth.OAuth2Provider); ok {
			oauth2Provider.RegisterResource(resource)
		} else {
			s.logger.Error("provider does not support resource registration",
				"provider", providerName)
		}
	}
}

// WithOAuth configures OAuth2 authentication for popular providers
func WithOAuth(providerName, clientID, clientSecret, redirectURL string, scopes []string) Option {
	return func(s *Server) {
		// Create token store with encryption
		encryptionKey := os.Getenv("OAUTH_ENCRYPTION_KEY")
		if encryptionKey == "" {
			// Generate a key if not provided (for development)
			var err error
			encryptionKey, err = auth.GenerateKey()
			if err != nil {
				s.logger.Error("failed to generate encryption key", "error", err)
				return
			}
			s.logger.Warn("using generated encryption key - set OAUTH_ENCRYPTION_KEY for production")
		}

		tokenStore, err := auth.NewFileTokenStore(".tokens", encryptionKey)
		if err != nil {
			s.logger.Error("failed to create token store", "error", err)
			return
		}

		// Create provider factory
		factory := auth.NewProviderFactory(tokenStore)

		// Create OAuth2 provider
		provider, err := factory.Create(providerName, clientID, clientSecret, redirectURL, scopes)
		if err != nil {
			s.logger.Error("failed to create OAuth2 provider",
				"provider", providerName,
				"error", err)
			return
		}

		// Register provider
		if err := s.authManager.Register("default", provider); err != nil {
			s.logger.Error("failed to register OAuth2 provider", "error", err)
			return
		}

		// Set provider on backend
		if s.backend != nil {
			s.backend.SetAuthProvider(provider)
		}

		s.logger.Info("OAuth2 provider configured",
			"provider", providerName,
			"scopes", scopes)
	}
}

// WithOAuth2Token sets a pre-configured OAuth2 token
func WithOAuth2Token(providerName string, token *auth.OAuth2Token) Option {
	return func(s *Server) {
		provider, err := s.authManager.Get(providerName)
		if err != nil {
			s.logger.Error("provider not found",
				"name", providerName,
				"error", err)
			return
		}

		oauth2Provider, ok := provider.(*auth.OAuth2Provider)
		if !ok {
			s.logger.Error("provider is not an OAuth2 provider",
				"name", providerName)
			return
		}

		ctx := context.Background()
		if err := oauth2Provider.SetToken(ctx, token); err != nil {
			s.logger.Error("failed to set token", "error", err)
			return
		}

		s.logger.Info("OAuth2 token configured", "provider", providerName)
	}
}

// ============================================================
// OAuth Convenience Functions
// ============================================================

// WithGitHub is a convenience function for GitHub OAuth2
func WithGitHub(clientID, clientSecret, redirectURL string, scopes []string) Option {
	return WithOAuth("github", clientID, clientSecret, redirectURL, scopes)
}

// WithGoogle is a convenience function for Google OAuth2
func WithGoogle(clientID, clientSecret, redirectURL string, scopes []string) Option {
	return WithOAuth("google", clientID, clientSecret, redirectURL, scopes)
}

// WithFacebook is a convenience function for Facebook OAuth2
func WithFacebook(clientID, clientSecret, redirectURL string, scopes []string) Option {
	return WithOAuth("facebook", clientID, clientSecret, redirectURL, scopes)
}

// WithMicrosoft is a convenience function for Microsoft OAuth2
func WithMicrosoft(clientID, clientSecret, redirectURL string, scopes []string) Option {
	return WithOAuth("microsoft", clientID, clientSecret, redirectURL, scopes)
}

// WithSlack is a convenience function for Slack OAuth2
func WithSlack(clientID, clientSecret, redirectURL string, scopes []string) Option {
	return WithOAuth("slack", clientID, clientSecret, redirectURL, scopes)
}
