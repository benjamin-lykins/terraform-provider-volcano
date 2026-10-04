package client

import "context"

// ServiceKey bypasses Row-Level Security; it is intended for backend/admin
// use only and must never be exposed to a frontend.
type ServiceKey struct {
	ID          string   `json:"id"`
	ProjectID   string   `json:"project_id"`
	Name        string   `json:"name"`
	KeyPrefix   string   `json:"key_prefix,omitempty"`
	KeyValue    string   `json:"key_value"`
	Permissions []string `json:"permissions,omitempty"`
	CreatedAt   string   `json:"created_at,omitempty"`
	UpdatedAt   string   `json:"updated_at,omitempty"`
}

type CreateServiceKeyRequest struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions,omitempty"`
}

func (c *Client) CreateServiceKey(ctx context.Context, projectID string, in CreateServiceKeyRequest) (*ServiceKey, error) {
	var out ServiceKey
	if err := c.Request(ctx, "POST", "/projects/"+EncodePathSegment(projectID)+"/service-keys", nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetServiceKey(ctx context.Context, projectID, keyID string) (*ServiceKey, error) {
	var out ServiceKey
	if err := c.Request(ctx, "GET", "/projects/"+EncodePathSegment(projectID)+"/service-keys/"+EncodePathSegment(keyID), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteServiceKey(ctx context.Context, projectID, keyID string) error {
	return c.Request(ctx, "DELETE", "/projects/"+EncodePathSegment(projectID)+"/service-keys/"+EncodePathSegment(keyID), nil, nil, nil)
}
