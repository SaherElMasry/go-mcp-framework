// internal/backend/handlers_issues.go
package backend

import (
	"context"
	"fmt"

	"github.com/SaherElMasry/go-mcp-framework/examples/github-server/internal/github"
)

// ============================================================================
// Issue Handlers
// ============================================================================

func (b *GitHubBackend) handleListIssues(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	owner, ok := args["owner"].(string)
	if !ok {
		return nil, fmt.Errorf("owner is required")
	}

	repo, ok := args["repo"].(string)
	if !ok {
		return nil, fmt.Errorf("repo is required")
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

	issues, err := b.client.ListIssues(ctx, owner, repo, opts)
	if err != nil {
		return nil, err
	}

	// Convert to simplified format
	result := make([]map[string]interface{}, len(issues))
	for i, issue := range issues {
		result[i] = formatIssue(&issue)
	}

	return map[string]interface{}{
		"issues": result,
		"count":  len(result),
	}, nil
}

func (b *GitHubBackend) handleGetIssue(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	owner, ok := args["owner"].(string)
	if !ok {
		return nil, fmt.Errorf("owner is required")
	}

	repo, ok := args["repo"].(string)
	if !ok {
		return nil, fmt.Errorf("repo is required")
	}

	number, ok := args["number"].(float64)
	if !ok {
		return nil, fmt.Errorf("number is required")
	}

	issue, err := b.client.GetIssue(ctx, owner, repo, int(number))
	if err != nil {
		return nil, err
	}

	return formatIssue(issue), nil
}

func (b *GitHubBackend) handleCreateIssue(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	owner, ok := args["owner"].(string)
	if !ok {
		return nil, fmt.Errorf("owner is required")
	}

	repo, ok := args["repo"].(string)
	if !ok {
		return nil, fmt.Errorf("repo is required")
	}

	title, ok := args["title"].(string)
	if !ok {
		return nil, fmt.Errorf("title is required")
	}

	req := github.CreateIssueRequest{
		Title: title,
	}

	if body, ok := args["body"].(string); ok {
		req.Body = body
	}

	issue, err := b.client.CreateIssue(ctx, owner, repo, req)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"number":  issue.Number,
		"title":   issue.Title,
		"url":     issue.HTMLURL,
		"state":   issue.State,
		"created": true,
	}, nil
}

// formatIssue converts an Issue to a simplified map
func formatIssue(issue *github.Issue) map[string]interface{} {
	labels := make([]string, len(issue.Labels))
	for i, label := range issue.Labels {
		labels[i] = label.Name
	}

	return map[string]interface{}{
		"number":     issue.Number,
		"title":      issue.Title,
		"body":       issue.Body,
		"state":      issue.State,
		"user":       issue.User.Login,
		"labels":     labels,
		"comments":   issue.Comments,
		"created_at": issue.CreatedAt,
		"updated_at": issue.UpdatedAt,
		"url":        issue.HTMLURL,
	}
}
