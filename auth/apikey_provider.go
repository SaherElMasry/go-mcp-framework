// framework/auth/apikey.go
package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// APIKeyProvider authenticates using API keys
type APIKeyProvider struct {
	*BaseProvider
	apiKey string
	header string // Header name (default: "X-API-Key")
}

// APIKeyConfig holds API key provider configuration
type APIKeyConfig struct {
	APIKey string `yaml:"api_key" json:"api_key"`
	Header string `yaml:"header" json:"header"` // Optional custom header name
}

// NewAPIKeyProvider creates a new API key provider
func NewAPIKeyProvider(name string, config APIKeyConfig) *APIKeyProvider {
	header := config.Header
	if header == "" {
		header = "X-API-Key"
	}

	return &APIKeyProvider{
		BaseProvider: NewBaseProvider(name),
		apiKey:       config.APIKey,
		header:       header,
	}
}

// GetResource returns an authenticated HTTP client
func (p *APIKeyProvider) GetResource(ctx context.Context, resourceID string) (Resource, error) {
	// Get resource config
	config, err := p.GetResourceConfig(resourceID)
	if err != nil {
		return nil, NewAuthError(p.Name(), resourceID, "get_resource", err)
	}

	// Create HTTP client with API key
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &apiKeyTransport{
			base:   http.DefaultTransport,
			apiKey: p.apiKey,
			header: p.header,
		},
	}

	baseURL, ok := config.Config["base_url"].(string)
	if !ok {
		return nil, NewAuthError(p.Name(), resourceID, "get_resource",
			fmt.Errorf("missing base_url in resource config"))
	}

	return &APIKeyResource{
		client:     client,
		baseURL:    baseURL,
		resourceID: resourceID,
	}, nil
}

// Validate checks if the API key is set
func (p *APIKeyProvider) Validate(ctx context.Context) error {
	if p.apiKey == "" {
		return NewAuthError(p.Name(), "", "validate", ErrInvalidCredentials)
	}
	return nil
}

// apiKeyTransport adds API key to all requests
type apiKeyTransport struct {
	base   http.RoundTripper
	apiKey string
	header string
}

func (t *apiKeyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone request to avoid modifying original
	req = req.Clone(req.Context())
	req.Header.Set(t.header, t.apiKey)
	return t.base.RoundTrip(req)
}

// APIKeyResource wraps an HTTP client
type APIKeyResource struct {
	client     *http.Client
	baseURL    string
	resourceID string
}

func (r *APIKeyResource) Close() error {
	r.client.CloseIdleConnections()
	return nil
}

func (r *APIKeyResource) Type() string {
	return "api"
}

// Client returns the HTTP client for making requests
func (r *APIKeyResource) Client() *http.Client {
	return r.client
}

// BaseURL returns the base URL for the API
func (r *APIKeyResource) BaseURL() string {
	return r.baseURL
}
