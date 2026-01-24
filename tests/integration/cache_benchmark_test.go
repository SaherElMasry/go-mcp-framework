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

// Benchmark: Compare performance with and without cache
func BenchmarkCache_Performance(b *testing.B) {
	b.Run("WithoutCache", func(b *testing.B) {
		benchmarkWithCache(b, false)
	})

	b.Run("WithCache", func(b *testing.B) {
		benchmarkWithCache(b, true)
	})
}

func benchmarkWithCache(b *testing.B, enableCache bool) {
	// Create backend
	backend := createBenchmarkBackend(b)

	// Create server
	var server *framework.Server
	if enableCache {
		server = framework.NewServer(
			framework.WithBackend(backend),
			framework.WithCache("short", 60),
		)
	} else {
		server = framework.NewServer(
			framework.WithBackend(backend),
		)
	}

	// Initialize
	ctx := context.Background()
	if err := server.Initialize(ctx); err != nil {
		b.Fatalf("initialization failed: %v", err)
	}

	// Setup handler
	handler := protocol.NewHandler(backend, server.GetLogger())
	if enableCache {
		handler.SetCache(
			server.GetCache(),
			server.GetKeyGenerator(),
			server.GetCacheConfig(),
		)
	}

	// Create request
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "bench_tool",
			"arguments": map[string]interface{}{
				"data": "test",
			},
		},
	}

	reqJSON, _ := json.Marshal(request)

	// Run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.Handle(ctx, reqJSON, "test")
	}

	// Report cache stats if enabled
	if enableCache {
		stats := server.GetCache().Stats()
		b.ReportMetric(float64(stats.Hits), "cache_hits")
		b.ReportMetric(stats.HitRate*100, "hit_rate_%")
	}
}

func createBenchmarkBackend(b *testing.B) backend.ServerBackend {
	base := backend.NewBaseBackend("benchmark")

	tool := backend.NewTool("bench_tool").
		Description("Benchmark tool").
		StringParam("data", "Test data", true).
		WithCache(true, 1*time.Minute).
		Build()

	base.RegisterTool(tool, func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		// Simulate real work (5ms)
		time.Sleep(5 * time.Millisecond)
		return map[string]interface{}{"result": "ok"}, nil
	})

	return base
}
