package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestStorageObjectUploadDeleteAndVisibility(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/storage/avatars/users/abc/avatar.png":
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				t.Fatalf("parsing multipart form: %v", err)
			}
			_ = json.NewEncoder(w).Encode(StorageObject{ID: "o1", BucketID: "b1", Name: "users/abc/avatar.png", MimeType: "image/png", Size: 1024, IsPublic: false})
		case r.Method == http.MethodPatch && r.URL.Path == "/storage/avatars/users/abc/avatar.png/visibility":
			var body setStorageObjectVisibilityRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			_ = json.NewEncoder(w).Encode(StorageObject{ID: "o1", BucketID: "b1", Name: "users/abc/avatar.png", IsPublic: body.IsPublic})
		case r.Method == http.MethodDelete && r.URL.Path == "/storage/avatars/users/abc/avatar.png":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	obj, err := c.UploadStorageObject(context.Background(), "avatars", "users/abc/avatar.png", "avatar.png", []byte{0x89, 0x50})
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if obj.Size != 1024 {
		t.Fatalf("unexpected object: %+v", obj)
	}

	updated, err := c.SetStorageObjectVisibility(context.Background(), "avatars", "users/abc/avatar.png", true)
	if err != nil || !updated.IsPublic {
		t.Fatalf("unexpected visibility update: %+v, err: %v", updated, err)
	}

	if err := c.DeleteStorageObject(context.Background(), "avatars", "users/abc/avatar.png"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestGetStorageObjectMatchesExactBucketAndName(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/p1/storage/objects" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("search") != "avatar.png" {
			t.Fatalf("expected search query, got: %s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(Page[StorageObject]{Data: []StorageObject{
			{ID: "o1", BucketName: "other-bucket", Name: "avatar.png"},
			{ID: "o2", BucketName: "avatars", Name: "avatar.png"},
		}})
	})
	defer closeFn()

	got, err := c.GetStorageObject(context.Background(), "p1", "avatars", "avatar.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "o2" {
		t.Fatalf("expected exact bucket match o2, got: %+v", got)
	}
}

func TestGetStorageObjectNotFound(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Page[StorageObject]{Data: nil})
	})
	defer closeFn()

	_, err := c.GetStorageObject(context.Background(), "p1", "avatars", "missing.png")
	if !IsNotFound(err) {
		t.Fatalf("expected not found error, got: %v", err)
	}
}
