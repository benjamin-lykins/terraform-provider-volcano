package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestSetSharedVariables(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/projects/p1/shared-variables" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body SetSharedVariablesRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if len(body.SharedVariables) != 2 {
			t.Fatalf("unexpected body: %+v", body)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer closeFn()

	if err := c.SetSharedVariables(context.Background(), "p1", []string{"API_KEY", "DEBUG"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
