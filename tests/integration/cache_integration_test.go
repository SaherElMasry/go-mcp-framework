package integration_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/SaherElMasry/go-mcp-framework/backend"
	"github.com/SaherElMasry/go-mcp-framework/framework"
	"github.com/SaherElMasry/go-mcp-framework/protocol"
)

// Test: Complete end-to-end cache integration
func TestCache_EndToEndIntegration(t *testing.T) {
	// Create a simple backend with cacheable tool
	backend := createTestBackend(t)

	// Create server with cache enabled
	server := framework.NewServer(
		framework.WithBackend(backend),
		framework.WithCache("short", 60), // 60 seconds
	)

	// Initialize server
	ctx := context.Background()
	if err := server.Initialize(ctx); err != nil {
		t.Fatalf("server initialization failed: %v", err)
	}
	defer server.GetCache().Close()

	// Get the protocol handler
	handler := protocol.NewHandler(backend, server.GetLogger())
	handler.SetCache(
		server.GetCache(),
		server.GetKeyGenerator(),
		server.GetCacheConfig(),
	)

	// Create test request
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "test_tool",
			"arguments": map[string]interface{}{
				"input": "test",
			},
		},
	}

	reqJSON, _ := json.Marshal(request)

	// First call - should execute tool
	start1 := time.Now()
	resp1, err := handler.Handle(ctx, reqJSON, "test")
	duration1 := time.Since(start1)

	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	// Second call - should use cache
	start2 := time.Now()
	resp2, err := handler.Handle(ctx, reqJSON, "test")
	duration2 := time.Since(start2)

	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	// Parse and compare results (not raw JSON which may have formatting differences)
	var result1, result2 protocol.Response
	json.Unmarshal(resp1, &result1)
	json.Unmarshal(resp2, &result2)

	// Compare the actual result data
	r1JSON, _ := json.Marshal(result1.Result)
	r2JSON, _ := json.Marshal(result2.Result)

	if string(r1JSON) != string(r2JSON) {
		t.Errorf("cached result differs from original\nFirst:  %s\nSecond: %s", r1JSON, r2JSON)
	} else {
		t.Logf("✓ Results match perfectly!")
	}

	// Verify cache made it faster
	if duration2 >= duration1 {
		t.Logf("WARNING: Cache didn't improve speed (first: %v, second: %v)", duration1, duration2)
	} else {
		speedup := float64(duration1) / float64(duration2)
		t.Logf("✓ Cache speedup: %.2fx faster (first: %v, second: %v)", speedup, duration1, duration2)
	}

	// Verify cache statistics
	stats := server.GetCache().Stats()

	if stats.Hits != 1 {
		t.Errorf("cache hits = %d, want 1", stats.Hits)
	}

	if stats.Sets != 1 {
		t.Errorf("cache sets = %d, want 1", stats.Sets)
	}

	if stats.HitRate != 0.5 { // 1 hit out of 2 gets
		t.Errorf("hit rate = %.2f, want 0.50", stats.HitRate)
	}

	t.Logf("✓ Cache stats: Hits=%d, Sets=%d, HitRate=%.2f%%",
		stats.Hits, stats.Sets, stats.HitRate*100)
}

// Helper: Create test backend
func createTestBackend(t *testing.T) backend.ServerBackend {
	base := backend.NewBaseBackend("test")

	// Create cacheable tool
	tool := backend.NewTool("test_tool").
		Description("Test tool for cache").
		StringParam("input", "Test input", true).
		WithCache(true, 5*time.Minute). // Cacheable!
		Build()

	// Register with handler that simulates some work
	base.RegisterTool(tool, func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		// Simulate work
		time.Sleep(10 * time.Millisecond)

		return map[string]interface{}{
			"result": "processed",
			"input":  args["input"],
			"time":   time.Now().Unix(),
		}, nil
	})

	return base
}
