package client

import "context"

type SetFrontendSharedVariablesRequest struct {
	FrontendSharedVariables []string `json:"frontend_shared_variables"`
}

// SetFrontendSharedVariables replaces the project's complete shared
// frontend-variable list. There is no GET for this resource; callers must
// track the list themselves (as Terraform state does).
func (c *Client) SetFrontendSharedVariables(ctx context.Context, projectID string, names []string) error {
	path := "/projects/" + EncodePathSegment(projectID) + "/frontend-shared-variables"
	return c.Request(ctx, "PUT", path, nil, SetFrontendSharedVariablesRequest{FrontendSharedVariables: names}, nil)
}
