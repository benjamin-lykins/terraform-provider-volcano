package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestRealtimeConfigGetUpdate(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/projects/p1/realtime/config":
			_ = json.NewEncoder(w).Encode(RealtimeConfig{ProjectID: "p1", Enabled: true})
		case r.Method == http.MethodPut && r.URL.Path == "/projects/p1/realtime/config":
			var body UpdateRealtimeConfigRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			_ = json.NewEncoder(w).Encode(RealtimeConfig{ProjectID: "p1", Enabled: true, BroadcastEnabled: *body.BroadcastEnabled})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	cfg, err := c.GetRealtimeConfig(context.Background(), "p1")
	if err != nil || !cfg.Enabled {
		t.Fatalf("unexpected config: %+v, err: %v", cfg, err)
	}

	broadcast := true
	updated, err := c.UpdateRealtimeConfig(context.Background(), "p1", UpdateRealtimeConfigRequest{BroadcastEnabled: &broadcast})
	if err != nil || !updated.BroadcastEnabled {
		t.Fatalf("unexpected update: %+v, err: %v", updated, err)
	}
}
