package client

import "context"

type ProjectGitDeploySettings struct {
	AutoDeployEnabled bool   `json:"auto_deploy_enabled"`
	DeployFunctions   bool   `json:"deploy_functions"`
	FrontendName      string `json:"frontend_name,omitempty"`
	FrontendAppRoot   string `json:"frontend_app_root,omitempty"`
	UpdatedAt         string `json:"updated_at,omitempty"`
}

func (c *Client) GetProjectGitDeploySettings(ctx context.Context, projectID string) (*ProjectGitDeploySettings, error) {
	var out ProjectGitDeploySettings
	path := "/projects/" + EncodePathSegment(projectID) + "/git-deploy-settings"
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) SetProjectGitDeploySettings(ctx context.Context, projectID string, in ProjectGitDeploySettings) (*ProjectGitDeploySettings, error) {
	var out ProjectGitDeploySettings
	path := "/projects/" + EncodePathSegment(projectID) + "/git-deploy-settings"
	if err := c.Request(ctx, "PUT", path, nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
