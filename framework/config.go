package framework

import (
	"flag"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads configuration from file, environment, and flags
func LoadConfig(configFile string) (*Config, error) {
	config := defaultConfig()

	if configFile != "" {
		if err := loadConfigFile(configFile, config); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	loadConfigFromEnv(config)
	loadConfigFromFlags(config)

	return config, nil
}

func defaultConfig() *Config {
	return &Config{
		Backend: BackendConfig{
			Type:   "simple",
			Config: map[string]interface{}{},
		},
		Transport: TransportConfig{
			Type: "stdio",
			HTTP: HTTPConfig{
				Address:        ":8080",
				ReadTimeout:    30 * time.Second,
				WriteTimeout:   30 * time.Second,
				MaxRequestSize: 10 * 1024 * 1024,
			},
		},
		Observability: ObservabilityConfig{
			Enabled:        false,
			MetricsAddress: ":9091",
		},
		Logging: LoggingConfig{
			Level:     "info",
			Format:    "json",
			AddSource: true,
		},
	}
}

func loadConfigFile(path string, config *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, config)
}

func loadConfigFromEnv(config *Config) {
	if v := os.Getenv("MCP_BACKEND_TYPE"); v != "" {
		config.Backend.Type = v
	}
	if v := os.Getenv("MCP_TRANSPORT"); v != "" {
		config.Transport.Type = v
	}
	if v := os.Getenv("MCP_LOG_LEVEL"); v != "" {
		config.Logging.Level = v
	}
}

func loadConfigFromFlags(config *Config) {
	flag.StringVar(&config.Backend.Type, "backend", config.Backend.Type, "Backend type")
	flag.StringVar(&config.Transport.Type, "transport", config.Transport.Type, "Transport type")
	flag.StringVar(&config.Transport.HTTP.Address, "http-addr", config.Transport.HTTP.Address, "HTTP address")
	flag.StringVar(&config.Logging.Level, "log-level", config.Logging.Level, "Log level")
	flag.Parse()
}
