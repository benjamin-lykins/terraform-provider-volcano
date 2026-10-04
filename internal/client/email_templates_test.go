package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestEmailTemplateLifecycle(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/projects/p1/email-templates":
			_ = json.NewEncoder(w).Encode(EmailTemplate{ID: "t1", ProjectID: "p1", TemplateType: "confirmation", Subject: "Confirm", HTMLBody: "<p>hi</p>", TextBody: "hi"})
		case r.Method == http.MethodPut && r.URL.Path == "/projects/p1/email-templates/confirmation":
			var body UpdateEmailTemplateRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			_ = json.NewEncoder(w).Encode(EmailTemplate{ID: "t1", ProjectID: "p1", TemplateType: "confirmation", Subject: *body.Subject, HTMLBody: "<p>hi</p>", TextBody: "hi"})
		case r.Method == http.MethodDelete && r.URL.Path == "/projects/p1/email-templates/confirmation":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	tmpl, err := c.CreateEmailTemplate(context.Background(), "p1", CreateEmailTemplateRequest{
		TemplateType: "confirmation", Subject: "Confirm", HTMLBody: "<p>hi</p>", TextBody: "hi",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	newSubject := "Please confirm"
	updated, err := c.UpdateEmailTemplate(context.Background(), "p1", tmpl.TemplateType, UpdateEmailTemplateRequest{Subject: &newSubject})
	if err != nil || updated.Subject != newSubject {
		t.Fatalf("unexpected update: %+v, err: %v", updated, err)
	}

	if err := c.DeleteEmailTemplate(context.Background(), "p1", "confirmation"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestListEmailTemplateDefaults(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/email-templates/defaults" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []EmailTemplateDefault{{TemplateType: "welcome", Subject: "Welcome!"}},
		})
	})
	defer closeFn()

	defaults, err := c.ListEmailTemplateDefaults(context.Background())
	if err != nil || len(defaults) != 1 || defaults[0].TemplateType != "welcome" {
		t.Fatalf("unexpected defaults: %+v, err: %v", defaults, err)
	}
}
