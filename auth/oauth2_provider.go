// auth/oauth2_provider.go
package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

// OAuth2Provider manages OAuth2 authentication
type OAuth2Provider struct {
	*BaseProvider
	config     *oauth2.Config
	tokenStore TokenStore
	token      *OAuth2Token
}

// OAuth2Token represents an OAuth2 token
type OAuth2Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// OAuth2Config holds OAuth2 provider configuration
type OAuth2Config struct {
	ClientID     string   `yaml:"client_id" json:"client_id"`
	ClientSecret string   `yaml:"client_secret" json:"client_secret"`
	RedirectURL  string   `yaml:"redirect_url" json:"redirect_url"`
	Scopes       []string `yaml:"scopes" json:"scopes"`
	AuthURL      string   `yaml:"auth_url" json:"auth_url"`
	TokenURL     string   `yaml:"token_url" json:"token_url"`
}

// NewOAuth2Provider creates a new OAuth2 provider
func NewOAuth2Provider(name string, config OAuth2Config, tokenStore TokenStore) *OAuth2Provider {
	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       config.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.AuthURL,
			TokenURL: config.TokenURL,
		},
	}

	return &OAuth2Provider{
		BaseProvider: NewBaseProvider(name),
		config:       oauth2Config,
		tokenStore:   tokenStore,
	}
}

// GetResource returns an authenticated HTTP client
func (p *OAuth2Provider) GetResource(ctx context.Context, resourceID string) (Resource, error) {
	// Get resource config
	config, err := p.GetResourceConfig(resourceID)
	if err != nil {
		return nil, NewAuthError(p.Name(), resourceID, "get_resource", err)
	}

	// Ensure we have a valid token
	if err := p.ensureValidToken(ctx); err != nil {
		return nil, NewAuthError(p.Name(), resourceID, "get_resource", err)
	}

	// Create OAuth2 HTTP client
	token := &oauth2.Token{
		AccessToken:  p.token.AccessToken,
		RefreshToken: p.token.RefreshToken,
		TokenType:    p.token.TokenType,
		Expiry:       p.token.ExpiresAt,
	}

	client := p.config.Client(ctx, token)

	baseURL, ok := config.Config["base_url"].(string)
	if !ok {
		return nil, NewAuthError(p.Name(), resourceID, "get_resource",
			fmt.Errorf("missing base_url in resource config"))
	}

	return &OAuth2Resource{
		client:     client,
		baseURL:    baseURL,
		resourceID: resourceID,
	}, nil
}

// Validate checks if we have valid credentials
func (p *OAuth2Provider) Validate(ctx context.Context) error {
	if p.token == nil {
		return NewAuthError(p.Name(), "", "validate", ErrInvalidCredentials)
	}

	// Check if token is expired
	if time.Now().After(p.token.ExpiresAt) {
		// Try to refresh
		return p.Refresh(ctx)
	}

	return nil
}

// Refresh refreshes the OAuth2 token
func (p *OAuth2Provider) Refresh(ctx context.Context) error {
	if p.token == nil || p.token.RefreshToken == "" {
		return NewAuthError(p.Name(), "", "refresh", ErrRefreshFailed)
	}

	token := &oauth2.Token{
		RefreshToken: p.token.RefreshToken,
	}

	tokenSource := p.config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return NewAuthError(p.Name(), "", "refresh", err)
	}

	// Update stored token
	p.token = &OAuth2Token{
		AccessToken:  newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
		TokenType:    newToken.TokenType,
		ExpiresAt:    newToken.Expiry,
	}

	// Save to token store
	if p.tokenStore != nil {
		if err := p.tokenStore.Save(ctx, p.Name(), p.token); err != nil {
			return NewAuthError(p.Name(), "", "refresh", err)
		}
	}

	return nil
}

// SetToken sets the OAuth2 token
func (p *OAuth2Provider) SetToken(ctx context.Context, token *OAuth2Token) error {
	p.token = token

	// Save to token store
	if p.tokenStore != nil {
		if err := p.tokenStore.Save(ctx, p.Name(), token); err != nil {
			return NewAuthError(p.Name(), "", "set_token", err)
		}
	}

	return nil
}

// GetAuthURL returns the OAuth2 authorization URL
func (p *OAuth2Provider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state)
}

// Exchange exchanges an authorization code for a token
func (p *OAuth2Provider) Exchange(ctx context.Context, code string) error {
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return NewAuthError(p.Name(), "", "exchange", err)
	}

	p.token = &OAuth2Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresAt:    token.Expiry,
	}

	// Save to token store
	if p.tokenStore != nil {
		if err := p.tokenStore.Save(ctx, p.Name(), p.token); err != nil {
			return NewAuthError(p.Name(), "", "exchange", err)
		}
	}

	return nil
}

// ensureValidToken ensures we have a valid token, refreshing if needed
func (p *OAuth2Provider) ensureValidToken(ctx context.Context) error {
	if p.token == nil {
		// Try to load from token store
		if p.tokenStore != nil {
			token, err := p.tokenStore.Load(ctx, p.Name())
			if err == nil {
				p.token = token
			}
		}

		if p.token == nil {
			return ErrInvalidCredentials
		}
	}

	// Check if expired
	if time.Now().After(p.token.ExpiresAt) {
		return p.Refresh(ctx)
	}

	return nil
}

// Close closes the provider
func (p *OAuth2Provider) Close() error {
	if p.tokenStore != nil {
		return p.tokenStore.Close()
	}
	return nil
}

// OAuth2Resource wraps an OAuth2 HTTP client
type OAuth2Resource struct {
	client     *http.Client
	baseURL    string
	resourceID string
}

func (r *OAuth2Resource) Close() error {
	r.client.CloseIdleConnections()
	return nil
}

func (r *OAuth2Resource) Type() string {
	return "oauth2"
}

// Client returns the HTTP client
func (r *OAuth2Resource) Client() *http.Client {
	return r.client
}

// BaseURL returns the base URL
func (r *OAuth2Resource) BaseURL() string {
	return r.baseURL
}
