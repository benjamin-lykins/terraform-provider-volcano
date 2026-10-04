package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetPasswordPolicy(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/password-policy" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(PasswordPolicy{EffectiveMinLength: 8, MaxLength: 128, RequireLowercase: true})
	})
	defer closeFn()

	policy, err := c.GetPasswordPolicy(context.Background())
	if err != nil || policy.EffectiveMinLength != 8 {
		t.Fatalf("unexpected policy: %+v, err: %v", policy, err)
	}
}
