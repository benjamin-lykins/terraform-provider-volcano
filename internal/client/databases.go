package client

import "context"

// Database is a provisioned PostgreSQL database within a project.
type Database struct {
	ID               string `json:"id"`
	ProjectID        string `json:"project_id"`
	Name             string `json:"name"`
	Region           string `json:"region"`
	PGVersion        string `json:"pg_version"`
	DatabaseType     string `json:"database_type"`
	Status           string `json:"status"`
	ConnectionString string `json:"connection_string,omitempty"`
	StorageBytes     int64  `json:"storage_bytes,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
	UpdatedAt        string `json:"updated_at,omitempty"`
}

type CreateDatabaseRequest struct {
	Name         string `json:"name"`
	Region       string `json:"region"`
	PGVersion    string `json:"pg_version"`
	DatabaseType string `json:"database_type,omitempty"`
}

type UpdateDatabaseTypeRequest struct {
	DatabaseType string `json:"database_type"`
}

func (c *Client) CreateDatabase(ctx context.Context, projectID string, in CreateDatabaseRequest) (*Database, error) {
	var out Database
	if err := c.Request(ctx, "POST", "/projects/"+EncodePathSegment(projectID)+"/databases", nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetDatabase(ctx context.Context, projectID, name string) (*Database, error) {
	var out Database
	if err := c.Request(ctx, "GET", "/projects/"+EncodePathSegment(projectID)+"/databases/"+EncodePathSegment(name), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateDatabaseType(ctx context.Context, projectID, name string, in UpdateDatabaseTypeRequest) (*Database, error) {
	var out Database
	if err := c.Request(ctx, "PATCH", "/projects/"+EncodePathSegment(projectID)+"/databases/"+EncodePathSegment(name)+"/type", nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteDatabase(ctx context.Context, projectID, name string) error {
	return c.Request(ctx, "DELETE", "/projects/"+EncodePathSegment(projectID)+"/databases/"+EncodePathSegment(name), nil, nil, nil)
}

// Region describes a hosting region offered by the platform.
type Region struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c *Client) ListDatabaseRegions(ctx context.Context) ([]Region, error) {
	var out []Region
	if err := c.Request(ctx, "GET", "/databases/regions", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// PostgresVersion describes a PostgreSQL major version offered by the platform.
type PostgresVersion struct {
	Version    string `json:"version"`
	Name       string `json:"name"`
	Default    bool   `json:"default"`
	Deprecated bool   `json:"deprecated"`
}

func (c *Client) ListPostgresVersions(ctx context.Context) ([]PostgresVersion, error) {
	var out []PostgresVersion
	if err := c.Request(ctx, "GET", "/databases/postgres-versions", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
