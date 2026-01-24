package backend

import (
	"context"
	"time"
)

// ToolDefinition describes a tool's interface
type ToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  []Parameter `json:"inputSchema"`
	Streaming   bool        `json:"streaming,omitempty"` // Existing: Mark streaming tools

	// NEW: Cache configuration
	Cache ToolCacheConfig `json:"cache,omitempty"`
}

// Parameter describes a tool parameter
type Parameter struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Enum        []string    `json:"enum,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Minimum     *int        `json:"minimum,omitempty"`
	Maximum     *int        `json:"maximum,omitempty"`
}

// ToolHandler is the function signature for regular tools
type ToolHandler func(ctx context.Context, args map[string]interface{}) (interface{}, error)

// ============================================================
// NEW: Cache Configuration
// ============================================================

// ToolCacheConfig defines caching behavior for a tool
type ToolCacheConfig struct {
	// Cacheable indicates if this tool's results can be cached
	// Default: false (SAFE DEFAULT - must opt-in)
	//
	// Set to true ONLY if:
	// - Tool is read-only (no side effects)
	// - Tool is deterministic (same input = same output)
	// - Tool doesn't access volatile external state
	//
	// NEVER cache:
	// - create*, update*, delete* operations
	// - Tools that write to databases/APIs
	// - Tools that read current time, random data, etc.
	// - Non-deterministic tools
	Cacheable bool `json:"cacheable"`

	// TTL overrides the default cache TTL for this tool
	// If nil, uses the global cache TTL
	// Example: 5*time.Minute for metadata lookups
	TTL *time.Duration `json:"ttl,omitempty"`

	// Tags for cache categorization (optional, future use)
	Tags []string `json:"tags,omitempty"`
}

// IsCacheable returns whether this tool can be cached
func (t *ToolDefinition) IsCacheable() bool {
	return t.Cache.Cacheable
}

// GetCacheTTL returns the cache TTL for this tool
// Falls back to defaultTTL if not specified
func (t *ToolDefinition) GetCacheTTL(defaultTTL time.Duration) time.Duration {
	if t.Cache.TTL != nil {
		return *t.Cache.TTL
	}
	return defaultTTL
}

// HasCacheTags checks if tool has specific cache tags
func (t *ToolDefinition) HasCacheTags(tags ...string) bool {
	if len(t.Cache.Tags) == 0 {
		return false
	}

	tagSet := make(map[string]bool)
	for _, tag := range t.Cache.Tags {
		tagSet[tag] = true
	}

	for _, tag := range tags {
		if !tagSet[tag] {
			return false
		}
	}
	return true
}

// ============================================================
// Cache Configuration Helpers
// ============================================================

// MakeCacheable returns a ToolCacheConfig for cacheable tools
func MakeCacheable(ttl time.Duration) ToolCacheConfig {
	return ToolCacheConfig{
		Cacheable: true,
		TTL:       &ttl,
	}
}

// MakeNonCacheable returns a ToolCacheConfig for non-cacheable tools
func MakeNonCacheable() ToolCacheConfig {
	return ToolCacheConfig{
		Cacheable: false,
	}
}

// DefaultCacheConfig returns the safe default (non-cacheable)
func DefaultCacheConfig() ToolCacheConfig {
	return ToolCacheConfig{
		Cacheable: false, // Safe default
	}
}
