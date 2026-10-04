package client

import "context"

// ProjectSourceExportStatus is the current status of a project's
// source-export/git-takeover transition.
type ProjectSourceExportStatus struct {
	Mode                string `json:"mode,omitempty"`
	ExportedAt          string `json:"exported_at,omitempty"`
	TransitionStartedAt string `json:"transition_started_at,omitempty"`
	HandedOverAt        string `json:"handed_over_at,omitempty"`
}

// ProjectSourceExportResult is returned once, when the export is created.
type ProjectSourceExportResult struct {
	RepoFullName string   `json:"repo_full_name,omitempty"`
	Branch       string   `json:"branch,omitempty"`
	CommitSHA    string   `json:"commit_sha,omitempty"`
	FileCount    int64    `json:"file_count,omitempty"`
	Skipped      []string `json:"skipped,omitempty"`
	Omitted      []string `json:"omitted,omitempty"`
}

type CreateProjectSourceExportRequest struct {
	ProductionBranch string `json:"production_branch"`
}

func (c *Client) CreateProjectSourceExport(ctx context.Context, projectID string, in CreateProjectSourceExportRequest) (*ProjectSourceExportResult, error) {
	var out ProjectSourceExportResult
	path := "/projects/" + EncodePathSegment(projectID) + "/source-export"
	if err := c.Request(ctx, "POST", path, nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetProjectSourceExportStatus(ctx context.Context, projectID string) (*ProjectSourceExportStatus, error) {
	var out ProjectSourceExportStatus
	path := "/projects/" + EncodePathSegment(projectID) + "/source-export"
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteProjectSourceExport(ctx context.Context, projectID string) error {
	path := "/projects/" + EncodePathSegment(projectID) + "/source-export"
	return c.Request(ctx, "DELETE", path, nil, nil, nil)
}
