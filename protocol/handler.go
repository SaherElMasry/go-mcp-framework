package protocol

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/SaherElMasry/go-mcp-framework/backend"
)

// Handler handles JSON-RPC requests
type Handler struct {
	backend backend.ServerBackend
	logger  *slog.Logger
}

// NewHandler creates a new protocol handler
func NewHandler(backend backend.ServerBackend, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return &Handler{
		backend: backend,
		logger:  logger,
	}
}

// Handle processes a JSON-RPC request
func (h *Handler) Handle(ctx context.Context, data []byte, transportType string) ([]byte, error) {
	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return h.errorResponse(nil, NewParseError(err))
	}

	h.logger.Debug("handling request",
		"method", req.Method,
		"id", req.ID,
		"transport", transportType)

	var resp Response
	resp.JSONRPC = "2.0"
	resp.ID = req.ID

	switch req.Method {
	case "tools/list":
		result, err := h.handleToolsList(ctx)
		if err != nil {
			resp.Error = err
		} else {
			resp.Result = result
		}

	case "tools/call":
		result, err := h.handleToolsCall(ctx, req.Params)
		if err != nil {
			resp.Error = err
		} else {
			resp.Result = result
		}

	default:
		resp.Error = NewMethodNotFound(req.Method)
	}

	return json.Marshal(resp)
}

// handleToolsList handles the tools/list method
func (h *Handler) handleToolsList(ctx context.Context) (interface{}, *Error) {
	tools := h.backend.ListTools()

	toolInfos := make([]ToolInfo, len(tools))
	for i, tool := range tools {
		toolInfos[i] = ToolInfo{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: h.convertParametersToSchema(tool.Parameters),
		}
	}

	return map[string]interface{}{
		"tools": toolInfos,
	}, nil
}

// handleToolsCall handles the tools/call method
func (h *Handler) handleToolsCall(ctx context.Context, params map[string]interface{}) (interface{}, *Error) {
	toolName, ok := params["name"].(string)
	if !ok {
		return nil, NewInvalidParams("missing or invalid 'name' parameter")
	}

	args, ok := params["arguments"].(map[string]interface{})
	if !ok {
		args = make(map[string]interface{})
	}

	// Execute tool
	result, err := h.backend.CallTool(ctx, toolName, args)
	if err != nil {
		return nil, NewInternalError(err)
	}

	// Convert result to MCP format
	return h.convertToToolCallResult(result), nil
}

// convertParametersToSchema converts tool parameters to JSON Schema
func (h *Handler) convertParametersToSchema(params []backend.Parameter) map[string]interface{} {
	properties := make(map[string]interface{})
	required := make([]string, 0)

	for _, param := range params {
		prop := map[string]interface{}{
			"type":        param.Type,
			"description": param.Description,
		}

		if len(param.Enum) > 0 {
			prop["enum"] = param.Enum
		}

		if param.Default != nil {
			prop["default"] = param.Default
		}

		if param.Minimum != nil {
			prop["minimum"] = *param.Minimum
		}

		if param.Maximum != nil {
			prop["maximum"] = *param.Maximum
		}

		properties[param.Name] = prop

		if param.Required {
			required = append(required, param.Name)
		}
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

// convertToToolCallResult converts a result to MCP ToolCallResult format
func (h *Handler) convertToToolCallResult(result interface{}) ToolCallResult {
	// Convert result to JSON string
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return ToolCallResult{
			Content: []ContentItem{
				{
					Type: "text",
					Text: fmt.Sprintf("%v", result),
				},
			},
		}
	}

	return ToolCallResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: string(resultJSON),
			},
		},
	}
}

// errorResponse creates an error response
func (h *Handler) errorResponse(id interface{}, err *Error) ([]byte, error) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Error:   err,
	}
	return json.Marshal(resp)
}
