package client

import "context"

// StorageBucketPolicy is a row-level access rule for a bucket, evaluated
// at request time for a specific operation.
type StorageBucketPolicy struct {
	ID         string `json:"id"`
	BucketID   string `json:"bucket_id"`
	Name       string `json:"name"`
	Operation  string `json:"operation"`
	Definition string `json:"definition"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

type CreateStorageBucketPolicyRequest struct {
	Name       string `json:"name"`
	Operation  string `json:"operation"`
	Definition string `json:"definition"`
}

func (c *Client) CreateStorageBucketPolicy(ctx context.Context, projectID, bucketName string, in CreateStorageBucketPolicyRequest) (*StorageBucketPolicy, error) {
	var out StorageBucketPolicy
	path := "/projects/" + EncodePathSegment(projectID) + "/storage/buckets/" + EncodePathSegment(bucketName) + "/policies"
	if err := c.Request(ctx, "POST", path, nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListStorageBucketPolicies is used for Read, since there is no
// single-policy GET endpoint.
func (c *Client) ListStorageBucketPolicies(ctx context.Context, projectID, bucketName string) ([]StorageBucketPolicy, error) {
	var out []StorageBucketPolicy
	path := "/projects/" + EncodePathSegment(projectID) + "/storage/buckets/" + EncodePathSegment(bucketName) + "/policies"
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) DeleteStorageBucketPolicy(ctx context.Context, projectID, bucketName, policyID string) error {
	path := "/projects/" + EncodePathSegment(projectID) + "/storage/buckets/" + EncodePathSegment(bucketName) + "/policies/" + EncodePathSegment(policyID)
	return c.Request(ctx, "DELETE", path, nil, nil, nil)
}
