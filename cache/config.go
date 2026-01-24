package cache

import (
	"fmt"
	"time"
)

// Type represents the cache type
type Type string

const (
	// TypeShort represents in-memory cache with TTL in seconds
	TypeShort Type = "short"

	// TypeLong represents file-based cache with TTL in minutes
	TypeLong Type = "long"
)

// Config holds cache configuration
type Config struct {
	// Type of cache ("short" or "long")
	Type Type `json:"type" yaml:"type"`

	// TTL is the time to live for cached entries
	// For TypeShort: seconds
	// For TypeLong: minutes
	TTL int `json:"ttl" yaml:"ttl"`

	// MaxSize is the maximum number of entries (for memory cache)
	// Ignored for file-based cache
	MaxSize int `json:"max_size" yaml:"max_size"`

	// Directory for file-based cache
	// Ignored for memory cache
	Directory string `json:"directory" yaml:"directory"`

	// Enabled enables/disables caching globally
	// Default: false (SAFE DEFAULT - must opt-in)
	// This is intentional for safety - caching is disabled by default
	Enabled bool `json:"enabled" yaml:"enabled"`

	// ToolTTL provides per-tool TTL overrides
	// Key: tool name, Value: TTL duration
	ToolTTL map[string]time.Duration `json:"tool_ttl,omitempty" yaml:"tool_ttl,omitempty"`
}

// DefaultConfig returns the default cache configuration
// Cache is DISABLED by default for safety
// Tools must explicitly opt-in to caching
func DefaultConfig() *Config {
	return &Config{
		Type:      TypeShort,
		TTL:       60,   // 60 seconds
		MaxSize:   1000, // 1000 entries
		Directory: ".mcp-cache",
		Enabled:   false, // ⚠️ DISABLED BY DEFAULT (safe default)
		ToolTTL:   make(map[string]time.Duration),
	}
}

// Validate validates the cache configuration
// If cache is disabled, no validation is performed
func (c *Config) Validate() error {
	if !c.Enabled {
		return nil // No validation needed if disabled
	}

	// Validate type
	if c.Type != TypeShort && c.Type != TypeLong {
		return fmt.Errorf("invalid cache type: %s (must be 'short' or 'long')", c.Type)
	}

	// Validate TTL
	if c.TTL <= 0 {
		return fmt.Errorf("TTL must be positive, got %d", c.TTL)
	}

	// Validate MaxSize for memory cache
	if c.Type == TypeShort && c.MaxSize <= 0 {
		return fmt.Errorf("max_size must be positive for memory cache, got %d", c.MaxSize)
	}

	// Validate Directory for file cache
	if c.Type == TypeLong && c.Directory == "" {
		return fmt.Errorf("directory is required for file cache")
	}

	return nil
}

// GetTTLDuration returns the default TTL as a time.Duration
func (c *Config) GetTTLDuration() time.Duration {
	switch c.Type {
	case TypeShort:
		return time.Duration(c.TTL) * time.Second
	case TypeLong:
		return time.Duration(c.TTL) * time.Minute
	default:
		return time.Duration(c.TTL) * time.Second
	}
}

// GetToolTTL returns TTL for a specific tool
// Returns default TTL if no override exists
func (c *Config) GetToolTTL(toolName string) time.Duration {
	if ttl, ok := c.ToolTTL[toolName]; ok {
		return ttl
	}
	return c.GetTTLDuration()
}

// SetToolTTL sets a TTL override for a specific tool
func (c *Config) SetToolTTL(toolName string, ttl time.Duration) {
	if c.ToolTTL == nil {
		c.ToolTTL = make(map[string]time.Duration)
	}
	c.ToolTTL[toolName] = ttl
}

// Enable enables caching
func (c *Config) Enable() {
	c.Enabled = true
}

// Disable disables caching
func (c *Config) Disable() {
	c.Enabled = false
}

// IsEnabled returns whether caching is enabled
func (c *Config) IsEnabled() bool {
	return c.Enabled
}
