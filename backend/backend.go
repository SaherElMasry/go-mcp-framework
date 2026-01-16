package backend

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// ServerBackend is the main interface that all backends must implement
type ServerBackend interface {
	Name() string
	ListTools(ctx context.Context) ([]ToolDefinition, error)
	ExecuteTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error)
	ListResources(ctx context.Context) ([]ResourceDefinition, error)
	ReadResource(ctx context.Context, uri string) (string, error)
	Initialize(ctx context.Context, config map[string]interface{}) error
	Close() error
}

// BackendFactory is a function that creates a new backend instance
type BackendFactory func() ServerBackend

var (
	registry   = make(map[string]BackendFactory)
	registryMu sync.RWMutex
)

// Register registers a backend factory with the given name
func Register(name string, factory BackendFactory) {
	if name == "" {
		panic("backend: Register called with empty name")
	}
	if factory == nil {
		panic(fmt.Sprintf("backend: Register called with nil factory for %q", name))
	}

	registryMu.Lock()
	defer registryMu.Unlock()

	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("backend: Register called twice for %q", name))
	}

	registry[name] = factory
}

// New creates a new backend instance by name
func New(name string) (ServerBackend, error) {
	registryMu.RLock()
	factory, exists := registry[name]
	registryMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unknown backend: %q (available: %v)", name, List())
	}

	return factory(), nil
}

// List returns all registered backend names
func List() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
