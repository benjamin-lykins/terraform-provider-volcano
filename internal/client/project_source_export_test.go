package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestProjectSourceExportLifecycle(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/projects/p1/source-export":
			_ = json.NewEncoder(w).Encode(ProjectSourceExportResult{RepoFullName: "acme/app", Branch: "main", CommitSHA: "abc123", FileCount: 10})
		case r.Method == http.MethodGet && r.URL.Path == "/projects/p1/source-export":
			_ = json.NewEncoder(w).Encode(ProjectSourceExportStatus{Mode: "handed_over", ExportedAt: "2026-01-01T00:00:00Z"})
		case r.Method == http.MethodDelete && r.URL.Path == "/projects/p1/source-export":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	result, err := c.CreateProjectSourceExport(context.Background(), "p1", CreateProjectSourceExportRequest{ProductionBranch: "main"})
	if err != nil || result.FileCount != 10 {
		t.Fatalf("unexpected result: %+v, err: %v", result, err)
	}

	status, err := c.GetProjectSourceExportStatus(context.Background(), "p1")
	if err != nil || status.Mode != "handed_over" {
		t.Fatalf("unexpected status: %+v, err: %v", status, err)
	}

	if err := c.DeleteProjectSourceExport(context.Background(), "p1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
