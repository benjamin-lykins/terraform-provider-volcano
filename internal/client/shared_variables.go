package client

import "context"

type SetSharedVariablesRequest struct {
	SharedVariables []string `json:"shared_variables"`
}

// SetSharedVariables replaces the project's complete shared
// function-variable list (membership only, not values). There is no GET
// for this resource.
func (c *Client) SetSharedVariables(ctx context.Context, projectID string, names []string) error {
	path := "/projects/" + EncodePathSegment(projectID) + "/shared-variables"
	return c.Request(ctx, "PUT", path, nil, SetSharedVariablesRequest{SharedVariables: names}, nil)
}
