package client

import (
	"context"
	"net/url"
)

// Project mirrors the project object returned by the Volcano API.
type Project struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Status          string   `json:"status"`
	Plan            string   `json:"plan,omitempty"`
	AllRegions      bool     `json:"all_regions"`
	SelectedRegions []string `json:"selected_regions,omitempty"`
	LogoURL         string   `json:"logo_url,omitempty"`
	CreatedAt       string   `json:"created_at,omitempty"`
	UpdatedAt       string   `json:"updated_at,omitempty"`
	LastInvokedAt   string   `json:"last_invoked_at,omitempty"`
}

type CreateProjectRequest struct {
	Name            string   `json:"name"`
	AllRegions      *bool    `json:"all_regions,omitempty"`
	SelectedRegions []string `json:"selected_regions,omitempty"`
	TemplateID      string   `json:"template_id,omitempty"`
	InitialPrompt   string   `json:"initialPrompt,omitempty"`
}

type UpdateProjectRequest struct {
	Name            *string  `json:"name,omitempty"`
	AllRegions      *bool    `json:"all_regions,omitempty"`
	SelectedRegions []string `json:"selected_regions,omitempty"`
}

type listProjectsResponse struct {
	Page[Project]
}

func (c *Client) CreateProject(ctx context.Context, in CreateProjectRequest) (*Project, error) {
	var out Project
	if err := c.Request(ctx, "POST", "/projects", nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetProject(ctx context.Context, id string) (*Project, error) {
	var out Project
	if err := c.Request(ctx, "GET", "/projects/"+EncodePathSegment(id), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateProject(ctx context.Context, id string, in UpdateProjectRequest) (*Project, error) {
	var out Project
	if err := c.Request(ctx, "PATCH", "/projects/"+EncodePathSegment(id), nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteProject(ctx context.Context, id string) error {
	return c.Request(ctx, "DELETE", "/projects/"+EncodePathSegment(id), nil, nil, nil)
}

// ListProjectsOptions controls pagination/filtering for ListProjects.
type ListProjectsOptions struct {
	Search string
}

func (c *Client) ListProjects(ctx context.Context, opts ListProjectsOptions) ([]Project, error) {
	q := url.Values{}
	if opts.Search != "" {
		q.Set("search", opts.Search)
	}
	var out listProjectsResponse
	if err := c.Request(ctx, "GET", "/projects", q, nil, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}
