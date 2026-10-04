package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestDatabaseBranchLifecycle(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/projects/p1/databases/main/branches":
			_ = json.NewEncoder(w).Encode(DatabaseBranch{ID: "b1", ProjectID: "p1", DatabaseID: "d1", Name: "feature-x", TTLSeconds: 3600})
		case r.Method == http.MethodPatch && r.URL.Path == "/projects/p1/databases/main/branches/feature-x":
			var body UpdateDatabaseBranchRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			_ = json.NewEncoder(w).Encode(DatabaseBranch{ID: "b1", ProjectID: "p1", DatabaseID: "d1", Name: "feature-x", TTLSeconds: body.TTLSeconds})
		case r.Method == http.MethodDelete && r.URL.Path == "/projects/p1/databases/main/branches/feature-x":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	ttl := int64(3600)
	branch, err := c.CreateDatabaseBranch(context.Background(), "p1", "main", CreateDatabaseBranchRequest{Name: "feature-x", TTLSeconds: &ttl})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	updated, err := c.UpdateDatabaseBranch(context.Background(), "p1", "main", branch.Name, UpdateDatabaseBranchRequest{TTLSeconds: 7200})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.TTLSeconds != 7200 {
		t.Fatalf("unexpected branch after update: %+v", updated)
	}

	if err := c.DeleteDatabaseBranch(context.Background(), "p1", "main", "feature-x"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
