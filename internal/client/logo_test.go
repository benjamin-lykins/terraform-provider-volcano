package client

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestUploadProjectLogo(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/projects/p1/logo" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Fatalf("expected multipart content type, got: %s", r.Header.Get("Content-Type"))
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("parsing multipart form: %v", err)
		}
		file, header, err := r.FormFile("logo")
		if err != nil {
			t.Fatalf("reading form file: %v", err)
		}
		defer file.Close()
		if header.Filename != "logo.png" {
			t.Fatalf("unexpected filename: %s", header.Filename)
		}
		_ = json.NewEncoder(w).Encode(Project{ID: "p1", Name: "demo", LogoURL: "/projects/p1/logo?v=1"})
	})
	defer closeFn()

	project, err := c.UploadProjectLogo(context.Background(), "p1", "logo.png", []byte{0x89, 0x50, 0x4e, 0x47})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project.LogoURL != "/projects/p1/logo?v=1" {
		t.Fatalf("unexpected project: %+v", project)
	}
}

func TestDeleteProjectLogo(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/projects/p1/logo" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{})
	})
	defer closeFn()

	if err := c.DeleteProjectLogo(context.Background(), "p1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
