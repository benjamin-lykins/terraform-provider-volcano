package client

import "context"

// StorageBucket is a named container for files within a project.
type StorageBucket struct {
	ID               string   `json:"id"`
	ProjectID        string   `json:"project_id"`
	Name             string   `json:"name"`
	FileSizeLimit    *int64   `json:"file_size_limit,omitempty"`
	AllowedMimeTypes []string `json:"allowed_mime_types,omitempty"`
	CreatedAt        string   `json:"created_at,omitempty"`
	UpdatedAt        string   `json:"updated_at,omitempty"`
}

type CreateStorageBucketRequest struct {
	Name             string   `json:"name"`
	FileSizeLimit    *int64   `json:"file_size_limit,omitempty"`
	AllowedMimeTypes []string `json:"allowed_mime_types,omitempty"`
}

type UpdateStorageBucketRequest struct {
	FileSizeLimit    *int64   `json:"file_size_limit,omitempty"`
	AllowedMimeTypes []string `json:"allowed_mime_types,omitempty"`
}

func (c *Client) CreateStorageBucket(ctx context.Context, projectID string, in CreateStorageBucketRequest) (*StorageBucket, error) {
	var out StorageBucket
	path := "/projects/" + EncodePathSegment(projectID) + "/storage/buckets"
	if err := c.Request(ctx, "POST", path, nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetStorageBucket(ctx context.Context, projectID, name string) (*StorageBucket, error) {
	var out StorageBucket
	path := "/projects/" + EncodePathSegment(projectID) + "/storage/buckets/" + EncodePathSegment(name)
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateStorageBucket(ctx context.Context, projectID, name string, in UpdateStorageBucketRequest) (*StorageBucket, error) {
	var out StorageBucket
	path := "/projects/" + EncodePathSegment(projectID) + "/storage/buckets/" + EncodePathSegment(name)
	if err := c.Request(ctx, "PATCH", path, nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteStorageBucket(ctx context.Context, projectID, name string) error {
	path := "/projects/" + EncodePathSegment(projectID) + "/storage/buckets/" + EncodePathSegment(name)
	return c.Request(ctx, "DELETE", path, nil, nil, nil)
}
