package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestProjectAuthConfigGetUpdate(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/projects/p1/auth/config":
			_ = json.NewEncoder(w).Encode(ProjectAuthConfig{ProjectID: "p1", EnableSignup: true, MinPasswordLength: 8})
		case r.Method == http.MethodPut && r.URL.Path == "/projects/p1/auth/config":
			var body UpdateProjectAuthConfigRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body.MinPasswordLength == nil || *body.MinPasswordLength != 12 {
				t.Fatalf("unexpected body: %+v", body)
			}
			_ = json.NewEncoder(w).Encode(ProjectAuthConfig{ProjectID: "p1", EnableSignup: true, MinPasswordLength: 12})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	cfg, err := c.GetProjectAuthConfig(context.Background(), "p1")
	if err != nil || cfg.MinPasswordLength != 8 {
		t.Fatalf("unexpected config: %+v, err: %v", cfg, err)
	}

	minLen := int64(12)
	updated, err := c.UpdateProjectAuthConfig(context.Background(), "p1", UpdateProjectAuthConfigRequest{MinPasswordLength: &minLen})
	if err != nil || updated.MinPasswordLength != 12 {
		t.Fatalf("unexpected update: %+v, err: %v", updated, err)
	}
}
