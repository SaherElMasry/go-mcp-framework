// framework/auth/manager.go
package auth

import (
	"context"
	"fmt"
	"sync"
)

// Manager manages multiple auth providers
type Manager struct {
	providers map[string]AuthProvider
	mu        sync.RWMutex
}

// NewManager creates a new auth manager
func NewManager() *Manager {
	return &Manager{
		providers: make(map[string]AuthProvider),
	}
}

// Register registers an auth provider with a unique name
func (m *Manager) Register(name string, provider AuthProvider) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.providers[name]; exists {
		return fmt.Errorf("provider %q already registered", name)
	}

	m.providers[name] = provider
	return nil
}

// Get retrieves a registered provider by name
func (m *Manager) Get(name string) (AuthProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, exists := m.providers[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, name)
	}

	return provider, nil
}

// GetResource gets a resource from a specific provider
func (m *Manager) GetResource(ctx context.Context, providerName, resourceID string) (Resource, error) {
	provider, err := m.Get(providerName)
	if err != nil {
		return nil, err
	}

	return provider.GetResource(ctx, resourceID)
}

// ValidateAll validates all registered providers
func (m *Manager) ValidateAll(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, provider := range m.providers {
		if err := provider.Validate(ctx); err != nil {
			return NewAuthError(name, "", "validate", err)
		}
	}

	return nil
}

// Close closes all providers
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for name, provider := range m.providers {
		if err := provider.Close(); err != nil {
			errs = append(errs, fmt.Errorf("provider %q: %w", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing providers: %v", errs)
	}

	return nil
}

// List returns names of all registered providers
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.providers))
	for name := range m.providers {
		names = append(names, name)
	}

	return names
}
