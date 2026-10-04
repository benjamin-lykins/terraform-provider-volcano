package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestCreateProjectAccessToken(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/p1/access-tokens" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(CreatedProjectAccessToken{
			ProjectAccessToken: ProjectAccessToken{ID: "tok1", ProjectID: "p1", Name: "ci", Scope: "full", Status: "active"},
			Token:              "pt-secret-value",
		})
	})
	defer closeFn()

	got, err := c.CreateProjectAccessToken(context.Background(), "p1", CreateProjectAccessTokenRequest{Name: "ci", Scope: "full"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Token != "pt-secret-value" || got.ID != "tok1" {
		t.Fatalf("unexpected token: %+v", got)
	}
}

func TestDeleteProjectAccessTokenNotFound(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	})
	defer closeFn()

	err := c.DeleteProjectAccessToken(context.Background(), "p1", "missing")
	if !IsNotFound(err) {
		t.Fatalf("expected not found error, got: %v", err)
	}
}
