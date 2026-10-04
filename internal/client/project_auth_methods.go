package client

import "context"

type AuthMethodStatus struct {
	Method  string `json:"method,omitempty"`
	Name    string `json:"name,omitempty"`
	Enabled bool   `json:"enabled"`
}

type OAuthProviderMethodStatus struct {
	Provider    string   `json:"provider"`
	Method      string   `json:"method,omitempty"`
	Name        string   `json:"name,omitempty"`
	Enabled     bool     `json:"enabled"`
	RedirectURL string   `json:"redirect_url,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
}

// ProjectAuthMethods is the project's enabled-sign-in-method summary.
type ProjectAuthMethods struct {
	Anonymous        AuthMethodStatus            `json:"anonymous"`
	EmailPassword    AuthMethodStatus            `json:"email_password"`
	OAuthProviders   []OAuthProviderMethodStatus `json:"oauth_providers,omitempty"`
	AvailableMethods []string                    `json:"available_methods,omitempty"`
}

type OAuthProviderToggle struct {
	Provider string `json:"provider"`
	Enabled  bool   `json:"enabled"`
}

type UpdateProjectAuthMethodsRequest struct {
	EnableAnonymous     *bool                 `json:"enable_anonymous,omitempty"`
	EnableEmailPassword *bool                 `json:"enable_email_password,omitempty"`
	OAuthProviders      []OAuthProviderToggle `json:"oauth_providers,omitempty"`
}

func (c *Client) GetProjectAuthMethods(ctx context.Context, projectID string) (*ProjectAuthMethods, error) {
	var out ProjectAuthMethods
	path := "/projects/" + EncodePathSegment(projectID) + "/auth/methods"
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateProjectAuthMethods(ctx context.Context, projectID string, in UpdateProjectAuthMethodsRequest) (*ProjectAuthMethods, error) {
	var out ProjectAuthMethods
	path := "/projects/" + EncodePathSegment(projectID) + "/auth/methods"
	if err := c.Request(ctx, "PUT", path, nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
