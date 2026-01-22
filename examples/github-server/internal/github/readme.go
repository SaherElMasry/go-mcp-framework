// internal/github/readme.go
package github

import (
	"context"
	"encoding/base64"
	"fmt"
)

// GetReadme gets the README for a repository
func (c *Client) GetReadme(ctx context.Context, owner, repo string) (*ReadmeContent, error) {
	var readme ReadmeContent
	path := fmt.Sprintf("/repos/%s/%s/readme", owner, repo)

	if err := c.get(ctx, path, &readme); err != nil {
		return nil, err
	}

	return &readme, nil
}

// GetReadmeText gets and decodes the README text
func (c *Client) GetReadmeText(ctx context.Context, owner, repo string) (string, error) {
	readme, err := c.GetReadme(ctx, owner, repo)
	if err != nil {
		return "", err
	}

	return DecodeReadme(readme)
}

// DecodeReadme decodes base64 README content
func DecodeReadme(readme *ReadmeContent) (string, error) {
	if readme.Encoding != "base64" {
		return "", fmt.Errorf("unexpected encoding: %s", readme.Encoding)
	}

	decoded, err := base64.StdEncoding.DecodeString(readme.Content)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	return string(decoded), nil
}
