package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestCreateAndSetDefaultAnonKey(t *testing.T) {
	setDefaultCalled := false
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/projects/p1/anon-keys":
			_ = json.NewEncoder(w).Encode(AnonKey{ID: "k1", ProjectID: "p1", Name: "frontend", KeyValue: "jwt-value"})
		case r.Method == http.MethodPost && r.URL.Path == "/projects/p1/anon-keys/k1/set-default":
			setDefaultCalled = true
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	key, err := c.CreateAnonKey(context.Background(), "p1", CreateAnonKeyRequest{Name: "frontend"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key.KeyValue != "jwt-value" {
		t.Fatalf("unexpected key: %+v", key)
	}

	if err := c.SetDefaultAnonKey(context.Background(), "p1", key.ID); err != nil {
		t.Fatalf("unexpected error setting default: %v", err)
	}
	if !setDefaultCalled {
		t.Fatal("expected set-default request to be made")
	}
}
