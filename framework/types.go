package framework

import "time"

// Config represents the complete server configuration
type Config struct {
	Backend       BackendConfig       `yaml:"backend" json:"backend"`
	Transport     TransportConfig     `yaml:"transport" json:"transport"`
	Observability ObservabilityConfig `yaml:"observability" json:"observability"`
	Logging       LoggingConfig       `yaml:"logging" json:"logging"`
}

// BackendConfig configures the backend
type BackendConfig struct {
	Type   string                 `yaml:"type" json:"type"`
	Config map[string]interface{} `yaml:"config" json:"config"`
}

// TransportConfig configures the transport
type TransportConfig struct {
	Type string     `yaml:"type" json:"type"`
	HTTP HTTPConfig `yaml:"http" json:"http"`
}

// HTTPConfig configures the HTTP transport
type HTTPConfig struct {
	Address        string        `yaml:"address" json:"address"`
	ReadTimeout    time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout" json:"write_timeout"`
	MaxRequestSize int64         `yaml:"max_request_size" json:"max_request_size"`
	AllowedOrigins []string      `yaml:"allowed_origins" json:"allowed_origins"`
}

// ObservabilityConfig configures observability features
type ObservabilityConfig struct {
	Enabled        bool   `yaml:"enabled" json:"enabled"`
	MetricsAddress string `yaml:"metrics_address" json:"metrics_address"`
}

// LoggingConfig configures logging
type LoggingConfig struct {
	Level     string `yaml:"level" json:"level"`
	Format    string `yaml:"format" json:"format"`
	AddSource bool   `yaml:"add_source" json:"add_source"`
}
