package engine

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// ExecutorState represents the execution state
type ExecutorState string

const (
	StateInit     ExecutorState = "init"
	StateRunning  ExecutorState = "running"
	StateDone     ExecutorState = "done"
	StateError    ExecutorState = "error"
	StateCanceled ExecutorState = "canceled"
)

// ExecutorConfig configures the executor
type ExecutorConfig struct {
	BufferSize    int
	Timeout       time.Duration
	MaxEvents     int64
	MaxConcurrent int // v2 feature: semaphore-based concurrency control
}

// DefaultExecutorConfig returns default configuration
func DefaultExecutorConfig() ExecutorConfig {
	return ExecutorConfig{
		BufferSize:    100,
		Timeout:       5 * time.Minute,
		MaxEvents:     10000,
		MaxConcurrent: 16,
	}
}

// StreamingToolHandler is the function signature for streaming tools
type StreamingToolHandler func(ctx context.Context, args map[string]interface{}, emit Emitter) error

// Executor manages streaming tool execution
type Executor struct {
	config    ExecutorConfig
	logger    *slog.Logger
	state     atomic.Value // ExecutorState
	mu        sync.RWMutex
	sem       chan struct{} // v2 semaphore for concurrency control
	closeOnce sync.Once     // Ensure channels closed only once
}

// NewExecutor creates a new executor
func NewExecutor(config ExecutorConfig, logger *slog.Logger) *Executor {
	if logger == nil {
		logger = slog.Default()
	}

	// Create semaphore for concurrency control
	sem := make(chan struct{}, config.MaxConcurrent)

	e := &Executor{
		config: config,
		logger: logger,
		sem:    sem,
	}

	e.state.Store(StateInit)

	return e
}

// Execute runs a streaming tool and returns an event channel
func (e *Executor) Execute(
	ctx context.Context,
	toolName string,
	requestID string,
	args map[string]interface{},
	handler StreamingToolHandler,
) <-chan Event {
	// Create output channel
	events := make(chan Event, e.config.BufferSize)

	// Run in goroutine
	go func() {
		defer close(events) // Always close on exit

		// Acquire semaphore (v2 concurrency control)
		select {
		case e.sem <- struct{}{}:
			defer func() { <-e.sem }() // Release semaphore when done
		case <-ctx.Done():
			e.emitEventSafe(events, NewErrorEvent(ctx.Err(), "", false))
			return
		}

		e.run(ctx, toolName, requestID, args, handler, events)
	}()

	return events
}

// run executes the tool
func (e *Executor) run(
	ctx context.Context,
	toolName string,
	requestID string,
	args map[string]interface{},
	handler StreamingToolHandler,
	events chan<- Event,
) {
	// Set state to running
	e.state.Store(StateRunning)
	startTime := time.Now()

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, e.config.Timeout)
	defer cancel()

	// Emit start event
	e.emitEventSafe(events, NewStartEvent(toolName, requestID, args))

	// Create emitter
	emitter := newEmitter(execCtx, events)
	defer emitter.close()

	// Event counter
	var eventCount int64

	// Execute handler
	err := handler(execCtx, args, emitter)

	// Get event count
	atomic.AddInt64(&eventCount, emitter.sequence)

	duration := time.Since(startTime)

	// Emit result
	if err != nil {
		e.state.Store(StateError)
		e.emitEventSafe(events, NewErrorEvent(err, "", false))

		e.logger.Error("tool execution failed",
			"tool", toolName,
			"request_id", requestID,
			"error", err.Error(),
			"duration", duration,
		)
	} else {
		e.state.Store(StateDone)
		e.emitEventSafe(events, NewEndEvent(duration, eventCount, ""))

		e.logger.Info("tool execution completed",
			"tool", toolName,
			"request_id", requestID,
			"duration", duration,
			"events", eventCount,
		)
	}
}

// emitEventSafe safely emits an event without panicking on closed channel
func (e *Executor) emitEventSafe(events chan<- Event, event Event) {
	defer func() {
		if r := recover(); r != nil {
			// Channel was already closed, ignore
			e.logger.Debug("attempted to emit on closed channel", "event_type", event.Type)
		}
	}()

	select {
	case events <- event:
		// Successfully sent
	default:
		// Channel full or closed, skip
		e.logger.Debug("event channel full or closed", "event_type", event.Type)
	}
}

// State returns the current execution state
func (e *Executor) State() ExecutorState {
	return e.state.Load().(ExecutorState)
}

// String implements fmt.Stringer
func (s ExecutorState) String() string {
	return string(s)
}
