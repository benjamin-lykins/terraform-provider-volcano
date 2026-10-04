package client

import "context"

// Variable is a project-scoped environment variable propagated to
// functions (and, when shared, to frontends).
type Variable struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	Shared    bool   `json:"shared,omitempty"`
	Status    string `json:"status,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type CreateVariableRequest struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Shared *bool  `json:"shared,omitempty"`
}

type UpdateVariableRequest struct {
	Value  *string `json:"value,omitempty"`
	Shared *bool   `json:"shared,omitempty"`
}

func (c *Client) CreateVariable(ctx context.Context, projectID string, in CreateVariableRequest) (*Variable, error) {
	var out Variable
	if err := c.Request(ctx, "POST", "/projects/"+EncodePathSegment(projectID)+"/variables", nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetVariable(ctx context.Context, projectID, name string) (*Variable, error) {
	var out Variable
	if err := c.Request(ctx, "GET", "/projects/"+EncodePathSegment(projectID)+"/variables/"+EncodePathSegment(name), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateVariable(ctx context.Context, projectID, name string, in UpdateVariableRequest) (*Variable, error) {
	var out Variable
	if err := c.Request(ctx, "PUT", "/projects/"+EncodePathSegment(projectID)+"/variables/"+EncodePathSegment(name), nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteVariable(ctx context.Context, projectID, name string) error {
	return c.Request(ctx, "DELETE", "/projects/"+EncodePathSegment(projectID)+"/variables/"+EncodePathSegment(name), nil, nil, nil)
}
