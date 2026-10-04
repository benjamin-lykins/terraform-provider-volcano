package client

import (
	"context"
	"encoding/json"
)

// Function is a deployed Volcano function.
type Function struct {
	ID                  string   `json:"id"`
	ProjectID           string   `json:"project_id"`
	Name                string   `json:"name"`
	Handler             string   `json:"handler,omitempty"`
	Runtime             string   `json:"runtime"`
	Status              string   `json:"status,omitempty"`
	InvokeURL           string   `json:"invoke_url,omitempty"`
	HTTPAuthMode        string   `json:"http_auth_mode,omitempty"`
	InvocationMode      string   `json:"invocation_mode,omitempty"`
	IsPublic            bool     `json:"is_public"`
	HasOpenAPISpec      bool     `json:"has_openapi_spec,omitempty"`
	OpenAPISpec         string   `json:"openapi_spec,omitempty"`
	DeployedRegions     []string `json:"deployed_regions,omitempty"`
	CurrentDeploymentID string   `json:"current_deployment_id,omitempty"`
	CreatedAt           string   `json:"created_at,omitempty"`
	UpdatedAt           string   `json:"updated_at,omitempty"`
}

// CreateFunctionRequest creates a function from a local source archive
// (CodePath, read by the caller and passed as CodeContent/CodeFilename).
type CreateFunctionRequest struct {
	Name           string
	Runtime        string
	CodeFilename   string
	CodeContent    []byte
	Handler        string
	HTTPAuthMode   string
	InvocationMode string
	IsPublic       *bool
	OpenAPISpec    string
	VariableScope  string
	Variables      []string
}

func (c *Client) CreateFunction(ctx context.Context, projectID string, in CreateFunctionRequest) (*Function, error) {
	fields := []MultipartField{
		{Name: "name", Value: in.Name},
		{Name: "runtime", Value: in.Runtime},
	}
	if in.Handler != "" {
		fields = append(fields, MultipartField{Name: "handler", Value: in.Handler})
	}
	if in.HTTPAuthMode != "" {
		fields = append(fields, MultipartField{Name: "http_auth_mode", Value: in.HTTPAuthMode})
	}
	if in.InvocationMode != "" {
		fields = append(fields, MultipartField{Name: "invocation_mode", Value: in.InvocationMode})
	}
	if in.IsPublic != nil {
		fields = append(fields, MultipartField{Name: "is_public", Value: boolString(*in.IsPublic)})
	}
	if in.OpenAPISpec != "" {
		fields = append(fields, MultipartField{Name: "openapi_spec", Value: in.OpenAPISpec})
	}
	if in.VariableScope != "" {
		fields = append(fields, MultipartField{Name: "variable_scope", Value: in.VariableScope})
	}
	if in.Variables != nil {
		b, err := json.Marshal(in.Variables)
		if err != nil {
			return nil, err
		}
		fields = append(fields, MultipartField{Name: "variables", Value: string(b)})
	}

	var out Function
	file := MultipartFile{FieldName: "code", Filename: in.CodeFilename, Content: in.CodeContent}
	if err := c.MultipartRequest(ctx, "POST", "/projects/"+EncodePathSegment(projectID)+"/functions", fields, file, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetFunction(ctx context.Context, projectID, functionID string) (*Function, error) {
	var out Function
	path := "/projects/" + EncodePathSegment(projectID) + "/functions/" + EncodePathSegment(functionID)
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateFunctionRequest covers the metadata PATCH can change. Code,
// handler, runtime, and variable scoping are immutable after creation.
type UpdateFunctionRequest struct {
	HTTPAuthMode   *string `json:"http_auth_mode,omitempty"`
	InvocationMode *string `json:"invocation_mode,omitempty"`
	IsPublic       *bool   `json:"is_public,omitempty"`
	OpenAPISpec    *string `json:"openapi_spec,omitempty"`
}

func (c *Client) UpdateFunction(ctx context.Context, projectID, functionID string, in UpdateFunctionRequest) (*Function, error) {
	var out Function
	path := "/projects/" + EncodePathSegment(projectID) + "/functions/" + EncodePathSegment(functionID)
	if err := c.Request(ctx, "PATCH", path, nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteFunction(ctx context.Context, projectID, functionID string) error {
	path := "/projects/" + EncodePathSegment(projectID) + "/functions/" + EncodePathSegment(functionID)
	return c.Request(ctx, "DELETE", path, nil, nil, nil)
}

func boolString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
