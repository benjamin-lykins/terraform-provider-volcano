package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestListUserGitConnectionsInstallationsAndRepositories(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user/git/connections":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"connections": []UserGitConnection{{ID: "conn1", Provider: "github", ProviderLogin: "octocat"}},
			})
		case "/user/git/connections/conn1/installations":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"installations": []UserGitInstallation{{ID: 42, AccountLogin: "acme"}},
			})
		case "/user/git/connections/conn1/installations/42/repositories":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"repositories": []UserGitRepository{{ID: 1, FullName: "acme/app"}},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})
	defer closeFn()

	connections, err := c.ListUserGitConnections(context.Background())
	if err != nil || len(connections) != 1 || connections[0].ID != "conn1" {
		t.Fatalf("unexpected connections: %+v, err: %v", connections, err)
	}

	installations, err := c.ListUserGitInstallations(context.Background(), "conn1")
	if err != nil || len(installations) != 1 || installations[0].ID != 42 {
		t.Fatalf("unexpected installations: %+v, err: %v", installations, err)
	}

	repos, err := c.ListUserGitInstallationRepositories(context.Background(), "conn1", 42)
	if err != nil || len(repos) != 1 || repos[0].FullName != "acme/app" {
		t.Fatalf("unexpected repos: %+v, err: %v", repos, err)
	}
}
