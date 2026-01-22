package auth

import (
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
	"golang.org/x/oauth2/slack"
)

// ProviderFactory creates OAuth2 providers for popular services
type ProviderFactory struct {
	tokenStore TokenStore
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory(tokenStore TokenStore) *ProviderFactory {
	return &ProviderFactory{
		tokenStore: tokenStore,
	}
}

// Create creates an OAuth2 provider for a service
func (f *ProviderFactory) Create(providerName, clientID, clientSecret, redirectURL string, scopes []string) (*OAuth2Provider, error) {
	var endpoint oauth2.Endpoint
	var defaultScopes []string

	switch providerName {
	case "github":
		endpoint = github.Endpoint
		if scopes == nil {
			defaultScopes = []string{"repo", "user"}
		}

	case "google":
		endpoint = google.Endpoint
		if scopes == nil {
			defaultScopes = []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			}
		}

	case "facebook":
		endpoint = facebook.Endpoint
		if scopes == nil {
			defaultScopes = []string{"public_profile", "email"}
		}

	case "microsoft":
		endpoint = microsoft.AzureADEndpoint("")
		if scopes == nil {
			defaultScopes = []string{
				"https://graph.microsoft.com/User.Read",
			}
		}

	case "slack":
		endpoint = slack.Endpoint
		if scopes == nil {
			defaultScopes = []string{"channels:read", "chat:write"}
		}

	default:
		return nil, fmt.Errorf("unsupported OAuth provider: %s", providerName)
	}

	if scopes == nil {
		scopes = defaultScopes
	}

	config := OAuth2Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		AuthURL:      endpoint.AuthURL,
		TokenURL:     endpoint.TokenURL,
	}

	return NewOAuth2Provider(providerName, config, f.tokenStore), nil
}

// GetDefaultScopes returns default scopes for a provider
func GetDefaultScopes(providerName string) []string {
	switch providerName {
	case "github":
		return []string{"repo", "user"}
	case "google":
		return []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		}
	case "facebook":
		return []string{"public_profile", "email"}
	case "microsoft":
		return []string{"https://graph.microsoft.com/User.Read"}
	case "slack":
		return []string{"channels:read", "chat:write"}
	default:
		return nil
	}
}
