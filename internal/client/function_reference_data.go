package client

import "context"

// FunctionRegion describes a region that functions can deploy to.
type FunctionRegion struct {
	Code  string `json:"code"`
	Label string `json:"label"`
	Flag  string `json:"flag,omitempty"`
}

func (c *Client) ListFunctionRegions(ctx context.Context) ([]FunctionRegion, error) {
	var out []FunctionRegion
	if err := c.Request(ctx, "GET", "/functions/regions", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// FunctionRuntime describes a language runtime available to functions and
// durable functions.
type FunctionRuntime struct {
	Name           string `json:"name"`
	Label          string `json:"label,omitempty"`
	Language       string `json:"language,omitempty"`
	Default        bool   `json:"default,omitempty"`
	DurableCapable bool   `json:"durable_capable,omitempty"`
}

type listFunctionRuntimesResponse struct {
	Runtimes []FunctionRuntime `json:"runtimes"`
}

func (c *Client) ListFunctionRuntimes(ctx context.Context) ([]FunctionRuntime, error) {
	var out listFunctionRuntimesResponse
	if err := c.Request(ctx, "GET", "/functions/runtimes", nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Runtimes, nil
}

// FunctionDeployment is one historical deployment of a function.
type FunctionDeployment struct {
	ID        string `json:"id"`
	Status    string `json:"status,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

func (c *Client) ListFunctionDeployments(ctx context.Context, projectID, functionID string) ([]FunctionDeployment, error) {
	var out Page[FunctionDeployment]
	path := "/projects/" + EncodePathSegment(projectID) + "/functions/" + EncodePathSegment(functionID) + "/deployments"
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

func (c *Client) ListDurableFunctionDeployments(ctx context.Context, projectID, functionID string) ([]FunctionDeployment, error) {
	var out Page[FunctionDeployment]
	path := "/projects/" + EncodePathSegment(projectID) + "/durable-functions/" + EncodePathSegment(functionID) + "/deployments"
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}
