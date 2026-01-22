// examples/weather-server/cmd/server/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/SaherElMasry/go-mcp-framework/backend"
	"github.com/SaherElMasry/go-mcp-framework/framework"

	// Import your weather backend
	"github.com/SaherElMasry/go-mcp-framework/examples/weather-server/internal/weather"
)

const banner = `
╔══════════════════════════════════════════════════════════════╗
║                                                              ║
║     🌤️  Weather MCP Server v2.0.0 (Framework v0.3.0)       ║
║                                                              ║
║     Production-ready MCP server with:                       ║
║     ✅ Framework v0.3.0 auth integration                    ║
║     ✅ Streaming support (SSE)                              ║
║     ✅ Prometheus metrics                                   ║
║     ✅ Caching & rate limiting                              ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝
`

func main() {
	// Print banner
	fmt.Println(banner)

	// Get API key from environment
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		log.Fatal("❌ WEATHER_API_KEY environment variable not set!\n" +
			"   Get your free API key from: https://www.weatherapi.com/signup.aspx\n" +
			"   Then run: export WEATHER_API_KEY=your-key-here")
	}

	// Register the weather backend
	backend.Register("weather", func() backend.ServerBackend {
		return weather.NewWeatherBackend()
	})

	// Create backend config with API key
	backendConfig := map[string]interface{}{
		"api_key":  apiKey,
		"base_url": "https://api.weatherapi.com/v1",
	}

	// Create server
	server := framework.NewServer(
		framework.WithBackendType("weather"),
		framework.WithHTTPAddress(":8080"),

		// Streaming
		framework.WithStreaming(true),
		framework.WithStreamingBufferSize(100),
		framework.WithMaxConcurrent(8),

		// Observability
		framework.WithObservability(true),
		framework.WithMetricsAddress(":9091"),

		// Logging
		framework.WithLogLevel("info"),
	)

	// Get the backend and initialize it with our config
	// This happens automatically during server.Run(), but we need to
	// make sure the API key is passed in the config
	if srv := server.GetBackend(); srv != nil {
		// The backend will be initialized by the framework with backendConfig
		// We need to ensure the config is passed properly
		// For now, we'll use WithConfig to pass backend config
	}

	// Better approach: Create a complete config
	cfg := framework.DefaultConfig()
	cfg.Backend.Type = "weather"
	cfg.Backend.Config = backendConfig
	cfg.Transport.Type = "http"
	cfg.Transport.HTTP.Address = ":8080"
	cfg.Streaming.Enabled = true
	cfg.Streaming.BufferSize = 100
	cfg.Streaming.MaxConcurrent = 8
	cfg.Observability.Enabled = true
	cfg.Observability.MetricsAddress = ":9091"
	cfg.Logging.Level = "info"

	// Create server with complete config
	server = framework.NewServer(
		framework.WithConfig(cfg),
	)

	// Print startup information
	printStartupInfo(apiKey)

	// Run the server
	ctx := context.Background()
	fmt.Println("🚀 Starting Weather MCP Server...")

	if err := server.Run(ctx); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}

	fmt.Println("✅ Server stopped gracefully")
}

func printStartupInfo(apiKey string) {
	fmt.Println("📡 Server Configuration:")
	fmt.Println("   ├─ HTTP Endpoint:    http://localhost:8080")
	fmt.Println("   ├─ Metrics:          http://localhost:9091/metrics")
	fmt.Println("   ├─ Health Check:     http://localhost:9091/health")
	fmt.Println("   ├─ Streaming:        Enabled (SSE)")
	fmt.Println("   ├─ Auth:             API Key (framework v0.3.0)")
	fmt.Printf("   ├─ API Key:          %s***\n", apiKey[:min(8, len(apiKey))])
	fmt.Println("   └─ Max Concurrent:   8 requests")
	fmt.Println()
	fmt.Println("🔧 Available Tools:")
	fmt.Println("   1. get_current_weather    - Current weather conditions")
	fmt.Println("   2. get_forecast           - Multi-day forecast (1-10 days)")
	fmt.Println("   3. search_locations       - Search locations (streaming)")
	fmt.Println("   4. get_astronomy          - Sunrise/sunset/moon data")
	fmt.Println("   5. bulk_weather_check     - Bulk weather check (streaming)")
	fmt.Println()
	fmt.Println("💡 Example Requests:")
	fmt.Println()
	fmt.Println("   # List available tools")
	fmt.Println("   curl -X POST http://localhost:8080 \\")
	fmt.Println("     -H 'Content-Type: application/json' \\")
	fmt.Println("     -d '{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/list\"}'")
	fmt.Println()
	fmt.Println("   # Get current weather")
	fmt.Println("   curl -X POST http://localhost:8080 \\")
	fmt.Println("     -H 'Content-Type: application/json' \\")
	fmt.Println("     -d '{")
	fmt.Println("       \"jsonrpc\":\"2.0\",")
	fmt.Println("       \"id\":1,")
	fmt.Println("       \"method\":\"tools/call\",")
	fmt.Println("       \"params\":{")
	fmt.Println("         \"name\":\"get_current_weather\",")
	fmt.Println("         \"arguments\":{\"location\":\"London\"}")
	fmt.Println("       }")
	fmt.Println("     }'")
	fmt.Println()
	fmt.Println("🌊 Stream Example (SSE):")
	fmt.Println("   curl -N http://localhost:8080/stream \\")
	fmt.Println("     -H 'Content-Type: application/json' \\")
	fmt.Println("     -d '{")
	fmt.Println("       \"tool\":\"search_locations\",")
	fmt.Println("       \"arguments\":{\"query\":\"London\"}")
	fmt.Println("     }'")
	fmt.Println()
	fmt.Println("📊 View Metrics:")
	fmt.Println("   curl http://localhost:9091/metrics")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop the server")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}
