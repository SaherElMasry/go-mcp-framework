package backend

import (
	"context"
	"fmt"
	"sync"
)

// ServerBackend interface for all backends
type ServerBackend interface {
	Name() string
	Initialize(ctx context.Context, config map[string]interface{}) error
	Close() error
	ListTools() []ToolDefinition
	GetTool(name string) (ToolDefinition, bool)
	CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error)

	// NEW: Streaming support
	CallStreamingTool(ctx context.Context, name string, args map[string]interface{}, emit StreamingEmitter) error
	IsStreamingTool(name string) bool
}

// StreamingEmitter is defined here to avoid circular imports
// The actual engine.Emitter will implement this
type StreamingEmitter interface {
	EmitData(data interface{}) error
	EmitProgress(current, total int64, message string) error
	Context() context.Context
}

// Registry for backend factories
var (
	registry   = make(map[string]BackendFactory)
	registryMu sync.RWMutex
)

// BackendFactory creates a backend instance
type BackendFactory func() ServerBackend

// Register registers a backend factory
func Register(name string, factory BackendFactory) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[name] = factory
}

// Get retrieves a backend factory
func Get(name string) (BackendFactory, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	factory, ok := registry[name]
	return factory, ok
}

// List returns all registered backend names
func List() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

// Create creates a backend instance
func Create(name string) (ServerBackend, error) {
	factory, ok := Get(name)
	if !ok {
		return nil, fmt.Errorf("backend not found: %s", name)
	}
	return factory(), nil
}
