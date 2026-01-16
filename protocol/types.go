package protocol

import "encoding/json"

// Request represents a JSON-RPC 2.0 request
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response represents a JSON-RPC 2.0 response
type Response struct {
	JSONRPC string       `json:"jsonrpc"`
	ID      *int         `json:"id"`
	Result  interface{}  `json:"result,omitempty"`
	Error   *ErrorObject `json:"error,omitempty"`
}

// ErrorObject represents a JSON-RPC 2.0 error
type ErrorObject struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// InitializeParams represents parameters for initialize method
type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
}

// ClientInfo represents client information
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// CallToolParams represents parameters for tools/call method
type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// ReadResourceParams represents parameters for resources/read method
type ReadResourceParams struct {
	URI string `json:"uri"`
}

// InitializeResult represents the initialize method result
type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	ServerInfo      ServerInfo             `json:"serverInfo"`
	Capabilities    map[string]interface{} `json:"capabilities"`
}

// ServerInfo represents server information
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// CallToolResult represents the tools/call method result
type CallToolResult struct {
	Content []ContentBlock `json:"content"`
}

// ContentBlock represents a piece of content
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// ReadResourceResult represents the resources/read method result
type ReadResourceResult struct {
	Contents []ResourceContent `json:"contents"`
}

// ResourceContent represents resource content
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
}
