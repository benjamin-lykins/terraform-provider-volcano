package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestVariableLifecycle(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/projects/p1/variables":
			_ = json.NewEncoder(w).Encode(Variable{ID: "v1", ProjectID: "p1", Name: "API_KEY", Value: "secret"})
		case r.Method == http.MethodPut && r.URL.Path == "/projects/p1/variables/API_KEY":
			var body UpdateVariableRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body.Value == nil || *body.Value != "rotated" {
				t.Fatalf("unexpected update body: %+v", body)
			}
			_ = json.NewEncoder(w).Encode(Variable{ID: "v1", ProjectID: "p1", Name: "API_KEY", Value: "rotated"})
		case r.Method == http.MethodDelete && r.URL.Path == "/projects/p1/variables/API_KEY":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	if _, err := c.CreateVariable(context.Background(), "p1", CreateVariableRequest{Name: "API_KEY", Value: "secret"}); err != nil {
		t.Fatalf("create: %v", err)
	}

	newValue := "rotated"
	updated, err := c.UpdateVariable(context.Background(), "p1", "API_KEY", UpdateVariableRequest{Value: &newValue})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Value != "rotated" {
		t.Fatalf("unexpected variable after update: %+v", updated)
	}

	if err := c.DeleteVariable(context.Background(), "p1", "API_KEY"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
