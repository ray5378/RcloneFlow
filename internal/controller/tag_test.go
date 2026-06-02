package controller

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"rcloneflow/internal/service"
	"rcloneflow/internal/store"
)

func TestHandleTags(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tag_ctrl_test_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	defer db.Close()

	tagSvc := service.NewTagService(db)
	db.AddTask(store.Task{Name: "备份数据库"})
	db.AddTask(store.Task{Name: "备份照片"})
	tagSvc.RecalcTags()

	ctrl := NewTagController(tagSvc)

	req := httptest.NewRequest("GET", "/api/tags", nil)
	w := httptest.NewRecorder()
	ctrl.HandleTags(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	if !strings.Contains(body, `"action"`) {
		t.Fatal("response should contain action tags")
	}
	if !strings.Contains(body, `"sync"`) || !strings.Contains(body, `"copy"`) || !strings.Contains(body, `"move"`) {
		t.Fatal("response should contain sync/copy/move tags")
	}
	if !strings.Contains(body, `"selected"`) {
		t.Fatal("response should contain selected field")
	}
}

func TestHandleSelectTag(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tag_ctrl_sel_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	defer db.Close()

	tagSvc := service.NewTagService(db)
	db.AddTask(store.Task{Name: "备份数据库"})
	db.AddTask(store.Task{Name: "备份照片"})
	tagSvc.RecalcTags()

	ctrl := NewTagController(tagSvc)

	req := httptest.NewRequest("POST", "/api/tags/select", strings.NewReader(`{"tag":"备份","selected":true}`))
	w := httptest.NewRecorder()
	ctrl.HandleTagActions(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	tags, _ := tagSvc.ListTags()
	for _, tg := range tags {
		if tg.Tag == "备份" && !tg.Selected {
			t.Fatal("备份 should be selected after select call")
		}
	}
}

func TestHandleCreateTag(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tag_ctrl_create_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	defer db.Close()

	tagSvc := service.NewTagService(db)
	ctrl := NewTagController(tagSvc)

	req := httptest.NewRequest("POST", "/api/tags/create", strings.NewReader(`{"tag":"自定义标签"}`))
	w := httptest.NewRecorder()
	ctrl.HandleTagActions(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	tags, _ := tagSvc.ListTags()
	found := false
	for _, tg := range tags {
		if tg.Tag == "自定义标签" && tg.Selected {
			found = true
		}
	}
	if !found {
		t.Fatal("自定义标签 should exist and be selected")
	}
}

func TestHandleDeleteTag(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tag_ctrl_del_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	defer db.Close()

	tagSvc := service.NewTagService(db)
	tagSvc.CreateManualTag("temp_tag")
	ctrl := NewTagController(tagSvc)

	req := httptest.NewRequest("POST", "/api/tags/delete", strings.NewReader(`{"tag":"temp_tag"}`))
	w := httptest.NewRecorder()
	ctrl.HandleTagActions(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	tags, _ := tagSvc.ListTags()
	for _, tg := range tags {
		if tg.Tag == "temp_tag" {
			t.Fatal("temp_tag should be deleted")
		}
	}
}
