package service

import (
	"testing"

	"rcloneflow/internal/store"
)

func setupServiceDB(t *testing.T) *store.DB {
	t.Helper()
	tmpDir := t.TempDir()
	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestTagService_ExtractEnglishTokens(t *testing.T) {
	db := setupServiceDB(t)
	svc := NewTagService(db)

	tokens := svc.extractEnglishTokens("copy-files-to-backup")
	found := make(map[string]bool)
	for _, t := range tokens {
		found[t.token] = true
	}

	if !found["copy"] {
		t.Error("expected token 'copy'")
	}
	if !found["files"] {
		t.Error("expected token 'files'")
	}
	if !found["backup"] {
		t.Error("expected token 'backup'")
	}
}

func TestTagService_ExtractChineseTokens(t *testing.T) {
	db := setupServiceDB(t)
	svc := NewTagService(db)

	tokens := svc.extractChineseTokens("电影备份")
	found := make(map[string]bool)
	for _, t := range tokens {
		found[t.token] = true
	}

	if !found["电影"] {
		t.Error("expected token '电影'")
	}
	if !found["备份"] {
		t.Error("expected token '备份'")
	}
}

func TestTagService_ExtractTokens_Mixed(t *testing.T) {
	db := setupServiceDB(t)
	svc := NewTagService(db)

	tokens := svc.extractTokens("backup-电影")
	found := make(map[string]bool)
	for _, t := range tokens {
		found[t.token] = true
	}

	if !found["backup"] {
		t.Error("expected token 'backup'")
	}
	if !found["电影"] {
		t.Error("expected token '电影'")
	}
}

func TestTagService_RecalcTags(t *testing.T) {
	db := setupServiceDB(t)
	svc := NewTagService(db)

	// Create tasks with common tokens
	_, _ = db.AddTask(store.Task{Name: "copy-files", Mode: "copy"})
	_, _ = db.AddTask(store.Task{Name: "copy-backup", Mode: "copy"})
	_, _ = db.AddTask(store.Task{Name: "move-files", Mode: "move"})

	err := svc.RecalcTags()
	if err != nil {
		t.Fatalf("RecalcTags() error = %v", err)
	}

	tags, err := svc.ListTags()
	if err != nil {
		t.Fatalf("ListTags() error = %v", err)
	}

	// Should have action tags (copy, move, sync) and keyword tags for common tokens
	tagNames := make(map[string]bool)
	for _, tag := range tags {
		tagNames[tag.Tag] = true
	}

	if !tagNames["copy"] {
		t.Error("expected 'copy' tag")
	}
	if !tagNames["move"] {
		t.Error("expected 'move' tag")
	}
	if !tagNames["sync"] {
		t.Error("expected 'sync' tag")
	}
}

func TestTagService_CreateManualTag(t *testing.T) {
	db := setupServiceDB(t)
	svc := NewTagService(db)

	err := svc.CreateManualTag("important")
	if err != nil {
		t.Fatalf("CreateManualTag() error = %v", err)
	}

	tags, _ := svc.ListTags()
	found := false
	for _, tag := range tags {
		if tag.Tag == "important" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'important' tag")
	}
}

func TestTagService_SelectTag(t *testing.T) {
	db := setupServiceDB(t)
	svc := NewTagService(db)

	_ = svc.CreateManualTag("test-tag")

	err := svc.SelectTag("test-tag", false)
	if err != nil {
		t.Fatalf("SelectTag() error = %v", err)
	}

	tags, _ := svc.ListTags()
	for _, tag := range tags {
		if tag.Tag == "test-tag" {
			if tag.Selected {
				t.Error("tag should not be selected")
			}
			break
		}
	}
}

func TestTagService_DeleteManualTag(t *testing.T) {
	db := setupServiceDB(t)
	svc := NewTagService(db)

	_ = svc.CreateManualTag("to-delete")
	err := svc.DeleteManualTag("to-delete")
	if err != nil {
		t.Fatalf("DeleteManualTag() error = %v", err)
	}

	tags, _ := svc.ListTags()
	for _, tag := range tags {
		if tag.Tag == "to-delete" {
			t.Error("tag should have been deleted")
		}
	}
}

func TestContainsChinese(t *testing.T) {
	if !containsChinese("电影") {
		t.Error("containsChinese('电影') should be true")
	}
	if containsChinese("movie") {
		t.Error("containsChinese('movie') should be false")
	}
	if containsChinese("123") {
		t.Error("containsChinese('123') should be false")
	}
}

func TestCleanupService_Replan(t *testing.T) {
	db := setupServiceDB(t)
	runSvc := NewRunService(NewStoreRunAdapter(db))

	svc := NewCleanupService(runSvc, 24, 7)
	svc.Replan(12, 30)

	// Verify replan doesn't panic and updates interval (stored as time.Duration)
	if svc.interval != 12*3600*1000000000 {
		t.Errorf("interval = %d, want %d", svc.interval, 12*3600*1000000000)
	}
}
