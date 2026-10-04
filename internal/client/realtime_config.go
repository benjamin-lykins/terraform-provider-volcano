package client

import "context"

type RealtimeConfig struct {
	ProjectID              string `json:"project_id,omitempty"`
	Enabled                bool   `json:"enabled"`
	BroadcastEnabled       bool   `json:"broadcast_enabled"`
	PresenceEnabled        bool   `json:"presence_enabled"`
	PostgresChangesEnabled bool   `json:"postgres_changes_enabled"`
	CreatedAt              string `json:"created_at,omitempty"`
	UpdatedAt              string `json:"updated_at,omitempty"`
}

type UpdateRealtimeConfigRequest struct {
	Enabled                *bool `json:"enabled,omitempty"`
	BroadcastEnabled       *bool `json:"broadcast_enabled,omitempty"`
	PresenceEnabled        *bool `json:"presence_enabled,omitempty"`
	PostgresChangesEnabled *bool `json:"postgres_changes_enabled,omitempty"`
}

func (c *Client) GetRealtimeConfig(ctx context.Context, projectID string) (*RealtimeConfig, error) {
	var out RealtimeConfig
	path := "/projects/" + EncodePathSegment(projectID) + "/realtime/config"
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateRealtimeConfig(ctx context.Context, projectID string, in UpdateRealtimeConfigRequest) (*RealtimeConfig, error) {
	var out RealtimeConfig
	path := "/projects/" + EncodePathSegment(projectID) + "/realtime/config"
	if err := c.Request(ctx, "PUT", path, nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
