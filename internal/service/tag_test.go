package service

import (
	"os"
	"testing"

	"rcloneflow/internal/store"
)

func setupTagTest(t *testing.T) (*TagService, *store.DB) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tag_svc_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	svc := NewTagService(db)

	tasks := []store.Task{
		{Name: "每日备份到阿里云"},
		{Name: "每周备份到阿里云"},
		{Name: "同步照片到GoogleDrive"},
		{Name: "Daily Backup NAS"},
		{Name: "Weekly Backup NAS"},
		{Name: "Move Old Logs"},
	}
	for _, tk := range tasks {
		_, err := db.AddTask(tk)
		if err != nil {
			t.Fatalf("create task: %v", err)
		}
	}

	return svc, db
}

func TestRecalcTags_KeywordTags(t *testing.T) {
	svc, db := setupTagTest(t)
	defer db.Close()

	if err := svc.RecalcTags(); err != nil {
		t.Fatalf("RecalcTags: %v", err)
	}

	tags, _ := svc.ListTags()
	tagSet := make(map[string]string)
	for _, tg := range tags {
		tagSet[tg.Tag] = tg.Type
	}

	if tagSet["sync"] != "action" {
		t.Fatal("missing action tag sync")
	}
	if tagSet["copy"] != "action" {
		t.Fatal("missing action tag copy")
	}
	if tagSet["move"] != "action" {
		t.Fatal("missing action tag move")
	}
	if tagSet["备份"] != "keyword" {
		t.Fatal("missing keyword tag 备份")
	}
	if tagSet["阿里云"] != "keyword" {
		t.Fatalf("missing keyword tag 阿里云, got tags: %v", tagSet)
	}
	if tagSet["backup"] != "keyword" {
		t.Fatalf("missing keyword tag backup, got tags: %v", tagSet)
	}
	if tagSet["nas"] != "keyword" {
		t.Fatalf("missing keyword tag nas, got tags: %v", tagSet)
	}
}

func TestRecalcTags_NoDuplicateShortTokens(t *testing.T) {
	svc, db := setupTagTest(t)
	defer db.Close()

	if err := svc.RecalcTags(); err != nil {
		t.Fatalf("RecalcTags: %v", err)
	}

	tags, _ := svc.ListTags()

	hasBackup := false
	hasAliyun := false
	hasNAS := false
	for _, tg := range tags {
		if tg.Tag == "备份" {
			hasBackup = true
		}
		if tg.Tag == "阿里云" {
			hasAliyun = true
		}
		if tg.Tag == "nas" {
			hasNAS = true
		}
	}
	if !hasBackup {
		t.Fatal("expected keyword tag 备份")
	}
	if !hasAliyun {
		t.Fatal("expected keyword tag 阿里云")
	}
	if !hasNAS {
		t.Fatal("expected keyword tag nas")
	}
}