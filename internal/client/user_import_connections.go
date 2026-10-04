package client

import "context"

// UserImportConnection is an account-level link to a third-party platform
// (e.g. another hosting provider) used to import projects from it.
type UserImportConnection struct {
	ID                  string   `json:"id"`
	Provider            string   `json:"provider"`
	AccountID           string   `json:"account_id,omitempty"`
	AccountName         string   `json:"account_name,omitempty"`
	ConfigurationID     string   `json:"configuration_id,omitempty"`
	GrantedScopes       []string `json:"granted_scopes,omitempty"`
	Status              string   `json:"status,omitempty"`
	LastAuthenticatedAt string   `json:"last_authenticated_at,omitempty"`
	ExpiresAt           string   `json:"expires_at,omitempty"`
	CreatedAt           string   `json:"created_at,omitempty"`
	UpdatedAt           string   `json:"updated_at,omitempty"`
}

func (c *Client) ListUserImportConnections(ctx context.Context) ([]UserImportConnection, error) {
	var out struct {
		Connections []UserImportConnection `json:"connections"`
	}
	if err := c.Request(ctx, "GET", "/user/imports/connections", nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Connections, nil
}

// ImportSource is a project available to import from a connected
// third-party platform.
type ImportSource struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Framework string `json:"framework,omitempty"`
	AccountID string `json:"account_id,omitempty"`
}

func (c *Client) ListImportSources(ctx context.Context, provider string) ([]ImportSource, error) {
	var out struct {
		Sources []ImportSource `json:"sources"`
	}
	path := "/imports/" + EncodePathSegment(provider) + "/sources"
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Sources, nil
}
