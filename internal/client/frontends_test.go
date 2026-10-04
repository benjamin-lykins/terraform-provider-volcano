package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestFrontendLifecycle(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/projects/p1/frontends":
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				t.Fatalf("parsing multipart form: %v", err)
			}
			_ = json.NewEncoder(w).Encode(Frontend{ID: "fe1", ProjectID: "p1", Name: r.FormValue("name"), Framework: "nextjs", Status: "provisioning"})
		case r.Method == http.MethodGet && r.URL.Path == "/projects/p1/frontends/fe1":
			_ = json.NewEncoder(w).Encode(Frontend{ID: "fe1", ProjectID: "p1", Name: "web", Framework: "nextjs", Status: "active", SiteURL: "https://web.volcano.app"})
		case r.Method == http.MethodDelete && r.URL.Path == "/projects/p1/frontends/fe1":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	created, err := c.CreateFrontend(context.Background(), "p1", CreateFrontendRequest{Name: "web", ArchiveName: "web.zip", ArchiveContent: []byte("PK")})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := c.GetFrontend(context.Background(), "p1", created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.SiteURL != "https://web.volcano.app" {
		t.Fatalf("unexpected frontend: %+v", got)
	}

	if err := c.DeleteFrontend(context.Background(), "p1", "fe1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestFrontendDomainLifecycle(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/projects/p1/frontends/fe1/domain":
			var body CreateFrontendDomainRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body.TLS.Mode != "byoc" {
				t.Fatalf("unexpected tls mode: %s", body.TLS.Mode)
			}
			_ = json.NewEncoder(w).Encode(FrontendDomain{Domain: body.Domain, TLSMode: "byoc", DomainStatus: "pending_verification", VerificationStatus: "pending"})
		case r.Method == http.MethodGet && r.URL.Path == "/projects/p1/frontends/fe1/domain":
			_ = json.NewEncoder(w).Encode(FrontendDomain{Domain: "example.com", TLSMode: "byoc", DomainStatus: "active", VerificationStatus: "verified"})
		case r.Method == http.MethodDelete && r.URL.Path == "/projects/p1/frontends/fe1/domain":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	_, err := c.CreateFrontendDomain(context.Background(), "p1", "fe1", CreateFrontendDomainRequest{
		Domain: "example.com",
		TLS:    FrontendDomainTLS{Mode: "byoc", CertificatePEM: "cert", PrivateKeyPEM: "key"},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := c.GetFrontendDomain(context.Background(), "p1", "fe1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.DomainStatus != "active" {
		t.Fatalf("unexpected domain: %+v", got)
	}

	if err := c.DeleteFrontendDomain(context.Background(), "p1", "fe1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestFrontendDomainGetReturnsNilWhenNone(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("null"))
	})
	defer closeFn()

	got, err := c.GetFrontendDomain(context.Background(), "p1", "fe1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil domain, got: %+v", got)
	}
}

func TestFrontendFunctionRouteLifecycle(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/projects/p1/frontends/fe1/function-routes":
			_ = json.NewEncoder(w).Encode(FrontendFunctionRoute{ID: "rt1", ProjectID: "p1", FrontendID: "fe1", FunctionID: "f1", PathPrefix: "/api"})
		case r.Method == http.MethodGet && r.URL.Path == "/projects/p1/frontends/fe1/function-routes":
			_ = json.NewEncoder(w).Encode(Page[FrontendFunctionRoute]{Data: []FrontendFunctionRoute{{ID: "rt1", FunctionID: "f1", PathPrefix: "/api"}}})
		case r.Method == http.MethodPut && r.URL.Path == "/projects/p1/frontends/fe1/function-routes/rt1":
			_ = json.NewEncoder(w).Encode(FrontendFunctionRoute{ID: "rt1", ProjectID: "p1", FrontendID: "fe1", FunctionID: "f1", PathPrefix: "/api/v2"})
		case r.Method == http.MethodDelete && r.URL.Path == "/projects/p1/frontends/fe1/function-routes/rt1":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	route, err := c.CreateFrontendFunctionRoute(context.Background(), "p1", "fe1", FrontendFunctionRouteRequest{FunctionID: "f1", PathPrefix: "/api"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	routes, err := c.ListFrontendFunctionRoutes(context.Background(), "p1", "fe1")
	if err != nil || len(routes) != 1 {
		t.Fatalf("unexpected routes: %+v, err: %v", routes, err)
	}

	updated, err := c.UpdateFrontendFunctionRoute(context.Background(), "p1", "fe1", route.ID, FrontendFunctionRouteRequest{FunctionID: "f1", PathPrefix: "/api/v2"})
	if err != nil || updated.PathPrefix != "/api/v2" {
		t.Fatalf("unexpected update: %+v, err: %v", updated, err)
	}

	if err := c.DeleteFrontendFunctionRoute(context.Background(), "p1", "fe1", "rt1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestSetFrontendSharedVariables(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/projects/p1/frontend-shared-variables" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body SetFrontendSharedVariablesRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if len(body.FrontendSharedVariables) != 2 {
			t.Fatalf("unexpected body: %+v", body)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer closeFn()

	if err := c.SetFrontendSharedVariables(context.Background(), "p1", []string{"API_URL", "APP_NAME"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
