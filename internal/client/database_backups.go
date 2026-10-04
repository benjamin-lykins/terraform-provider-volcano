package client

import "context"

// DatabaseBackup is a named, on-demand or scheduled snapshot of a database.
type DatabaseBackup struct {
	Name      string `json:"name"`
	Source    string `json:"source,omitempty"`
	SizeBytes int64  `json:"size_bytes,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

type CreateDatabaseBackupRequest struct {
	Name string `json:"name"`
}

func (c *Client) CreateDatabaseBackup(ctx context.Context, projectID, databaseName string, in CreateDatabaseBackupRequest) (*DatabaseBackup, error) {
	var out DatabaseBackup
	path := "/projects/" + EncodePathSegment(projectID) + "/databases/" + EncodePathSegment(databaseName) + "/backups"
	if err := c.Request(ctx, "POST", path, nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetDatabaseBackup(ctx context.Context, projectID, databaseName, backupName string) (*DatabaseBackup, error) {
	var out DatabaseBackup
	path := "/projects/" + EncodePathSegment(projectID) + "/databases/" + EncodePathSegment(databaseName) + "/backups/" + EncodePathSegment(backupName)
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteDatabaseBackup(ctx context.Context, projectID, databaseName, backupName string) error {
	path := "/projects/" + EncodePathSegment(projectID) + "/databases/" + EncodePathSegment(databaseName) + "/backups/" + EncodePathSegment(backupName)
	return c.Request(ctx, "DELETE", path, nil, nil, nil)
}

// BackupScheduleEntry is one recurrence of a database's automated backup schedule.
type BackupScheduleEntry struct {
	Frequency        string `json:"frequency"`
	Hour             int64  `json:"hour"`
	Day              *int64 `json:"day,omitempty"`
	RetentionSeconds *int64 `json:"retention_seconds,omitempty"`
}

type BackupSchedule struct {
	Entries []BackupScheduleEntry `json:"entries"`
}

func (c *Client) GetDatabaseBackupSchedule(ctx context.Context, projectID, databaseName string) (*BackupSchedule, error) {
	var out BackupSchedule
	path := "/projects/" + EncodePathSegment(projectID) + "/databases/" + EncodePathSegment(databaseName) + "/backup-schedule"
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) SetDatabaseBackupSchedule(ctx context.Context, projectID, databaseName string, in BackupSchedule) (*BackupSchedule, error) {
	var out BackupSchedule
	path := "/projects/" + EncodePathSegment(projectID) + "/databases/" + EncodePathSegment(databaseName) + "/backup-schedule"
	if err := c.Request(ctx, "PUT", path, nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DatabaseRestore is a (possibly still in-progress) point-in-time or
// named-backup restore operation.
type DatabaseRestore struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	DatabaseID  string `json:"database_id"`
	Kind        string `json:"kind,omitempty"`
	BackupName  string `json:"backup_name,omitempty"`
	RestoreTo   string `json:"restore_to,omitempty"`
	Status      string `json:"status,omitempty"`
	Error       string `json:"error,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
}

type CreateDatabaseRestoreRequest struct {
	BackupName string `json:"backup_name,omitempty"`
	RestoreTo  string `json:"restore_to,omitempty"`
}

func (c *Client) CreateDatabaseRestore(ctx context.Context, projectID, databaseName string, in CreateDatabaseRestoreRequest) (*DatabaseRestore, error) {
	var out DatabaseRestore
	path := "/projects/" + EncodePathSegment(projectID) + "/databases/" + EncodePathSegment(databaseName) + "/restores"
	if err := c.Request(ctx, "POST", path, nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetDatabaseRestore(ctx context.Context, projectID, databaseName, restoreID string) (*DatabaseRestore, error) {
	var out DatabaseRestore
	path := "/projects/" + EncodePathSegment(projectID) + "/databases/" + EncodePathSegment(databaseName) + "/restores/" + EncodePathSegment(restoreID)
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
