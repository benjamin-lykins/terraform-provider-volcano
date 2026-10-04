package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestDatabaseLifecycle(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/projects/p1/databases":
			_ = json.NewEncoder(w).Encode(Database{ID: "d1", ProjectID: "p1", Name: "main", Region: "us-east-1", PGVersion: "16", DatabaseType: "volcano-db-xs", Status: "provisioning"})
		case r.Method == http.MethodPatch && r.URL.Path == "/projects/p1/databases/main/type":
			var body UpdateDatabaseTypeRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			_ = json.NewEncoder(w).Encode(Database{ID: "d1", ProjectID: "p1", Name: "main", Region: "us-east-1", PGVersion: "16", DatabaseType: body.DatabaseType, Status: "active"})
		case r.Method == http.MethodDelete && r.URL.Path == "/projects/p1/databases/main":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	db, err := c.CreateDatabase(context.Background(), "p1", CreateDatabaseRequest{Name: "main", Region: "us-east-1", PGVersion: "16"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	resized, err := c.UpdateDatabaseType(context.Background(), "p1", db.Name, UpdateDatabaseTypeRequest{DatabaseType: "volcano-db-m"})
	if err != nil {
		t.Fatalf("resize: %v", err)
	}
	if resized.DatabaseType != "volcano-db-m" {
		t.Fatalf("unexpected database after resize: %+v", resized)
	}

	if err := c.DeleteDatabase(context.Background(), "p1", "main"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestListDatabaseRegionsAndPostgresVersions(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/databases/regions":
			_ = json.NewEncoder(w).Encode([]Region{{ID: "us-east-1", Name: "US East (N. Virginia)"}})
		case "/databases/postgres-versions":
			_ = json.NewEncoder(w).Encode([]PostgresVersion{{Version: "16", Name: "PostgreSQL 16", Default: true}})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})
	defer closeFn()

	regions, err := c.ListDatabaseRegions(context.Background())
	if err != nil || len(regions) != 1 || regions[0].ID != "us-east-1" {
		t.Fatalf("unexpected regions: %+v, err: %v", regions, err)
	}

	versions, err := c.ListPostgresVersions(context.Background())
	if err != nil || len(versions) != 1 || !versions[0].Default {
		t.Fatalf("unexpected versions: %+v, err: %v", versions, err)
	}
}
