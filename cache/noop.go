package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// NoOpCache is a cache implementation that does nothing
// Used when caching is disabled
type NoOpCache struct {
	stats CacheStats
}

// NewNoOpCache creates a new no-op cache
func NewNoOpCache() *NoOpCache {
	return &NoOpCache{
		stats: CacheStats{
			MaxSize: 0,
			Size:    0,
		},
	}
}

// Get always returns cache miss
func (c *NoOpCache) Get(ctx context.Context, key string) (*Entry, error) {
	c.stats.Misses++
	return nil, fmt.Errorf("cache disabled")
}

// Set does nothing
func (c *NoOpCache) Set(ctx context.Context, key string, value json.RawMessage, ttl time.Duration) error {
	// Silently ignore - cache is disabled
	return nil
}

// Delete does nothing
func (c *NoOpCache) Delete(ctx context.Context, key string) error {
	return fmt.Errorf("cache disabled")
}

// Clear does nothing
func (c *NoOpCache) Clear(ctx context.Context) error {
	return nil
}

// Stats returns empty stats
func (c *NoOpCache) Stats() CacheStats {
	return c.stats
}

// Close does nothing
func (c *NoOpCache) Close() error {
	return nil
}
