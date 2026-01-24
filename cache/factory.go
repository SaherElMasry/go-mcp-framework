package cache

import (
	"fmt"
)

// New creates a new cache instance based on configuration
// Returns NoOpCache if caching is disabled
func New(config *Config) (Cache, error) {
	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid cache config: %w", err)
	}

	// Return NoOp if disabled
	if !config.Enabled {
		return NewNoOpCache(), nil
	}

	// Create cache based on type
	switch config.Type {
	case TypeShort:
		// Memory cache with TTL in seconds
		ttl := config.GetTTLDuration()
		return NewMemoryCache(config.MaxSize, ttl), nil

	case TypeLong:
		// File cache with TTL in minutes (to be implemented in Week 3)
		return nil, fmt.Errorf("file cache not yet implemented - use 'short' for now")

	default:
		return nil, fmt.Errorf("unknown cache type: %s", config.Type)
	}
}

// MustNew creates a cache or panics on error
// Useful for initialization where failure should be fatal
func MustNew(config *Config) Cache {
	cache, err := New(config)
	if err != nil {
		panic(fmt.Sprintf("failed to create cache: %v", err))
	}
	return cache
}
