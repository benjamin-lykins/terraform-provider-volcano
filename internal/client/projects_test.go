package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) (*Client, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c := New(srv.URL, "pk-test-token", srv.Client())
	return c, srv.Close
}

func TestCreateProject(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/projects" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer pk-test-token" {
			t.Fatalf("unexpected Authorization header: %s", got)
		}
		var body CreateProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding request body: %v", err)
		}
		if body.Name != "my-app" {
			t.Fatalf("unexpected name in request: %s", body.Name)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(Project{
			ID:         "11111111-1111-1111-1111-111111111111",
			Name:       body.Name,
			Status:     "active",
			AllRegions: true,
		})
	})
	defer closeFn()

	got, err := c.CreateProject(context.Background(), CreateProjectRequest{Name: "my-app"})
	if err != nil {
		t.Fatalf("CreateProject returned error: %v", err)
	}
	if got.ID != "11111111-1111-1111-1111-111111111111" || got.Status != "active" {
		t.Fatalf("unexpected project: %+v", got)
	}
}

func TestGetProjectNotFound(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "project not found"})
	})
	defer closeFn()

	_, err := c.GetProject(context.Background(), "missing-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Fatalf("expected IsNotFound(err) to be true, got error: %v", err)
	}
}

func TestUpdateProject(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/projects/abc" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body UpdateProjectRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Name == nil || *body.Name != "renamed" {
			t.Fatalf("unexpected update body: %+v", body)
		}
		_ = json.NewEncoder(w).Encode(Project{ID: "abc", Name: "renamed", Status: "active"})
	})
	defer closeFn()

	name := "renamed"
	got, err := c.UpdateProject(context.Background(), "abc", UpdateProjectRequest{Name: &name})
	if err != nil {
		t.Fatalf("UpdateProject returned error: %v", err)
	}
	if got.Name != "renamed" {
		t.Fatalf("unexpected project after update: %+v", got)
	}
}

func TestDeleteProject(t *testing.T) {
	called := false
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodDelete || r.URL.Path != "/projects/abc" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer closeFn()

	if err := c.DeleteProject(context.Background(), "abc"); err != nil {
		t.Fatalf("DeleteProject returned error: %v", err)
	}
	if !called {
		t.Fatal("expected DELETE request to be made")
	}
}

func TestListProjects(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("search") != "staging" {
			t.Fatalf("expected search query param, got: %s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(listProjectsResponse{
			Page: Page[Project]{Data: []Project{{ID: "1", Name: "staging-app"}}},
		})
	})
	defer closeFn()

	got, err := c.ListProjects(context.Background(), ListProjectsOptions{Search: "staging"})
	if err != nil {
		t.Fatalf("ListProjects returned error: %v", err)
	}
	if len(got) != 1 || got[0].Name != "staging-app" {
		t.Fatalf("unexpected projects: %+v", got)
	}
}
