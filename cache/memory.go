package cache

import (
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// MemoryCache implements an in-memory LRU cache with TTL support
// Thread-safe with O(1) get/set operations
//
// Architecture:
//
//	map[string]*list.Element → O(1) lookup by key
//	doubly-linked list        → O(1) LRU eviction
//	[Front = MRU] ← → ← → [Back = LRU]
//
// On Get:  Move to front (most recently used)
// On Set:  Add to front, evict from back if full
type MemoryCache struct {
	maxSize int           // Maximum number of entries
	ttl     time.Duration // Default TTL for entries

	mu      sync.RWMutex             // Protects all fields below
	entries map[string]*list.Element // Key → list element
	lru     *list.List               // LRU eviction list

	stats CacheStats // Cache statistics
}

// cacheItem represents an item in the LRU list
type cacheItem struct {
	key   string // Cache key
	entry *Entry // Cached entry
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(maxSize int, ttl time.Duration) *MemoryCache {
	return &MemoryCache{
		maxSize: maxSize,
		ttl:     ttl,
		entries: make(map[string]*list.Element),
		lru:     list.New(),
		stats: CacheStats{
			MaxSize: maxSize,
		},
	}
}

// Get retrieves a cached entry
// Returns error if key not found or entry expired
func (c *MemoryCache) Get(ctx context.Context, key string) (*Entry, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if key exists
	element, exists := c.entries[key]
	if !exists {
		c.stats.Misses++
		c.updateHitRate()
		return nil, fmt.Errorf("cache miss: key not found")
	}

	item := element.Value.(*cacheItem)

	// Check expiration
	if item.entry.IsExpired() {
		// Remove expired entry
		c.removeElement(element)
		c.stats.Misses++
		c.stats.Evictions++
		c.updateHitRate()
		return nil, fmt.Errorf("cache miss: entry expired")
	}

	// Move to front (most recently used)
	c.lru.MoveToFront(element)

	// Update hit count
	item.entry.Hits++
	c.stats.Hits++
	c.updateHitRate()

	return item.entry, nil
}

// Set stores an entry in the cache
func (c *MemoryCache) Set(ctx context.Context, key string, value json.RawMessage, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Use provided TTL or default
	if ttl == 0 {
		ttl = c.ttl
	}

	// Create entry
	entry := &Entry{
		Key:       key,
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now(),
		Hits:      0,
	}

	// Check if entry already exists
	if element, exists := c.entries[key]; exists {
		// Update existing entry
		item := element.Value.(*cacheItem)
		item.entry = entry
		c.lru.MoveToFront(element)
	} else {
		// Add new entry
		item := &cacheItem{
			key:   key,
			entry: entry,
		}
		element := c.lru.PushFront(item)
		c.entries[key] = element

		// Evict oldest if needed
		if c.lru.Len() > c.maxSize {
			c.evictOldest()
		}
	}

	c.stats.Sets++
	c.stats.Size = len(c.entries)

	return nil
}

// Delete removes an entry from the cache
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	element, exists := c.entries[key]
	if !exists {
		return fmt.Errorf("key not found")
	}

	c.removeElement(element)
	c.stats.Deletes++
	c.stats.Size = len(c.entries)

	return nil
}

// Clear removes all entries
func (c *MemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*list.Element)
	c.lru.Init()
	c.stats.Size = 0

	return nil
}

// Stats returns cache statistics
func (c *MemoryCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to avoid race conditions
	return c.stats
}

// Close closes the cache (clears all entries)
func (c *MemoryCache) Close() error {
	return c.Clear(context.Background())
}

// evictOldest removes the least recently used entry
func (c *MemoryCache) evictOldest() {
	element := c.lru.Back()
	if element != nil {
		c.removeElement(element)
		c.stats.Evictions++
	}
}

// removeElement removes an element from the cache
func (c *MemoryCache) removeElement(element *list.Element) {
	item := element.Value.(*cacheItem)
	delete(c.entries, item.key)
	c.lru.Remove(element)
}

// updateHitRate calculates the cache hit rate
func (c *MemoryCache) updateHitRate() {
	total := c.stats.Hits + c.stats.Misses
	if total > 0 {
		c.stats.HitRate = float64(c.stats.Hits) / float64(total)
	}
}

// CleanExpired removes all expired entries
// Returns the number of entries removed
func (c *MemoryCache) CleanExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	removed := 0

	// Collect keys to remove (can't iterate and delete simultaneously)
	var keysToRemove []string
	for key, element := range c.entries {
		item := element.Value.(*cacheItem)
		if item.entry.IsExpired() {
			keysToRemove = append(keysToRemove, key)
		}
	}

	// Remove expired entries
	for _, key := range keysToRemove {
		if element, exists := c.entries[key]; exists {
			c.removeElement(element)
			removed++
		}
	}

	c.stats.Evictions += int64(removed)
	c.stats.Size = len(c.entries)

	return removed
}

// Len returns the current number of entries
func (c *MemoryCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}
