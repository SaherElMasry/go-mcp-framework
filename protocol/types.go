package protocol

// Request represents a JSON-RPC 2.0 request
type Request struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// Response represents a JSON-RPC 2.0 response
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Error represents a JSON-RPC 2.0 error
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Notification represents a JSON-RPC 2.0 notification (no ID)
type Notification struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// ToolInfo represents tool information for MCP
type ToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolCallResult represents the result of a tool call
type ToolCallResult struct {
	Content []ContentItem          `json:"content"`
	IsError bool                   `json:"isError,omitempty"`
	Meta    map[string]interface{} `json:"_meta,omitempty"`
}

// ContentItem represents a piece of content in the result
type ContentItem struct {
	Type        string                 `json:"type"`
	Text        string                 `json:"text,omitempty"`
	Data        interface{}            `json:"data,omitempty"`
	MimeType    string                 `json:"mimeType,omitempty"`
	Annotations map[string]interface{} `json:"annotations,omitempty"`
}
