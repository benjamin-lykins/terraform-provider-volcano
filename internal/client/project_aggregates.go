package client

import "context"

// Deployment is one entry in a project's (or the account's) deployment
// history, across every resource type that deploys (functions, durable
// functions, frontends).
type Deployment struct {
	ID           string `json:"id"`
	ProjectID    string `json:"project_id"`
	Resource     string `json:"resource,omitempty"`
	Operation    string `json:"operation,omitempty"`
	Status       string `json:"status,omitempty"`
	Progress     int64  `json:"progress,omitempty"`
	DeploySource string `json:"deploy_source,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
	CompletedAt  string `json:"completed_at,omitempty"`
}

func (c *Client) ListProjectDeployments(ctx context.Context, projectID string) ([]Deployment, error) {
	var out Page[Deployment]
	path := "/projects/" + EncodePathSegment(projectID) + "/deployments"
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

func (c *Client) ListAccountDeployments(ctx context.Context) ([]Deployment, error) {
	var out Page[Deployment]
	if err := c.Request(ctx, "GET", "/deployments", nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

func (c *Client) ListProjectDomains(ctx context.Context, projectID string) ([]FrontendDomain, error) {
	var out Page[FrontendDomain]
	path := "/projects/" + EncodePathSegment(projectID) + "/domains"
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

func (c *Client) ListProjectSchedulers(ctx context.Context, projectID string) ([]FunctionScheduler, error) {
	var out Page[FunctionScheduler]
	path := "/projects/" + EncodePathSegment(projectID) + "/schedulers"
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

// ProjectHealthCheck is one evaluated health signal for a project
// resource (lifecycle/storage checks, e.g. an idle function or a bucket
// approaching its size limit).
type ProjectHealthCheck struct {
	ID         string `json:"id"`
	Category   string `json:"category,omitempty"`
	Status     string `json:"status,omitempty"`
	ReasonCode string `json:"reason_code,omitempty"`
	Scope      struct {
		Resource struct {
			Type string `json:"type,omitempty"`
			ID   string `json:"id,omitempty"`
			Name string `json:"name,omitempty"`
		} `json:"resource"`
	} `json:"scope"`
}

type ProjectHealth struct {
	ProjectID    string               `json:"project_id,omitempty"`
	Status       string               `json:"status,omitempty"`
	DataStatus   string               `json:"data_status,omitempty"`
	ObservedAt   string               `json:"observed_at,omitempty"`
	FreshThrough string               `json:"fresh_through,omitempty"`
	Checks       []ProjectHealthCheck `json:"checks,omitempty"`
	Findings     []ProjectHealthCheck `json:"findings,omitempty"`
}

func (c *Client) GetProjectHealth(ctx context.Context, projectID string) (*ProjectHealth, error) {
	var out ProjectHealth
	path := "/projects/" + EncodePathSegment(projectID) + "/health"
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
