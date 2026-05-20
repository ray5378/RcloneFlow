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

func TestRecalcTags_ActionTags(t *testing.T) {
	svc, db := setupTagTest(t)
	defer db.Close()

	if err := svc.RecalcTags(); err != nil {
		t.Fatalf("RecalcTags: %v", err)
	}

	tags, _ := svc.ListTags()
	tagSet := make(map[string]bool)
	for _, tg := range tags {
		tagSet[tg.Tag] = tg.Selected
	}

	if !tagSet["sync"] {
		t.Fatal("action tag sync should be selected")
	}
	if !tagSet["copy"] {
		t.Fatal("action tag copy should be selected")
	}
	if !tagSet["move"] {
		t.Fatal("action tag move should be selected")
	}
}

func TestRecalcTags_AutoKeywords(t *testing.T) {
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
		if tg.Tag == "备份" && !tg.Selected {
			hasBackup = true
		}
		if tg.Tag == "阿里云" && !tg.Selected {
			hasAliyun = true
		}
		if tg.Tag == "nas" && !tg.Selected {
			hasNAS = true
		}
	}
	if !hasBackup {
		t.Fatal("expected auto keyword tag 备份 (not selected)")
	}
	if !hasAliyun {
		t.Fatal("expected auto keyword tag 阿里云 (not selected)")
	}
	if !hasNAS {
		t.Fatal("expected auto keyword tag nas (not selected)")
	}
}

func TestSelectTag(t *testing.T) {
	svc, db := setupTagTest(t)
	defer db.Close()

	svc.RecalcTags()

	if err := svc.SelectTag("备份", true); err != nil {
		t.Fatalf("SelectTag: %v", err)
	}

	tags, _ := svc.ListTags()
	for _, tg := range tags {
		if tg.Tag == "备份" && tg.Selected {
			return
		}
	}
	t.Fatal("备份 should be selected after SelectTag true")
}

func TestCreateAndDeleteManualTag(t *testing.T) {
	svc, db := setupTagTest(t)
	defer db.Close()

	if err := svc.CreateManualTag("我的标签"); err != nil {
		t.Fatalf("CreateManualTag: %v", err)
	}

	tags, _ := svc.ListTags()
	found := false
	for _, tg := range tags {
		if tg.Tag == "我的标签" && tg.Selected && tg.Type == "keyword" {
			found = true
		}
	}
	if !found {
		t.Fatal("manual tag not found or wrong state")
	}

	if err := svc.DeleteManualTag("我的标签"); err != nil {
		t.Fatalf("DeleteManualTag: %v", err)
	}

	tags, _ = svc.ListTags()
	for _, tg := range tags {
		if tg.Tag == "我的标签" {
			t.Fatal("manual tag should be deleted")
		}
	}
}