package backend

import (
	"context"
	"fmt"
	"sync"

	"github.com/SaherElMasry/go-mcp-framework/auth"
)

// BaseBackend provides common functionality for backends
type BaseBackend struct {
	name              string
	tools             map[string]ToolDefinition
	handlers          map[string]ToolHandler
	streamingHandlers map[string]StreamingHandler // NEW
	resources         []Resource                  //****
	prompts           []Prompt                    ///**** and backend.go files to delete them

	// === NEW: Auth Support ===
	authProvider auth.AuthProvider
	authManager  *auth.Manager
	mu           sync.RWMutex // Protects auth fields
}

// StreamingHandler is the function signature for streaming tools
type StreamingHandler func(ctx context.Context, args map[string]interface{}, emit StreamingEmitter) error

// NewBaseBackend creates a new base backend
func NewBaseBackend(name string) *BaseBackend {
	return &BaseBackend{
		name:              name,
		tools:             make(map[string]ToolDefinition),
		handlers:          make(map[string]ToolHandler),
		streamingHandlers: make(map[string]StreamingHandler), // NEW
		resources:         []Resource{},                      //v3
		prompts:           []Prompt{},                        //v3
	}
}

// Name returns the backend name
func (b *BaseBackend) Name() string {
	return b.name
}

// Initialize initializes the backend
func (b *BaseBackend) Initialize(ctx context.Context, config map[string]interface{}) error {
	return nil
}

// Close closes the backend
func (b *BaseBackend) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// === NEW: Close auth provider ===
	if b.authProvider != nil {
		if err := b.authProvider.Close(); err != nil {
			return fmt.Errorf("failed to close auth provider: %w", err)
		}
	}

	// === NEW: Close auth manager ===
	if b.authManager != nil {
		if err := b.authManager.Close(); err != nil {
			return fmt.Errorf("failed to close auth manager: %w", err)
		}
	}
	return nil
}

// RegisterTool registers a regular tool
func (b *BaseBackend) RegisterTool(tool ToolDefinition, handler ToolHandler) {
	tool.Streaming = false
	b.tools[tool.Name] = tool
	b.handlers[tool.Name] = handler
}

// RegisterStreamingTool registers a streaming tool (NEW)
func (b *BaseBackend) RegisterStreamingTool(tool ToolDefinition, handler StreamingHandler) {
	tool.Streaming = true
	b.tools[tool.Name] = tool
	b.streamingHandlers[tool.Name] = handler
}

// ListTools returns all registered tools
func (b *BaseBackend) ListTools() []ToolDefinition {
	tools := make([]ToolDefinition, 0, len(b.tools))
	for _, tool := range b.tools {
		tools = append(tools, tool)
	}
	return tools
}

// GetTool retrieves a tool definition
func (b *BaseBackend) GetTool(name string) (ToolDefinition, bool) {
	tool, ok := b.tools[name]
	return tool, ok
}

// CallTool executes a regular tool
func (b *BaseBackend) CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
	handler, ok := b.handlers[name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	// === NEW: Optional auth validation before tool execution ===
	// Tools can choose to use auth or not
	b.mu.RLock()
	authProvider := b.authProvider
	b.mu.RUnlock()

	if authProvider != nil {
		// Validate auth (non-blocking - tool can still decide)
		if err := authProvider.Validate(ctx); err != nil {
			// Just log, don't fail - let tool decide if auth is required
			// Logger would go here if available
		}
	}
	return handler(ctx, args)
}

// IsStreamingTool checks if a tool supports streaming (NEW)
func (b *BaseBackend) IsStreamingTool(name string) bool {
	_, ok := b.streamingHandlers[name]
	return ok
}

// CallStreamingTool executes a streaming tool (NEW)
func (b *BaseBackend) CallStreamingTool(ctx context.Context, name string, args map[string]interface{}, emit StreamingEmitter) error {
	handler, ok := b.streamingHandlers[name]
	if !ok {
		return fmt.Errorf("streaming tool not found: %s", name)
	}
	return handler(ctx, args, emit)
}

// ============================================================
// Resource Management
// ============================================================

// ListResources returns all registered resources
func (b *BaseBackend) ListResources() []Resource {
	return b.resources
}

// RegisterResource registers a resource
func (b *BaseBackend) RegisterResource(resource Resource) {
	b.resources = append(b.resources, resource)
}

// ============================================================
// Prompt Management
// ============================================================

// ListPrompts returns all registered prompts
func (b *BaseBackend) ListPrompts() []Prompt {
	return b.prompts
}

// RegisterPrompt registers a prompt
func (b *BaseBackend) RegisterPrompt(prompt Prompt) {
	b.prompts = append(b.prompts, prompt)
}

// ============================================================
// AUTH SUPPORT - NEW! 🔐
// ============================================================

// SetAuthProvider sets the primary auth provider for this backend
func (b *BaseBackend) SetAuthProvider(provider auth.AuthProvider) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.authProvider = provider
}

// GetAuthProvider returns the primary auth provider
func (b *BaseBackend) GetAuthProvider() auth.AuthProvider {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.authProvider
}

// SetAuthManager sets the auth manager (for multi-provider scenarios)
func (b *BaseBackend) SetAuthManager(manager *auth.Manager) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.authManager = manager
}

// GetAuthManager returns the auth manager
func (b *BaseBackend) GetAuthManager() *auth.Manager {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.authManager
}

// ============================================================
// Helper Methods for Tools to Use Auth
// ============================================================

// GetAuthenticatedClient is a helper for tools to get an authenticated HTTP client
// Example usage in a tool:
//
//	client, err := b.GetAuthenticatedClient(ctx, "github-api")
func (b *BaseBackend) GetAuthenticatedClient(ctx context.Context, resourceID string) (interface{}, error) {
	b.mu.RLock()
	provider := b.authProvider
	b.mu.RUnlock()

	if provider == nil {
		return nil, fmt.Errorf("no auth provider configured")
	}

	resource, err := provider.GetResource(ctx, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated resource: %w", err)
	}

	return resource, nil
}

// GetAuthenticatedResource is a generic helper to get any authenticated resource
// Returns the resource which can be type-asserted to specific types:
//   - *auth.APIKeyResource for API key auth
//   - *auth.OAuth2Resource for OAuth2
//   - *auth.DatabaseResource for database connections
func (b *BaseBackend) GetAuthenticatedResource(ctx context.Context, providerName, resourceID string) (auth.Resource, error) {
	b.mu.RLock()
	manager := b.authManager
	b.mu.RUnlock()

	if manager == nil {
		// Fall back to default provider
		b.mu.RLock()
		provider := b.authProvider
		b.mu.RUnlock()

		if provider == nil {
			return nil, fmt.Errorf("no auth configured")
		}

		return provider.GetResource(ctx, resourceID)
	}

	// Use manager for multi-provider scenarios
	return manager.GetResource(ctx, providerName, resourceID)
}

// ValidateAuth validates the current auth configuration
// Tools can call this at the start of execution
func (b *BaseBackend) ValidateAuth(ctx context.Context) error {
	b.mu.RLock()
	provider := b.authProvider
	b.mu.RUnlock()

	if provider == nil {
		return fmt.Errorf("no auth provider configured")
	}

	return provider.Validate(ctx)
}

// RefreshAuth refreshes authentication credentials (e.g., OAuth2 tokens)
func (b *BaseBackend) RefreshAuth(ctx context.Context) error {
	b.mu.RLock()
	provider := b.authProvider
	b.mu.RUnlock()

	if provider == nil {
		return fmt.Errorf("no auth provider configured")
	}

	return provider.Refresh(ctx)
}
