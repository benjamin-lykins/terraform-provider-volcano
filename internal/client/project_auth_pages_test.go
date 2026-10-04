package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestProjectAuthThemeAndLayoutViaAppearance(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/projects/p1/auth/pages/theme":
			var body setThemeRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			_ = json.NewEncoder(w).Encode(body)
		case r.Method == http.MethodPut && r.URL.Path == "/projects/p1/auth/pages/login/layout":
			var body setPageLayoutRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			_ = json.NewEncoder(w).Encode(body)
		case r.Method == http.MethodGet && r.URL.Path == "/projects/p1/auth/pages/appearance":
			_ = json.NewEncoder(w).Encode(AuthPagesAppearance{
				Theme:   Theme{Colors: ThemeColors{Background: "#ffffff", Surface: "#eeeeee", Text: "#000000", Accent: "#0000ff", AccentText: "#ffffff"}, Font: "system"},
				Layouts: map[string]string{"login": "centered"},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/projects/p1/auth/pages/theme":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/projects/p1/auth/pages/login/layout":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	theme, err := c.SetProjectAuthTheme(context.Background(), "p1", Theme{
		Colors: ThemeColors{Background: "#ffffff", Surface: "#eeeeee", Text: "#000000", Accent: "#0000ff", AccentText: "#ffffff"},
		Font:   "system",
	})
	if err != nil || theme.Font != "system" {
		t.Fatalf("unexpected theme: %+v, err: %v", theme, err)
	}

	layout, err := c.SetProjectAuthPageLayout(context.Background(), "p1", "login", "centered")
	if err != nil || layout != "centered" {
		t.Fatalf("unexpected layout: %v, err: %v", layout, err)
	}

	appearance, err := c.GetProjectAuthPagesAppearance(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if appearance.Layouts["login"] != "centered" {
		t.Fatalf("unexpected appearance: %+v", appearance)
	}

	if err := c.DeleteProjectAuthTheme(context.Background(), "p1"); err != nil {
		t.Fatalf("delete theme: %v", err)
	}
	if err := c.DeleteProjectAuthPageLayout(context.Background(), "p1", "login"); err != nil {
		t.Fatalf("delete layout: %v", err)
	}
}

func TestProjectAuthHostedPageLifecycle(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/projects/p1/auth/hosted-pages/login":
			var body AuthHostedPage
			_ = json.NewDecoder(r.Body).Decode(&body)
			_ = json.NewEncoder(w).Encode(authHostedPageResponse{Page: body})
		case r.Method == http.MethodGet && r.URL.Path == "/projects/p1/auth/hosted-pages/login":
			_ = json.NewEncoder(w).Encode(authHostedPageResponse{Page: AuthHostedPage{HTML: "<html></html>", CSS: "body{}"}})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	set, err := c.SetProjectAuthHostedPage(context.Background(), "p1", "login", "<html></html>", "body{}")
	if err != nil || set.HTML != "<html></html>" {
		t.Fatalf("unexpected set: %+v, err: %v", set, err)
	}

	got, err := c.GetProjectAuthHostedPage(context.Background(), "p1", "login")
	if err != nil || got.CSS != "body{}" {
		t.Fatalf("unexpected get: %+v, err: %v", got, err)
	}
}
