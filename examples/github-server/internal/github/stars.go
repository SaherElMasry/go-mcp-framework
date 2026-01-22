// internal/github/stars.go
package github

import (
	"context"
	"fmt"
)

// StarRepo stars a repository
func (c *Client) StarRepo(ctx context.Context, owner, repo string) error {
	path := fmt.Sprintf("/user/starred/%s/%s", owner, repo)
	return c.put(ctx, path)
}

// UnstarRepo unstars a repository
func (c *Client) UnstarRepo(ctx context.Context, owner, repo string) error {
	path := fmt.Sprintf("/user/starred/%s/%s", owner, repo)
	return c.delete(ctx, path)
}

// IsStarred checks if a repository is starred
func (c *Client) IsStarred(ctx context.Context, owner, repo string) (bool, error) {
	path := fmt.Sprintf("/user/starred/%s/%s", owner, repo)

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// 204 = starred, 404 = not starred
	if resp.StatusCode == 204 {
		return true, nil
	} else if resp.StatusCode == 404 {
		return false, nil
	}

	// Unexpected status code
	return false, checkResponse(resp)
}

// ListStarredOptions specifies options for listing starred repos
type ListStarredOptions struct {
	Sort      string // created, updated
	Direction string // asc, desc
	PerPage   int
	Page      int
}

// ListStarred lists repositories the user has starred
func (c *Client) ListStarred(ctx context.Context, opts *ListStarredOptions) ([]Repository, error) {
	if opts == nil {
		opts = &ListStarredOptions{
			Sort:    "updated",
			PerPage: 30,
		}
	}

	path := fmt.Sprintf("/user/starred?sort=%s&per_page=%d",
		opts.Sort, opts.PerPage)

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
