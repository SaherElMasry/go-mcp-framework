// internal/github/repos.go
package github

import (
	"context"
	"fmt"
)

// GetUser gets the authenticated user
func (c *Client) GetUser(ctx context.Context) (*User, error) {
	var user User
	if err := c.get(ctx, "/user", &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// ListReposOptions specifies options for listing repositories
type ListReposOptions struct {
	Visibility string // all, public, private
	Sort       string // created, updated, pushed, full_name
	Direction  string // asc, desc
	PerPage    int    // results per page (max 100)
	Page       int    // page number
}

// ListRepos lists repositories for the authenticated user
func (c *Client) ListRepos(ctx context.Context, opts *ListReposOptions) ([]Repository, error) {
	if opts == nil {
		opts = &ListReposOptions{
			Sort:    "updated",
			PerPage: 30,
		}
	}

	path := fmt.Sprintf("/user/repos?sort=%s&per_page=%d",
		opts.Sort, opts.PerPage)

	if opts.Visibility != "" {
		path += "&visibility=" + opts.Visibility
	}
	if opts.Direction != "" {
		path += "&direction=" + opts.Direction
	}
	if opts.Page > 0 {
		path += fmt.Sprintf("&page=%d", opts.Page)
	}

	var repos []Repository
	if err := c.get(ctx, path, &repos); err != nil {
		return nil, err
	}

	return repos, nil
}

// GetRepo gets a specific repository
func (c *Client) GetRepo(ctx context.Context, owner, repo string) (*Repository, error) {
	var repository Repository
	path := fmt.Sprintf("/repos/%s/%s", owner, repo)

	if err := c.get(ctx, path, &repository); err != nil {
		return nil, err
	}

	return &repository, nil
}

// CreateRepoRequest specifies options for creating a repository
type CreateRepoRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Private     bool   `json:"private,omitempty"`
	AutoInit    bool   `json:"auto_init,omitempty"` // Initialize with README
}

// CreateRepo creates a new repository for the authenticated user
func (c *Client) CreateRepo(ctx context.Context, req CreateRepoRequest) (*Repository, error) {
	var repo Repository
	if err := c.post(ctx, "/user/repos", req, &repo); err != nil {
		return nil, err
	}

	return &repo, nil
}

// DeleteRepo deletes a repository
func (c *Client) DeleteRepo(ctx context.Context, owner, repo string) error {
	path := fmt.Sprintf("/repos/%s/%s", owner, repo)
	return c.delete(ctx, path)
}
