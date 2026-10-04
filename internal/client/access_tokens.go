package client

import "context"

// ProjectAccessToken is a control-plane credential scoped to one project.
// The plaintext secret is returned only once, in CreatedProjectAccessToken.
type ProjectAccessToken struct {
	ID         string `json:"id"`
	ProjectID  string `json:"project_id"`
	Name       string `json:"name"`
	Scope      string `json:"scope"`
	Status     string `json:"status,omitempty"`
	ExpiresAt  string `json:"expires_at,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	LastUsedAt string `json:"last_used_at,omitempty"`
}

type CreatedProjectAccessToken struct {
	ProjectAccessToken
	Token string `json:"token"`
}

type CreateProjectAccessTokenRequest struct {
	Name      string `json:"name"`
	Scope     string `json:"scope"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

func (c *Client) CreateProjectAccessToken(ctx context.Context, projectID string, in CreateProjectAccessTokenRequest) (*CreatedProjectAccessToken, error) {
	var out CreatedProjectAccessToken
	if err := c.Request(ctx, "POST", "/projects/"+EncodePathSegment(projectID)+"/access-tokens", nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetProjectAccessToken(ctx context.Context, projectID, tokenID string) (*ProjectAccessToken, error) {
	var out ProjectAccessToken
	if err := c.Request(ctx, "GET", "/projects/"+EncodePathSegment(projectID)+"/access-tokens/"+EncodePathSegment(tokenID), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteProjectAccessToken(ctx context.Context, projectID, tokenID string) error {
	return c.Request(ctx, "DELETE", "/projects/"+EncodePathSegment(projectID)+"/access-tokens/"+EncodePathSegment(tokenID), nil, nil, nil)
}
