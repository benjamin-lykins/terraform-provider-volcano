package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestServiceKeyLifecycle(t *testing.T) {
	deleted := false
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/projects/p1/service-keys":
			_ = json.NewEncoder(w).Encode(ServiceKey{ID: "sk1", ProjectID: "p1", Name: "jobs", KeyValue: "jwt", Permissions: []string{"*"}})
		case r.Method == http.MethodGet && r.URL.Path == "/projects/p1/service-keys/sk1":
			_ = json.NewEncoder(w).Encode(ServiceKey{ID: "sk1", ProjectID: "p1", Name: "jobs", KeyValue: "jwt", Permissions: []string{"*"}})
		case r.Method == http.MethodDelete && r.URL.Path == "/projects/p1/service-keys/sk1":
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	created, err := c.CreateServiceKey(context.Background(), "p1", CreateServiceKeyRequest{Name: "jobs"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := c.GetServiceKey(context.Background(), "p1", created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "jobs" {
		t.Fatalf("unexpected key: %+v", got)
	}

	if err := c.DeleteServiceKey(context.Background(), "p1", created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if !deleted {
		t.Fatal("expected delete request to be made")
	}
}
