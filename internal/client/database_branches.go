package client

import "context"

// DatabaseBranch is a short-lived fork of a database, used for
// development/testing and automatically expiring after ttl_seconds.
type DatabaseBranch struct {
	ID               string `json:"id"`
	ProjectID        string `json:"project_id"`
	DatabaseID       string `json:"database_id"`
	Name             string `json:"name"`
	Status           string `json:"status,omitempty"`
	ConnectionString string `json:"connection_string,omitempty"`
	TTLSeconds       int64  `json:"ttl_seconds,omitempty"`
	ExpiresAt        string `json:"expires_at,omitempty"`
	StorageBytes     int64  `json:"storage_bytes,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
	UpdatedAt        string `json:"updated_at,omitempty"`
}

type CreateDatabaseBranchRequest struct {
	Name       string `json:"name"`
	TTLSeconds *int64 `json:"ttl_seconds,omitempty"`
}

type UpdateDatabaseBranchRequest struct {
	TTLSeconds int64 `json:"ttl_seconds"`
}

func (c *Client) CreateDatabaseBranch(ctx context.Context, projectID, databaseName string, in CreateDatabaseBranchRequest) (*DatabaseBranch, error) {
	var out DatabaseBranch
	if err := c.Request(ctx, "POST", "/projects/"+EncodePathSegment(projectID)+"/databases/"+EncodePathSegment(databaseName)+"/branches", nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetDatabaseBranch(ctx context.Context, projectID, databaseName, branchName string) (*DatabaseBranch, error) {
	var out DatabaseBranch
	path := "/projects/" + EncodePathSegment(projectID) + "/databases/" + EncodePathSegment(databaseName) + "/branches/" + EncodePathSegment(branchName)
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateDatabaseBranch(ctx context.Context, projectID, databaseName, branchName string, in UpdateDatabaseBranchRequest) (*DatabaseBranch, error) {
	var out DatabaseBranch
	path := "/projects/" + EncodePathSegment(projectID) + "/databases/" + EncodePathSegment(databaseName) + "/branches/" + EncodePathSegment(branchName)
	if err := c.Request(ctx, "PATCH", path, nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteDatabaseBranch(ctx context.Context, projectID, databaseName, branchName string) error {
	path := "/projects/" + EncodePathSegment(projectID) + "/databases/" + EncodePathSegment(databaseName) + "/branches/" + EncodePathSegment(branchName)
	return c.Request(ctx, "DELETE", path, nil, nil, nil)
}
