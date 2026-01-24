// Package cache provides response caching for MCP tools
package cache

import (
	"context"
	"encoding/json"
	"time"
)

// Cache defines the interface for response caching
// All implementations must be thread-safe
type Cache interface {
	// Get retrieves a cached response
	// Returns error if key not found or entry expired
	Get(ctx context.Context, key string) (*Entry, error)

	// Set stores a response in the cache
	// value MUST be JSON-serializable (json.RawMessage)
	Set(ctx context.Context, key string, value json.RawMessage, ttl time.Duration) error

	// Delete removes a cached entry
	// Returns error if key not found
	Delete(ctx context.Context, key string) error

	// Clear removes all cached entries
	Clear(ctx context.Context) error

	// Stats returns cache statistics
	Stats() CacheStats

	// Close closes the cache and releases resources
	Close() error
}

// Entry represents a cached item
// Uses json.RawMessage to ensure JSON-safe storage
type Entry struct {
	Key       string          `json:"key"`        // Cache key (SHA-256 hash)
	Value     json.RawMessage `json:"value"`      // Cached JSON response
	ExpiresAt time.Time       `json:"expires_at"` // Expiration timestamp
	CreatedAt time.Time       `json:"created_at"` // Creation timestamp
	Hits      int64           `json:"hits"`       // Number of cache hits
}

// CacheStats holds cache statistics
type CacheStats struct {
	Hits      int64   `json:"hits"`      // Total cache hits
	Misses    int64   `json:"misses"`    // Total cache misses
	Sets      int64   `json:"sets"`      // Total set operations
	Deletes   int64   `json:"deletes"`   // Total delete operations
	Evictions int64   `json:"evictions"` // Total evictions (LRU/TTL)
	Size      int     `json:"size"`      // Current number of entries
	MaxSize   int     `json:"max_size"`  // Maximum capacity
	HitRate   float64 `json:"hit_rate"`  // Hit rate (hits / (hits + misses))
}

// IsExpired checks if the entry has expired
func (e *Entry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// TTL returns the remaining time to live
// Returns 0 if expired
func (e *Entry) TTL() time.Duration {
	remaining := time.Until(e.ExpiresAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Unmarshal unmarshals the cached JSON value into v
// v must be a pointer to the target type
func (e *Entry) Unmarshal(v interface{}) error {
	return json.Unmarshal(e.Value, v)
}

// Age returns how long ago the entry was created
func (e *Entry) Age() time.Duration {
	return time.Since(e.CreatedAt)
}
