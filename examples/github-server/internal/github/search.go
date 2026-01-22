// internal/github/search.go
package github

import (
	"context"
	"fmt"
	"net/url"
)

// SearchReposOptions specifies options for searching repositories
type SearchReposOptions struct {
	Query   string // Search query
	Sort    string // stars, forks, updated
	Order   string // asc, desc
	PerPage int
	Page    int
}

// SearchRepos searches for repositories
func (c *Client) SearchRepos(ctx context.Context, opts SearchReposOptions) (*SearchResult, error) {
	if opts.Query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	if opts.PerPage == 0 {
		opts.PerPage = 30
	}

	path := fmt.Sprintf("/search/repositories?q=%s&per_page=%d",
		url.QueryEscape(opts.Query), opts.PerPage)

	if opts.Sort != "" {
		path += "&sort=" + opts.Sort
	}
	if opts.Order != "" {
		path += "&order=" + opts.Order
	}
	if opts.Page > 0 {
		path += fmt.Sprintf("&page=%d", opts.Page)
	}

	var result SearchResult
	if err := c.get(ctx, path, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SearchCodeOptions specifies options for searching code
type SearchCodeOptions struct {
	Query   string // Search query
	Sort    string // indexed
	Order   string // asc, desc
	PerPage int
	Page    int
}

// CodeSearchResult represents code search results
type CodeSearchResult struct {
	TotalCount int `json:"total_count"`
	Items      []struct {
		Name       string     `json:"name"`
		Path       string     `json:"path"`
		SHA        string     `json:"sha"`
		HTMLURL    string     `json:"html_url"`
		Repository Repository `json:"repository"`
	} `json:"items"`
}

// SearchCode searches for code
func (c *Client) SearchCode(ctx context.Context, opts SearchCodeOptions) (*CodeSearchResult, error) {
	if opts.Query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	if opts.PerPage == 0 {
		opts.PerPage = 30
	}

	path := fmt.Sprintf("/search/code?q=%s&per_page=%d",
		url.QueryEscape(opts.Query), opts.PerPage)

	if opts.Sort != "" {
		path += "&sort=" + opts.Sort
	}
	if opts.Order != "" {
		path += "&order=" + opts.Order
	}
	if opts.Page > 0 {
		path += fmt.Sprintf("&page=%d", opts.Page)
	}

	var result CodeSearchResult
	if err := c.get(ctx, path, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
