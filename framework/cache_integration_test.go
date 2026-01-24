package framework_test

import (
	"testing"
	"time"

	"github.com/SaherElMasry/go-mcp-framework/cache"
	"github.com/SaherElMasry/go-mcp-framework/framework"
)

// Test: Server with cache enabled
func TestServer_WithCache(t *testing.T) {
	server := framework.NewServer(
		framework.WithCache("short", 60),
	)

	cacheConfig := server.GetCacheConfig()
	if cacheConfig == nil {
		t.Fatal("cache config should not be nil")
	}

	if !cacheConfig.Enabled {
		t.Error("cache should be enabled")
	}

	if cacheConfig.Type != cache.TypeShort {
		t.Errorf("Type = %v, want %v", cacheConfig.Type, cache.TypeShort)
	}

	if cacheConfig.TTL != 60 {
		t.Errorf("TTL = %d, want 60", cacheConfig.TTL)
	}
}

// Test: Server with cache disabled
func TestServer_WithCacheDisabled(t *testing.T) {
	server := framework.NewServer(
		framework.WithCacheDisabled(),
	)

	cacheConfig := server.GetCacheConfig()
	if cacheConfig == nil {
		t.Fatal("cache config should not be nil")
	}

	if cacheConfig.Enabled {
		t.Error("cache should be disabled")
	}
}

// Test: Server with per-tool TTL
func TestServer_WithToolCacheTTL(t *testing.T) {
	server := framework.NewServer(
		framework.WithCache("short", 60),
		framework.WithToolCacheTTL("search", 30*time.Second),
		framework.WithToolCacheTTL("list_files", 5*time.Minute),
	)

	cacheConfig := server.GetCacheConfig()
	if cacheConfig == nil {
		t.Fatal("cache config should not be nil")
	}

	if cacheConfig.GetToolTTL("search") != 30*time.Second {
		t.Error("search TTL not set correctly")
	}

	if cacheConfig.GetToolTTL("list_files") != 5*time.Minute {
		t.Error("list_files TTL not set correctly")
	}

	// Tool without override should use default
	if cacheConfig.GetToolTTL("unknown") != 60*time.Second {
		t.Error("unknown tool should use default TTL")
	}
}

// Test: Server with custom cache size
func TestServer_WithCacheSize(t *testing.T) {
	server := framework.NewServer(
		framework.WithCache("short", 60),
		framework.WithCacheSize(5000),
	)

	cacheConfig := server.GetCacheConfig()
	if cacheConfig == nil {
		t.Fatal("cache config should not be nil")
	}

	if cacheConfig.MaxSize != 5000 {
		t.Errorf("MaxSize = %d, want 5000", cacheConfig.MaxSize)
	}
}

// Test: Server without cache (default)
func TestServer_NoCacheByDefault(t *testing.T) {
	server := framework.NewServer()

	cacheConfig := server.GetCacheConfig()
	// Should be nil by default (no cache configured)
	if cacheConfig != nil {
		t.Error("cache config should be nil by default")
	}
}
