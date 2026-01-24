package protocol_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/SaherElMasry/go-mcp-framework/backend"
	"github.com/SaherElMasry/go-mcp-framework/cache"
	"github.com/SaherElMasry/go-mcp-framework/protocol"
)

// mockBackend for testing
type mockBackend struct {
	*backend.BaseBackend
	callCount int
}

func newMockBackend() *mockBackend {
	base := backend.NewBaseBackend("mock")
	mb := &mockBackend{
		BaseBackend: base,
		callCount:   0,
	}

	// Register cacheable tool
	tool := backend.NewTool("read_file").
		Description("Reads a file").
		StringParam("path", "File path", true).
		WithCache(true, 5*time.Minute).
		Build()

	mb.RegisterTool(tool, func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		mb.callCount++
		return map[string]interface{}{
			"content": "file contents",
			"size":    100,
		}, nil
	})

	// Register non-cacheable tool
	tool2 := backend.NewTool("create_file").
		Description("Creates a file").
		StringParam("path", "File path", true).
		NonCacheable().
		Build()

	mb.RegisterTool(tool2, func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		mb.callCount++
		return map[string]interface{}{
			"status": "created",
		}, nil
	})

	return mb
}

// Test: Cacheable tool is cached
func TestHandler_CacheableToolIsCached(t *testing.T) {
	mb := newMockBackend()
	handler := protocol.NewHandler(mb, nil)

	// Configure cache
	cacheConfig := &cache.Config{
		Type:    cache.TypeShort,
		TTL:     60,
		MaxSize: 100,
		Enabled: true,
	}
	c, _ := cache.New(cacheConfig)
	keyGen := cache.NewKeyGenerator()
	handler.SetCache(c, keyGen, cacheConfig)

	ctx := context.Background()

	// Create request
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "read_file",
			"arguments": map[string]interface{}{
				"path": "/test/file.txt",
			},
		},
	}

	reqJSON, _ := json.Marshal(req)

	// First call - should execute and cache
	_, err := handler.Handle(ctx, reqJSON, "test")
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}

	if mb.callCount != 1 {
		t.Errorf("first call: callCount = %d, want 1", mb.callCount)
	}

	// Second call - should use cache (callCount should NOT increase)
	_, err = handler.Handle(ctx, reqJSON, "test")
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}

	// THE KEY TEST: callCount should still be 1 (cache hit!)
	if mb.callCount != 1 {
		t.Errorf("second call: callCount = %d, want 1 (should be cached)", mb.callCount)
	}

	t.Logf("✓ Cache works! Tool executed once, second call used cache")
}

// Test: Non-cacheable tool is never cached
func TestHandler_NonCacheableToolNotCached(t *testing.T) {
	mb := newMockBackend()
	handler := protocol.NewHandler(mb, nil)

	// Configure cache
	cacheConfig := &cache.Config{
		Type:    cache.TypeShort,
		TTL:     60,
		MaxSize: 100,
		Enabled: true,
	}
	c, _ := cache.New(cacheConfig)
	keyGen := cache.NewKeyGenerator()
	handler.SetCache(c, keyGen, cacheConfig)

	ctx := context.Background()

	// Create request for non-cacheable tool
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "create_file",
			"arguments": map[string]interface{}{
				"path": "/test/newfile.txt",
			},
		},
	}

	reqJSON, _ := json.Marshal(req)

	// First call
	_, err := handler.Handle(ctx, reqJSON, "test")
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}

	// Second call - should execute again (not cached)
	_, err = handler.Handle(ctx, reqJSON, "test")
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}

	// Non-cacheable tool should execute twice
	if mb.callCount != 2 {
		t.Errorf("callCount = %d, want 2 (should not be cached)", mb.callCount)
	}

	t.Logf("✓ Non-cacheable tool bypassed cache correctly")
}

// Test: Different args produce different cache keys
func TestHandler_DifferentArgsNotCached(t *testing.T) {
	mb := newMockBackend()
	handler := protocol.NewHandler(mb, nil)

	// Configure cache
	cacheConfig := &cache.Config{
		Type:    cache.TypeShort,
		TTL:     60,
		MaxSize: 100,
		Enabled: true,
	}
	c, _ := cache.New(cacheConfig)
	keyGen := cache.NewKeyGenerator()
	handler.SetCache(c, keyGen, cacheConfig)

	ctx := context.Background()

	// First request
	req1 := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "read_file",
			"arguments": map[string]interface{}{
				"path": "/test/file1.txt",
			},
		},
	}

	// Second request with different args
	req2 := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "read_file",
			"arguments": map[string]interface{}{
				"path": "/test/file2.txt",
			},
		},
	}

	reqJSON1, _ := json.Marshal(req1)
	reqJSON2, _ := json.Marshal(req2)

	// Execute both
	handler.Handle(ctx, reqJSON1, "test")
	handler.Handle(ctx, reqJSON2, "test")

	// Should have executed twice (different cache keys)
	if mb.callCount != 2 {
		t.Errorf("callCount = %d, want 2 (different args = different keys)", mb.callCount)
	}

	t.Logf("✓ Different args correctly produced different cache keys")
}

// Test: Handler works without cache
func TestHandler_WorksWithoutCache(t *testing.T) {
	mb := newMockBackend()
	handler := protocol.NewHandler(mb, nil)
	// Don't set cache - handler should work fine

	ctx := context.Background()

	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "read_file",
			"arguments": map[string]interface{}{
				"path": "/test/file.txt",
			},
		},
	}

	reqJSON, _ := json.Marshal(req)

	// Should work without cache
	_, err := handler.Handle(ctx, reqJSON, "test")
	if err != nil {
		t.Fatalf("error without cache: %v", err)
	}

	if mb.callCount != 1 {
		t.Errorf("callCount = %d, want 1", mb.callCount)
	}

	t.Logf("✓ Handler works correctly without cache")
}
