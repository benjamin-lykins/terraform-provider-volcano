package client

import (
	"context"
	"net/url"
)

// StorageObject is a single file stored in a bucket.
type StorageObject struct {
	ID         string         `json:"id"`
	BucketID   string         `json:"bucket_id"`
	BucketName string         `json:"bucket_name,omitempty"`
	Name       string         `json:"name"`
	MimeType   string         `json:"mime_type,omitempty"`
	Size       int64          `json:"size,omitempty"`
	ETag       string         `json:"etag,omitempty"`
	IsPublic   bool           `json:"is_public"`
	PublicURL  string         `json:"public_url,omitempty"`
	OwnerID    string         `json:"owner_id,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	CreatedAt  string         `json:"created_at,omitempty"`
	UpdatedAt  string         `json:"updated_at,omitempty"`
}

// UploadStorageObject uploads a complete, non-chunked object in one call
// (the simple multipart/form-data path; the API's resumable session
// headers for very large files are not used by this client).
func (c *Client) UploadStorageObject(ctx context.Context, bucketName, objectPath, filename string, content []byte) (*StorageObject, error) {
	var out StorageObject
	path := "/storage/" + EncodePathSegment(bucketName) + "/" + EncodeObjectPath(objectPath)
	file := MultipartFile{FieldName: "file", Filename: filename, Content: content}
	if err := c.MultipartRequest(ctx, "POST", path, nil, file, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteStorageObject(ctx context.Context, bucketName, objectPath string) error {
	path := "/storage/" + EncodePathSegment(bucketName) + "/" + EncodeObjectPath(objectPath)
	return c.Request(ctx, "DELETE", path, nil, nil, nil)
}

type setStorageObjectVisibilityRequest struct {
	IsPublic bool `json:"is_public"`
}

func (c *Client) SetStorageObjectVisibility(ctx context.Context, bucketName, objectPath string, isPublic bool) (*StorageObject, error) {
	var out StorageObject
	path := "/storage/" + EncodePathSegment(bucketName) + "/" + EncodeObjectPath(objectPath) + "/visibility"
	if err := c.Request(ctx, "PATCH", path, nil, setStorageObjectVisibilityRequest{IsPublic: isPublic}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListStorageObjectsOptions filters the project-wide object listing used
// both by the objects data source and (with Search set to the object
// path) to read a single object's metadata, since there is no per-object
// GET endpoint.
type ListStorageObjectsOptions struct {
	Search  string
	OwnerID string
}

func (c *Client) ListStorageObjects(ctx context.Context, projectID string, opts ListStorageObjectsOptions) ([]StorageObject, error) {
	q := url.Values{}
	if opts.Search != "" {
		q.Set("search", opts.Search)
	}
	if opts.OwnerID != "" {
		q.Set("owner_id", opts.OwnerID)
	}
	var out Page[StorageObject]
	path := "/projects/" + EncodePathSegment(projectID) + "/storage/objects"
	if err := c.Request(ctx, "GET", path, q, nil, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

// GetStorageObject finds a single object by exact bucket+path match,
// since the data-plane GET for an object returns its bytes, not JSON
// metadata. It paginates the search results (narrowed server-side by the
// path as a search term) until it finds an exact match or runs out of
// pages.
func (c *Client) GetStorageObject(ctx context.Context, projectID, bucketName, objectPath string) (*StorageObject, error) {
	objects, err := c.ListStorageObjects(ctx, projectID, ListStorageObjectsOptions{Search: objectPath})
	if err != nil {
		return nil, err
	}
	for _, o := range objects {
		if o.BucketName == bucketName && o.Name == objectPath {
			return &o, nil
		}
	}
	return nil, &APIError{StatusCode: 404, Message: "object not found"}
}
