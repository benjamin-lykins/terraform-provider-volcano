package client

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestFunctionLifecycle(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/projects/p1/functions":
			if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
				t.Fatalf("expected multipart request, got content-type: %s", r.Header.Get("Content-Type"))
			}
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				t.Fatalf("parsing multipart form: %v", err)
			}
			if r.FormValue("name") != "my-fn" || r.FormValue("runtime") != "nodejs24.x" {
				t.Fatalf("unexpected form values: name=%s runtime=%s", r.FormValue("name"), r.FormValue("runtime"))
			}
			_ = json.NewEncoder(w).Encode(Function{ID: "f1", ProjectID: "p1", Name: "my-fn", Runtime: "nodejs24.x", Status: "provisioning"})
		case r.Method == http.MethodPatch && r.URL.Path == "/projects/p1/functions/f1":
			var body UpdateFunctionRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			_ = json.NewEncoder(w).Encode(Function{ID: "f1", ProjectID: "p1", Name: "my-fn", Runtime: "nodejs24.x", Status: "active", IsPublic: *body.IsPublic})
		case r.Method == http.MethodDelete && r.URL.Path == "/projects/p1/functions/f1":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	fn, err := c.CreateFunction(context.Background(), "p1", CreateFunctionRequest{
		Name: "my-fn", Runtime: "nodejs24.x", CodeFilename: "code.zip", CodeContent: []byte("PK\x03\x04"),
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	isPublic := true
	updated, err := c.UpdateFunction(context.Background(), "p1", fn.ID, UpdateFunctionRequest{IsPublic: &isPublic})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if !updated.IsPublic {
		t.Fatalf("unexpected function after update: %+v", updated)
	}

	if err := c.DeleteFunction(context.Background(), "p1", "f1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestListFunctionRegionsAndRuntimes(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/functions/regions":
			_ = json.NewEncoder(w).Encode([]FunctionRegion{{Code: "us-east-1", Label: "NA-East", Flag: "🇺🇸"}})
		case "/functions/runtimes":
			_ = json.NewEncoder(w).Encode(listFunctionRuntimesResponse{Runtimes: []FunctionRuntime{{Name: "nodejs24.x", Default: true}}})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})
	defer closeFn()

	regions, err := c.ListFunctionRegions(context.Background())
	if err != nil || len(regions) != 1 || regions[0].Code != "us-east-1" {
		t.Fatalf("unexpected regions: %+v, err: %v", regions, err)
	}

	runtimes, err := c.ListFunctionRuntimes(context.Background())
	if err != nil || len(runtimes) != 1 || !runtimes[0].Default {
		t.Fatalf("unexpected runtimes: %+v, err: %v", runtimes, err)
	}
}
