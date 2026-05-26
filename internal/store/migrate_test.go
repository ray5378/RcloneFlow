package store

import (
	"os"
	"testing"
)

func openMigrateDB(t *testing.T) *DB {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "rcloneflow_migrate_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })
	db, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	return db
}

func TestMigrate_Idempotent(t *testing.T) {
	db := openMigrateDB(t)
	defer db.Close()

	if err := db.migrate(); err != nil {
		t.Fatalf("first migrate() error = %v", err)
	}

	if err := db.migrate(); err != nil {
		t.Fatalf("second migrate() should be idempotent: %v", err)
	}

	var count int
	if err := db.db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("failed to count migrations: %v", err)
	}
	var version int
	if err := db.db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version); err != nil {
		t.Fatalf("failed to get max version: %v", err)
	}
	if count != version {
		t.Errorf("migration count (%d) != max version (%d)", count, version)
	}
}

func TestMigrate_SchemaIntegrity(t *testing.T) {
	db := openMigrateDB(t)
	defer db.Close()

	tables := []string{"tasks", "schedules", "runs", "users", "task_tags", "schema_migrations"}
	for _, table := range tables {
		var count int
		err := db.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil {
			t.Fatalf("failed to check table %s: %v", table, err)
		}
		if count == 0 {
			t.Errorf("expected table %s to exist", table)
		}
	}

	indexes := []string{
		"idx_runs_task_id",
		"idx_runs_created_at",
		"idx_runs_status",
		"idx_schedules_task_id",
		"idx_tasks_name_unique",
	}
	for _, idx := range indexes {
		var count int
		err := db.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?", idx).Scan(&count)
		if err != nil {
			t.Fatalf("failed to check index %s: %v", idx, err)
		}
		if count == 0 {
			t.Errorf("expected index %s to exist", idx)
		}
	}
}

func TestMigrate_TaskColumnsIntegrity(t *testing.T) {
	db := openMigrateDB(t)
	defer db.Close()

	task := Task{
		Name:         "schema-test",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/src",
		TargetRemote: "dst",
		TargetPath:   "/dst",
		BisyncOptions: []byte(`{"resync": true}`),
	}

	created, err := db.AddTask(task)
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}

	got, _ := db.GetTask(created.ID)
	if got.Name != "schema-test" {
		t.Errorf("expected Name schema-test, got %s", got.Name)
	}
	if got.SortOrder == 0 {
		t.Error("expected non-zero SortOrder")
	}
	if got.BisyncOptions == nil {
		t.Error("expected non-nil BisyncOptions column exists")
	}
}

func TestMigrate_UserColumnsIntegrity(t *testing.T) {
	db := openMigrateDB(t)
	defer db.Close()

	u, err := db.CreateUser("test", "pw")
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	got, ok := db.GetUserByID(u.ID)
	if !ok {
		t.Fatal("expected GetUserByID to return true")
	}

	var passwordChanged int
	if err := db.db.QueryRow("SELECT password_changed FROM users WHERE id = ?", u.ID).Scan(&passwordChanged); err != nil {
		t.Fatalf("password_changed column missing: %v", err)
	}
	if passwordChanged != 0 {
		t.Errorf("expected password_changed=0, got %d", passwordChanged)
	}

	if !got.PasswordChanged && passwordChanged == 0 {
		_ = 1
	}
}
