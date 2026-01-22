// internal/backend/handlers_streaming.go
package backend

import (
	"context"
	"fmt"

	"github.com/SaherElMasry/go-mcp-framework/examples/github-server/internal/github"
)

// ============================================================================
// Streaming Handlers
//
// These handlers emit results progressively for better UX with large datasets
// ============================================================================

func (b *GitHubBackend) handleListReposStreaming(ctx context.Context, args map[string]interface{}, emit StreamingEmitter) error {
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

	// Emit progress
	if err := emit.EmitProgress(0, 0, "Fetching repositories..."); err != nil {
		return err
	}

	repos, err := b.client.ListRepos(ctx, opts)
	if err != nil {
		return err
	}

	// Emit each repository
	total := int64(len(repos))
	for i, repo := range repos {
		// Check if context was canceled
		select {
		case <-emit.Context().Done():
			return emit.Context().Err()
		default:
		}

		// Emit repository
		if err := emit.EmitData(formatRepository(&repo)); err != nil {
			return err
		}

		// Emit progress
		current := int64(i + 1)
		if err := emit.EmitProgress(current, total, fmt.Sprintf("Processed %d/%d repositories", current, total)); err != nil {
			return err
		}
	}

	// Final progress
	return emit.EmitProgress(total, total, "Complete")
}

func (b *GitHubBackend) handleListIssuesStreaming(ctx context.Context, args map[string]interface{}, emit StreamingEmitter) error {
	owner, ok := args["owner"].(string)
	if !ok {
		return fmt.Errorf("owner is required")
	}

	repo, ok := args["repo"].(string)
	if !ok {
		return fmt.Errorf("repo is required")
	}

	opts := &github.ListIssuesOptions{
		State:   "open",
		PerPage: 30,
	}

	if state, ok := args["state"].(string); ok {
		opts.State = state
	}
	if perPage, ok := args["per_page"].(float64); ok {
		opts.PerPage = int(perPage)
	}

	// Emit progress
	if err := emit.EmitProgress(0, 0, "Fetching issues..."); err != nil {
		return err
	}

	issues, err := b.client.ListIssues(ctx, owner, repo, opts)
	if err != nil {
		return err
	}

	// Emit each issue
	total := int64(len(issues))
	for i, issue := range issues {
		select {
		case <-emit.Context().Done():
			return emit.Context().Err()
		default:
		}

		if err := emit.EmitData(formatIssue(&issue)); err != nil {
			return err
		}

		current := int64(i + 1)
		if err := emit.EmitProgress(current, total, fmt.Sprintf("Processed %d/%d issues", current, total)); err != nil {
			return err
		}
	}

	return emit.EmitProgress(total, total, "Complete")
}

func (b *GitHubBackend) handleSearchReposStreaming(ctx context.Context, args map[string]interface{}, emit StreamingEmitter) error {
	query, ok := args["query"].(string)
	if !ok {
		return fmt.Errorf("query is required")
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

	// Emit progress
	if err := emit.EmitProgress(0, 0, "Searching repositories..."); err != nil {
		return err
	}

	result, err := b.client.SearchRepos(ctx, opts)
	if err != nil {
		return err
	}

	// Emit metadata
	if err := emit.EmitData(map[string]interface{}{
		"total_count": result.TotalCount,
		"type":        "metadata",
	}); err != nil {
		return err
	}

	// Emit each repository
	total := int64(len(result.Items))
	for i, repo := range result.Items {
		select {
		case <-emit.Context().Done():
			return emit.Context().Err()
		default:
		}

		repoData := formatRepository(&repo)
		repoData["type"] = "repository"

		if err := emit.EmitData(repoData); err != nil {
			return err
		}

		current := int64(i + 1)
		if err := emit.EmitProgress(current, total, fmt.Sprintf("Processed %d/%d results", current, total)); err != nil {
			return err
		}
	}

	return emit.EmitProgress(total, total, "Complete")
}
