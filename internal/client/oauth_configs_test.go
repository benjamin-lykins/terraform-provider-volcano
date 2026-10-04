package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestOAuthConfigLifecycle(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/projects/p1/oauth/configs":
			_ = json.NewEncoder(w).Encode(OAuthConfig{ID: "o1", Provider: "google", ClientID: "cid", ClientSecret: "secret", Enabled: true})
		case r.Method == http.MethodPut && r.URL.Path == "/projects/p1/oauth/configs/google":
			var body UpdateOAuthConfigRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			_ = json.NewEncoder(w).Encode(OAuthConfig{ID: "o1", Provider: "google", ClientID: body.ClientID, Enabled: *body.Enabled})
		case r.Method == http.MethodDelete && r.URL.Path == "/projects/p1/oauth/configs/google":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	cfg, err := c.CreateOAuthConfig(context.Background(), "p1", CreateOAuthConfigRequest{Provider: "google", ClientID: "cid", ClientSecret: "secret"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	enabled := false
	updated, err := c.UpdateOAuthConfig(context.Background(), "p1", cfg.Provider, UpdateOAuthConfigRequest{ClientID: "cid2", Enabled: &enabled})
	if err != nil || updated.Enabled || updated.ClientID != "cid2" {
		t.Fatalf("unexpected update: %+v, err: %v", updated, err)
	}

	if err := c.DeleteOAuthConfig(context.Background(), "p1", "google"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestListOAuthProviderInfo(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/p1/oauth/providers" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"providers": []OAuthProviderInfo{{ID: "google", Name: "Google", DefaultScopes: []string{"email", "profile"}}},
		})
	})
	defer closeFn()

	providers, err := c.ListOAuthProviderInfo(context.Background(), "p1")
	if err != nil || len(providers) != 1 || providers[0].ID != "google" {
		t.Fatalf("unexpected providers: %+v, err: %v", providers, err)
	}
}
