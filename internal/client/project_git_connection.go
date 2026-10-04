package client

import "context"

// ProjectGitConnection ties a project to a repository reachable through
// one of the account's git connections (see ListUserGitConnections).
type ProjectGitConnection struct {
	RepoInstallationID int64  `json:"repo_installation_id"`
	RepoID             int64  `json:"repo_id"`
	RepoFullName       string `json:"repo_full_name"`
	RootDirectory      string `json:"root_directory,omitempty"`
	ProductionBranch   string `json:"production_branch,omitempty"`
	UpdatedAt          string `json:"updated_at,omitempty"`
}

type SetProjectGitConnectionRequest struct {
	ConnectionID     string `json:"connection_id"`
	InstallationID   int64  `json:"installation_id"`
	RepositoryID     int64  `json:"repository_id,omitempty"`
	RepoFullName     string `json:"repo_full_name,omitempty"`
	RootDirectory    string `json:"root_directory,omitempty"`
	ProductionBranch string `json:"production_branch,omitempty"`
}

func (c *Client) GetProjectGitConnection(ctx context.Context, projectID string) (*ProjectGitConnection, error) {
	var out ProjectGitConnection
	path := "/projects/" + EncodePathSegment(projectID) + "/git-connection"
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) SetProjectGitConnection(ctx context.Context, projectID string, in SetProjectGitConnectionRequest) (*ProjectGitConnection, error) {
	var out ProjectGitConnection
	path := "/projects/" + EncodePathSegment(projectID) + "/git-connection"
	if err := c.Request(ctx, "PUT", path, nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteProjectGitConnection(ctx context.Context, projectID string) error {
	path := "/projects/" + EncodePathSegment(projectID) + "/git-connection"
	return c.Request(ctx, "DELETE", path, nil, nil, nil)
}

func (c *Client) SetProjectGitProductionBranch(ctx context.Context, projectID, branch string) (*ProjectGitConnection, error) {
	var out ProjectGitConnection
	path := "/projects/" + EncodePathSegment(projectID) + "/git-connection/production-branch"
	if err := c.Request(ctx, "PUT", path, nil, map[string]string{"production_branch": branch}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
