package http

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SaherElMasry/go-mcp-framework/backend"
	"github.com/SaherElMasry/go-mcp-framework/engine"
)

// Mock backend for testing
type mockStreamingBackend struct {
	*backend.BaseBackend
}

func newMockStreamingBackend() *mockStreamingBackend {
	base := backend.NewBaseBackend("mock")

	mock := &mockStreamingBackend{
		BaseBackend: base,
	}

	// Register a simple streaming tool
	tool := backend.NewTool("test_stream").
		Description("Test streaming tool").
		Streaming(true).
		Build()

	handler := func(ctx context.Context, args map[string]interface{}, emit backend.StreamingEmitter) error {
		// Emit some test events
		emit.EmitProgress(1, 3, "step 1")
		emit.EmitData("chunk1")

		emit.EmitProgress(2, 3, "step 2")
		emit.EmitData("chunk2")

		emit.EmitProgress(3, 3, "step 3")
		emit.EmitData("chunk3")

		return nil
	}

	mock.RegisterStreamingTool(tool, handler)

	return mock
}

func TestSSEHandler_Basic(t *testing.T) {
	// Setup
	backend := newMockStreamingBackend()
	config := engine.DefaultExecutorConfig()
	executor := engine.NewExecutor(config, slog.Default())

	handler := NewSSEHandler(executor, backend, slog.Default(), 5*time.Second)

	// Create test request
	req := httptest.NewRequest("POST", "/stream?tool=test_stream", strings.NewReader("{}"))
	w := httptest.NewRecorder()

	// Execute
	handler.ServeHTTP(w, req)

	// Verify response
	resp := w.Result()
	defer resp.Body.Close()

	// Check headers
	if resp.Header.Get("Content-Type") != "text/event-stream" {
		t.Errorf("Expected Content-Type text/event-stream, got %s", resp.Header.Get("Content-Type"))
	}

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	bodyStr := string(body)

	// Verify SSE format
	if !strings.Contains(bodyStr, "event: start") {
		t.Error("Missing start event")
	}

	if !strings.Contains(bodyStr, "event: data") {
		t.Error("Missing data events")
	}

	if !strings.Contains(bodyStr, "event: progress") {
		t.Error("Missing progress events")
	}

	if !strings.Contains(bodyStr, "event: end") {
		t.Error("Missing end event")
	}

	// Verify data chunks
	if !strings.Contains(bodyStr, "chunk1") {
		t.Error("Missing chunk1 in data")
	}

	if !strings.Contains(bodyStr, "chunk2") {
		t.Error("Missing chunk2 in data")
	}

	if !strings.Contains(bodyStr, "chunk3") {
		t.Error("Missing chunk3 in data")
	}
}

func TestSSEHandler_MissingTool(t *testing.T) {
	backend := newMockStreamingBackend()
	config := engine.DefaultExecutorConfig()
	executor := engine.NewExecutor(config, slog.Default())

	handler := NewSSEHandler(executor, backend, slog.Default(), 5*time.Second)

	// Request without tool parameter
	req := httptest.NewRequest("POST", "/stream", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Should contain error event
	if !strings.Contains(bodyStr, "event: error") {
		t.Error("Expected error event for missing tool")
	}

	if !strings.Contains(bodyStr, "MISSING_TOOL") {
		t.Error("Expected MISSING_TOOL error code")
	}
}

func TestSSEHandler_ToolNotFound(t *testing.T) {
	backend := newMockStreamingBackend()
	config := engine.DefaultExecutorConfig()
	executor := engine.NewExecutor(config, slog.Default())

	handler := NewSSEHandler(executor, backend, slog.Default(), 5*time.Second)

	// Request with non-existent tool
	req := httptest.NewRequest("POST", "/stream?tool=nonexistent", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Should contain error event
	if !strings.Contains(bodyStr, "event: error") {
		t.Error("Expected error event for non-existent tool")
	}

	if !strings.Contains(bodyStr, "TOOL_NOT_FOUND") {
		t.Error("Expected TOOL_NOT_FOUND error code")
	}
}

func TestSSEHandler_MethodNotAllowed(t *testing.T) {
	backend := newMockStreamingBackend()
	config := engine.DefaultExecutorConfig()
	executor := engine.NewExecutor(config, slog.Default())

	handler := NewSSEHandler(executor, backend, slog.Default(), 5*time.Second)

	// GET request (should fail)
	req := httptest.NewRequest("GET", "/stream?tool=test_stream", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}
}

func TestSSEHandler_WithArguments(t *testing.T) {
	// Setup backend that uses arguments
	base := backend.NewBaseBackend("mock-args")

	tool := backend.NewTool("echo").
		Description("Echo arguments").
		StringParam("message", "Message to echo", true).
		Streaming(true).
		Build()

	handler := func(ctx context.Context, args map[string]interface{}, emit backend.StreamingEmitter) error {
		message := args["message"].(string)
		emit.EmitData(map[string]string{"echo": message})
		return nil
	}

	base.RegisterStreamingTool(tool, handler)

	config := engine.DefaultExecutorConfig()
	executor := engine.NewExecutor(config, slog.Default())

	sseHandler := NewSSEHandler(executor, base, slog.Default(), 5*time.Second)

	// Request with arguments
	req := httptest.NewRequest("POST", "/stream?tool=echo",
		strings.NewReader(`{"message":"Hello, World!"}`))
	w := httptest.NewRecorder()

	sseHandler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Verify echo message is in response
	if !strings.Contains(bodyStr, "Hello, World!") {
		t.Error("Expected echoed message in response")
	}
}
