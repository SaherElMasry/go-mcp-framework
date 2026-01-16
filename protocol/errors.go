package protocol

import "fmt"

const (
	ErrCodeParseError       = -32700
	ErrCodeInvalidRequest   = -32600
	ErrCodeMethodNotFound   = -32601
	ErrCodeInvalidParams    = -32602
	ErrCodeInternalError    = -32603
	ErrCodeServerError      = -32000
	ErrCodeToolNotFound     = -32001
	ErrCodeToolFailed       = -32002
	ErrCodeResourceNotFound = -32003
	ErrCodeSecurityError    = -32004
	ErrCodeValidationError  = -32005
)

// Error represents an MCP protocol error
type Error struct {
	Code    int
	Message string
	Data    interface{}
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Data != nil {
		return fmt.Sprintf("MCP error %d: %s (data: %v)", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("MCP error %d: %s", e.Code, e.Message)
}

// NewError creates a new Error
func NewError(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

// ParseError creates a parse error
func ParseError(message string) *Error {
	return NewError(ErrCodeParseError, message)
}

// InvalidRequest creates an invalid request error
func InvalidRequest(message string) *Error {
	return NewError(ErrCodeInvalidRequest, message)
}

// MethodNotFound creates a method not found error
func MethodNotFound(method string) *Error {
	return NewError(ErrCodeMethodNotFound, fmt.Sprintf("method not found: %s", method))
}

// InvalidParams creates an invalid params error
func InvalidParams(message string) *Error {
	return NewError(ErrCodeInvalidParams, message)
}

// InternalError creates an internal error
func InternalError(message string) *Error {
	return NewError(ErrCodeInternalError, message)
}

func categorizeError(err error) *Error {
	if err == nil {
		return nil
	}

	if protocolErr, ok := err.(*Error); ok {
		return protocolErr
	}

	errStr := err.Error()

	switch {
	case contains(errStr, "not found"):
		return NewError(ErrCodeToolNotFound, errStr)
	case contains(errStr, "security"), contains(errStr, "path traversal"):
		return NewError(ErrCodeSecurityError, errStr)
	case contains(errStr, "invalid"), contains(errStr, "validation"):
		return NewError(ErrCodeValidationError, errStr)
	default:
		return NewError(ErrCodeServerError, errStr)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
