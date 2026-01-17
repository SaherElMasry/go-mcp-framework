package framework

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete server configuration
type Config struct {
	Backend       BackendConfig       `yaml:"backend"`
	Transport     TransportConfig     `yaml:"transport"`
	Observability ObservabilityConfig `yaml:"observability"`
	Logging       LoggingConfig       `yaml:"logging"`
	Streaming     StreamingConfig     `yaml:"streaming"` // NEW
}

// BackendConfig configures the backend
type BackendConfig struct {
	Type   string                 `yaml:"type"`
	Config map[string]interface{} `yaml:"config"`
}

// TransportConfig configures the transport layer
type TransportConfig struct {
	Type  string     `yaml:"type"`
	HTTP  HTTPConfig `yaml:"http"`
	Stdio struct{}   `yaml:"stdio"`
}

// HTTPConfig configures HTTP transport
type HTTPConfig struct {
	Address        string        `yaml:"address"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
	MaxRequestSize int64         `yaml:"max_request_size"`
	AllowedOrigins []string      `yaml:"allowed_origins"`
}

// ObservabilityConfig configures observability features
type ObservabilityConfig struct {
	Enabled        bool   `yaml:"enabled"`
	MetricsAddress string `yaml:"metrics_address"`
}

// LoggingConfig configures logging
type LoggingConfig struct {
	Level     string `yaml:"level"`
	Format    string `yaml:"format"`
	AddSource bool   `yaml:"add_source"`
}

// StreamingConfig configures streaming execution (NEW - v2 feature)
type StreamingConfig struct {
	Enabled       bool          `yaml:"enabled"`
	BufferSize    int           `yaml:"buffer_size"`
	Timeout       time.Duration `yaml:"timeout"`
	MaxEvents     int64         `yaml:"max_events"`
	MaxConcurrent int           `yaml:"max_concurrent"` // NEW: v2 semaphore
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Backend: BackendConfig{
			Type: "filesystem",
			Config: map[string]interface{}{
				"workspace_root": "./workspace",
				"max_file_size":  10485760,
				"read_only":      false,
			},
		},
		Transport: TransportConfig{
			Type: "http",
			HTTP: HTTPConfig{
				Address:        ":8080",
				ReadTimeout:    30 * time.Second,
				WriteTimeout:   30 * time.Second,
				MaxRequestSize: 10485760,
				AllowedOrigins: []string{"*"},
			},
		},
		Observability: ObservabilityConfig{
			Enabled:        true,
			MetricsAddress: ":9091",
		},
		Logging: LoggingConfig{
			Level:     "info",
			Format:    "json",
			AddSource: false,
		},
		Streaming: StreamingConfig{
			Enabled:       true,
			BufferSize:    100,
			Timeout:       5 * time.Minute,
			MaxEvents:     10000,
			MaxConcurrent: 16, // NEW: v2 concurrency control
		},
	}
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables
	expanded := os.ExpandEnv(string(data))

	config := DefaultConfig()
	if err := yaml.Unmarshal([]byte(expanded), config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Backend.Type == "" {
		return fmt.Errorf("backend type is required")
	}

	if c.Transport.Type == "" {
		return fmt.Errorf("transport type is required")
	}

	if c.Transport.Type == "http" && c.Transport.HTTP.Address == "" {
		return fmt.Errorf("HTTP address is required when using HTTP transport")
	}

	// NEW: Validate streaming config
	if c.Streaming.Enabled {
		if c.Streaming.BufferSize <= 0 {
			return fmt.Errorf("streaming buffer size must be positive")
		}
		if c.Streaming.MaxConcurrent <= 0 {
			return fmt.Errorf("max concurrent executions must be positive")
		}
	}

	return nil
}
