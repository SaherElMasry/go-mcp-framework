package main

import (
	"context"
	"log"

	"github.com/SaherElMasry/go-mcp-framework/examples/grep-server/backend"
	"github.com/SaherElMasry/go-mcp-framework/framework"
)

func main() {
	log.Println("🔍 Grep-Like MCP Server")
	log.Println("=======================")
	log.Println()

	// Create grep backend
	grepBackend := backend.NewGrepBackend()

	// Create server
	server := framework.NewServer(
		framework.WithBackend(grepBackend),
		framework.WithTransport("http"),
		framework.WithHTTPAddress(":8080"),
		framework.WithStreaming(true),
		framework.WithMaxConcurrent(8),
		framework.WithObservability(true),
		framework.WithMetricsAddress(":9091"), // Metrics on port 9091
		framework.WithLogLevel("info"),
	)

	log.Println("Server starting...")
	log.Println("  - HTTP: http://localhost:8080")
	log.Println("  - Metrics: http://localhost:9091/metrics")
	log.Println()
	log.Println("Tools available:")
	log.Println("  - grep_html: Search HTML files for patterns (streaming)")
	log.Println("  - search_csv: Search CSV by field (streaming)")
	log.Println()
	log.Println("Endpoints:")
	log.Println("  - POST /rpc (JSON-RPC)")
	log.Println("  - POST /stream?tool=grep_html (SSE)")
	log.Println("  - POST /stream?tool=search_csv (SSE)")
	log.Println("  - GET /health")
	log.Println()

	// Run server
	if err := server.Run(context.Background()); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
