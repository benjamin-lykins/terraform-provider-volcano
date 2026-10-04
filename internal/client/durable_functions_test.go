package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestDurableFunctionLifecycle(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/projects/p1/durable-functions":
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				t.Fatalf("parsing multipart form: %v", err)
			}
			_ = json.NewEncoder(w).Encode(DurableFunction{ID: "df1", ProjectID: "p1", Name: r.FormValue("name"), Runtime: r.FormValue("runtime"), Durable: true, Status: "provisioning"})
		case r.Method == http.MethodGet && r.URL.Path == "/projects/p1/durable-functions/df1":
			_ = json.NewEncoder(w).Encode(DurableFunction{ID: "df1", ProjectID: "p1", Name: "workflow", Runtime: "python3.13", Durable: true, Status: "active"})
		case r.Method == http.MethodDelete && r.URL.Path == "/projects/p1/durable-functions/df1":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	created, err := c.CreateDurableFunction(context.Background(), "p1", CreateDurableFunctionRequest{
		Name: "workflow", Runtime: "python3.13", CodeFilename: "code.zip", CodeContent: []byte("PK"),
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := c.GetDurableFunction(context.Background(), "p1", created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != "active" {
		t.Fatalf("unexpected durable function: %+v", got)
	}

	if err := c.DeleteDurableFunction(context.Background(), "p1", "df1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
