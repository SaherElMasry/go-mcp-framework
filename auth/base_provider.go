// framework/auth/base_provider.go
package auth

import (
	"context"
	"fmt"
	"sync"
)

// BaseProvider provides common functionality for auth providers
type BaseProvider struct {
	name      string
	resources map[string]ResourceConfig
	mu        sync.RWMutex
}

// NewBaseProvider creates a new base provider
func NewBaseProvider(name string) *BaseProvider {
	return &BaseProvider{
		name:      name,
		resources: make(map[string]ResourceConfig),
	}
}

// Name returns the provider name
func (p *BaseProvider) Name() string {
	return p.name
}

// RegisterResource registers a resource configuration
func (p *BaseProvider) RegisterResource(config ResourceConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.resources[config.ID] = config
}

// GetResourceConfig retrieves a resource configuration
func (p *BaseProvider) GetResourceConfig(resourceID string) (ResourceConfig, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	config, exists := p.resources[resourceID]
	if !exists {
		return ResourceConfig{}, fmt.Errorf("%w: %s", ErrResourceNotFound, resourceID)
	}

	return config, nil
}

// ListResources returns all registered resource IDs
func (p *BaseProvider) ListResources() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	ids := make([]string, 0, len(p.resources))
	for id := range p.resources {
		ids = append(ids, id)
	}

	return ids
}

// Close is a no-op base implementation
func (p *BaseProvider) Close() error {
	return nil
}

// Refresh is a no-op base implementation
func (p *BaseProvider) Refresh(ctx context.Context) error {
	return nil
}

// Validate is a no-op base implementation
func (p *BaseProvider) Validate(ctx context.Context) error {
	return nil
}
