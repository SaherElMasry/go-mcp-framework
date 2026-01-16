package protocol

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/SaherElMasry/go-mcp-framework/backend"
)

// Handler processes MCP protocol messages
type Handler interface {
	Handle(ctx context.Context, requestBytes []byte, transport string) ([]byte, error)
}

// BaseHandler implements the core MCP protocol handler
type BaseHandler struct {
	backend    backend.ServerBackend
	logger     *slog.Logger
	serverInfo ServerInfo
}

// NewHandler creates a new protocol handler
func NewHandler(b backend.ServerBackend, logger *slog.Logger) Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return &BaseHandler{
		backend: b,
		logger:  logger,
		serverInfo: ServerInfo{
			Name:    "MCP Server",
			Version: "1.0.0",
		},
	}
}

// Handle processes a JSON-RPC request
func (h *BaseHandler) Handle(ctx context.Context, requestBytes []byte, transport string) ([]byte, error) {
	var req Request
	if err := json.Unmarshal(requestBytes, &req); err != nil {
		h.logger.Error("failed to parse request", "error", err)
		return h.buildErrorResponse(nil, ParseError("invalid JSON"))
	}

	h.logger.Debug("received request", "method", req.Method, "id", req.ID)

	result, err := h.routeRequest(ctx, &req)

	if err != nil {
		return h.buildErrorResponseFromError(req.ID, err)
	}

	return h.buildSuccessResponse(req.ID, result)
}

func (h *BaseHandler) routeRequest(ctx context.Context, req *Request) (interface{}, error) {
	switch req.Method {
	case "initialize":
		return h.handleInitialize(ctx, req)
	case "tools/list":
		return h.handleListTools(ctx, req)
	case "tools/call":
		return h.handleCallTool(ctx, req)
	case "resources/list":
		return h.handleListResources(ctx, req)
	case "resources/read":
		return h.handleReadResource(ctx, req)
	case "ping":
		return h.handlePing(ctx, req)
	default:
		return nil, MethodNotFound(req.Method)
	}
}

func (h *BaseHandler) handleInitialize(ctx context.Context, req *Request) (interface{}, error) {
	var params InitializeParams
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, InvalidParams("invalid initialize parameters")
		}
	}

	h.logger.Info("client connected", "client", params.ClientInfo.Name)

	return InitializeResult{
		ProtocolVersion: "2024-11-05",
		ServerInfo:      h.serverInfo,
		Capabilities: map[string]interface{}{
			"tools":     map[string]interface{}{},
			"resources": map[string]interface{}{},
		},
	}, nil
}

func (h *BaseHandler) handleListTools(ctx context.Context, req *Request) (interface{}, error) {
	tools, err := h.backend.ListTools(ctx)
	if err != nil {
		h.logger.Error("failed to list tools", "error", err)
		return nil, InternalError(fmt.Sprintf("failed to list tools: %v", err))
	}

	return map[string]interface{}{"tools": tools}, nil
}

func (h *BaseHandler) handleCallTool(ctx context.Context, req *Request) (interface{}, error) {
	var params CallToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, InvalidParams("invalid tools/call parameters")
	}

	if params.Name == "" {
		return nil, InvalidParams("tool name is required")
	}

	h.logger.Info("executing tool", "tool", params.Name)

	result, err := h.backend.ExecuteTool(ctx, params.Name, params.Arguments)
	if err != nil {
		h.logger.Error("tool execution failed", "tool", params.Name, "error", err)
		return nil, categorizeError(err)
	}

	return CallToolResult{
		Content: []ContentBlock{
			{
				Type: "text",
				Text: formatToolResult(result),
			},
		},
	}, nil
}

func (h *BaseHandler) handleListResources(ctx context.Context, req *Request) (interface{}, error) {
	resources, err := h.backend.ListResources(ctx)
	if err != nil {
		h.logger.Error("failed to list resources", "error", err)
		return nil, InternalError(fmt.Sprintf("failed to list resources: %v", err))
	}

	return map[string]interface{}{"resources": resources}, nil
}

func (h *BaseHandler) handleReadResource(ctx context.Context, req *Request) (interface{}, error) {
	var params ReadResourceParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, InvalidParams("invalid resources/read parameters")
	}

	if params.URI == "" {
		return nil, InvalidParams("resource URI is required")
	}

	h.logger.Info("reading resource", "uri", params.URI)

	content, err := h.backend.ReadResource(ctx, params.URI)
	if err != nil {
		h.logger.Error("resource read failed", "uri", params.URI, "error", err)
		return nil, categorizeError(err)
	}

	return ReadResourceResult{
		Contents: []ResourceContent{
			{
				URI:      params.URI,
				MimeType: "text/plain",
				Text:     content,
			},
		},
	}, nil
}

func (h *BaseHandler) handlePing(ctx context.Context, req *Request) (interface{}, error) {
	return map[string]string{"status": "ok"}, nil
}

func (h *BaseHandler) buildSuccessResponse(id *int, result interface{}) ([]byte, error) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	return json.Marshal(resp)
}

func (h *BaseHandler) buildErrorResponse(id *int, err *Error) ([]byte, error) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &ErrorObject{
			Code:    err.Code,
			Message: err.Message,
			Data:    err.Data,
		},
	}
	return json.Marshal(resp)
}

func (h *BaseHandler) buildErrorResponseFromError(id *int, err error) ([]byte, error) {
	if protocolErr, ok := err.(*Error); ok {
		return h.buildErrorResponse(id, protocolErr)
	}

	protocolErr := categorizeError(err)
	return h.buildErrorResponse(id, protocolErr)
}

func formatToolResult(result interface{}) string {
	if result == nil {
		return ""
	}

	switch v := result.(type) {
	case string:
		return v
	case map[string]interface{}, []interface{}:
		bytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Sprintf("%v", result)
		}
		return string(bytes)
	default:
		return fmt.Sprintf("%v", result)
	}
}
