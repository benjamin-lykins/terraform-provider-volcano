package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestListProjectAndAccountDeployments(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/projects/p1/deployments":
			_ = json.NewEncoder(w).Encode(Page[Deployment]{Data: []Deployment{{ID: "d1", ProjectID: "p1", Resource: "function"}}})
		case "/deployments":
			_ = json.NewEncoder(w).Encode(Page[Deployment]{Data: []Deployment{{ID: "d2", ProjectID: "p2", Resource: "frontend"}}})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})
	defer closeFn()

	projectDeployments, err := c.ListProjectDeployments(context.Background(), "p1")
	if err != nil || len(projectDeployments) != 1 || projectDeployments[0].ID != "d1" {
		t.Fatalf("unexpected project deployments: %+v, err: %v", projectDeployments, err)
	}

	accountDeployments, err := c.ListAccountDeployments(context.Background())
	if err != nil || len(accountDeployments) != 1 || accountDeployments[0].ID != "d2" {
		t.Fatalf("unexpected account deployments: %+v, err: %v", accountDeployments, err)
	}
}

func TestListProjectDomainsAndSchedulers(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/projects/p1/domains":
			_ = json.NewEncoder(w).Encode(Page[FrontendDomain]{Data: []FrontendDomain{{Domain: "example.com", DomainStatus: "active"}}})
		case "/projects/p1/schedulers":
			_ = json.NewEncoder(w).Encode(Page[FunctionScheduler]{Data: []FunctionScheduler{{ID: "s1", Name: "nightly"}}})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})
	defer closeFn()

	domains, err := c.ListProjectDomains(context.Background(), "p1")
	if err != nil || len(domains) != 1 || domains[0].Domain != "example.com" {
		t.Fatalf("unexpected domains: %+v, err: %v", domains, err)
	}

	schedulers, err := c.ListProjectSchedulers(context.Background(), "p1")
	if err != nil || len(schedulers) != 1 || schedulers[0].Name != "nightly" {
		t.Fatalf("unexpected schedulers: %+v, err: %v", schedulers, err)
	}
}

func TestGetProjectHealth(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/p1/health" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(ProjectHealth{
			ProjectID: "p1",
			Status:    "healthy",
			Checks:    []ProjectHealthCheck{{ID: "c1", Category: "storage", Status: "healthy"}},
		})
	})
	defer closeFn()

	health, err := c.GetProjectHealth(context.Background(), "p1")
	if err != nil || health.Status != "healthy" || len(health.Checks) != 1 {
		t.Fatalf("unexpected health: %+v, err: %v", health, err)
	}
}
