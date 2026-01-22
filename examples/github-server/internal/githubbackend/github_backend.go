// internal/backend/github_backend.go
package backend

import (
	"context"
	"fmt"

	"github.com/SaherElMasry/go-mcp-framework/examples/github-server/internal/config"
	"github.com/SaherElMasry/go-mcp-framework/examples/github-server/internal/github"
)

// GitHubBackend implements the MCP backend for GitHub
type GitHubBackend struct {
	client *github.Client
	config *config.Config
	tools  map[string]Tool
}

// NewGitHubBackend creates a new GitHub backend
func NewGitHubBackend(cfg *config.Config) *GitHubBackend {
	backend := &GitHubBackend{
		client: github.NewClient(cfg),
		config: cfg,
		tools:  make(map[string]Tool),
	}

	// Register all tools
	backend.registerTools()

	return backend
}

// Name returns the backend name
func (b *GitHubBackend) Name() string {
	return "github"
}

// Description returns the backend description
func (b *GitHubBackend) Description() string {
	return "GitHub API integration for managing repositories, issues, and more"
}

// Initialize initializes the backend
func (b *GitHubBackend) Initialize(ctx context.Context, config map[string]interface{}) error {
	// Verify we can connect to GitHub
	_, err := b.client.GetUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to authenticate with GitHub: %w", err)
	}
	return nil
}

// Close closes the backend
func (b *GitHubBackend) Close() error {
	// Nothing to clean up
	return nil
}

// GetTools returns all tools
func (b *GitHubBackend) GetTools() []Tool {
	tools := make([]Tool, 0, len(b.tools))
	for _, tool := range b.tools {
		tools = append(tools, tool)
	}
	return tools
}

// GetTool returns a specific tool
func (b *GitHubBackend) GetTool(name string) (Tool, bool) {
	tool, ok := b.tools[name]
	return tool, ok
}

// CallTool executes a tool
func (b *GitHubBackend) CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
	_, ok := b.tools[name]
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}

	// Route to the appropriate handler
	switch name {
	case "get_user":
		return b.handleGetUser(ctx, args)
	case "list_repos":
		return b.handleListRepos(ctx, args)
	case "get_repo":
		return b.handleGetRepo(ctx, args)
	case "create_repo":
		return b.handleCreateRepo(ctx, args)
	case "get_readme":
		return b.handleGetReadme(ctx, args)
	case "list_issues":
		return b.handleListIssues(ctx, args)
	case "get_issue":
		return b.handleGetIssue(ctx, args)
	case "create_issue":
		return b.handleCreateIssue(ctx, args)
	case "star_repo":
		return b.handleStarRepo(ctx, args)
	case "unstar_repo":
		return b.handleUnstarRepo(ctx, args)
	case "is_starred":
		return b.handleIsStarred(ctx, args)
	case "search_repos":
		return b.handleSearchRepos(ctx, args)
	case "get_rate_limit":
		return b.handleGetRateLimit(ctx, args)
	default:
		return nil, fmt.Errorf("handler not implemented for tool: %s", name)
	}
}

// IsStreamingTool checks if a tool supports streaming
func (b *GitHubBackend) IsStreamingTool(name string) bool {
	// These tools will support streaming
	streamingTools := map[string]bool{
		"list_repos":   true,
		"list_issues":  true,
		"search_repos": true,
	}
	return streamingTools[name]
}

// CallStreamingTool executes a streaming tool
func (b *GitHubBackend) CallStreamingTool(ctx context.Context, name string, args map[string]interface{}, emit StreamingEmitter) error {
	switch name {
	case "list_repos":
		return b.handleListReposStreaming(ctx, args, emit)
	case "list_issues":
		return b.handleListIssuesStreaming(ctx, args, emit)
	case "search_repos":
		return b.handleSearchReposStreaming(ctx, args, emit)
	default:
		return fmt.Errorf("streaming not supported for tool: %s", name)
	}
}

// GetCapabilities returns capabilities (not using capabilities yet)
func (b *GitHubBackend) GetCapabilities() []interface{} {
	return nil
}

// GetCapability returns a specific capability
func (b *GitHubBackend) GetCapability(name string) (interface{}, bool) {
	return nil, false
}

// HasCapabilities returns whether backend uses capabilities
func (b *GitHubBackend) HasCapabilities() bool {
	return false
}

// Tool represents an MCP tool
type Tool struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
}

// StreamingEmitter interface for streaming responses
type StreamingEmitter interface {
	EmitData(data interface{}) error
	EmitProgress(current, total int64, message string) error
	Context() context.Context
}

// registerTools registers all available tools
func (b *GitHubBackend) registerTools() {
	// User tools
	b.tools["get_user"] = Tool{
		Name:        "get_user",
		Description: "Get authenticated user information",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	}

	// Repository tools
	b.tools["list_repos"] = Tool{
		Name:        "list_repos",
		Description: "List repositories for the authenticated user",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"visibility": map[string]interface{}{
					"type":        "string",
					"description": "Filter by visibility: all, public, private",
					"enum":        []string{"all", "public", "private"},
				},
				"sort": map[string]interface{}{
					"type":        "string",
					"description": "Sort by: created, updated, pushed, full_name",
					"enum":        []string{"created", "updated", "pushed", "full_name"},
					"default":     "updated",
				},
				"per_page": map[string]interface{}{
					"type":        "number",
					"description": "Results per page (max 100)",
					"default":     30,
				},
			},
		},
	}

	b.tools["get_repo"] = Tool{
		Name:        "get_repo",
		Description: "Get details for a specific repository",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"owner": map[string]interface{}{
					"type":        "string",
					"description": "Repository owner",
				},
				"repo": map[string]interface{}{
					"type":        "string",
					"description": "Repository name",
				},
			},
			"required": []string{"owner", "repo"},
		},
	}

	b.tools["create_repo"] = Tool{
		Name:        "create_repo",
		Description: "Create a new repository",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Repository name",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Repository description",
				},
				"private": map[string]interface{}{
					"type":        "boolean",
					"description": "Create as private repository",
					"default":     false,
				},
			},
			"required": []string{"name"},
		},
	}

	b.tools["get_readme"] = Tool{
		Name:        "get_readme",
		Description: "Get README content for a repository",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"owner": map[string]interface{}{
					"type":        "string",
					"description": "Repository owner",
				},
				"repo": map[string]interface{}{
					"type":        "string",
					"description": "Repository name",
				},
			},
			"required": []string{"owner", "repo"},
		},
	}

	// Issue tools
	b.tools["list_issues"] = Tool{
		Name:        "list_issues",
		Description: "List issues in a repository",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"owner": map[string]interface{}{
					"type":        "string",
					"description": "Repository owner",
				},
				"repo": map[string]interface{}{
					"type":        "string",
					"description": "Repository name",
				},
				"state": map[string]interface{}{
					"type":        "string",
					"description": "Issue state: open, closed, all",
					"enum":        []string{"open", "closed", "all"},
					"default":     "open",
				},
				"per_page": map[string]interface{}{
					"type":    "number",
					"default": 30,
				},
			},
			"required": []string{"owner", "repo"},
		},
	}

	b.tools["get_issue"] = Tool{
		Name:        "get_issue",
		Description: "Get a specific issue",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"owner": map[string]interface{}{
					"type": "string",
				},
				"repo": map[string]interface{}{
					"type": "string",
				},
				"number": map[string]interface{}{
					"type": "number",
				},
			},
			"required": []string{"owner", "repo", "number"},
		},
	}

	b.tools["create_issue"] = Tool{
		Name:        "create_issue",
		Description: "Create a new issue",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"owner": map[string]interface{}{
					"type": "string",
				},
				"repo": map[string]interface{}{
					"type": "string",
				},
				"title": map[string]interface{}{
					"type": "string",
				},
				"body": map[string]interface{}{
					"type": "string",
				},
			},
			"required": []string{"owner", "repo", "title"},
		},
	}

	// Star tools
	b.tools["star_repo"] = Tool{
		Name:        "star_repo",
		Description: "Star a repository",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"owner": map[string]interface{}{"type": "string"},
				"repo":  map[string]interface{}{"type": "string"},
			},
			"required": []string{"owner", "repo"},
		},
	}

	b.tools["unstar_repo"] = Tool{
		Name:        "unstar_repo",
		Description: "Unstar a repository",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"owner": map[string]interface{}{"type": "string"},
				"repo":  map[string]interface{}{"type": "string"},
			},
			"required": []string{"owner", "repo"},
		},
	}

	b.tools["is_starred"] = Tool{
		Name:        "is_starred",
		Description: "Check if repository is starred",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"owner": map[string]interface{}{"type": "string"},
				"repo":  map[string]interface{}{"type": "string"},
			},
			"required": []string{"owner", "repo"},
		},
	}

	// Search tools
	b.tools["search_repos"] = Tool{
		Name:        "search_repos",
		Description: "Search for repositories",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query (e.g., 'language:go stars:>1000')",
				},
				"sort": map[string]interface{}{
					"type": "string",
					"enum": []string{"stars", "forks", "updated"},
				},
				"per_page": map[string]interface{}{
					"type":    "number",
					"default": 30,
				},
			},
			"required": []string{"query"},
		},
	}

	// Meta tools
	b.tools["get_rate_limit"] = Tool{
		Name:        "get_rate_limit",
		Description: "Get current API rate limit status",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	}
}
