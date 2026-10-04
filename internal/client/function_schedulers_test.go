package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestFunctionSchedulerLifecycle(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/projects/p1/functions/f1/schedulers":
			var body createSchedulerBody
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body.Schedule.CronExpression != "*/5 * * * *" {
				t.Fatalf("unexpected schedule: %+v", body.Schedule)
			}
			_ = json.NewEncoder(w).Encode(FunctionScheduler{ID: "s1", ProjectID: "p1", FunctionID: "f1", Name: body.Name, Enabled: true, CronExpression: body.Schedule.CronExpression})
		case r.Method == http.MethodPatch && r.URL.Path == "/projects/p1/functions/f1/schedulers/s1":
			var body updateSchedulerBody
			_ = json.NewDecoder(r.Body).Decode(&body)
			_ = json.NewEncoder(w).Encode(FunctionScheduler{ID: "s1", ProjectID: "p1", FunctionID: "f1", Name: "renamed", Enabled: false})
		case r.Method == http.MethodDelete && r.URL.Path == "/projects/p1/functions/f1/schedulers/s1":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	s, err := c.CreateFunctionScheduler(context.Background(), "p1", "f1", CreateFunctionSchedulerRequest{
		Name: "every-5-min", CronExpression: "*/5 * * * *",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	name := "renamed"
	enabled := false
	updated, err := c.UpdateFunctionScheduler(context.Background(), "p1", "f1", s.ID, UpdateFunctionSchedulerRequest{Name: &name, Enabled: &enabled})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "renamed" || updated.Enabled {
		t.Fatalf("unexpected scheduler after update: %+v", updated)
	}

	if err := c.DeleteFunctionScheduler(context.Background(), "p1", "f1", "s1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestDurableFunctionSchedulerUsesDurableFunctionsPath(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/p1/durable-functions/df1/schedulers" {
			t.Fatalf("expected durable-functions path, got: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(FunctionScheduler{ID: "s1", ProjectID: "p1", FunctionID: "df1", Name: "nightly"})
	})
	defer closeFn()

	if _, err := c.CreateDurableFunctionScheduler(context.Background(), "p1", "df1", CreateFunctionSchedulerRequest{Name: "nightly", CronExpression: "0 0 * * *"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
