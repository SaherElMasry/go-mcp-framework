// internal/github/types.go
package github

import "time"

// User represents a GitHub user
type User struct {
	Login       string    `json:"login"`
	ID          int       `json:"id"`
	NodeID      string    `json:"node_id"`
	AvatarURL   string    `json:"avatar_url"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Bio         string    `json:"bio"`
	Company     string    `json:"company"`
	Location    string    `json:"location"`
	PublicRepos int       `json:"public_repos"`
	Followers   int       `json:"followers"`
	Following   int       `json:"following"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Repository represents a GitHub repository
type Repository struct {
	ID            int       `json:"id"`
	NodeID        string    `json:"node_id"`
	Name          string    `json:"name"`
	FullName      string    `json:"full_name"`
	Description   string    `json:"description"`
	Private       bool      `json:"private"`
	Fork          bool      `json:"fork"`
	Owner         User      `json:"owner"`
	HTMLURL       string    `json:"html_url"`
	CloneURL      string    `json:"clone_url"`
	GitURL        string    `json:"git_url"`
	SSHURL        string    `json:"ssh_url"`
	Language      string    `json:"language"`
	Size          int       `json:"size"`
	Stars         int       `json:"stargazers_count"`
	Watchers      int       `json:"watchers_count"`
	Forks         int       `json:"forks_count"`
	OpenIssues    int       `json:"open_issues_count"`
	DefaultBranch string    `json:"default_branch"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	PushedAt      time.Time `json:"pushed_at"`
}

// Issue represents a GitHub issue
type Issue struct {
	ID        int        `json:"id"`
	NodeID    string     `json:"node_id"`
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	State     string     `json:"state"` // open, closed
	Locked    bool       `json:"locked"`
	User      User       `json:"user"`
	Labels    []Label    `json:"labels"`
	Assignees []User     `json:"assignees"`
	Comments  int        `json:"comments"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at,omitempty"`
	HTMLURL   string     `json:"html_url"`
}

// Label represents a GitHub label
type Label struct {
	ID          int    `json:"id"`
	NodeID      string `json:"node_id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// CreateIssueRequest is the request to create an issue
type CreateIssueRequest struct {
	Title     string   `json:"title"`
	Body      string   `json:"body,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
	Labels    []string `json:"labels,omitempty"`
}

// ReadmeContent represents repository README
type ReadmeContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	SHA         string `json:"sha"`
	Size        int    `json:"size"`
	Content     string `json:"content"`  // Base64 encoded
	Encoding    string `json:"encoding"` // "base64"
	DownloadURL string `json:"download_url"`
	HTMLURL     string `json:"html_url"`
}

// SearchResult represents search results
type SearchResult struct {
	TotalCount int          `json:"total_count"`
	Items      []Repository `json:"items"`
}
