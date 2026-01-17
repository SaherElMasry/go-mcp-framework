package protocol

import "fmt"

// Standard JSON-RPC 2.0 error codes
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// NewError creates a new protocol error
func NewError(code int, message string, data interface{}) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// NewParseError creates a parse error
func NewParseError(err error) *Error {
	return NewError(ParseError, "Parse error", err.Error())
}

// NewInvalidRequest creates an invalid request error
func NewInvalidRequest(message string) *Error {
	return NewError(InvalidRequest, "Invalid request", message)
}

// NewMethodNotFound creates a method not found error
func NewMethodNotFound(method string) *Error {
	return NewError(MethodNotFound, "Method not found", method)
}

// NewInvalidParams creates an invalid params error
func NewInvalidParams(message string) *Error {
	return NewError(InvalidParams, "Invalid params", message)
}

// NewInternalError creates an internal error
func NewInternalError(err error) *Error {
	return NewError(InternalError, "Internal error", err.Error())
}

// Error implements the error interface
func (e *Error) Error() string {
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}
