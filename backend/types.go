package backend

import (
	"context"
)

// ToolDefinition describes a tool's interface
type ToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  []Parameter `json:"inputSchema"`
	Streaming   bool        `json:"streaming,omitempty"` // NEW: Mark streaming tools
}

// Parameter describes a tool parameter
type Parameter struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Enum        []string    `json:"enum,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Minimum     *int        `json:"minimum,omitempty"`
	Maximum     *int        `json:"maximum,omitempty"`
}

// ToolHandler is the function signature for regular tools
type ToolHandler func(ctx context.Context, args map[string]interface{}) (interface{}, error)

// No need to import engine here - will be in backend.go
