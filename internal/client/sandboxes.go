package client

import "context"

// Sandbox is a standing, reusable Linux sandbox environment (distinct
// from the ephemeral execution sessions run inside one).
type Sandbox struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	Preset    string `json:"preset"`
	MemoryMB  int64  `json:"memory_mb,omitempty"`
	Status    string `json:"status,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

type CreateSandboxRequest struct {
	Name     string `json:"name"`
	Preset   string `json:"preset"`
	MemoryMB *int64 `json:"memory_mb,omitempty"`
}

type UpdateSandboxRequest struct {
	Name string `json:"name"`
}

func (c *Client) CreateSandbox(ctx context.Context, projectID string, in CreateSandboxRequest) (*Sandbox, error) {
	var out Sandbox
	path := "/projects/" + EncodePathSegment(projectID) + "/sandboxes"
	if err := c.Request(ctx, "POST", path, nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetSandbox(ctx context.Context, projectID, sandboxID string) (*Sandbox, error) {
	var out Sandbox
	path := "/projects/" + EncodePathSegment(projectID) + "/sandboxes/" + EncodePathSegment(sandboxID)
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateSandbox(ctx context.Context, projectID, sandboxID string, in UpdateSandboxRequest) (*Sandbox, error) {
	var out Sandbox
	path := "/projects/" + EncodePathSegment(projectID) + "/sandboxes/" + EncodePathSegment(sandboxID)
	if err := c.Request(ctx, "PATCH", path, nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteSandbox(ctx context.Context, projectID, sandboxID string) error {
	path := "/projects/" + EncodePathSegment(projectID) + "/sandboxes/" + EncodePathSegment(sandboxID)
	return c.Request(ctx, "DELETE", path, nil, nil, nil)
}

// SandboxPreset describes a sandbox size/runtime preset.
type SandboxPreset struct {
	ID       string   `json:"id"`
	Runtime  string   `json:"runtime,omitempty"`
	Version  string   `json:"version,omitempty"`
	MemoryMB int64    `json:"memory_mb,omitempty"`
	Regions  []string `json:"regions,omitempty"`
}

func (c *Client) ListSandboxPresets(ctx context.Context) ([]SandboxPreset, error) {
	var out struct {
		Data []SandboxPreset `json:"data"`
	}
	if err := c.Request(ctx, "GET", "/sandboxes/presets", nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

func (c *Client) ListSandboxDeployments(ctx context.Context, projectID, sandboxID string) ([]FunctionDeployment, error) {
	var out Page[FunctionDeployment]
	path := "/projects/" + EncodePathSegment(projectID) + "/sandboxes/" + EncodePathSegment(sandboxID) + "/deployments"
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}
