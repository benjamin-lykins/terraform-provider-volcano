package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestDatabaseBackupLifecycle(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/projects/p1/databases/main/backups":
			_ = json.NewEncoder(w).Encode(DatabaseBackup{Name: "before_migration", Source: "manual", SizeBytes: 1024})
		case r.Method == http.MethodGet && r.URL.Path == "/projects/p1/databases/main/backups/before_migration":
			_ = json.NewEncoder(w).Encode(DatabaseBackup{Name: "before_migration", Source: "manual", SizeBytes: 1024})
		case r.Method == http.MethodDelete && r.URL.Path == "/projects/p1/databases/main/backups/before_migration":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	if _, err := c.CreateDatabaseBackup(context.Background(), "p1", "main", CreateDatabaseBackupRequest{Name: "before_migration"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := c.GetDatabaseBackup(context.Background(), "p1", "main", "before_migration"); err != nil {
		t.Fatalf("get: %v", err)
	}
	if err := c.DeleteDatabaseBackup(context.Background(), "p1", "main", "before_migration"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestDatabaseBackupSchedule(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/projects/p1/databases/main/backup-schedule" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body BackupSchedule
		_ = json.NewDecoder(r.Body).Decode(&body)
		_ = json.NewEncoder(w).Encode(body)
	})
	defer closeFn()

	day := int64(1)
	schedule, err := c.SetDatabaseBackupSchedule(context.Background(), "p1", "main", BackupSchedule{
		Entries: []BackupScheduleEntry{{Frequency: "weekly", Hour: 3, Day: &day}},
	})
	if err != nil {
		t.Fatalf("set schedule: %v", err)
	}
	if len(schedule.Entries) != 1 || schedule.Entries[0].Frequency != "weekly" {
		t.Fatalf("unexpected schedule: %+v", schedule)
	}
}

func TestDatabaseRestore(t *testing.T) {
	c, closeFn := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/projects/p1/databases/main/restores":
			_ = json.NewEncoder(w).Encode(DatabaseRestore{ID: "r1", ProjectID: "p1", DatabaseID: "d1", Status: "in_progress", BackupName: "before_migration"})
		case r.Method == http.MethodGet && r.URL.Path == "/projects/p1/databases/main/restores/r1":
			_ = json.NewEncoder(w).Encode(DatabaseRestore{ID: "r1", ProjectID: "p1", DatabaseID: "d1", Status: "completed", BackupName: "before_migration"})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer closeFn()

	restore, err := c.CreateDatabaseRestore(context.Background(), "p1", "main", CreateDatabaseRestoreRequest{BackupName: "before_migration"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := c.GetDatabaseRestore(context.Background(), "p1", "main", restore.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != "completed" {
		t.Fatalf("unexpected restore: %+v", got)
	}
}
