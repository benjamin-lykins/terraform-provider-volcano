package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestStorageBucketLifecycle(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/projects/p1/storage/buckets":
			_ = json.NewEncoder(w).Encode(StorageBucket{ID: "b1", ProjectID: "p1", Name: "avatars"})
		case r.Method == http.MethodPatch && r.URL.Path == "/projects/p1/storage/buckets/avatars":
			var body UpdateStorageBucketRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			_ = json.NewEncoder(w).Encode(StorageBucket{ID: "b1", ProjectID: "p1", Name: "avatars", FileSizeLimit: body.FileSizeLimit})
		case r.Method == http.MethodDelete && r.URL.Path == "/projects/p1/storage/buckets/avatars":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	bucket, err := c.CreateStorageBucket(context.Background(), "p1", CreateStorageBucketRequest{Name: "avatars"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	limit := int64(1 << 20)
	updated, err := c.UpdateStorageBucket(context.Background(), "p1", bucket.Name, UpdateStorageBucketRequest{FileSizeLimit: &limit})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.FileSizeLimit == nil || *updated.FileSizeLimit != limit {
		t.Fatalf("unexpected bucket after update: %+v", updated)
	}

	if err := c.DeleteStorageBucket(context.Background(), "p1", "avatars"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestStorageBucketPolicyLifecycle(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/projects/p1/storage/buckets/avatars/policies":
			_ = json.NewEncoder(w).Encode(StorageBucketPolicy{ID: "pol1", BucketID: "b1", Name: "owner-only", Operation: "SELECT", Definition: "auth.uid() = owner_id"})
		case r.Method == http.MethodGet && r.URL.Path == "/projects/p1/storage/buckets/avatars/policies":
			_ = json.NewEncoder(w).Encode([]StorageBucketPolicy{{ID: "pol1", BucketID: "b1", Name: "owner-only", Operation: "SELECT", Definition: "auth.uid() = owner_id"}})
		case r.Method == http.MethodDelete && r.URL.Path == "/projects/p1/storage/buckets/avatars/policies/pol1":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	policy, err := c.CreateStorageBucketPolicy(context.Background(), "p1", "avatars", CreateStorageBucketPolicyRequest{
		Name: "owner-only", Operation: "SELECT", Definition: "auth.uid() = owner_id",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	policies, err := c.ListStorageBucketPolicies(context.Background(), "p1", "avatars")
	if err != nil || len(policies) != 1 {
		t.Fatalf("unexpected policies: %+v, err: %v", policies, err)
	}

	if err := c.DeleteStorageBucketPolicy(context.Background(), "p1", "avatars", policy.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
