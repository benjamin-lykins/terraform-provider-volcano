package client

import (
	"context"
	"encoding/json"
)

// DurableFunction is a long-running function that checkpoints its
// progress and resumes from the last checkpoint. It has no PATCH: once
// created, every attribute (including code) is immutable.
type DurableFunction struct {
	ID                  string   `json:"id"`
	ProjectID           string   `json:"project_id"`
	Name                string   `json:"name"`
	Handler             string   `json:"handler,omitempty"`
	Runtime             string   `json:"runtime"`
	Status              string   `json:"status,omitempty"`
	IsPublic            bool     `json:"is_public"`
	Durable             bool     `json:"durable,omitempty"`
	DeployedRegions     []string `json:"deployed_regions,omitempty"`
	CurrentDeploymentID string   `json:"current_deployment_id,omitempty"`
	CreatedAt           string   `json:"created_at,omitempty"`
	UpdatedAt           string   `json:"updated_at,omitempty"`
}

type CreateDurableFunctionRequest struct {
	Name          string
	Runtime       string
	CodeFilename  string
	CodeContent   []byte
	Handler       string
	IsPublic      *bool
	VariableScope string
	Variables     []string
}

func (c *Client) CreateDurableFunction(ctx context.Context, projectID string, in CreateDurableFunctionRequest) (*DurableFunction, error) {
	fields := []MultipartField{
		{Name: "name", Value: in.Name},
		{Name: "runtime", Value: in.Runtime},
	}
	if in.Handler != "" {
		fields = append(fields, MultipartField{Name: "handler", Value: in.Handler})
	}
	if in.IsPublic != nil {
		fields = append(fields, MultipartField{Name: "is_public", Value: boolString(*in.IsPublic)})
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

	var out DurableFunction
	file := MultipartFile{FieldName: "code", Filename: in.CodeFilename, Content: in.CodeContent}
	if err := c.MultipartRequest(ctx, "POST", "/projects/"+EncodePathSegment(projectID)+"/durable-functions", fields, file, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetDurableFunction(ctx context.Context, projectID, functionID string) (*DurableFunction, error) {
	var out DurableFunction
	path := "/projects/" + EncodePathSegment(projectID) + "/durable-functions/" + EncodePathSegment(functionID)
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteDurableFunction(ctx context.Context, projectID, functionID string) error {
	path := "/projects/" + EncodePathSegment(projectID) + "/durable-functions/" + EncodePathSegment(functionID)
	return c.Request(ctx, "DELETE", path, nil, nil, nil)
}
