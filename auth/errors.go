// framework/auth/errors.go
package auth

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidCredentials indicates credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrResourceNotFound indicates the requested resource doesn't exist
	ErrResourceNotFound = errors.New("resource not found")

	// ErrProviderNotFound indicates the auth provider doesn't exist
	ErrProviderNotFound = errors.New("auth provider not found")

	// ErrTokenExpired indicates the auth token has expired
	ErrTokenExpired = errors.New("token expired")

	// ErrRefreshFailed indicates credential refresh failed
	ErrRefreshFailed = errors.New("credential refresh failed")

	// ErrValidationFailed indicates credential validation failed
	ErrValidationFailed = errors.New("validation failed")
)

// AuthError wraps errors with additional context
type AuthError struct {
	Provider string
	Resource string
	Op       string // Operation that failed
	Err      error  // Underlying error
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("auth: provider=%s resource=%s op=%s: %v",
		e.Provider, e.Resource, e.Op, e.Err)
}

func (e *AuthError) Unwrap() error {
	return e.Err
}

// NewAuthError creates a new AuthError
func NewAuthError(provider, resource, op string, err error) error {
	return &AuthError{
		Provider: provider,
		Resource: resource,
		Op:       op,
		Err:      err,
	}
}
