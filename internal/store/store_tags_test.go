package store

import (
	"os"
	"testing"
)

func openTestDB(t *testing.T) *DB {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tags_test_*")
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

func TestListTags_Empty(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	tags, err := db.ListTags()
	if err != nil {
		t.Fatalf("ListTags error: %v", err)
	}
	if len(tags) != 0 {
		t.Fatalf("expected 0 tags, got %d", len(tags))
	}
}

func TestMergeTags_ThenListTags(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	entries := []TagEntry{
		{Tag: "sync", Type: "action", Selected: true},
		{Tag: "copy", Type: "action", Selected: true},
		{Tag: "move", Type: "action", Selected: true},
		{Tag: "备份", Type: "keyword", Selected: false},
		{Tag: "backup", Type: "keyword", Selected: false},
	}

	if err := db.MergeTags(entries); err != nil {
		t.Fatalf("MergeTags error: %v", err)
	}

	tags, err := db.ListTags()
	if err != nil {
		t.Fatalf("ListTags error: %v", err)
	}
	if len(tags) != 5 {
		t.Fatalf("expected 5 tags, got %d", len(tags))
	}

	for _, tg := range tags {
		switch tg.Tag {
		case "sync", "copy", "move":
			if tg.Type != "action" || !tg.Selected {
				t.Fatalf("action tag %s: type=%s selected=%v", tg.Tag, tg.Type, tg.Selected)
			}
		case "备份", "backup":
			if tg.Type != "keyword" || tg.Selected {
				t.Fatalf("keyword tag %s: type=%s selected=%v", tg.Tag, tg.Type, tg.Selected)
			}
		}
	}
}

func TestMergeTags_ResetsSelected(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	db.MergeTags([]TagEntry{
		{Tag: "备份", Type: "keyword", Selected: false},
	})

	db.SetTagSelected("备份", true)

	db.MergeTags([]TagEntry{
		{Tag: "备份", Type: "keyword", Selected: false},
		{Tag: "新的关键词", Type: "keyword", Selected: false},
	})

	tags, _ := db.ListTags()
	for _, tg := range tags {
		if tg.Tag == "备份" && tg.Selected {
			t.Fatal("备份 should be reset to selected=false after MergeTags (all keywords are re-derived)")
		}
	}
}

func TestSetTagSelected(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	db.CreateManualTag("mytag")

	if err := db.SetTagSelected("mytag", false); err != nil {
		t.Fatalf("SetTagSelected false: %v", err)
	}
	tags, _ := db.ListTags()
	for _, tg := range tags {
		if tg.Tag == "mytag" && tg.Selected {
			t.Fatal("mytag should be unselected")
		}
	}

	if err := db.SetTagSelected("mytag", true); err != nil {
		t.Fatalf("SetTagSelected true: %v", err)
	}
	tags, _ = db.ListTags()
	for _, tg := range tags {
		if tg.Tag == "mytag" && !tg.Selected {
			t.Fatal("mytag should be selected")
		}
	}
}

func TestCreateManualTag(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	if err := db.CreateManualTag("手动标签"); err != nil {
		t.Fatalf("CreateManualTag: %v", err)
	}

	tags, _ := db.ListTags()
	for _, tg := range tags {
		if tg.Tag == "手动标签" {
			if tg.Type != "keyword" || !tg.Selected {
				t.Fatal("manual tag should be keyword with selected=true")
			}
			return
		}
	}
	t.Fatal("manual tag not found")
}

func TestCreateManualTag_Duplicate(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	db.CreateManualTag("dup")
	db.CreateManualTag("dup")

	tags, _ := db.ListTags()
	count := 0
	for _, tg := range tags {
		if tg.Tag == "dup" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected 1 tag, got %d", count)
	}
}

func TestDeleteManualTag(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	db.CreateManualTag("todelete")
	db.DeleteManualTag("todelete")

	tags, _ := db.ListTags()
	for _, tg := range tags {
		if tg.Tag == "todelete" {
			t.Fatal("tag should be deleted")
		}
	}
}

func TestMergeTags_OverwriteAutoKeywords(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	db.MergeTags([]TagEntry{
		{Tag: "old_kw", Type: "keyword", Selected: false},
	})

	db.MergeTags([]TagEntry{
		{Tag: "new_kw", Type: "keyword", Selected: false},
	})

	tags, _ := db.ListTags()
	foundOld := false
	foundNew := false
	for _, tg := range tags {
		if tg.Tag == "old_kw" {
			foundOld = true
		}
		if tg.Tag == "new_kw" {
			foundNew = true
		}
	}
	if foundOld {
		t.Fatal("old keyword should be removed")
	}
	if !foundNew {
		t.Fatal("new keyword should exist")
	}
}