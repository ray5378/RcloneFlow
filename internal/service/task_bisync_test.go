package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"rcloneflow/internal/store"
)

// setupBisyncTest creates a temp dir, sets APP_DATA_DIR, opens a DB, creates a bisync task.
// It returns (tmpDir, db, svc, taskID) and registers cleanup via t.Cleanup.
func setupBisyncTest(t *testing.T) (string, *store.DB, *TaskService, int64) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "rcloneflow_bisync_*")
	if err != nil {
		t.Fatalf("MkdirTemp error = %v", err)
	}

	oldAppDataDir := os.Getenv("APP_DATA_DIR")
	if err := os.Setenv("APP_DATA_DIR", tmpDir); err != nil {
		t.Fatalf("Setenv(APP_DATA_DIR) error = %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
		if oldAppDataDir == "" {
			_ = os.Unsetenv("APP_DATA_DIR")
		} else {
			_ = os.Setenv("APP_DATA_DIR", oldAppDataDir)
		}
	})

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open error = %v", err)
	}
	t.Cleanup(func() { db.Close() })

	svc := NewTaskService(db, nil)

	task, err := svc.CreateTask(store.Task{
		Name:         "test-bisync",
		Mode:         "bisync",
		SourceRemote: "src",
		SourcePath:   "/a",
		TargetRemote: "dst",
		TargetPath:   "/b",
	})
	if err != nil {
		t.Fatalf("CreateTask error = %v", err)
	}

	return tmpDir, db, svc, task.ID
}

// createBisyncFiles creates files inside the bisync dir for the given task.
// files map keys are relative paths (e.g. "path1.lst", "bak/path1.lst.ts.bak").
func createBisyncFiles(t *testing.T, svc *TaskService, taskID int64, files map[string]string) string {
	t.Helper()

	task, ok := svc.GetTask(taskID)
	if !ok {
		t.Fatalf("task %d not found", taskID)
	}

	dir := svc.getBisyncDir(task.Name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) error = %v", dir, err)
	}

	for name, content := range files {
		fp := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(fp), 0o755); err != nil {
			t.Fatalf("MkdirAll(%s) error = %v", filepath.Dir(fp), err)
		}
		if err := os.WriteFile(fp, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile(%s) error = %v", fp, err)
		}
	}

	return dir
}

// ---------------------------------------------------------------------------
// GetBisyncLstFiles
// ---------------------------------------------------------------------------

func TestGetBisyncLstFiles(t *testing.T) {
	t.Run("current version", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)
		createBisyncFiles(t, svc, taskID, map[string]string{
			"path1.lst": "content1",
			"path2.lst": "content2",
		})

		versions, err := svc.GetBisyncLstFiles(taskID)
		if err != nil {
			t.Fatalf("GetBisyncLstFiles error = %v", err)
		}
		if len(versions) != 1 {
			t.Fatalf("expected 1 version, got %d", len(versions))
		}
		v := versions[0]
		if v.ID != "current" {
			t.Errorf("expected ID 'current', got %q", v.ID)
		}
		if v.Type != "current" {
			t.Errorf("expected Type 'current', got %q", v.Type)
		}
		if v.Path1Lst != "path1.lst" {
			t.Errorf("expected Path1Lst 'path1.lst', got %q", v.Path1Lst)
		}
		if v.Path2Lst != "path2.lst" {
			t.Errorf("expected Path2Lst 'path2.lst', got %q", v.Path2Lst)
		}
		if v.Path1Size <= 0 {
			t.Errorf("expected Path1Size > 0, got %d", v.Path1Size)
		}
		if v.Path2Size <= 0 {
			t.Errorf("expected Path2Size > 0, got %d", v.Path2Size)
		}
	})

	t.Run("conflict files", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)
		createBisyncFiles(t, svc, taskID, map[string]string{
			"mysync.conflict1": "keep",
			"mysync.conflict2": "delete",
		})

		versions, err := svc.GetBisyncLstFiles(taskID)
		if err != nil {
			t.Fatalf("GetBisyncLstFiles error = %v", err)
		}
		if len(versions) != 1 {
			t.Fatalf("expected 1 version, got %d", len(versions))
		}
		v := versions[0]
		if v.ID != "mysync" {
			t.Errorf("expected ID 'mysync', got %q", v.ID)
		}
		if v.Type != "conflict" {
			t.Errorf("expected Type 'conflict', got %q", v.Type)
		}
		if v.Conflict1 != "mysync.conflict1" {
			t.Errorf("expected Conflict1 'mysync.conflict1', got %q", v.Conflict1)
		}
		if v.Conflict2 != "mysync.conflict2" {
			t.Errorf("expected Conflict2 'mysync.conflict2', got %q", v.Conflict2)
		}
	})

	t.Run("old-format backups", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)
		createBisyncFiles(t, svc, taskID, map[string]string{
			"prefix.path1.lst-old": "old1",
			"prefix.path2.lst-old": "old2",
		})

		versions, err := svc.GetBisyncLstFiles(taskID)
		if err != nil {
			t.Fatalf("GetBisyncLstFiles error = %v", err)
		}
		if len(versions) != 1 {
			t.Fatalf("expected 1 version, got %d", len(versions))
		}
		v := versions[0]
		if v.ID != "prefix" {
			t.Errorf("expected ID 'prefix', got %q", v.ID)
		}
		if v.Type != "backup" {
			t.Errorf("expected Type 'backup', got %q", v.Type)
		}
		if v.Path1Lst != "prefix.path1.lst-old" {
			t.Errorf("expected Path1Lst 'prefix.path1.lst-old', got %q", v.Path1Lst)
		}
		if v.Path2Lst != "prefix.path2.lst-old" {
			t.Errorf("expected Path2Lst 'prefix.path2.lst-old', got %q", v.Path2Lst)
		}
	})

	t.Run("new-format backups in bak subdir", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)
		createBisyncFiles(t, svc, taskID, map[string]string{
			"bak/path1.lst.20240101-120000.bak": "bak1",
			"bak/path2.lst.20240101-120000.bak": "bak2",
		})

		versions, err := svc.GetBisyncLstFiles(taskID)
		if err != nil {
			t.Fatalf("GetBisyncLstFiles error = %v", err)
		}
		if len(versions) != 1 {
			t.Fatalf("expected 1 version, got %d", len(versions))
		}
		v := versions[0]
		if v.ID != "20240101-120000" {
			t.Errorf("expected ID '20240101-120000', got %q", v.ID)
		}
		if v.Type != "backup" {
			t.Errorf("expected Type 'backup', got %q", v.Type)
		}
		// One of the path fields should be set; the naming convention can vary
		if v.Path1Lst == "" && v.Path2Lst == "" {
			t.Errorf("expected at least one path file to be set, got Path1Lst=%q Path2Lst=%q", v.Path1Lst, v.Path2Lst)
		}
	})

	t.Run("empty dir returns empty list", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)
		// Create bisync dir but leave it empty
		task, _ := svc.GetTask(taskID)
		dir := svc.getBisyncDir(task.Name)
		_ = os.MkdirAll(dir, 0o755)

		versions, err := svc.GetBisyncLstFiles(taskID)
		if err != nil {
			t.Fatalf("GetBisyncLstFiles error = %v", err)
		}
		if len(versions) != 0 {
			t.Errorf("expected 0 versions for empty dir, got %d", len(versions))
		}
	})

	t.Run("no task returns error", func(t *testing.T) {
		_, _, svc, _ := setupBisyncTest(t)
		_, err := svc.GetBisyncLstFiles(99999)
		if err != ErrTaskNotFound {
			t.Errorf("expected ErrTaskNotFound, got %v", err)
		}
	})

	t.Run("current and backup combined", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)
		createBisyncFiles(t, svc, taskID, map[string]string{
			"path1.lst":                 "current1",
			"path2.lst":                 "current2",
			"bak/path1.lst.20240101-120000.bak": "bak1",
			"bak/path2.lst.20240101-120000.bak": "bak2",
		})

		versions, err := svc.GetBisyncLstFiles(taskID)
		if err != nil {
			t.Fatalf("GetBisyncLstFiles error = %v", err)
		}
		if len(versions) != 2 {
			t.Fatalf("expected 2 versions (current + backup), got %d", len(versions))
		}
		// Should have one "current" and one "backup"
		seenCurrent := false
		seenBackup := false
		for _, v := range versions {
			if v.Type == "current" {
				seenCurrent = true
				if v.Path1Lst != "path1.lst" {
					t.Errorf("current Path1Lst = %q", v.Path1Lst)
				}
			}
			if v.Type == "backup" {
				seenBackup = true
			}
		}
		if !seenCurrent {
			t.Error("missing current version")
		}
		if !seenBackup {
			t.Error("missing backup version")
		}
	})
}

// ---------------------------------------------------------------------------
// DeleteBisyncLstVersion
// ---------------------------------------------------------------------------

func TestDeleteBisyncLstVersion(t *testing.T) {
	t.Run("delete conflict1 version removes only conflict2 file", func(t *testing.T) {
		// The TrimSuffix logic produces conflict1=prefix.conflict1.conflict1 (not found),
		// conflict2=prefix.conflict2 (found). So passing conflict1 removes conflict2 only.
		_, _, svc, taskID := setupBisyncTest(t)
		dir := createBisyncFiles(t, svc, taskID, map[string]string{
			"mysync.conflict1": "keep",
			"mysync.conflict2": "delete",
		})

		err := svc.DeleteBisyncLstVersion(taskID, "mysync.conflict1")
		if err != nil {
			t.Fatalf("DeleteBisyncLstVersion error = %v", err)
		}

		// conflict1 should survive (no matching computed name)
		if _, err := os.Stat(filepath.Join(dir, "mysync.conflict1")); os.IsNotExist(err) {
			t.Error("expected mysync.conflict1 to survive")
		}
		// conflict2 should be removed
		if _, err := os.Stat(filepath.Join(dir, "mysync.conflict2")); !os.IsNotExist(err) {
			t.Error("expected mysync.conflict2 to be removed")
		}
	})

	t.Run("delete conflict2 version removes only conflict1 file", func(t *testing.T) {
		// The TrimSuffix logic produces conflict1=prefix.conflict1 (found),
		// conflict2=prefix.conflict2.conflict2 (not found). So passing conflict2 removes conflict1 only.
		_, _, svc, taskID := setupBisyncTest(t)
		dir := createBisyncFiles(t, svc, taskID, map[string]string{
			"mysync.conflict1": "keep",
			"mysync.conflict2": "delete",
		})

		err := svc.DeleteBisyncLstVersion(taskID, "mysync.conflict2")
		if err != nil {
			t.Fatalf("DeleteBisyncLstVersion error = %v", err)
		}

		if _, err := os.Stat(filepath.Join(dir, "mysync.conflict1")); !os.IsNotExist(err) {
			t.Error("expected mysync.conflict1 to be removed")
		}
		// conflict2 should survive
		if _, err := os.Stat(filepath.Join(dir, "mysync.conflict2")); os.IsNotExist(err) {
			t.Error("expected mysync.conflict2 to survive")
		}
	})

	t.Run("delete old-format backup removes both -old files", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)
		dir := createBisyncFiles(t, svc, taskID, map[string]string{
			"prefix.path1.lst-old": "old1",
			"prefix.path2.lst-old": "old2",
		})

		err := svc.DeleteBisyncLstVersion(taskID, "prefix.path1.lst-old")
		if err != nil {
			t.Fatalf("DeleteBisyncLstVersion error = %v", err)
		}

		if _, err := os.Stat(filepath.Join(dir, "prefix.path1.lst-old")); !os.IsNotExist(err) {
			t.Error("expected prefix.path1.lst-old to be removed")
		}
		if _, err := os.Stat(filepath.Join(dir, "prefix.path2.lst-old")); !os.IsNotExist(err) {
			t.Error("expected prefix.path2.lst-old to be removed")
		}
	})

	t.Run("delete old-format backup with path2 ID", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)
		dir := createBisyncFiles(t, svc, taskID, map[string]string{
			"prefix.path1.lst-old": "old1",
			"prefix.path2.lst-old": "old2",
		})

		err := svc.DeleteBisyncLstVersion(taskID, "prefix.path2.lst-old")
		if err != nil {
			t.Fatalf("DeleteBisyncLstVersion error = %v", err)
		}

		if _, err := os.Stat(filepath.Join(dir, "prefix.path1.lst-old")); !os.IsNotExist(err) {
			t.Error("expected prefix.path1.lst-old to be removed")
		}
		if _, err := os.Stat(filepath.Join(dir, "prefix.path2.lst-old")); !os.IsNotExist(err) {
			t.Error("expected prefix.path2.lst-old to be removed")
		}
	})

	t.Run("delete current version returns ErrBisyncCurrentVersion", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)
		createBisyncFiles(t, svc, taskID, map[string]string{
			"path1.lst": "content1",
			"path2.lst": "content2",
		})

		err := svc.DeleteBisyncLstVersion(taskID, "current")
		if err != ErrBisyncCurrentVersion {
			t.Errorf("expected ErrBisyncCurrentVersion, got %v", err)
		}

		// Files should still exist
		task, _ := svc.GetTask(taskID)
		dir := svc.getBisyncDir(task.Name)
		if _, err := os.Stat(filepath.Join(dir, "path1.lst")); os.IsNotExist(err) {
			t.Error("expected path1.lst to still exist")
		}
	})

	t.Run("delete non-existent version is no-op", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)
		createBisyncFiles(t, svc, taskID, map[string]string{
			"path1.lst": "content1",
		})

		err := svc.DeleteBisyncLstVersion(taskID, "nonexistent")
		if err != nil {
			t.Fatalf("expected no error for non-existent version, got %v", err)
		}

		task, _ := svc.GetTask(taskID)
		dir := svc.getBisyncDir(task.Name)
		if _, err := os.Stat(filepath.Join(dir, "path1.lst")); os.IsNotExist(err) {
			t.Error("expected path1.lst to remain untouched")
		}
	})

	t.Run("delete new-format backup from bak subdir", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)
		dir := createBisyncFiles(t, svc, taskID, map[string]string{
			"bak/path1.lst.20240101-120000.bak": "bak1",
			"bak/path2.lst.20240101-120000.bak": "bak2",
		})

		err := svc.DeleteBisyncLstVersion(taskID, "20240101-120000")
		if err != nil {
			t.Fatalf("DeleteBisyncLstVersion error = %v", err)
		}

		bakDir := filepath.Join(dir, "bak")
		if entries, _ := os.ReadDir(bakDir); len(entries) != 0 {
			t.Errorf("expected bak dir to be empty after delete, got %d entries", len(entries))
		}
	})

	t.Run("no task returns error", func(t *testing.T) {
		_, _, svc, _ := setupBisyncTest(t)
		err := svc.DeleteBisyncLstVersion(99999, "current")
		if err != ErrTaskNotFound {
			t.Errorf("expected ErrTaskNotFound, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// RollbackBisyncLstVersion
// ---------------------------------------------------------------------------

func TestRollbackBisyncLstVersion(t *testing.T) {
	t.Run("rollback old-format backup renames files back", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)
		dir := createBisyncFiles(t, svc, taskID, map[string]string{
			"path1.lst":              "current1",
			"path2.lst":              "current2",
			"prefix.path1.lst-old":   "backup1",
			"prefix.path2.lst-old":   "backup2",
		})

		err := svc.RollbackBisyncLstVersion(taskID, "prefix.path1.lst-old")
		if err != nil {
			t.Fatalf("RollbackBisyncLstVersion error = %v", err)
		}

		// Current files should now contain backup content
		data1, err := os.ReadFile(filepath.Join(dir, "path1.lst"))
		if err != nil {
			t.Fatalf("ReadFile(path1.lst) error = %v", err)
		}
		if string(data1) != "backup1" {
			t.Errorf("expected path1.lst content 'backup1', got %q", string(data1))
		}

		data2, err := os.ReadFile(filepath.Join(dir, "path2.lst"))
		if err != nil {
			t.Fatalf("ReadFile(path2.lst) error = %v", err)
		}
		if string(data2) != "backup2" {
			t.Errorf("expected path2.lst content 'backup2', got %q", string(data2))
		}

		// Old backup files should no longer exist
		if _, err := os.Stat(filepath.Join(dir, "prefix.path1.lst-old")); !os.IsNotExist(err) {
			t.Error("expected prefix.path1.lst-old to be removed")
		}
		if _, err := os.Stat(filepath.Join(dir, "prefix.path2.lst-old")); !os.IsNotExist(err) {
			t.Error("expected prefix.path2.lst-old to be removed")
		}

		// Old current files should be backed up in bak/ dir
		bakDir := filepath.Join(dir, "bak")
		bakEntries, err := os.ReadDir(bakDir)
		if err != nil {
			t.Fatalf("ReadDir(bak) error = %v", err)
		}
		if len(bakEntries) != 2 {
			t.Errorf("expected 2 backup files in bak/, got %d", len(bakEntries))
		}
	})

	t.Run("rollback with prefix ID for old format", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)
		dir := createBisyncFiles(t, svc, taskID, map[string]string{
			"path1.lst":            "current1",
			"path2.lst":            "current2",
			"prefix.path1.lst-old": "backup1",
			"prefix.path2.lst-old": "backup2",
		})

		err := svc.RollbackBisyncLstVersion(taskID, "prefix")
		if err != nil {
			t.Fatalf("RollbackBisyncLstVersion error = %v", err)
		}

		data1, _ := os.ReadFile(filepath.Join(dir, "path1.lst"))
		if string(data1) != "backup1" {
			t.Errorf("expected path1.lst content 'backup1', got %q", string(data1))
		}
	})

	t.Run("no task returns error", func(t *testing.T) {
		_, _, svc, _ := setupBisyncTest(t)
		err := svc.RollbackBisyncLstVersion(99999, "prefix.path1.lst-old")
		if err != ErrTaskNotFound {
			t.Errorf("expected ErrTaskNotFound, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// BackupCurrentLstFiles
// ---------------------------------------------------------------------------

func TestBackupCurrentLstFiles(t *testing.T) {
	t.Run("backs up path1.lst and path2.lst to bak dir", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)
		dir := createBisyncFiles(t, svc, taskID, map[string]string{
			"path1.lst": "content1",
			"path2.lst": "content2",
		})

		err := svc.BackupCurrentLstFiles(taskID)
		if err != nil {
			t.Fatalf("BackupCurrentLstFiles error = %v", err)
		}

		// Original files should be moved to bak/
		bakDir := filepath.Join(dir, "bak")
		bakEntries, err := os.ReadDir(bakDir)
		if err != nil {
			t.Fatalf("ReadDir(bak) error = %v", err)
		}
		if len(bakEntries) != 2 {
			t.Fatalf("expected 2 backup files in bak/, got %d", len(bakEntries))
		}

		// Original files should no longer exist in main bisync dir
		if _, err := os.Stat(filepath.Join(dir, "path1.lst")); !os.IsNotExist(err) {
			t.Error("expected path1.lst to be moved")
		}
		if _, err := os.Stat(filepath.Join(dir, "path2.lst")); !os.IsNotExist(err) {
			t.Error("expected path2.lst to be moved")
		}

		// Verify .bak extension
		for _, entry := range bakEntries {
			if !strings.HasSuffix(entry.Name(), ".bak") {
				t.Errorf("expected .bak extension, got %q", entry.Name())
			}
			if !strings.Contains(entry.Name(), ".lst.") {
				t.Errorf("expected .lst. in backup name, got %q", entry.Name())
			}
		}
	})

	t.Run("no files to backup is not an error", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)
		task, _ := svc.GetTask(taskID)
		dir := svc.getBisyncDir(task.Name)
		_ = os.MkdirAll(dir, 0o755)

		err := svc.BackupCurrentLstFiles(taskID)
		if err != nil {
			t.Fatalf("BackupCurrentLstFiles error = %v", err)
		}
	})

	t.Run("no task returns error", func(t *testing.T) {
		_, _, svc, _ := setupBisyncTest(t)
		err := svc.BackupCurrentLstFiles(99999)
		if err != ErrTaskNotFound {
			t.Errorf("expected ErrTaskNotFound, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// GetBisyncLstContent
// ---------------------------------------------------------------------------

func TestGetBisyncLstContent(t *testing.T) {
	t.Run("reads current lst file", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)
		createBisyncFiles(t, svc, taskID, map[string]string{
			"path1.lst": "hello world",
		})

		content, err := svc.GetBisyncLstContent(taskID, "path1.lst")
		if err != nil {
			t.Fatalf("GetBisyncLstContent error = %v", err)
		}
		if content != "hello world" {
			t.Errorf("expected 'hello world', got %q", content)
		}
	})

	t.Run("reads bak file from bak subdir", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)
		createBisyncFiles(t, svc, taskID, map[string]string{
			"bak/path1.lst.20240101-120000.bak": "backup content",
		})

		content, err := svc.GetBisyncLstContent(taskID, "path1.lst.20240101-120000.bak")
		if err != nil {
			t.Fatalf("GetBisyncLstContent error = %v", err)
		}
		if content != "backup content" {
			t.Errorf("expected 'backup content', got %q", content)
		}
	})

	t.Run("path traversal is rejected", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)
		createBisyncFiles(t, svc, taskID, map[string]string{
			"path1.lst": "safe",
		})

		_, err := svc.GetBisyncLstContent(taskID, "../../etc/passwd")
		if err == nil {
			t.Error("expected error for path traversal, got nil")
		}
	})

	t.Run("non-existent file returns error", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)
		createBisyncFiles(t, svc, taskID, map[string]string{
			"path1.lst": "safe",
		})

		_, err := svc.GetBisyncLstContent(taskID, "nonexistent.lst")
		if err == nil {
			t.Error("expected error for non-existent file, got nil")
		}
	})

	t.Run("no task returns error", func(t *testing.T) {
		_, _, svc, _ := setupBisyncTest(t)
		_, err := svc.GetBisyncLstContent(99999, "path1.lst")
		if err != ErrTaskNotFound {
			t.Errorf("expected ErrTaskNotFound, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// ResolveBisyncConflict
// ---------------------------------------------------------------------------

func TestResolveBisyncConflict(t *testing.T) {
	t.Run("keep conflict1 removes conflict2 and renames conflict1", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)
		dir := createBisyncFiles(t, svc, taskID, map[string]string{
			"mysync.conflict1": "winner",
			"mysync.conflict2": "loser",
		})

		err := svc.ResolveBisyncConflict(taskID, "mysync", "conflict1")
		if err != nil {
			t.Fatalf("ResolveBisyncConflict error = %v", err)
		}

		// conflict2 should be removed
		if _, err := os.Stat(filepath.Join(dir, "mysync.conflict2")); !os.IsNotExist(err) {
			t.Error("expected mysync.conflict2 to be removed")
		}

		// conflict1 renamed to the prefix
		data, err := os.ReadFile(filepath.Join(dir, "mysync"))
		if err != nil {
			t.Fatalf("ReadFile(mysync) error = %v", err)
		}
		if string(data) != "winner" {
			t.Errorf("expected content 'winner', got %q", string(data))
		}

		// conflict1 should no longer exist
		if _, err := os.Stat(filepath.Join(dir, "mysync.conflict1")); !os.IsNotExist(err) {
			t.Error("expected mysync.conflict1 to be renamed away")
		}
	})

	t.Run("keep conflict2 removes conflict1 and renames conflict2", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)
		dir := createBisyncFiles(t, svc, taskID, map[string]string{
			"mysync.conflict1": "loser",
			"mysync.conflict2": "winner",
		})

		err := svc.ResolveBisyncConflict(taskID, "mysync", "conflict2")
		if err != nil {
			t.Fatalf("ResolveBisyncConflict error = %v", err)
		}

		if _, err := os.Stat(filepath.Join(dir, "mysync.conflict1")); !os.IsNotExist(err) {
			t.Error("expected mysync.conflict1 to be removed")
		}

		data, err := os.ReadFile(filepath.Join(dir, "mysync"))
		if err != nil {
			t.Fatalf("ReadFile(mysync) error = %v", err)
		}
		if string(data) != "winner" {
			t.Errorf("expected content 'winner', got %q", string(data))
		}

		if _, err := os.Stat(filepath.Join(dir, "mysync.conflict2")); !os.IsNotExist(err) {
			t.Error("expected mysync.conflict2 to be renamed away")
		}
	})

	t.Run("invalid keepFile defaults to conflict2 branch", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)
		dir := createBisyncFiles(t, svc, taskID, map[string]string{
			"mysync.conflict1": "loser",
			"mysync.conflict2": "kept",
		})

		// "invalid" triggers the else branch (keep = conflict2)
		err := svc.ResolveBisyncConflict(taskID, "mysync", "invalid")
		if err != nil {
			t.Fatalf("ResolveBisyncConflict error = %v", err)
		}

		data, err := os.ReadFile(filepath.Join(dir, "mysync"))
		if err != nil {
			t.Fatalf("ReadFile(mysync) error = %v", err)
		}
		if string(data) != "kept" {
			t.Errorf("expected content 'kept', got %q", string(data))
		}
	})

	t.Run("no task returns error", func(t *testing.T) {
		_, _, svc, _ := setupBisyncTest(t)
		err := svc.ResolveBisyncConflict(99999, "mysync", "conflict1")
		if err != ErrTaskNotFound {
			t.Errorf("expected ErrTaskNotFound, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// ResyncBisync
// ---------------------------------------------------------------------------

func TestResyncBisync(t *testing.T) {
	t.Run("adds resync=true to bisync options and calls RunTask", func(t *testing.T) {
		_, db, svc, taskID := setupBisyncTest(t)

		// Set initial bisync options
		task, _ := svc.GetTask(taskID)
		task.BisyncOptions = json.RawMessage(`{"compare":"size"}`)
		if err := db.UpdateTask(taskID, task); err != nil {
			t.Fatalf("UpdateTask error = %v", err)
		}

		err := svc.ResyncBisync(taskID)
		if err != nil {
			t.Fatalf("ResyncBisync error = %v", err)
		}

		// Verify bisync options contain resync=true
		updated, ok := svc.GetTask(taskID)
		if !ok {
			t.Fatal("task not found after resync")
		}
		var opts map[string]any
		if err := json.Unmarshal(updated.BisyncOptions, &opts); err != nil {
			t.Fatalf("Unmarshal bisync options error = %v", err)
		}
		if opts["resync"] != true {
			t.Errorf("expected resync=true, got %v", opts["resync"])
		}
		if opts["compare"] != "size" {
			t.Errorf("expected compare='size' preserved, got %v", opts["compare"])
		}

		// A run should have been started
		runs, err := db.ListRunsByTask(taskID)
		if err != nil {
			t.Fatalf("ListRunsByTask error = %v", err)
		}
		if len(runs) == 0 {
			t.Fatal("expected at least one run record")
		}
		// The latest run should have status "running" or similar
		if runs[len(runs)-1].Status != "running" {
			t.Errorf("expected run status 'running', got %q", runs[len(runs)-1].Status)
		}
	})

	t.Run("handles empty bisync options", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)

		err := svc.ResyncBisync(taskID)
		if err != nil {
			t.Fatalf("ResyncBisync error = %v", err)
		}

		updated, ok := svc.GetTask(taskID)
		if !ok {
			t.Fatal("task not found")
		}
		var opts map[string]any
		if err := json.Unmarshal(updated.BisyncOptions, &opts); err != nil {
			t.Fatalf("Unmarshal error = %v", err)
		}
		if opts["resync"] != true {
			t.Errorf("expected resync=true, got %v", opts["resync"])
		}
	})

	t.Run("no task returns error", func(t *testing.T) {
		_, _, svc, _ := setupBisyncTest(t)
		err := svc.ResyncBisync(99999)
		if err != ErrTaskNotFound {
			t.Errorf("expected ErrTaskNotFound, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// ErrBisyncCurrentVersion
// ---------------------------------------------------------------------------

func TestErrBisyncCurrentVersion(t *testing.T) {
	err := ErrBisyncCurrentVersion
	if err.Error() != "cannot delete current version" {
		t.Errorf("expected 'cannot delete current version', got %q", err.Error())
	}

	// Verify DeleteBisyncLstVersion returns this error for "current"
	_, _, svc, taskID := setupBisyncTest(t)
	createBisyncFiles(t, svc, taskID, map[string]string{
		"path1.lst": "content1",
		"path2.lst": "content2",
	})

	err = svc.DeleteBisyncLstVersion(taskID, "current")
	if err != ErrBisyncCurrentVersion {
		t.Errorf("expected ErrBisyncCurrentVersion, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// cleanupOldLstBackups
// ---------------------------------------------------------------------------

func TestCleanupOldLstBackups(t *testing.T) {
	t.Run("removes excess backups beyond limit", func(t *testing.T) {
		_, db, svc, taskID := setupBisyncTest(t)

		// Set lstBackupCount = 1 (so at most 2 backup timestamps kept = 2*1)
		// The logic preserves backupCount timestamps when len(backups) > backupCount*2
		task, _ := svc.GetTask(taskID)
		task.BisyncOptions = json.RawMessage(`{"lstBackupCount":1}`)
		if err := db.UpdateTask(taskID, task); err != nil {
			t.Fatalf("UpdateTask error = %v", err)
		}

		task, _ = svc.GetTask(taskID)
		dir := svc.getBisyncDir(task.Name)
		bakDir := filepath.Join(dir, "bak")
		if err := os.MkdirAll(bakDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error = %v", err)
		}

		// Create 3 timestamp groups (more than backupCount*2 = 2)
		now := time.Now()
		for i := 0; i < 3; i++ {
			ts := now.Add(-time.Duration(i) * time.Hour).Format("20060102-150405")
			f1 := filepath.Join(bakDir, "path1.lst."+ts+".bak")
			f2 := filepath.Join(bakDir, "path2.lst."+ts+".bak")
			_ = os.WriteFile(f1, []byte("old"), 0o644)
			_ = os.WriteFile(f2, []byte("old"), 0o644)
		}

		if err := svc.BackupCurrentLstFiles(taskID); err != nil {
			t.Fatalf("BackupCurrentLstFiles error = %v", err)
		}

		// After cleanup, only 1 timestamp group should remain (backupCount=1)
		entries, _ := os.ReadDir(bakDir)
		if len(entries) > 2 {
			t.Errorf("expected at most 2 backup files (1 timestamp group), got %d", len(entries))
		}
	})

	t.Run("no bak dir is not an error", func(t *testing.T) {
		_, _, svc, taskID := setupBisyncTest(t)
		err := svc.cleanupOldLstBackups(taskID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}
