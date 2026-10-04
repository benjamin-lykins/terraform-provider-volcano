package client

import "context"

// OAuthConfig is a project's configuration for one OAuth provider.
type OAuthConfig struct {
	ID           string   `json:"id"`
	Provider     string   `json:"provider"`
	ClientID     string   `json:"client_id,omitempty"`
	ClientSecret string   `json:"client_secret,omitempty"`
	RedirectURL  string   `json:"redirect_url,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
	Enabled      bool     `json:"enabled"`
	CreatedAt    string   `json:"created_at,omitempty"`
	UpdatedAt    string   `json:"updated_at,omitempty"`
}

type CreateOAuthConfigRequest struct {
	Provider     string   `json:"provider"`
	ClientID     string   `json:"client_id,omitempty"`
	ClientSecret string   `json:"client_secret,omitempty"`
	RedirectURL  string   `json:"redirect_url,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
}

type UpdateOAuthConfigRequest struct {
	ClientID     string   `json:"client_id,omitempty"`
	ClientSecret string   `json:"client_secret,omitempty"`
	RedirectURL  string   `json:"redirect_url,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
	Enabled      *bool    `json:"enabled,omitempty"`
}

func (c *Client) CreateOAuthConfig(ctx context.Context, projectID string, in CreateOAuthConfigRequest) (*OAuthConfig, error) {
	var out OAuthConfig
	path := "/projects/" + EncodePathSegment(projectID) + "/oauth/configs"
	if err := c.Request(ctx, "POST", path, nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetOAuthConfig(ctx context.Context, projectID, provider string) (*OAuthConfig, error) {
	var out OAuthConfig
	path := "/projects/" + EncodePathSegment(projectID) + "/oauth/configs/" + EncodePathSegment(provider)
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateOAuthConfig(ctx context.Context, projectID, provider string, in UpdateOAuthConfigRequest) (*OAuthConfig, error) {
	var out OAuthConfig
	path := "/projects/" + EncodePathSegment(projectID) + "/oauth/configs/" + EncodePathSegment(provider)
	if err := c.Request(ctx, "PUT", path, nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteOAuthConfig(ctx context.Context, projectID, provider string) error {
	path := "/projects/" + EncodePathSegment(projectID) + "/oauth/configs/" + EncodePathSegment(provider)
	return c.Request(ctx, "DELETE", path, nil, nil, nil)
}

// OAuthProviderInfo describes an OAuth provider type a project could
// configure (not whether it is actually configured).
type OAuthProviderInfo struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	DefaultScopes []string `json:"default_scopes,omitempty"`
}

func (c *Client) ListOAuthProviderInfo(ctx context.Context, projectID string) ([]OAuthProviderInfo, error) {
	var out struct {
		Providers []OAuthProviderInfo `json:"providers"`
	}
	path := "/projects/" + EncodePathSegment(projectID) + "/oauth/providers"
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Providers, nil
}
