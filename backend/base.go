package backend

import (
	"context"
	"fmt"
	"sync"
)

// BaseBackend provides automatic tool/resource registration and routing
type BaseBackend struct {
	name      string
	tools     map[string]*ToolRegistration
	resources map[string]*ResourceRegistration
	mu        sync.RWMutex
}

// ToolRegistration contains a tool definition and its handler
type ToolRegistration struct {
	Definition ToolDefinition
	Handler    ToolHandler
}

// ResourceRegistration contains a resource definition and its provider
type ResourceRegistration struct {
	Definition ResourceDefinition
	Provider   ResourceProvider
}

// ToolHandler is a function that executes a tool
type ToolHandler func(ctx context.Context, args map[string]interface{}) (interface{}, error)

// ResourceProvider is a function that provides a resource
type ResourceProvider func(ctx context.Context) (string, error)

// NewBaseBackend creates a new base backend
func NewBaseBackend(name string) *BaseBackend {
	return &BaseBackend{
		name:      name,
		tools:     make(map[string]*ToolRegistration),
		resources: make(map[string]*ResourceRegistration),
	}
}

// Name returns the backend name
func (b *BaseBackend) Name() string {
	return b.name
}

// RegisterTool registers a tool with its handler
func (b *BaseBackend) RegisterTool(def ToolDefinition, handler ToolHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tools[def.Name] = &ToolRegistration{
		Definition: def,
		Handler:    handler,
	}
}

// RegisterResource registers a resource with its provider
func (b *BaseBackend) RegisterResource(def ResourceDefinition, provider ResourceProvider) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.resources[def.URI] = &ResourceRegistration{
		Definition: def,
		Provider:   provider,
	}
}

// ListTools returns all registered tools
func (b *BaseBackend) ListTools(ctx context.Context) ([]ToolDefinition, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	tools := make([]ToolDefinition, 0, len(b.tools))
	for _, reg := range b.tools {
		tools = append(tools, reg.Definition)
	}
	return tools, nil
}

// ExecuteTool executes a tool by name
func (b *BaseBackend) ExecuteTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
	b.mu.RLock()
	reg, exists := b.tools[name]
	b.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}

	return reg.Handler(ctx, args)
}

// ListResources returns all registered resources
func (b *BaseBackend) ListResources(ctx context.Context) ([]ResourceDefinition, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	resources := make([]ResourceDefinition, 0, len(b.resources))
	for _, reg := range b.resources {
		resources = append(resources, reg.Definition)
	}
	return resources, nil
}

// ReadResource reads a resource by URI
func (b *BaseBackend) ReadResource(ctx context.Context, uri string) (string, error) {
	b.mu.RLock()
	reg, exists := b.resources[uri]
	b.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("unknown resource: %s", uri)
	}

	return reg.Provider(ctx)
}

// Initialize is called when the backend is created (override if needed)
func (b *BaseBackend) Initialize(ctx context.Context, config map[string]interface{}) error {
	return nil
}

// Close is called when the server shuts down (override if needed)
func (b *BaseBackend) Close() error {
	return nil
}
