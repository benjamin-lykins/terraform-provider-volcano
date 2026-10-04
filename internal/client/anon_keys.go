package client

import "context"

// AnonKey is a project-specific public key safe to expose in frontend code.
type AnonKey struct {
	ID          string   `json:"id"`
	ProjectID   string   `json:"project_id"`
	Name        string   `json:"name"`
	KeyValue    string   `json:"key_value"`
	Permissions []string `json:"permissions,omitempty"`
	IsDefault   bool     `json:"is_default"`
	CreatedAt   string   `json:"created_at,omitempty"`
}

type CreateAnonKeyRequest struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions,omitempty"`
}

func (c *Client) CreateAnonKey(ctx context.Context, projectID string, in CreateAnonKeyRequest) (*AnonKey, error) {
	var out AnonKey
	if err := c.Request(ctx, "POST", "/projects/"+EncodePathSegment(projectID)+"/anon-keys", nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetAnonKey(ctx context.Context, projectID, keyID string) (*AnonKey, error) {
	var out AnonKey
	if err := c.Request(ctx, "GET", "/projects/"+EncodePathSegment(projectID)+"/anon-keys/"+EncodePathSegment(keyID), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteAnonKey(ctx context.Context, projectID, keyID string) error {
	return c.Request(ctx, "DELETE", "/projects/"+EncodePathSegment(projectID)+"/anon-keys/"+EncodePathSegment(keyID), nil, nil, nil)
}

func (c *Client) SetDefaultAnonKey(ctx context.Context, projectID, keyID string) error {
	return c.Request(ctx, "POST", "/projects/"+EncodePathSegment(projectID)+"/anon-keys/"+EncodePathSegment(keyID)+"/set-default", nil, nil, nil)
}
