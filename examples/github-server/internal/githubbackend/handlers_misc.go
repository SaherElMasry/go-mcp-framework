// internal/backend/handlers_misc.go
package backend

import (
	"context"
	"fmt"

	"github.com/SaherElMasry/go-mcp-framework/examples/github-server/internal/github"
)

// ============================================================================
// Star Handlers
// ============================================================================

func (b *GitHubBackend) handleStarRepo(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	owner, ok := args["owner"].(string)
	if !ok {
		return nil, fmt.Errorf("owner is required")
	}

	repo, ok := args["repo"].(string)
	if !ok {
		return nil, fmt.Errorf("repo is required")
	}

	err := b.client.StarRepo(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"starred": true,
		"repo":    fmt.Sprintf("%s/%s", owner, repo),
	}, nil
}

func (b *GitHubBackend) handleUnstarRepo(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	owner, ok := args["owner"].(string)
	if !ok {
		return nil, fmt.Errorf("owner is required")
	}

	repo, ok := args["repo"].(string)
	if !ok {
		return nil, fmt.Errorf("repo is required")
	}

	err := b.client.UnstarRepo(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"unstarred": true,
		"repo":      fmt.Sprintf("%s/%s", owner, repo),
	}, nil
}

func (b *GitHubBackend) handleIsStarred(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	owner, ok := args["owner"].(string)
	if !ok {
		return nil, fmt.Errorf("owner is required")
	}

	repo, ok := args["repo"].(string)
	if !ok {
		return nil, fmt.Errorf("repo is required")
	}

	starred, err := b.client.IsStarred(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"repo":    fmt.Sprintf("%s/%s", owner, repo),
		"starred": starred,
	}, nil
}

// ============================================================================
// Search Handlers
// ============================================================================

func (b *GitHubBackend) handleSearchRepos(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	query, ok := args["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query is required")
	}

	opts := github.SearchReposOptions{
		Query:   query,
		PerPage: 30,
	}

	if sort, ok := args["sort"].(string); ok {
		opts.Sort = sort
	}
	if perPage, ok := args["per_page"].(float64); ok {
		opts.PerPage = int(perPage)
	}

	result, err := b.client.SearchRepos(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Convert to simplified format
	repos := make([]map[string]interface{}, len(result.Items))
	for i, repo := range result.Items {
		repos[i] = formatRepository(&repo)
	}

	return map[string]interface{}{
		"total_count":  result.TotalCount,
		"repositories": repos,
		"count":        len(repos),
	}, nil
}

// ============================================================================
// Meta Handlers
// ============================================================================

func (b *GitHubBackend) handleGetRateLimit(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	rateLimit, err := b.client.GetRateLimit(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"limit":     rateLimit.Limit,
		"remaining": rateLimit.Remaining,
		"reset":     rateLimit.Reset,
		"used":      rateLimit.Limit - rateLimit.Remaining,
	}, nil
}
