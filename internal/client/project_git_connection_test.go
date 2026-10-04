package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestProjectGitConnectionLifecycle(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/projects/p1/git-connection":
			var body SetProjectGitConnectionRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			_ = json.NewEncoder(w).Encode(ProjectGitConnection{RepoInstallationID: body.InstallationID, RepoID: body.RepositoryID, RepoFullName: body.RepoFullName, ProductionBranch: body.ProductionBranch})
		case r.Method == http.MethodPut && r.URL.Path == "/projects/p1/git-connection/production-branch":
			_ = json.NewEncoder(w).Encode(ProjectGitConnection{ProductionBranch: "release"})
		case r.Method == http.MethodGet && r.URL.Path == "/projects/p1/git-connection":
			_ = json.NewEncoder(w).Encode(ProjectGitConnection{RepoFullName: "acme/app", ProductionBranch: "main"})
		case r.Method == http.MethodDelete && r.URL.Path == "/projects/p1/git-connection":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	gc, err := c.SetProjectGitConnection(context.Background(), "p1", SetProjectGitConnectionRequest{
		ConnectionID: "conn1", InstallationID: 42, RepoFullName: "acme/app", ProductionBranch: "main",
	})
	if err != nil || gc.RepoFullName != "acme/app" {
		t.Fatalf("unexpected connection: %+v, err: %v", gc, err)
	}

	branchUpdated, err := c.SetProjectGitProductionBranch(context.Background(), "p1", "release")
	if err != nil || branchUpdated.ProductionBranch != "release" {
		t.Fatalf("unexpected branch update: %+v, err: %v", branchUpdated, err)
	}

	got, err := c.GetProjectGitConnection(context.Background(), "p1")
	if err != nil || got.RepoFullName != "acme/app" {
		t.Fatalf("unexpected get: %+v, err: %v", got, err)
	}

	if err := c.DeleteProjectGitConnection(context.Background(), "p1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
