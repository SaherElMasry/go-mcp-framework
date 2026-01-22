// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	GitHub GitHubConfig `yaml:"github"`
}

// GitHubConfig holds GitHub-specific configuration
type GitHubConfig struct {
	// Personal Access Token for authentication
	Token string `yaml:"token"`

	// Base URL for GitHub API (default: https://api.github.com)
	BaseURL string `yaml:"base_url"`

	// Request timeout
	Timeout time.Duration `yaml:"timeout"`

	// User agent string (optional)
	UserAgent string `yaml:"user_agent"`
}

// Load loads configuration from a YAML file
func Load(filepath string) (*Config, error) {
	// Expand environment variables in filepath
	filepath = os.ExpandEnv(filepath)

	// Read file
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults
	if config.GitHub.BaseURL == "" {
		config.GitHub.BaseURL = "https://api.github.com"
	}

	if config.GitHub.Timeout == 0 {
		config.GitHub.Timeout = 30 * time.Second
	}

	if config.GitHub.UserAgent == "" {
		config.GitHub.UserAgent = "github-mcp-server/1.0"
	}

	// Validate
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.GitHub.Token == "" {
		return fmt.Errorf("github.token is required")
	}

	if c.GitHub.Token == "xxx" {
		return fmt.Errorf("github.token must be set to your actual token (not placeholder)")
	}

	// Basic token format validation
	// Personal Access Tokens start with "ghp_" for classic tokens
	// or "github_pat_" for fine-grained tokens
	if len(c.GitHub.Token) < 10 {
		return fmt.Errorf("github.token appears to be invalid (too short)")
	}

	if c.GitHub.BaseURL == "" {
		return fmt.Errorf("github.base_url is required")
	}

	if c.GitHub.Timeout < 0 {
		return fmt.Errorf("github.timeout must be positive")
	}

	return nil
}

// LoadFromEnv loads configuration with environment variable overrides
func LoadFromEnv(filepath string) (*Config, error) {
	config, err := Load(filepath)
	if err != nil {
		return nil, err
	}

	// Override with environment variables if set
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		config.GitHub.Token = token
	}

	if baseURL := os.Getenv("GITHUB_BASE_URL"); baseURL != "" {
		config.GitHub.BaseURL = baseURL
	}

	// Re-validate after env overrides
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}
