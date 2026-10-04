package client

import "context"

// FrontendFunctionRoute forwards requests under path_prefix on a frontend
// to a function.
type FrontendFunctionRoute struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	FrontendID  string `json:"frontend_id"`
	FunctionID  string `json:"function_id"`
	PathPrefix  string `json:"path_prefix"`
	StripPrefix bool   `json:"strip_prefix,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

type FrontendFunctionRouteRequest struct {
	FunctionID  string `json:"function_id"`
	PathPrefix  string `json:"path_prefix"`
	StripPrefix *bool  `json:"strip_prefix,omitempty"`
}

func (c *Client) CreateFrontendFunctionRoute(ctx context.Context, projectID, frontendID string, in FrontendFunctionRouteRequest) (*FrontendFunctionRoute, error) {
	var out FrontendFunctionRoute
	path := "/projects/" + EncodePathSegment(projectID) + "/frontends/" + EncodePathSegment(frontendID) + "/function-routes"
	if err := c.Request(ctx, "POST", path, nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListFrontendFunctionRoutes is used for Read, since there is no
// single-route GET endpoint.
func (c *Client) ListFrontendFunctionRoutes(ctx context.Context, projectID, frontendID string) ([]FrontendFunctionRoute, error) {
	var out Page[FrontendFunctionRoute]
	path := "/projects/" + EncodePathSegment(projectID) + "/frontends/" + EncodePathSegment(frontendID) + "/function-routes"
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

func (c *Client) UpdateFrontendFunctionRoute(ctx context.Context, projectID, frontendID, routeID string, in FrontendFunctionRouteRequest) (*FrontendFunctionRoute, error) {
	var out FrontendFunctionRoute
	path := "/projects/" + EncodePathSegment(projectID) + "/frontends/" + EncodePathSegment(frontendID) + "/function-routes/" + EncodePathSegment(routeID)
	if err := c.Request(ctx, "PUT", path, nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteFrontendFunctionRoute(ctx context.Context, projectID, frontendID, routeID string) error {
	path := "/projects/" + EncodePathSegment(projectID) + "/frontends/" + EncodePathSegment(frontendID) + "/function-routes/" + EncodePathSegment(routeID)
	return c.Request(ctx, "DELETE", path, nil, nil, nil)
}
