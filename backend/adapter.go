package backend

import (
	"fmt"
)

// FunctionBackend adapts a single streaming function to the Backend interface
// This enables v2-style simple registration
type FunctionBackend struct {
	*BaseBackend
	toolName string
	handler  StreamingHandler
}

// NewFunctionBackend creates a backend from a single streaming function
func NewFunctionBackend(name string, handler StreamingHandler) *FunctionBackend {
	base := NewBaseBackend(fmt.Sprintf("function-%s", name))

	fb := &FunctionBackend{
		BaseBackend: base,
		toolName:    name,
		handler:     handler,
	}

	// Auto-register the function as a streaming tool
	tool := NewTool(name).
		Description(fmt.Sprintf("Function tool: %s", name)).
		Streaming(true).
		Build()

	fb.RegisterStreamingTool(tool, handler)

	return fb
}

// GetToolName returns the tool name
func (fb *FunctionBackend) GetToolName() string {
	return fb.toolName
}
