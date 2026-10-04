package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestProjectAuthMethodsGetUpdate(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/projects/p1/auth/methods":
			_ = json.NewEncoder(w).Encode(ProjectAuthMethods{
				Anonymous:        AuthMethodStatus{Enabled: false},
				EmailPassword:    AuthMethodStatus{Enabled: true},
				AvailableMethods: []string{"email_password", "oauth_google"},
			})
		case r.Method == http.MethodPut && r.URL.Path == "/projects/p1/auth/methods":
			var body UpdateProjectAuthMethodsRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			if len(body.OAuthProviders) != 1 || body.OAuthProviders[0].Provider != "google" {
				t.Fatalf("unexpected body: %+v", body)
			}
			_ = json.NewEncoder(w).Encode(ProjectAuthMethods{
				Anonymous:      AuthMethodStatus{Enabled: true},
				EmailPassword:  AuthMethodStatus{Enabled: true},
				OAuthProviders: []OAuthProviderMethodStatus{{Provider: "google", Enabled: true}},
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	methods, err := c.GetProjectAuthMethods(context.Background(), "p1")
	if err != nil || methods.EmailPassword.Enabled != true {
		t.Fatalf("unexpected methods: %+v, err: %v", methods, err)
	}

	anon := true
	updated, err := c.UpdateProjectAuthMethods(context.Background(), "p1", UpdateProjectAuthMethodsRequest{
		EnableAnonymous: &anon,
		OAuthProviders:  []OAuthProviderToggle{{Provider: "google", Enabled: true}},
	})
	if err != nil || !updated.Anonymous.Enabled {
		t.Fatalf("unexpected update: %+v, err: %v", updated, err)
	}
}
