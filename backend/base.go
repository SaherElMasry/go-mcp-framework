package backend

import (
	"context"
	"fmt"
)

// BaseBackend provides common functionality for backends
type BaseBackend struct {
	name              string
	tools             map[string]ToolDefinition
	handlers          map[string]ToolHandler
	streamingHandlers map[string]StreamingHandler // NEW
}

// StreamingHandler is the function signature for streaming tools
type StreamingHandler func(ctx context.Context, args map[string]interface{}, emit StreamingEmitter) error

// NewBaseBackend creates a new base backend
func NewBaseBackend(name string) *BaseBackend {
	return &BaseBackend{
		name:              name,
		tools:             make(map[string]ToolDefinition),
		handlers:          make(map[string]ToolHandler),
		streamingHandlers: make(map[string]StreamingHandler), // NEW
	}
}

// Name returns the backend name
func (b *BaseBackend) Name() string {
	return b.name
}

// Initialize initializes the backend
func (b *BaseBackend) Initialize(ctx context.Context, config map[string]interface{}) error {
	return nil
}

// Close closes the backend
func (b *BaseBackend) Close() error {
	return nil
}

// RegisterTool registers a regular tool
func (b *BaseBackend) RegisterTool(tool ToolDefinition, handler ToolHandler) {
	tool.Streaming = false
	b.tools[tool.Name] = tool
	b.handlers[tool.Name] = handler
}

// RegisterStreamingTool registers a streaming tool (NEW)
func (b *BaseBackend) RegisterStreamingTool(tool ToolDefinition, handler StreamingHandler) {
	tool.Streaming = true
	b.tools[tool.Name] = tool
	b.streamingHandlers[tool.Name] = handler
}

// ListTools returns all registered tools
func (b *BaseBackend) ListTools() []ToolDefinition {
	tools := make([]ToolDefinition, 0, len(b.tools))
	for _, tool := range b.tools {
		tools = append(tools, tool)
	}
	return tools
}

// GetTool retrieves a tool definition
func (b *BaseBackend) GetTool(name string) (ToolDefinition, bool) {
	tool, ok := b.tools[name]
	return tool, ok
}

// CallTool executes a regular tool
func (b *BaseBackend) CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
	handler, ok := b.handlers[name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	return handler(ctx, args)
}

// IsStreamingTool checks if a tool supports streaming (NEW)
func (b *BaseBackend) IsStreamingTool(name string) bool {
	_, ok := b.streamingHandlers[name]
	return ok
}

// CallStreamingTool executes a streaming tool (NEW)
func (b *BaseBackend) CallStreamingTool(ctx context.Context, name string, args map[string]interface{}, emit StreamingEmitter) error {
	handler, ok := b.streamingHandlers[name]
	if !ok {
		return fmt.Errorf("streaming tool not found: %s", name)
	}
	return handler(ctx, args, emit)
}
