// internal/backend/handlers_repos.go
package backend

import (
	"context"
	"fmt"

	"github.com/SaherElMasry/go-mcp-framework/examples/github-server/internal/github"
)

// ============================================================================
// User Handlers
// ============================================================================

func (b *GitHubBackend) handleGetUser(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	user, err := b.client.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	// Return simplified user info
	return map[string]interface{}{
		"login":        user.Login,
		"name":         user.Name,
		"email":        user.Email,
		"bio":          user.Bio,
		"avatar_url":   user.AvatarURL,
		"public_repos": user.PublicRepos,
		"followers":    user.Followers,
		"following":    user.Following,
		"created_at":   user.CreatedAt,
	}, nil
}

// ============================================================================
// Repository Handlers
// ============================================================================

func (b *GitHubBackend) handleListRepos(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	opts := &github.ListReposOptions{
		Sort:    "updated",
		PerPage: 30,
	}

	// Parse arguments
	if visibility, ok := args["visibility"].(string); ok {
		opts.Visibility = visibility
	}
	if sort, ok := args["sort"].(string); ok {
		opts.Sort = sort
	}
	if perPage, ok := args["per_page"].(float64); ok {
		opts.PerPage = int(perPage)
	}

	repos, err := b.client.ListRepos(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Convert to simplified format
	result := make([]map[string]interface{}, len(repos))
	for i, repo := range repos {
		result[i] = formatRepository(&repo)
	}

	return map[string]interface{}{
		"repositories": result,
		"count":        len(result),
	}, nil
}

func (b *GitHubBackend) handleGetRepo(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	owner, ok := args["owner"].(string)
	if !ok {
		return nil, fmt.Errorf("owner is required")
	}

	repo, ok := args["repo"].(string)
	if !ok {
		return nil, fmt.Errorf("repo is required")
	}

	repository, err := b.client.GetRepo(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	return formatRepository(repository), nil
}

func (b *GitHubBackend) handleCreateRepo(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	name, ok := args["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name is required")
	}

	req := github.CreateRepoRequest{
		Name:     name,
		AutoInit: true, // Initialize with README
	}

	if description, ok := args["description"].(string); ok {
		req.Description = description
	}
	if private, ok := args["private"].(bool); ok {
		req.Private = private
	}

	repo, err := b.client.CreateRepo(ctx, req)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":      repo.Name,
		"full_name": repo.FullName,
		"url":       repo.HTMLURL,
		"clone_url": repo.CloneURL,
		"private":   repo.Private,
		"created":   true,
	}, nil
}

func (b *GitHubBackend) handleGetReadme(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	owner, ok := args["owner"].(string)
	if !ok {
		return nil, fmt.Errorf("owner is required")
	}

	repo, ok := args["repo"].(string)
	if !ok {
		return nil, fmt.Errorf("repo is required")
	}

	// Get README text (already decoded)
	text, err := b.client.GetReadmeText(ctx, owner, repo)
	if err != nil {
		if github.IsNotFound(err) {
			return map[string]interface{}{
				"found":   false,
				"message": "No README found in this repository",
			}, nil
		}
		return nil, err
	}

	return map[string]interface{}{
		"found":   true,
		"content": text,
		"size":    len(text),
	}, nil
}

// formatRepository converts a Repository to a simplified map
func formatRepository(repo *github.Repository) map[string]interface{} {
	return map[string]interface{}{
		"name":           repo.Name,
		"full_name":      repo.FullName,
		"description":    repo.Description,
		"private":        repo.Private,
		"url":            repo.HTMLURL,
		"clone_url":      repo.CloneURL,
		"language":       repo.Language,
		"stars":          repo.Stars,
		"forks":          repo.Forks,
		"open_issues":    repo.OpenIssues,
		"default_branch": repo.DefaultBranch,
		"created_at":     repo.CreatedAt,
		"updated_at":     repo.UpdatedAt,
		"owner":          repo.Owner.Login,
	}
}
