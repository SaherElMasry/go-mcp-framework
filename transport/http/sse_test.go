package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SaherElMasry/go-mcp-framework/auth"
	"github.com/SaherElMasry/go-mcp-framework/backend"
	"github.com/SaherElMasry/go-mcp-framework/engine"
)

// mockBackend implements backend.ServerBackend for testing
type mockBackend struct {
	Tools           map[string]backend.ToolDefinition
	StreamingResult error
}

func (m *mockBackend) Name() string { return "mock" }
func (m *mockBackend) Initialize(ctx context.Context, config map[string]interface{}) error {
	return nil
}
func (m *mockBackend) Close() error                        { return nil }
func (m *mockBackend) ListTools() []backend.ToolDefinition { return nil }
func (m *mockBackend) GetTool(name string) (backend.ToolDefinition, bool) {
	t, ok := m.Tools[name]
	return t, ok
}
func (m *mockBackend) CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
	return nil, nil
}
func (m *mockBackend) CallStreamingTool(ctx context.Context, name string, args map[string]interface{}, emit backend.StreamingEmitter) error {
	if m.StreamingResult != nil {
		return m.StreamingResult
	}
	emit.EmitData(map[string]string{"foo": "bar"})
	return nil
}
func (m *mockBackend) IsStreamingTool(name string) bool {
	t, ok := m.Tools[name]
	return ok && t.Streaming
}
func (m *mockBackend) ListResources() []backend.Resource          { return nil }
func (m *mockBackend) ListPrompts() []backend.Prompt              { return nil }
func (m *mockBackend) SetAuthProvider(provider auth.AuthProvider) {}
func (m *mockBackend) GetAuthProvider() auth.AuthProvider         { return nil }
func (m *mockBackend) SetAuthManager(manager *auth.Manager)       {}
func (m *mockBackend) GetAuthManager() *auth.Manager              { return nil }

func TestNewSSEHandler(t *testing.T) {
	h := NewSSEHandler(nil, nil, nil, 0)
	if h == nil {
		t.Fatal("Expected NewSSEHandler to return a value")
	}
	if h.logger == nil {
		t.Error("Expected default logger")
	}
	if h.timeout == 0 {
		t.Error("Expected default timeout")
	}
}

func TestSSEHandler_ServeHTTP_Validation(t *testing.T) {
	executor := engine.NewExecutor(engine.DefaultExecutorConfig(), nil)
	mb := &mockBackend{
		Tools: map[string]backend.ToolDefinition{
			"tool1": {Name: "tool1", Streaming: true},
			"tool2": {Name: "tool2", Streaming: false},
		},
	}
	h := NewSSEHandler(executor, mb, nil, 0)

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/stream", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected 405, got %d", w.Code)
		}
	})

	t.Run("MissingTool", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/stream", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if !strings.Contains(w.Body.String(), "missing_tool") {
			t.Errorf("Expected missing_tool error, got %s", w.Body.String())
		}
	})

	t.Run("ToolNotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/stream?tool=unknown", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if !strings.Contains(w.Body.String(), "tool_not_found") {
			t.Errorf("Expected tool_not_found error, got %s", w.Body.String())
		}
	})

	t.Run("NotStreamingTool", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/stream?tool=tool2", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if !strings.Contains(w.Body.String(), "not_streaming") {
			t.Errorf("Expected not_streaming error, got %s", w.Body.String())
		}
	})
}

type flushingRecorder struct {
	*httptest.ResponseRecorder
	flushed bool
}

func (r *flushingRecorder) Flush() {
	r.flushed = true
}

func TestSSEHandler_ServeHTTP_Success(t *testing.T) {
	executor := engine.NewExecutor(engine.DefaultExecutorConfig(), nil)
	mb := &mockBackend{
		Tools: map[string]backend.ToolDefinition{
			"tool1": {Name: "tool1", Streaming: true},
		},
	}
	h := NewSSEHandler(executor, mb, nil, 1*time.Second)

	args := map[string]interface{}{"input": "test"}
	argsBytes, _ := json.Marshal(args)
	req := httptest.NewRequest(http.MethodPost, "/stream?tool=tool1", bytes.NewReader(argsBytes))

	w := &flushingRecorder{ResponseRecorder: httptest.NewRecorder()}

	h.ServeHTTP(w, req)

	if w.Header().Get("Content-Type") != "text/event-stream" {
		t.Errorf("Expected text/event-stream, got %s", w.Header().Get("Content-Type"))
	}

	// Wait a bit for goroutines to finish streaming if necessary,
	// though ServeHTTP calls streamEvents which blocks until channel closed.

	body := w.Body.String()
	if !strings.Contains(body, "event: start") {
		t.Error("Missing start event")
	}
	if !strings.Contains(body, "event: data") {
		t.Error("Missing data event")
	}
	if !strings.Contains(body, "event: end") {
		t.Error("Missing end event")
	}
}

func TestSSEHandler_sendErrorEvent(t *testing.T) {
	h := NewSSEHandler(nil, nil, nil, 0)
	w := &flushingRecorder{ResponseRecorder: httptest.NewRecorder()}

	h.sendErrorEvent(w, w, "test_code", "test message")

	body := w.Body.String()
	if !strings.Contains(body, "event: error") {
		t.Error("Missing error event")
	}
	if !strings.Contains(body, "test message") {
		t.Error("Missing error message")
	}
	if !w.flushed {
		t.Error("Expected flusher to be called")
	}
}
