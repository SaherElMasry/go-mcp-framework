// framework/auth/auth.go
package auth

import (
	"context"
	"time"
)

// Resource represents an authenticated connection to a resource
// This could be a database connection, HTTP client, file handle, etc.
type Resource interface {
	// Close releases the resource
	Close() error

	// Type returns the resource type (e.g., "database", "api", "file")
	Type() string
}

// AuthProvider manages authentication to external resources
// Each provider handles a specific auth method (API key, OAuth, DB, etc.)
type AuthProvider interface {
	// GetResource returns an authenticated resource
	// resourceID identifies which resource to access (e.g., "main-db", "payment-api")
	GetResource(ctx context.Context, resourceID string) (Resource, error)

	// Validate checks if credentials are valid without fetching a resource
	Validate(ctx context.Context) error

	// Refresh refreshes credentials (useful for tokens that expire)
	Refresh(ctx context.Context) error

	// Close cleans up any resources held by the provider
	Close() error

	// Name returns the provider name for logging/debugging
	Name() string
}

// ProviderConfig holds configuration for an auth provider
type ProviderConfig struct {
	// Provider type (e.g., "database", "api-key", "oauth2")
	Type string `yaml:"type" json:"type"`

	// Configuration specific to the provider type
	Config map[string]interface{} `yaml:"config" json:"config"`

	// Resources managed by this provider
	Resources map[string]ResourceConfig `yaml:"resources" json:"resources"`
}

// ResourceConfig describes a single resource
type ResourceConfig struct {
	// Resource identifier
	ID string `yaml:"id" json:"id"`

	// Resource type (database, api, file, etc.)
	Type string `yaml:"type" json:"type"`

	// Resource-specific configuration
	Config map[string]interface{} `yaml:"config" json:"config"`
}

// Credentials holds authentication credentials
// IMPORTANT: This should never be logged or serialized to disk in plaintext
type Credentials struct {
	// Username or API key
	Username string

	// Password or secret
	Password string

	// Token for token-based auth
	Token string

	// Token expiry time
	ExpiresAt time.Time

	// Additional metadata
	Metadata map[string]string
}

// CredentialStore securely stores and retrieves credentials
type CredentialStore interface {
	// Store saves credentials securely
	Store(ctx context.Context, key string, creds *Credentials) error

	// Retrieve gets credentials
	Retrieve(ctx context.Context, key string) (*Credentials, error)

	// Delete removes credentials
	Delete(ctx context.Context, key string) error

	// Close cleans up the store
	Close() error
}
