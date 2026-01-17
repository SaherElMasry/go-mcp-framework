package engine

import (
	"context"
	"fmt"
	"sync/atomic"
)

// Emitter is the interface provided to streaming tools for emitting events
type Emitter interface {
	// EmitData sends a data chunk
	EmitData(data interface{}) error

	// EmitProgress sends a progress update
	EmitProgress(current, total int64, message string) error

	// Context returns the execution context (for cancellation)
	Context() context.Context
}

// emitterImpl is the internal implementation of Emitter
type emitterImpl struct {
	ctx      context.Context
	events   chan<- Event
	sequence int64
	closed   atomic.Bool
}

// newEmitter creates a new emitter instance
func newEmitter(ctx context.Context, events chan<- Event) *emitterImpl {
	return &emitterImpl{
		ctx:      ctx,
		events:   events,
		sequence: 0,
	}
}

// EmitData sends a data event
func (e *emitterImpl) EmitData(data interface{}) error {
	if e.closed.Load() {
		return fmt.Errorf("emitter is closed")
	}

	// Safely send event
	return e.sendEventSafe(NewDataEvent(data, atomic.AddInt64(&e.sequence, 1)))
}

// EmitProgress sends a progress event
func (e *emitterImpl) EmitProgress(current, total int64, message string) error {
	if e.closed.Load() {
		return fmt.Errorf("emitter is closed")
	}

	// Safely send event
	return e.sendEventSafe(NewProgressEvent(current, total, message))
}

// Context returns the execution context
func (e *emitterImpl) Context() context.Context {
	return e.ctx
}

// close marks the emitter as closed
func (e *emitterImpl) close() {
	e.closed.Store(true)
}

// sendEventSafe safely sends an event without panicking
func (e *emitterImpl) sendEventSafe(event Event) error {
	defer func() {
		if r := recover(); r != nil {
			// Recovered from panic (channel closed)
		}
	}()

	select {
	case <-e.ctx.Done():
		return e.ctx.Err()
	case e.events <- event:
		return nil
	default:
		// Channel full or closed
		return fmt.Errorf("event channel unavailable")
	}
}
