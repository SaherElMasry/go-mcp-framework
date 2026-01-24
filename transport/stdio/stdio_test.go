package stdio

import (
	"bufio"
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

// mockHandler implements transport.Handler for testing
type mockHandler struct {
	HandleResult      []byte
	HandleError       error
	ReceivedBody      []byte
	ReceivedTransport string
}

func (m *mockHandler) Handle(ctx context.Context, requestBytes []byte, transport string) ([]byte, error) {
	m.ReceivedBody = requestBytes
	m.ReceivedTransport = transport
	return m.HandleResult, m.HandleError
}

func TestNewStdioTransport(t *testing.T) {
	handler := &mockHandler{}
	tr := NewStdioTransport(handler, nil)

	if tr == nil {
		t.Fatal("Expected NewStdioTransport to return a value")
	}

	if tr.handler != handler {
		t.Errorf("Expected handler to be %v, got %v", handler, tr.handler)
	}

	if tr.logger == nil {
		t.Error("Expected logger to be initialized with default")
	}
}

func TestStdioTransport_Run_ContextCancel(t *testing.T) {
	handler := &mockHandler{}
	tr := &StdioTransport{
		handler: handler,
		logger:  slog.Default(),
		reader:  bufio.NewReader(strings.NewReader("")),
		writer:  bufio.NewWriter(&bytes.Buffer{}),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := tr.Run(ctx)
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func TestStdioTransport_Run_Success(t *testing.T) {
	reqBody := []byte(`{"jsonrpc":"2.0","method":"test","id":1}`)
	respBody := []byte(`{"jsonrpc":"2.0","result":"ok","id":1}`)

	handler := &mockHandler{
		HandleResult: respBody,
	}

	input := bytes.NewBuffer(append(reqBody, '\n'))
	output := &bytes.Buffer{}

	tr := &StdioTransport{
		handler: handler,
		logger:  slog.Default(),
		reader:  bufio.NewReader(input),
		writer:  bufio.NewWriter(output),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Run in goroutine because Run is an infinite loop (until EOF or cancel)
	errCh := make(chan error, 1)
	go func() {
		errCh <- tr.Run(ctx)
	}()

	// Wait for processing or timeout
	select {
	case err := <-errCh:
		if err != nil && err != context.Canceled {
			t.Errorf("Run failed with error: %v", err)
		}
	case <-time.After(50 * time.Millisecond):
		cancel()
	}

	if !bytes.Equal(handler.ReceivedBody, append(reqBody, '\n')) {
		t.Errorf("Handler received %q, expected %q", handler.ReceivedBody, append(reqBody, '\n'))
	}

	if handler.ReceivedTransport != "stdio" {
		t.Errorf("Expected transport label 'stdio', got %s", handler.ReceivedTransport)
	}

	// Output should contain response + newline
	expectedOutput := append(respBody, '\n')
	if !bytes.Contains(output.Bytes(), expectedOutput) {
		t.Errorf("Output %s does not contain expected %s", output.String(), expectedOutput)
	}
}

func TestStdioTransport_Run_EOF(t *testing.T) {
	tr := &StdioTransport{
		handler: &mockHandler{},
		logger:  slog.Default(),
		reader:  bufio.NewReader(strings.NewReader("")), // Immediate EOF
		writer:  bufio.NewWriter(&bytes.Buffer{}),
	}

	err := tr.Run(context.Background())
	if err != nil {
		t.Errorf("Expected nil error on EOF, got %v", err)
	}
}

func TestStdioTransport_Run_EmptyLines(t *testing.T) {
	handler := &mockHandler{}
	input := strings.NewReader("\n\n\n")
	output := &bytes.Buffer{}

	tr := &StdioTransport{
		handler: handler,
		logger:  slog.Default(),
		reader:  bufio.NewReader(input),
		writer:  bufio.NewWriter(output),
	}

	err := tr.Run(context.Background())
	if err != nil {
		t.Errorf("Expected nil error on EOF, got %v", err)
	}

	if len(handler.ReceivedBody) > 0 {
		t.Error("Handler should not have been called for empty lines")
	}
}
