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
}