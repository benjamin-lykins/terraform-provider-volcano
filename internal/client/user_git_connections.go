package client

import (
	"context"
	"strconv"
)

// UserGitConnection is an account-level link to a git provider (e.g. a
// GitHub App installation owner), used as the connection_id when wiring
// up a project's git connection.
type UserGitConnection struct {
	ID                  string `json:"id"`
	Provider            string `json:"provider"`
	ProviderLogin       string `json:"provider_login,omitempty"`
	ProviderUserID      string `json:"provider_user_id,omitempty"`
	Status              string `json:"status,omitempty"`
	LastAuthenticatedAt string `json:"last_authenticated_at,omitempty"`
	CreatedAt           string `json:"created_at,omitempty"`
	UpdatedAt           string `json:"updated_at,omitempty"`
}

func (c *Client) ListUserGitConnections(ctx context.Context) ([]UserGitConnection, error) {
	var out struct {
		Connections []UserGitConnection `json:"connections"`
	}
	if err := c.Request(ctx, "GET", "/user/git/connections", nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Connections, nil
}

// UserGitInstallation is a GitHub App installation reachable through a
// connection.
type UserGitInstallation struct {
	ID                  int64  `json:"id"`
	AccountLogin        string `json:"account_login,omitempty"`
	AccountType         string `json:"account_type,omitempty"`
	RepositorySelection string `json:"repository_selection,omitempty"`
}

func (c *Client) ListUserGitInstallations(ctx context.Context, connectionID string) ([]UserGitInstallation, error) {
	var out struct {
		Installations []UserGitInstallation `json:"installations"`
	}
	path := "/user/git/connections/" + EncodePathSegment(connectionID) + "/installations"
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Installations, nil
}

// UserGitRepository is a repository reachable through an installation.
type UserGitRepository struct {
	ID            int64  `json:"id"`
	FullName      string `json:"full_name"`
	DefaultBranch string `json:"default_branch,omitempty"`
	Private       bool   `json:"private,omitempty"`
	IsEmpty       bool   `json:"is_empty,omitempty"`
}

func (c *Client) ListUserGitInstallationRepositories(ctx context.Context, connectionID string, installationID int64) ([]UserGitRepository, error) {
	var out struct {
		Repositories []UserGitRepository `json:"repositories"`
	}
	path := "/user/git/connections/" + EncodePathSegment(connectionID) + "/installations/" + strconv.FormatInt(installationID, 10) + "/repositories"
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Repositories, nil
}
