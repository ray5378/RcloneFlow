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

func TestReplaceTags_Then_ListTags(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	input := map[string]string{
		"sync":   "action",
		"copy":   "action",
		"备份":    "keyword",
		"Backup": "keyword",
	}

	if err := db.ReplaceTags(input); err != nil {
		t.Fatalf("ReplaceTags error: %v", err)
	}

	tags, err := db.ListTags()
	if err != nil {
		t.Fatalf("ListTags error: %v", err)
	}
	if len(tags) != 4 {
		t.Fatalf("expected 4 tags, got %d", len(tags))
	}

	found := make(map[string]string)
	for _, tg := range tags {
		found[tg.Tag] = tg.Type
	}
	for k, v := range input {
		if found[k] != v {
			t.Fatalf("expected tag %s type %s, got %s", k, v, found[k])
		}
	}
}

func TestReplaceTags_Overwrite(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	db.ReplaceTags(map[string]string{"sync": "action", "backup": "keyword"})
	db.ReplaceTags(map[string]string{"copy": "action"})

	tags, _ := db.ListTags()
	if len(tags) != 1 {
		t.Fatalf("expected 1 tag after overwrite, got %d", len(tags))
	}
	if tags[0].Tag != "copy" {
		t.Fatalf("expected tag 'copy', got '%s'", tags[0].Tag)
	}
}