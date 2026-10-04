package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestListUserImportConnectionsAndSources(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user/imports/connections":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"connections": []UserImportConnection{{ID: "ic1", Provider: "other-platform"}},
			})
		case "/imports/other-platform/sources":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"sources": []ImportSource{{ID: "src1", Name: "my-app", Framework: "nextjs"}},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})
	defer closeFn()

	connections, err := c.ListUserImportConnections(context.Background())
	if err != nil || len(connections) != 1 || connections[0].ID != "ic1" {
		t.Fatalf("unexpected connections: %+v, err: %v", connections, err)
	}

	sources, err := c.ListImportSources(context.Background(), "other-platform")
	if err != nil || len(sources) != 1 || sources[0].Name != "my-app" {
		t.Fatalf("unexpected sources: %+v, err: %v", sources, err)
	}
}
