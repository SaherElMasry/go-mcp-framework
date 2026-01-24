package cache_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/SaherElMasry/go-mcp-framework/cache"
)

// Test: Constructor
func TestMemoryCache_NewMemoryCache(t *testing.T) {
	maxSize := 100
	ttl := 5 * time.Minute

	mc := cache.NewMemoryCache(maxSize, ttl)

	if mc == nil {
		t.Fatal("NewMemoryCache() returned nil")
	}

	stats := mc.Stats()
	if stats.MaxSize != maxSize {
		t.Errorf("MaxSize = %d, want %d", stats.MaxSize, maxSize)
	}

	if stats.Size != 0 {
		t.Errorf("initial Size = %d, want 0", stats.Size)
	}
}

// Test: Basic Set and Get
func TestMemoryCache_SetAndGet(t *testing.T) {
	mc := cache.NewMemoryCache(10, time.Minute)
	ctx := context.Background()

	key := "test-key"
	value := json.RawMessage(`{"message":"hello"}`)

	// Set value
	err := mc.Set(ctx, key, value, 0)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Get value
	entry, err := mc.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if string(entry.Value) != string(value) {
		t.Errorf("Value = %s, want %s", entry.Value, value)
	}

	// Verify stats
	stats := mc.Stats()
	if stats.Sets != 1 {
		t.Errorf("Sets = %d, want 1", stats.Sets)
	}

	if stats.Hits != 1 {
		t.Errorf("Hits = %d, want 1", stats.Hits)
	}

	if stats.Size != 1 {
		t.Errorf("Size = %d, want 1", stats.Size)
	}
}

// Test: Cache Miss
func TestMemoryCache_CacheMiss(t *testing.T) {
	mc := cache.NewMemoryCache(10, time.Minute)
	ctx := context.Background()

	// Get non-existent key
	_, err := mc.Get(ctx, "non-existent")
	if err == nil {
		t.Error("Get() for non-existent key should return error")
	}

	// Verify miss was recorded
	stats := mc.Stats()
	if stats.Misses != 1 {
		t.Errorf("Misses = %d, want 1", stats.Misses)
	}
}

// Test: Update Existing Entry
func TestMemoryCache_Update(t *testing.T) {
	mc := cache.NewMemoryCache(10, time.Minute)
	ctx := context.Background()

	key := "test-key"
	value1 := json.RawMessage(`{"version":1}`)
	value2 := json.RawMessage(`{"version":2}`)

	// Set initial value
	mc.Set(ctx, key, value1, 0)

	// Update value
	mc.Set(ctx, key, value2, 0)

	// Get updated value
	entry, _ := mc.Get(ctx, key)
	if string(entry.Value) != string(value2) {
		t.Errorf("Value = %s, want %s", entry.Value, value2)
	}

	// Should still have only 1 entry
	if mc.Len() != 1 {
		t.Errorf("Len() = %d, want 1", mc.Len())
	}
}

// Test: Delete
func TestMemoryCache_Delete(t *testing.T) {
	mc := cache.NewMemoryCache(10, time.Minute)
	ctx := context.Background()

	key := "test-key"
	value := json.RawMessage(`{"data":"test"}`)

	// Set and verify
	mc.Set(ctx, key, value, 0)
	if mc.Len() != 1 {
		t.Error("entry should exist")
	}

	// Delete
	err := mc.Delete(ctx, key)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deleted
	if mc.Len() != 0 {
		t.Error("entry should be deleted")
	}

	// Try to get deleted entry
	_, err = mc.Get(ctx, key)
	if err == nil {
		t.Error("Get() after delete should return error")
	}

	// Verify stats
	stats := mc.Stats()
	if stats.Deletes != 1 {
		t.Errorf("Deletes = %d, want 1", stats.Deletes)
	}
}

// Test: Delete Non-Existent
func TestMemoryCache_DeleteNonExistent(t *testing.T) {
	mc := cache.NewMemoryCache(10, time.Minute)
	ctx := context.Background()

	err := mc.Delete(ctx, "non-existent")
	if err == nil {
		t.Error("Delete() for non-existent key should return error")
	}
}

// Test: Clear
func TestMemoryCache_Clear(t *testing.T) {
	mc := cache.NewMemoryCache(10, time.Minute)
	ctx := context.Background()

	// Add multiple entries
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := json.RawMessage(fmt.Sprintf(`{"id":%d}`, i))
		mc.Set(ctx, key, value, 0)
	}

	if mc.Len() != 5 {
		t.Errorf("Len() = %d, want 5", mc.Len())
	}

	// Clear all
	err := mc.Clear(ctx)
	if err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	if mc.Len() != 0 {
		t.Errorf("Len() after Clear() = %d, want 0", mc.Len())
	}

	stats := mc.Stats()
	if stats.Size != 0 {
		t.Errorf("Size after Clear() = %d, want 0", stats.Size)
	}
}

// Test: LRU Eviction
func TestMemoryCache_LRUEviction(t *testing.T) {
	maxSize := 3
	mc := cache.NewMemoryCache(maxSize, time.Minute)
	ctx := context.Background()

	// Add 3 entries (fill cache)
	mc.Set(ctx, "key1", json.RawMessage(`{"id":1}`), 0)
	mc.Set(ctx, "key2", json.RawMessage(`{"id":2}`), 0)
	mc.Set(ctx, "key3", json.RawMessage(`{"id":3}`), 0)

	if mc.Len() != 3 {
		t.Errorf("Len() = %d, want 3", mc.Len())
	}

	// Add 4th entry - should evict oldest (key1)
	mc.Set(ctx, "key4", json.RawMessage(`{"id":4}`), 0)

	// Cache should still have 3 entries
	if mc.Len() != maxSize {
		t.Errorf("Len() = %d, want %d", mc.Len(), maxSize)
	}

	// key1 should be evicted
	_, err := mc.Get(ctx, "key1")
	if err == nil {
		t.Error("key1 should have been evicted")
	}

	// Other keys should still exist
	for _, key := range []string{"key2", "key3", "key4"} {
		if _, err := mc.Get(ctx, key); err != nil {
			t.Errorf("Get(%s) should succeed, got error: %v", key, err)
		}
	}

	// Verify eviction was recorded
	stats := mc.Stats()
	if stats.Evictions != 1 {
		t.Errorf("Evictions = %d, want 1", stats.Evictions)
	}
}

// Test: LRU Ordering (Access Updates Order)
func TestMemoryCache_LRUOrdering(t *testing.T) {
	mc := cache.NewMemoryCache(3, time.Minute)
	ctx := context.Background()

	// Add 3 entries
	mc.Set(ctx, "key1", json.RawMessage(`{"id":1}`), 0)
	mc.Set(ctx, "key2", json.RawMessage(`{"id":2}`), 0)
	mc.Set(ctx, "key3", json.RawMessage(`{"id":3}`), 0)

	// Access key1 (moves to front)
	mc.Get(ctx, "key1")

	// Add key4 - should evict key2 (now oldest)
	mc.Set(ctx, "key4", json.RawMessage(`{"id":4}`), 0)

	// key2 should be evicted (not key1)
	_, err := mc.Get(ctx, "key2")
	if err == nil {
		t.Error("key2 should have been evicted")
	}

	// key1 should still exist (was accessed recently)
	_, err = mc.Get(ctx, "key1")
	if err != nil {
		t.Error("key1 should still exist (was accessed recently)")
	}
}

// Test: TTL Expiration
func TestMemoryCache_TTLExpiration(t *testing.T) {
	mc := cache.NewMemoryCache(10, 100*time.Millisecond)
	ctx := context.Background()

	key := "test-key"
	value := json.RawMessage(`{"data":"test"}`)

	// Set with short TTL
	mc.Set(ctx, key, value, 50*time.Millisecond)

	// Should exist immediately
	_, err := mc.Get(ctx, key)
	if err != nil {
		t.Error("entry should exist immediately after set")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	_, err = mc.Get(ctx, key)
	if err == nil {
		t.Error("entry should be expired")
	}

	// Entry should be removed on access
	if mc.Len() != 0 {
		t.Error("expired entry should be removed")
	}

	// Verify expiration was recorded
	stats := mc.Stats()
	if stats.Evictions != 1 {
		t.Errorf("Evictions = %d, want 1 (expired entry)", stats.Evictions)
	}
}

// Test: Custom TTL
func TestMemoryCache_CustomTTL(t *testing.T) {
	mc := cache.NewMemoryCache(10, time.Hour) // Default: 1 hour
	ctx := context.Background()

	// Set with custom short TTL
	key := "short-ttl"
	value := json.RawMessage(`{"data":"test"}`)
	mc.Set(ctx, key, value, 50*time.Millisecond)

	// Get entry and check TTL
	entry, _ := mc.Get(ctx, key)
	ttl := entry.TTL()

	// TTL should be less than default (hour)
	if ttl > time.Minute {
		t.Errorf("custom TTL not applied, got %v", ttl)
	}
}

// Test: CleanExpired
func TestMemoryCache_CleanExpired(t *testing.T) {
	mc := cache.NewMemoryCache(10, 100*time.Millisecond)
	ctx := context.Background()

	// Add entries with different TTLs
	mc.Set(ctx, "key1", json.RawMessage(`{"id":1}`), 50*time.Millisecond) // Expires soon
	mc.Set(ctx, "key2", json.RawMessage(`{"id":2}`), 50*time.Millisecond) // Expires soon
	mc.Set(ctx, "key3", json.RawMessage(`{"id":3}`), 10*time.Second)      // Long TTL

	// Wait for some to expire
	time.Sleep(100 * time.Millisecond)

	// Clean expired entries
	removed := mc.CleanExpired()

	if removed != 2 {
		t.Errorf("CleanExpired() removed %d entries, want 2", removed)
	}

	if mc.Len() != 1 {
		t.Errorf("Len() = %d, want 1 (only key3 should remain)", mc.Len())
	}

	// key3 should still exist
	_, err := mc.Get(ctx, "key3")
	if err != nil {
		t.Error("key3 should still exist")
	}
}

// ===================================================================
// PART 2: CONCURRENCY, STATISTICS, EDGE CASES, BENCHMARKS
// ===================================================================

// Test: Hit Rate Calculation
func TestMemoryCache_HitRate(t *testing.T) {
	mc := cache.NewMemoryCache(10, time.Minute)
	ctx := context.Background()

	key := "test-key"
	value := json.RawMessage(`{"data":"test"}`)

	mc.Set(ctx, key, value, 0)

	// 3 hits
	mc.Get(ctx, key)
	mc.Get(ctx, key)
	mc.Get(ctx, key)

	// 2 misses
	mc.Get(ctx, "miss1")
	mc.Get(ctx, "miss2")

	stats := mc.Stats()

	// 3 hits, 2 misses = 60% hit rate
	expectedHitRate := 3.0 / 5.0
	if stats.HitRate != expectedHitRate {
		t.Errorf("HitRate = %f, want %f", stats.HitRate, expectedHitRate)
	}

	if stats.Hits != 3 {
		t.Errorf("Hits = %d, want 3", stats.Hits)
	}

	if stats.Misses != 2 {
		t.Errorf("Misses = %d, want 2", stats.Misses)
	}
}

// Test: Entry Hit Count
func TestMemoryCache_EntryHitCount(t *testing.T) {
	mc := cache.NewMemoryCache(10, time.Minute)
	ctx := context.Background()

	key := "test-key"
	value := json.RawMessage(`{"data":"test"}`)

	mc.Set(ctx, key, value, 0)

	// Access multiple times
	for i := 0; i < 5; i++ {
		mc.Get(ctx, key)
	}

	// Get entry and check hit count
	entry, _ := mc.Get(ctx, key)

	// Should be 6 (5 previous + 1 from this Get)
	if entry.Hits != 6 {
		t.Errorf("entry.Hits = %d, want 6", entry.Hits)
	}
}

// Test: Concurrent Writes
func TestMemoryCache_ConcurrentWrites(t *testing.T) {
	mc := cache.NewMemoryCache(1000, time.Minute)
	ctx := context.Background()

	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				value := json.RawMessage(fmt.Sprintf(`{"id":%d}`, j))
				mc.Set(ctx, key, value, 0)
			}
		}(i)
	}

	wg.Wait()

	// Should have 1000 entries (no race conditions)
	expectedEntries := numGoroutines * numOperations
	if mc.Len() != expectedEntries {
		t.Errorf("Len() = %d, want %d", mc.Len(), expectedEntries)
	}
}

// Test: Concurrent Read/Write
func TestMemoryCache_ConcurrentReadWrite(t *testing.T) {
	mc := cache.NewMemoryCache(100, time.Minute)
	ctx := context.Background()

	// Pre-populate
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := json.RawMessage(fmt.Sprintf(`{"id":%d}`, i))
		mc.Set(ctx, key, value, 0)
	}

	var wg sync.WaitGroup

	// Concurrent readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key-%d", j%50)
				mc.Get(ctx, key)
			}
		}()
	}

	// Concurrent writers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key-%d", j%50)
				value := json.RawMessage(fmt.Sprintf(`{"id":%d}`, id))
				mc.Set(ctx, key, value, 0)
			}
		}(i)
	}

	wg.Wait()

	// No crashes = success
	t.Log("Concurrent read/write completed successfully")
}

// Test: Concurrent Deletes
func TestMemoryCache_ConcurrentDeletes(t *testing.T) {
	mc := cache.NewMemoryCache(100, time.Minute)
	ctx := context.Background()

	// Pre-populate
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := json.RawMessage(fmt.Sprintf(`{"id":%d}`, i))
		mc.Set(ctx, key, value, 0)
	}

	var wg sync.WaitGroup

	// Concurrent deletes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				key := fmt.Sprintf("key-%d", id*10+j)
				mc.Delete(ctx, key)
			}
		}(i)
	}

	wg.Wait()

	// All should be deleted
	if mc.Len() != 0 {
		t.Errorf("Len() = %d, want 0 (all deleted)", mc.Len())
	}
}

// Test: Edge Case - Zero MaxSize
func TestMemoryCache_ZeroMaxSize(t *testing.T) {
	// MaxSize 0 should work but immediately evict
	mc := cache.NewMemoryCache(0, time.Minute)
	ctx := context.Background()

	err := mc.Set(ctx, "key", json.RawMessage(`{"data":"test"}`), 0)
	if err != nil {
		t.Errorf("Set() should succeed even with maxSize=0")
	}

	// Should immediately evict
	if mc.Len() != 0 {
		t.Error("cache with maxSize=0 should not store entries")
	}
}

// Test: Edge Case - Large Value
func TestMemoryCache_LargeValue(t *testing.T) {
	mc := cache.NewMemoryCache(10, time.Minute)
	ctx := context.Background()

	// Create large JSON value (1MB)
	largeData := make([]byte, 1024*1024)
	for i := range largeData {
		largeData[i] = 'a'
	}
	value := json.RawMessage(fmt.Sprintf(`{"data":"%s"}`, string(largeData)))

	err := mc.Set(ctx, "large-key", value, 0)
	if err != nil {
		t.Fatalf("Set() with large value error = %v", err)
	}

	entry, err := mc.Get(ctx, "large-key")
	if err != nil {
		t.Fatalf("Get() with large value error = %v", err)
	}

	if len(entry.Value) != len(value) {
		t.Errorf("large value size mismatch")
	}
}

// Test: Edge Case - Many Small Entries
func TestMemoryCache_ManySmallEntries(t *testing.T) {
	mc := cache.NewMemoryCache(1000, time.Minute)
	ctx := context.Background()

	// Add 1000 small entries
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := json.RawMessage(fmt.Sprintf(`{"id":%d}`, i))
		mc.Set(ctx, key, value, 0)
	}

	if mc.Len() != 1000 {
		t.Errorf("Len() = %d, want 1000", mc.Len())
	}

	// Verify all can be retrieved
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d", i)
		_, err := mc.Get(ctx, key)
		if err != nil {
			t.Errorf("Get(%s) should succeed", key)
			break
		}
	}
}

// Test: Edge Case - Rapid Expiration
func TestMemoryCache_RapidExpiration(t *testing.T) {
	mc := cache.NewMemoryCache(100, time.Millisecond)
	ctx := context.Background()

	// Add entries with very short TTL
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := json.RawMessage(fmt.Sprintf(`{"id":%d}`, i))
		mc.Set(ctx, key, value, 1*time.Millisecond)
	}

	// Wait for all to expire
	time.Sleep(10 * time.Millisecond)

	// Try to access - should all be expired
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key-%d", i)
		_, err := mc.Get(ctx, key)
		if err == nil {
			t.Errorf("Get(%s) should fail (expired)", key)
		}
	}

	// Cache should be empty after accessing expired entries
	if mc.Len() != 0 {
		t.Errorf("Len() = %d, want 0 (all expired)", mc.Len())
	}
}

// Test: Close
func TestMemoryCache_Close(t *testing.T) {
	mc := cache.NewMemoryCache(10, time.Minute)
	ctx := context.Background()

	// Add some entries
	mc.Set(ctx, "key1", json.RawMessage(`{"id":1}`), 0)
	mc.Set(ctx, "key2", json.RawMessage(`{"id":2}`), 0)

	// Close
	err := mc.Close()
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Should be empty
	if mc.Len() != 0 {
		t.Errorf("Len() after Close() = %d, want 0", mc.Len())
	}
}

// Test: Stats Copy (No Race)
func TestMemoryCache_StatsCopy(t *testing.T) {
	mc := cache.NewMemoryCache(10, time.Minute)
	ctx := context.Background()

	// Get stats
	stats1 := mc.Stats()

	// Modify cache
	mc.Set(ctx, "key", json.RawMessage(`{"data":"test"}`), 0)

	// Get stats again
	stats2 := mc.Stats()

	// stats1 should not be affected (it's a copy)
	if stats1.Sets != 0 {
		t.Error("stats1 should not be affected by later operations")
	}

	if stats2.Sets != 1 {
		t.Error("stats2 should reflect new operation")
	}
}

// Test: Multiple Operations
func TestMemoryCache_MultipleOperations(t *testing.T) {
	mc := cache.NewMemoryCache(10, time.Minute)
	ctx := context.Background()

	// Complex sequence of operations
	mc.Set(ctx, "key1", json.RawMessage(`{"id":1}`), 0)
	mc.Set(ctx, "key2", json.RawMessage(`{"id":2}`), 0)
	mc.Get(ctx, "key1")
	mc.Delete(ctx, "key2")
	mc.Set(ctx, "key3", json.RawMessage(`{"id":3}`), 0)
	mc.Get(ctx, "key1")
	mc.Get(ctx, "key3")
	mc.Get(ctx, "missing")

	stats := mc.Stats()

	// Verify stats
	if stats.Sets != 3 {
		t.Errorf("Sets = %d, want 3", stats.Sets)
	}

	if stats.Hits != 3 {
		t.Errorf("Hits = %d, want 3", stats.Hits)
	}

	if stats.Misses != 1 {
		t.Errorf("Misses = %d, want 1", stats.Misses)
	}

	if stats.Deletes != 1 {
		t.Errorf("Deletes = %d, want 1", stats.Deletes)
	}

	if mc.Len() != 2 {
		t.Errorf("Len() = %d, want 2", mc.Len())
	}
}

// ===================================================================
// BENCHMARKS
// ===================================================================

// Benchmark: Set
func BenchmarkMemoryCache_Set(b *testing.B) {
	mc := cache.NewMemoryCache(10000, time.Minute)
	ctx := context.Background()
	value := json.RawMessage(`{"data":"test"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		mc.Set(ctx, key, value, 0)
	}
}

// Benchmark: Get
func BenchmarkMemoryCache_Get(b *testing.B) {
	mc := cache.NewMemoryCache(10000, time.Minute)
	ctx := context.Background()

	// Pre-populate
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := json.RawMessage(`{"data":"test"}`)
		mc.Set(ctx, key, value, 0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%1000)
		mc.Get(ctx, key)
	}
}

// Benchmark: Set Parallel
func BenchmarkMemoryCache_SetParallel(b *testing.B) {
	mc := cache.NewMemoryCache(10000, time.Minute)
	ctx := context.Background()
	value := json.RawMessage(`{"data":"test"}`)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", i)
			mc.Set(ctx, key, value, 0)
			i++
		}
	})
}

// Benchmark: Get Parallel
func BenchmarkMemoryCache_GetParallel(b *testing.B) {
	mc := cache.NewMemoryCache(10000, time.Minute)
	ctx := context.Background()

	// Pre-populate
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := json.RawMessage(`{"data":"test"}`)
		mc.Set(ctx, key, value, 0)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", i%1000)
			mc.Get(ctx, key)
			i++
		}
	})
}

// Benchmark: CleanExpired
func BenchmarkMemoryCache_CleanExpired(b *testing.B) {
	mc := cache.NewMemoryCache(10000, time.Minute)
	ctx := context.Background()

	// Pre-populate with expired entries
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := json.RawMessage(`{"data":"test"}`)
		mc.Set(ctx, key, value, 1*time.Nanosecond) // Already expired
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mc.CleanExpired()

		// Re-populate for next iteration
		if i < b.N-1 {
			for j := 0; j < 1000; j++ {
				key := fmt.Sprintf("key-%d", j)
				value := json.RawMessage(`{"data":"test"}`)
				mc.Set(ctx, key, value, 1*time.Nanosecond)
			}
		}
	}
}

// Benchmark: LRU Eviction
func BenchmarkMemoryCache_LRUEviction(b *testing.B) {
	mc := cache.NewMemoryCache(1000, time.Minute)
	ctx := context.Background()

	// Fill cache
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := json.RawMessage(`{"data":"test"}`)
		mc.Set(ctx, key, value, 0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", 1000+i)
		value := json.RawMessage(`{"data":"test"}`)
		mc.Set(ctx, key, value, 0) // Triggers eviction
	}
}
