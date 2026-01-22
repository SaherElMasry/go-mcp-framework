package backend

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Final Mock emitter to satisfy the interface exactly
type mockHtmlEmitter struct{}

func (m *mockHtmlEmitter) Context() context.Context {
	return context.Background()
}

// These must return error to match your framework!
func (m *mockHtmlEmitter) EmitData(data interface{}) error {
	return nil
}

func (m *mockHtmlEmitter) EmitProgress(current, total int64, msg string) error {
	return nil
}
func BenchmarkHandleGrepHTML(b *testing.B) {
	// 1. Setup: Create a 1MB test HTML file
	tmpDir, _ := os.MkdirTemp("", "htmlbench")
	defer os.RemoveAll(tmpDir)

	htmlPath := filepath.Join(tmpDir, "bench.html")
	f, _ := os.Create(htmlPath)

	// Write 10,000 lines of HTML
	for i := 0; i < 10000; i++ {
		line := fmt.Sprintf("<div>User %d data line with some padding to make it realistic</div>\n", i)
		if i == 5000 {
			line = "<div>TARGET_MATCH_STRING</div>\n"
		}
		f.WriteString(line)
	}
	f.Close()

	// Initialize backend
	gb := &GrepBackend{
		// Assuming your GrepBackend has a Root field or similar to resolve paths
	}

	args := map[string]interface{}{
		"file_path": htmlPath,
		"pattern":   "target_match", // Lowercase to match our optimized logic
	}

	emitter := &mockHtmlEmitter{}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gb.handleGrepHTML(ctx, args, emitter)
	}
}
