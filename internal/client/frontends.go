package client

import "context"

// Frontend is a deployed Next.js frontend application.
type Frontend struct {
	ID                  string   `json:"id"`
	ProjectID           string   `json:"project_id"`
	Name                string   `json:"name"`
	Framework           string   `json:"framework,omitempty"`
	AppRoot             string   `json:"app_root,omitempty"`
	VariableScope       string   `json:"variable_scope,omitempty"`
	DeclaredVariables   []string `json:"declared_variables,omitempty"`
	Status              string   `json:"status,omitempty"`
	SiteURL             string   `json:"site_url,omitempty"`
	CustomDomain        string   `json:"custom_domain,omitempty"`
	CustomDomainStatus  string   `json:"custom_domain_status,omitempty"`
	DeployedRegions     []string `json:"deployed_regions,omitempty"`
	CurrentDeploymentID string   `json:"current_deployment_id,omitempty"`
	CreatedAt           string   `json:"created_at,omitempty"`
	UpdatedAt           string   `json:"updated_at,omitempty"`
}

type CreateFrontendRequest struct {
	Name           string
	ArchiveName    string
	ArchiveContent []byte
	AppRoot        string
	Framework      string
	VariableScope  string
	Variables      []string
}

func (c *Client) CreateFrontend(ctx context.Context, projectID string, in CreateFrontendRequest) (*Frontend, error) {
	fields := []MultipartField{{Name: "name", Value: in.Name}}
	if in.AppRoot != "" {
		fields = append(fields, MultipartField{Name: "app_root", Value: in.AppRoot})
	}
	if in.Framework != "" {
		fields = append(fields, MultipartField{Name: "framework", Value: in.Framework})
	}
	if in.VariableScope != "" {
		fields = append(fields, MultipartField{Name: "variable_scope", Value: in.VariableScope})
	}
	for _, v := range in.Variables {
		fields = append(fields, MultipartField{Name: "variables", Value: v})
	}

	var out Frontend
	file := MultipartFile{FieldName: "archive", Filename: in.ArchiveName, Content: in.ArchiveContent}
	if err := c.MultipartRequest(ctx, "POST", "/projects/"+EncodePathSegment(projectID)+"/frontends", fields, file, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetFrontend(ctx context.Context, projectID, frontendID string) (*Frontend, error) {
	var out Frontend
	path := "/projects/" + EncodePathSegment(projectID) + "/frontends/" + EncodePathSegment(frontendID)
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteFrontend(ctx context.Context, projectID, frontendID string) error {
	path := "/projects/" + EncodePathSegment(projectID) + "/frontends/" + EncodePathSegment(frontendID)
	return c.Request(ctx, "DELETE", path, nil, nil, nil)
}
