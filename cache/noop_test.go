package cache_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/SaherElMasry/go-mcp-framework/cache"
)

// Test: NoOpCache creation
func TestNoOpCache_New(t *testing.T) {
	nc := cache.NewNoOpCache()

	if nc == nil {
		t.Fatal("NewNoOpCache() returned nil")
	}

	stats := nc.Stats()
	if stats.MaxSize != 0 {
		t.Errorf("NoOpCache MaxSize = %d, want 0", stats.MaxSize)
	}

	if stats.Size != 0 {
		t.Errorf("NoOpCache Size = %d, want 0", stats.Size)
	}
}

// Test: Get always misses
func TestNoOpCache_Get(t *testing.T) {
	nc := cache.NewNoOpCache()
	ctx := context.Background()

	_, err := nc.Get(ctx, "any-key")
	if err == nil {
		t.Error("NoOpCache.Get() should always return error")
	}

	// Verify miss was recorded
	stats := nc.Stats()
	if stats.Misses != 1 {
		t.Errorf("Misses = %d, want 1", stats.Misses)
	}
}

// Test: Set does nothing
func TestNoOpCache_Set(t *testing.T) {
	nc := cache.NewNoOpCache()
	ctx := context.Background()

	value := json.RawMessage(`{"data":"test"}`)
	err := nc.Set(ctx, "key", value, time.Minute)

	if err != nil {
		t.Errorf("NoOpCache.Set() should not return error, got %v", err)
	}

	// Try to get - should still fail
	_, err = nc.Get(ctx, "key")
	if err == nil {
		t.Error("NoOpCache should not store values")
	}
}

// Test: Delete returns error
func TestNoOpCache_Delete(t *testing.T) {
	nc := cache.NewNoOpCache()
	ctx := context.Background()

	err := nc.Delete(ctx, "any-key")
	if err == nil {
		t.Error("NoOpCache.Delete() should return error")
	}
}

// Test: Clear succeeds but does nothing
func TestNoOpCache_Clear(t *testing.T) {
	nc := cache.NewNoOpCache()
	ctx := context.Background()

	err := nc.Clear(ctx)
	if err != nil {
		t.Errorf("NoOpCache.Clear() should succeed, got %v", err)
	}
}

// Test: Stats returns empty
func TestNoOpCache_Stats(t *testing.T) {
	nc := cache.NewNoOpCache()

	stats := nc.Stats()

	if stats.MaxSize != 0 {
		t.Errorf("MaxSize = %d, want 0", stats.MaxSize)
	}

	if stats.Size != 0 {
		t.Errorf("Size = %d, want 0", stats.Size)
	}

	if stats.Hits != 0 {
		t.Errorf("Hits = %d, want 0", stats.Hits)
	}
}

// Test: Close succeeds
func TestNoOpCache_Close(t *testing.T) {
	nc := cache.NewNoOpCache()

	err := nc.Close()
	if err != nil {
		t.Errorf("NoOpCache.Close() should succeed, got %v", err)
	}
}

// Test: Multiple operations
func TestNoOpCache_MultipleOperations(t *testing.T) {
	nc := cache.NewNoOpCache()
	ctx := context.Background()

	// Multiple sets (all ignored)
	for i := 0; i < 10; i++ {
		value := json.RawMessage(`{"data":"test"}`)
		nc.Set(ctx, "key", value, time.Minute)
	}

	// Multiple gets (all miss)
	for i := 0; i < 5; i++ {
		nc.Get(ctx, "key")
	}

	stats := nc.Stats()

	// Only misses should be recorded
	if stats.Misses != 5 {
		t.Errorf("Misses = %d, want 5", stats.Misses)
	}

	if stats.Hits != 0 {
		t.Errorf("Hits = %d, want 0", stats.Hits)
	}

	if stats.Sets != 0 {
		t.Errorf("Sets = %d, want 0", stats.Sets)
	}
}
