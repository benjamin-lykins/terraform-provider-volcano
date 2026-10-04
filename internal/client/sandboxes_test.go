package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestSandboxLifecycle(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/projects/p1/sandboxes":
			_ = json.NewEncoder(w).Encode(Sandbox{ID: "sb1", ProjectID: "p1", Name: "dev", Preset: "small", Status: "running"})
		case r.Method == http.MethodPatch && r.URL.Path == "/projects/p1/sandboxes/sb1":
			var body UpdateSandboxRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			_ = json.NewEncoder(w).Encode(Sandbox{ID: "sb1", ProjectID: "p1", Name: body.Name, Preset: "small", Status: "running"})
		case r.Method == http.MethodDelete && r.URL.Path == "/projects/p1/sandboxes/sb1":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	sb, err := c.CreateSandbox(context.Background(), "p1", CreateSandboxRequest{Name: "dev", Preset: "small"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	updated, err := c.UpdateSandbox(context.Background(), "p1", sb.ID, UpdateSandboxRequest{Name: "dev-renamed"})
	if err != nil || updated.Name != "dev-renamed" {
		t.Fatalf("unexpected update: %+v, err: %v", updated, err)
	}

	if err := c.DeleteSandbox(context.Background(), "p1", "sb1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestListSandboxPresets(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sandboxes/presets" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []SandboxPreset{{ID: "small", Runtime: "ubuntu", MemoryMB: 512}},
		})
	})
	defer closeFn()

	presets, err := c.ListSandboxPresets(context.Background())
	if err != nil || len(presets) != 1 || presets[0].ID != "small" {
		t.Fatalf("unexpected presets: %+v, err: %v", presets, err)
	}
}
