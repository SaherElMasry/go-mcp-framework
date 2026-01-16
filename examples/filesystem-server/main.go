package main

import (
	"context"
	"log"

	fsbackend "filesystem-server/backend"

	"github.com/SaherElMasry/go-mcp-framework/backend"
	"github.com/SaherElMasry/go-mcp-framework/framework"
)

func init() {
	// Register the filesystem backend
	backend.Register("filesystem", func() backend.ServerBackend {
		return fsbackend.NewFilesystemBackend()
	})
}

func main() {
	server := framework.NewServer(
		framework.WithBackendType("filesystem"),
		framework.WithTransport("http"),
		framework.WithHTTPAddress(":8080"),
		framework.WithLogLevel("info"),
		framework.WithObservability(true),
		framework.WithMetricsAddress(":9091"), // ✅ Now this works!
	)

	log.Println("📁 Filesystem MCP Server")
	log.Println("Server: http://localhost:8080")
	log.Println("Metrics: http://localhost:9091/metrics")
	log.Println()
	log.Println("File operations: file_create, file_read, file_write, file_update, file_delete, file_copy, file_search, file_show_content")
	log.Println("Folder operations: folder_create, folder_delete, folder_rename, folder_copy, folder_move, folder_list")
	log.Println()
	log.Println("Security: Sandboxed to workspace directory with path traversal prevention")

	if err := server.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
