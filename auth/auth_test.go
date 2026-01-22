// framework/auth/auth_test.go
package auth

import (
	"context"
	"testing"
)

func TestManager(t *testing.T) {
	manager := NewManager()

	// Create a mock provider
	mockProvider := &mockAuthProvider{name: "test-provider"}

	// Test registration
	err := manager.Register("test", mockProvider)
	if err != nil {
		t.Fatalf("failed to register provider: %v", err)
	}

	// Test duplicate registration
	err = manager.Register("test", mockProvider)
	if err == nil {
		t.Fatal("expected error on duplicate registration")
	}

	// Test retrieval
	provider, err := manager.Get("test")
	if err != nil {
		t.Fatalf("failed to get provider: %v", err)
	}
	if provider.Name() != "test-provider" {
		t.Errorf("expected provider name %q, got %q", "test-provider", provider.Name())
	}

	// Test non-existent provider
	_, err = manager.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent provider")
	}

	// Test list
	names := manager.List()
	if len(names) != 1 || names[0] != "test" {
		t.Errorf("expected [test], got %v", names)
	}
}

func TestAPIKeyProvider(t *testing.T) {
	config := APIKeyConfig{
		APIKey: "test-key-12345",
		Header: "X-API-Key",
	}

	provider := NewAPIKeyProvider("test-api", config)

	// Register a resource
	provider.RegisterResource(ResourceConfig{
		ID:   "api1",
		Type: "api",
		Config: map[string]interface{}{
			"base_url": "https://api.example.com",
		},
	})

	// Test validation
	ctx := context.Background()
	err := provider.Validate(ctx)
	if err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	// Test getting resource
	resource, err := provider.GetResource(ctx, "api1")
	if err != nil {
		t.Fatalf("failed to get resource: %v", err)
	}

	apiResource, ok := resource.(*APIKeyResource)
	if !ok {
		t.Fatal("expected APIKeyResource")
	}

	if apiResource.BaseURL() != "https://api.example.com" {
		t.Errorf("expected base URL %q, got %q", "https://api.example.com", apiResource.BaseURL())
	}

	// Clean up
	err = resource.Close()
	if err != nil {
		t.Errorf("failed to close resource: %v", err)
	}
}

// Mock provider for testing
type mockAuthProvider struct {
	name string
}

func (p *mockAuthProvider) Name() string { return p.name }

func (p *mockAuthProvider) GetResource(ctx context.Context, resourceID string) (Resource, error) {
	return &mockResource{}, nil
}

func (p *mockAuthProvider) Validate(ctx context.Context) error { return nil }
func (p *mockAuthProvider) Refresh(ctx context.Context) error  { return nil }
func (p *mockAuthProvider) Close() error                       { return nil }

type mockResource struct{}

func (r *mockResource) Close() error { return nil }
func (r *mockResource) Type() string { return "mock" }
