// internal/github/issues.go
package github

import (
	"context"
	"fmt"
)

// ListIssuesOptions specifies options for listing issues
type ListIssuesOptions struct {
	State   string // open, closed, all
	Labels  string // comma-separated label names
	Sort    string // created, updated, comments
	PerPage int
	Page    int
}

// ListIssues lists issues for a repository
func (c *Client) ListIssues(ctx context.Context, owner, repo string, opts *ListIssuesOptions) ([]Issue, error) {
	if opts == nil {
		opts = &ListIssuesOptions{
			State:   "open",
			PerPage: 30,
		}
	}

	path := fmt.Sprintf("/repos/%s/%s/issues?state=%s&per_page=%d",
		owner, repo, opts.State, opts.PerPage)

	if opts.Labels != "" {
		path += "&labels=" + opts.Labels
	}
	if opts.Sort != "" {
		path += "&sort=" + opts.Sort
	}
	if opts.Page > 0 {
		path += fmt.Sprintf("&page=%d", opts.Page)
	}

	var issues []Issue
	if err := c.get(ctx, path, &issues); err != nil {
		return nil, err
	}

	return issues, nil
}

// GetIssue gets a specific issue
func (c *Client) GetIssue(ctx context.Context, owner, repo string, number int) (*Issue, error) {
	var issue Issue
	path := fmt.Sprintf("/repos/%s/%s/issues/%d", owner, repo, number)

	if err := c.get(ctx, path, &issue); err != nil {
		return nil, err
	}

	return &issue, nil
}

// CreateIssue creates a new issue
func (c *Client) CreateIssue(ctx context.Context, owner, repo string, req CreateIssueRequest) (*Issue, error) {
	var issue Issue
	path := fmt.Sprintf("/repos/%s/%s/issues", owner, repo)

	if err := c.post(ctx, path, req, &issue); err != nil {
		return nil, err
	}

	return &issue, nil
}

// UpdateIssueRequest specifies fields to update
type UpdateIssueRequest struct {
	Title     *string  `json:"title,omitempty"`
	Body      *string  `json:"body,omitempty"`
	State     *string  `json:"state,omitempty"` // open, closed
	Labels    []string `json:"labels,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
}

// UpdateIssue updates an existing issue
func (c *Client) UpdateIssue(ctx context.Context, owner, repo string, number int, req UpdateIssueRequest) (*Issue, error) {
	var issue Issue
	path := fmt.Sprintf("/repos/%s/%s/issues/%d", owner, repo, number)

	if err := c.post(ctx, path, req, &issue); err != nil {
		return nil, err
	}

	return &issue, nil
}
