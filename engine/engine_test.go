package engine

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestExecutor_Execute_Success(t *testing.T) {
	config := ExecutorConfig{
		BufferSize:    10,
		Timeout:       5 * time.Second,
		MaxEvents:     100,
		MaxConcurrent: 2,
	}

	executor := NewExecutor(config, nil)

	handler := func(ctx context.Context, args map[string]interface{}, emit Emitter) error {
		emit.EmitProgress(1, 3, "Step 1")
		emit.EmitData(map[string]string{"result": "test"})
		emit.EmitProgress(3, 3, "Complete")
		return nil
	}

	events := executor.Execute(
		context.Background(),
		"test_tool",
		"req-123",
		map[string]interface{}{"test": "value"},
		handler,
	)

	eventCount := 0
	for evt := range events {
		eventCount++
		t.Logf("Event: %s", evt.Type.String())
	}

	if eventCount < 3 {
		t.Errorf("Expected at least 3 events, got %d", eventCount)
	}

	if executor.State() != StateDone {
		t.Errorf("Expected state done, got %s", executor.State())
	}
}

func TestExecutor_Execute_Error(t *testing.T) {
	config := DefaultExecutorConfig()
	executor := NewExecutor(config, nil)

	handler := func(ctx context.Context, args map[string]interface{}, emit Emitter) error {
		return errors.New("test error")
	}

	events := executor.Execute(
		context.Background(),
		"test_tool",
		"req-456",
		nil,
		handler,
	)

	hasError := false
	for evt := range events {
		if evt.Type == EventError {
			hasError = true
		}
	}

	if !hasError {
		t.Error("Expected error event")
	}

	if executor.State() != StateError {
		t.Errorf("Expected state error, got %s", executor.State())
	}
}

func TestExecutor_Concurrency(t *testing.T) {
	config := ExecutorConfig{
		BufferSize:    10,
		Timeout:       5 * time.Second,
		MaxEvents:     100,
		MaxConcurrent: 2, // Only 2 concurrent
	}

	executor := NewExecutor(config, nil)

	handler := func(ctx context.Context, args map[string]interface{}, emit Emitter) error {
		time.Sleep(100 * time.Millisecond)
		emit.EmitData(map[string]string{"done": "true"})
		return nil
	}

	start := time.Now()

	// Launch 4 executions
	var channels []<-chan Event
	for i := 0; i < 4; i++ {
		ch := executor.Execute(
			context.Background(),
			"test_tool",
			"req",
			nil,
			handler,
		)
		channels = append(channels, ch)
	}

	// Wait for all
	for _, ch := range channels {
		for range ch {
			// Drain events
		}
	}

	duration := time.Since(start)

	// With max_concurrent=2, 4 tasks should take ~200ms (2 batches)
	// Allow some tolerance
	if duration < 150*time.Millisecond || duration > 300*time.Millisecond {
		t.Logf("Duration: %v (expected ~200ms)", duration)
	}
}
