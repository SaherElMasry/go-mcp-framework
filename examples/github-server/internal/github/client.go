// internal/github/client.go
package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/SaherElMasry/go-mcp-framework/examples/github-server/internal/config"
)

// Client is the GitHub API client
type Client struct {
	baseURL   string
	token     string
	userAgent string
	http      *http.Client
}

// NewClient creates a new GitHub API client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		baseURL:   cfg.GitHub.BaseURL,
		token:     cfg.GitHub.Token,
		userAgent: cfg.GitHub.UserAgent,
		http: &http.Client{
			Timeout: cfg.GitHub.Timeout,
		},
	}
}

// doRequest performs an authenticated HTTP request
func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication and headers
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", c.userAgent)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute request
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// get performs a GET request and decodes JSON response
func (c *Client) get(ctx context.Context, path string, result interface{}) error {
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return err
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// post performs a POST request with JSON body
func (c *Client) post(ctx context.Context, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {

		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	resp, err := c.doRequest(ctx, "POST", path, bodyReader)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return err
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// put performs a PUT request
func (c *Client) put(ctx context.Context, path string) error {
	resp, err := c.doRequest(ctx, "PUT", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// delete performs a DELETE request
func (c *Client) delete(ctx context.Context, path string) error {
	resp, err := c.doRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

// checkResponse checks for API errors
func checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil // Success
	}

	// Read error body
	body, _ := io.ReadAll(resp.Body)

	// Try to parse GitHub error format
	var errResp struct {
		Message string `json:"message"`
		Errors  []struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Message != "" {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    errResp.Message,
			Response:   resp,
		}
	}

	// Fallback to raw body
	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    string(body),
		Response:   resp,
	}
}

// APIError represents a GitHub API error
type APIError struct {
	StatusCode int
	Message    string
	Response   *http.Response
}

func (e *APIError) Error() string {
	return fmt.Sprintf("GitHub API error %d: %s", e.StatusCode, e.Message)
}

// IsNotFound returns true if the error is a 404
func IsNotFound(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 404
	}
	return false
}

// IsRateLimited returns true if the error is rate limiting
func IsRateLimited(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 403 || apiErr.StatusCode == 429
	}
	return false
}

// RateLimit holds rate limit information
type RateLimit struct {
	Limit     int
	Remaining int
	Reset     time.Time
}

// GetRateLimit returns current rate limit status
func (c *Client) GetRateLimit(ctx context.Context) (*RateLimit, error) {
	var result struct {
		Rate struct {
			Limit     int   `json:"limit"`
			Remaining int   `json:"remaining"`
			Reset     int64 `json:"reset"`
		} `json:"rate"`
	}

	if err := c.get(ctx, "/rate_limit", &result); err != nil {
		return nil, err
	}

	return &RateLimit{
		Limit:     result.Rate.Limit,
		Remaining: result.Rate.Remaining,
		Reset:     time.Unix(result.Rate.Reset, 0),
	}, nil
}
