package transport

import "context"

// Transport represents a network transport for MCP
type Transport interface {
	Run(ctx context.Context) error
}

// Handler processes MCP requests
type Handler interface {
	Handle(ctx context.Context, requestBytes []byte, transport string) ([]byte, error)
}
