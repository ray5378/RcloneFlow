package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"rcloneflow/internal/service"
	"rcloneflow/internal/store"
)

type taskControllerRunSvcMock struct {
	runs []service.RunRecord
	err  error
}

func (m *taskControllerRunSvcMock) ListRuns(page, pageSize int) ([]service.RunRecord, int, error) {
	return m.runs, len(m.runs), m.err
}
func (m *taskControllerRunSvcMock) ListRunsByTask(taskId int64) ([]service.RunRecord, error) {
	out := make([]service.RunRecord, 0, len(m.runs))
	for _, r := range m.runs {
		if r.TaskID == taskId {
			out = append(out, r)
		}
	}
	return out, m.err
}
func (m *taskControllerRunSvcMock) ListActiveRuns() ([]service.RunRecord, error) {
	return m.runs, m.err
}
func (m *taskControllerRunSvcMock) GetActiveRunByTaskID(taskID int64) (service.RunRecord, error) {
	for _, r := range m.runs {
		if r.TaskID == taskID {
			return r, nil
		}
	}
	return service.RunRecord{}, m.err
}
func (m *taskControllerRunSvcMock) GetRun(id int64) (service.RunRecord, error) { return service.RunRecord{}, m.err }
func (m *taskControllerRunSvcMock) UpdateRun(id int64, updateFn func(*service.RunRecord)) {}
func (m *taskControllerRunSvcMock) DeleteRun(id int64) error                     { return m.err }
func (m *taskControllerRunSvcMock) DeleteAllRuns() error                         { return m.err }
func (m *taskControllerRunSvcMock) DeleteRunsByTask(taskId int64) error          { return m.err }
func (m *taskControllerRunSvcMock) CleanOldRuns(days int) (int64, error)         { return 0, m.err }

func newTaskControllerTestDB(t *testing.T) *store.DB {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "rcloneflow-task-controller-*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func newTaskControllerForDB(db *store.DB, runSvc service.RunServiceInterface) *TaskController {
	return NewTaskController(service.NewTaskService(db, nil), service.NewScheduleService(db), service.NewRunService(runSvc), nil)
}

func decodeJSONBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; body=%s", err, rec.Body.String())
	}
	return body
}

func TestTaskController_HandleTasks_CRUDAndPatch(t *testing.T) {
	db := newTaskControllerTestDB(t)
	ctrl := newTaskControllerForDB(db, &taskControllerRunSvcMock{})

	t.Run("GET returns tasks", func(t *testing.T) {
		_, err := db.AddTask(store.Task{
			Name:         "task-get",
			Mode:         "copy",
			SourceRemote: "src",
			SourcePath:   "/from",
			TargetRemote: "dst",
			TargetPath:   "/to",
		})
		if err != nil {
			t.Fatalf("AddTask() error = %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleTasks(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200", rec.Code)
		}
		var tasks []store.Task
		if err := json.Unmarshal(rec.Body.Bytes(), &tasks); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}
		if len(tasks) == 0 {
			t.Fatal("expected at least one task")
		}
	})

	t.Run("POST creates task", func(t *testing.T) {
		body := []byte(`{"name":"task-post","mode":"copy","sourceRemote":"src","sourcePath":"/from","targetRemote":"dst","targetPath":"/to"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		ctrl.HandleTasks(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
		resp := decodeJSONBody(t, rec)
		if resp["name"] != "task-post" {
			t.Fatalf("expected created task name task-post, got %#v", resp["name"])
		}
	})

	t.Run("POST duplicate name returns conflict", func(t *testing.T) {
		body := []byte(`{"name":"task-post","mode":"copy","sourceRemote":"src","sourcePath":"/from","targetRemote":"dst","targetPath":"/to"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		ctrl.HandleTasks(rec, req)
		if rec.Code != http.StatusConflict {
			t.Fatalf("status=%d, want 409 body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("PUT updates task", func(t *testing.T) {
		created, err := db.AddTask(store.Task{
			Name:         "task-put",
			Mode:         "copy",
			SourceRemote: "src",
			SourcePath:   "/from",
			TargetRemote: "dst",
			TargetPath:   "/to",
		})
		if err != nil {
			t.Fatalf("AddTask() error = %v", err)
		}
		body := []byte(`{"id":` + jsonNumber(created.ID) + `,"task":{"name":"task-put-updated","sourcePath":"/updated"}}`)
		req := httptest.NewRequest(http.MethodPut, "/api/tasks", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		ctrl.HandleTasks(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
		updated, ok := db.GetTask(created.ID)
		if !ok {
			t.Fatal("expected updated task to exist")
		}
		if updated.Name != "task-put-updated" || updated.SourcePath != "/updated" {
			t.Fatalf("unexpected updated task: %#v", updated)
		}
	})

	t.Run("PATCH updates options", func(t *testing.T) {
		created, err := db.AddTask(store.Task{
			Name:         "task-patch",
			Mode:         "copy",
			SourceRemote: "src",
			SourcePath:   "/from",
			TargetRemote: "dst",
			TargetPath:   "/to",
			Options:      json.RawMessage(`{"transfers":2}`),
		})
		if err != nil {
			t.Fatalf("AddTask() error = %v", err)
		}
		body := []byte(`{"id":` + jsonNumber(created.ID) + `,"options":{"checkers":8}}`)
		req := httptest.NewRequest(http.MethodPatch, "/api/tasks", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		ctrl.HandleTasks(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
		updated, ok := db.GetTask(created.ID)
		if !ok {
			t.Fatal("expected patched task to exist")
		}
		var opts map[string]any
		if err := json.Unmarshal(updated.Options, &opts); err != nil {
			t.Fatalf("json.Unmarshal(options) error = %v", err)
		}
		if opts["transfers"].(float64) != 2 || opts["checkers"].(float64) != 8 {
			t.Fatalf("unexpected merged options: %#v", opts)
		}
	})

	t.Run("PATCH reorder persists order", func(t *testing.T) {
		db := newTaskControllerTestDB(t)
		ctrl := newTaskControllerForDB(db, &taskControllerRunSvcMock{})

		first, err := db.AddTask(store.Task{
			Name:         "task-order-1",
			Mode:         "copy",
			SourceRemote: "src",
			SourcePath:   "/from-1",
			TargetRemote: "dst",
			TargetPath:   "/to-1",
		})
		if err != nil {
			t.Fatalf("AddTask(first) error = %v", err)
		}
		second, err := db.AddTask(store.Task{
			Name:         "task-order-2",
			Mode:         "copy",
			SourceRemote: "src",
			SourcePath:   "/from-2",
			TargetRemote: "dst",
			TargetPath:   "/to-2",
		})
		if err != nil {
			t.Fatalf("AddTask(second) error = %v", err)
		}
		third, err := db.AddTask(store.Task{
			Name:         "task-order-3",
			Mode:         "copy",
			SourceRemote: "src",
			SourcePath:   "/from-3",
			TargetRemote: "dst",
			TargetPath:   "/to-3",
		})
		if err != nil {
			t.Fatalf("AddTask(third) error = %v", err)
		}

		body := []byte(`{"order":[` + jsonNumber(third.ID) + `,` + jsonNumber(first.ID) + `,` + jsonNumber(second.ID) + `]}`)
		req := httptest.NewRequest(http.MethodPatch, "/api/tasks", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		ctrl.HandleTasks(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}

		tasks, err := db.ListTasks()
		if err != nil {
			t.Fatalf("ListTasks() error = %v", err)
		}
		gotIDs := []int64{tasks[0].ID, tasks[1].ID, tasks[2].ID}
		wantIDs := []int64{third.ID, first.ID, second.ID}
		for i := range wantIDs {
			if gotIDs[i] != wantIDs[i] {
				t.Fatalf("unexpected persisted order: got %v want %v", gotIDs, wantIDs)
			}
		}
	})

	t.Run("PATCH reorder appends missing ids", func(t *testing.T) {
		db := newTaskControllerTestDB(t)
		ctrl := newTaskControllerForDB(db, &taskControllerRunSvcMock{})

		first, err := db.AddTask(store.Task{Name: "task-partial-1", Mode: "copy", SourceRemote: "src", SourcePath: "/from-1", TargetRemote: "dst", TargetPath: "/to-1"})
		if err != nil {
			t.Fatalf("AddTask(first) error = %v", err)
		}
		second, err := db.AddTask(store.Task{Name: "task-partial-2", Mode: "copy", SourceRemote: "src", SourcePath: "/from-2", TargetRemote: "dst", TargetPath: "/to-2"})
		if err != nil {
			t.Fatalf("AddTask(second) error = %v", err)
		}
		third, err := db.AddTask(store.Task{Name: "task-partial-3", Mode: "copy", SourceRemote: "src", SourcePath: "/from-3", TargetRemote: "dst", TargetPath: "/to-3"})
		if err != nil {
			t.Fatalf("AddTask(third) error = %v", err)
		}
		fourth, err := db.AddTask(store.Task{Name: "task-partial-4", Mode: "copy", SourceRemote: "src", SourcePath: "/from-4", TargetRemote: "dst", TargetPath: "/to-4"})
		if err != nil {
			t.Fatalf("AddTask(fourth) error = %v", err)
		}

		body := []byte(`{"order":[` + jsonNumber(third.ID) + `,` + jsonNumber(first.ID) + `]}`)
		req := httptest.NewRequest(http.MethodPatch, "/api/tasks", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		ctrl.HandleTasks(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}

		tasks, err := db.ListTasks()
		if err != nil {
			t.Fatalf("ListTasks() error = %v", err)
		}
		gotIDs := []int64{tasks[0].ID, tasks[1].ID, tasks[2].ID, tasks[3].ID}
		wantIDs := []int64{third.ID, first.ID, second.ID, fourth.ID}
		for i := range wantIDs {
			if gotIDs[i] != wantIDs[i] {
				t.Fatalf("unexpected partial reorder result: got %v want %v", gotIDs, wantIDs)
			}
		}
	})
}

func TestTaskController_HandleTasks_ErrorBranches(t *testing.T) {
	db := newTaskControllerTestDB(t)
	ctrl := newTaskControllerForDB(db, &taskControllerRunSvcMock{})

	t.Run("POST invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString(`{"name":`))
		rec := httptest.NewRecorder()
		ctrl.HandleTasks(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status=%d, want 400", rec.Code)
		}
	})

	t.Run("PATCH invalid body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/tasks", bytes.NewBufferString(`{"id":`))
		rec := httptest.NewRecorder()
		ctrl.HandleTasks(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status=%d, want 400", rec.Code)
		}
	})

	t.Run("PATCH missing id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/tasks", bytes.NewBufferString(`{"options":{"transfers":1}}`))
		rec := httptest.NewRecorder()
		ctrl.HandleTasks(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status=%d, want 400 body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("PATCH not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/tasks", bytes.NewBufferString(`{"id":999,"options":{"transfers":1}}`))
		rec := httptest.NewRecorder()
		ctrl.HandleTasks(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want 500 body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodHead, "/api/tasks", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleTasks(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status=%d, want 405", rec.Code)
		}
	})
}

func TestTaskController_HandleBootstrap_BuildsActiveRunItems(t *testing.T) {
	db := newTaskControllerTestDB(t)
	_, err := db.AddTask(store.Task{
		Name:         "bootstrap-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/from",
		TargetRemote: "dst",
		TargetPath:   "/to",
	})
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}

	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "stderr.log")
	if err := os.WriteFile(logPath, []byte("2026/05/13 20:00:00 INFO  : file1.bin: Copied (new)\n2026/05/13 20:00:01 NOTICE: file2.bin: CAS compatible match after source cleanup\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(logPath) error = %v", err)
	}

	summary, _ := json.Marshal(map[string]any{
		"progress": map[string]any{
			"bytes":      float64(180),
			"totalBytes": float64(120),
			"speed":      float64(12),
			"eta":        float64(9),
		},
		"stderrFile": logPath,
		"finishWait": map[string]any{"enabled": true, "done": false},
		"preflight":  map[string]any{"totalCount": float64(10)},
	})

	ctrl := newTaskControllerForDB(db, &taskControllerRunSvcMock{runs: []service.RunRecord{{
		ID:        1,
		TaskID:    77,
		Status:    "running",
		StartedAt: time.Now().Add(-90 * time.Second).Format(time.RFC3339),
		Summary:   string(summary),
	}}})

	req := httptest.NewRequest(http.MethodGet, "/api/bootstrap", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleBootstrap(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
	}
	body := decodeJSONBody(t, rec)
	if _, ok := body["tasks"].([]any); !ok {
		t.Fatalf("expected tasks array, got %#v", body["tasks"])
	}
	activeRuns, ok := body["activeRuns"].([]any)
	if !ok || len(activeRuns) != 1 {
		t.Fatalf("expected one active run, got %#v", body["activeRuns"])
	}
	item, _ := activeRuns[0].(map[string]any)
	progress, _ := item["progress"].(map[string]any)
	if progress == nil {
		t.Fatalf("expected progress map, got %#v", item["progress"])
	}
	if got := progress["phase"]; got != "finalizing" {
		t.Fatalf("phase=%#v, want finalizing", got)
	}
	if got := progress["bytes"].(float64); got != 120 {
		t.Fatalf("bytes=%v, want clamped 120", got)
	}
	if got := progress["percentage"].(float64); got != 100 {
		t.Fatalf("percentage=%v, want clamped 100", got)
	}
	if got := progress["completedFiles"].(float64); got != 2 {
		t.Fatalf("completedFiles=%v, want 2 from log", got)
	}
	runRecord, _ := item["runRecord"].(map[string]any)
	if runRecord == nil {
		t.Fatalf("expected runRecord, got %#v", item["runRecord"])
	}
	if _, ok := runRecord["durationSeconds"]; !ok {
		t.Fatalf("expected durationSeconds in runRecord, got %#v", runRecord)
	}
	check, _ := item["progressCheck"].(map[string]any)
	if check == nil {
		t.Fatalf("expected progressCheck, got %#v", item["progressCheck"])
	}
	if ok, _ := check["ok"].(bool); !ok {
		t.Fatalf("expected progressCheck.ok=true, got %#v", check)
	}
}

func TestTaskController_HandleBootstrap_MethodAndErrorBranches(t *testing.T) {
	t.Run("method not allowed", func(t *testing.T) {
		db := newTaskControllerTestDB(t)
		ctrl := newTaskControllerForDB(db, &taskControllerRunSvcMock{})
		req := httptest.NewRequest(http.MethodPost, "/api/bootstrap", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleBootstrap(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status=%d, want 405", rec.Code)
		}
	})

	t.Run("active run error returns 500", func(t *testing.T) {
		db := newTaskControllerTestDB(t)
		ctrl := newTaskControllerForDB(db, &taskControllerRunSvcMock{err: os.ErrInvalid})
		req := httptest.NewRequest(http.MethodGet, "/api/bootstrap", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleBootstrap(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want 500 body=%s", rec.Code, rec.Body.String())
		}
	})
}

func TestTaskController_HandleTaskActions_DeleteAndRunBranches(t *testing.T) {
	t.Run("DELETE success", func(t *testing.T) {
		db := newTaskControllerTestDB(t)
		ctrl := newTaskControllerForDB(db, &taskControllerRunSvcMock{})
		created, err := db.AddTask(store.Task{
			Name:         "delete-ok",
			Mode:         "copy",
			SourceRemote: "src",
			SourcePath:   "/from",
			TargetRemote: "dst",
			TargetPath:   "/to",
		})
		if err != nil {
			t.Fatalf("AddTask() error = %v", err)
		}
		req := httptest.NewRequest(http.MethodDelete, "/api/tasks/"+jsonNumber(created.ID), nil)
		rec := httptest.NewRecorder()
		ctrl.HandleTaskActions(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
		body := decodeJSONBody(t, rec)
		if deleted, _ := body["deleted"].(bool); !deleted {
			t.Fatalf("expected deleted=true, got %#v", body)
		}
		if _, ok := db.GetTask(created.ID); ok {
			t.Fatal("expected task to be deleted")
		}
	})

	t.Run("DELETE invalid id", func(t *testing.T) {
		db := newTaskControllerTestDB(t)
		ctrl := newTaskControllerForDB(db, &taskControllerRunSvcMock{})
		req := httptest.NewRequest(http.MethodDelete, "/api/tasks/not-a-number", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleTaskActions(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status=%d, want 400 body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("DELETE not found returns 500", func(t *testing.T) {
		db := newTaskControllerTestDB(t)
		ctrl := newTaskControllerForDB(db, &taskControllerRunSvcMock{})
		req := httptest.NewRequest(http.MethodDelete, "/api/tasks/999", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleTaskActions(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want 500 body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("POST run already_running", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "rcloneflow-task-run-action-*")
		if err != nil {
			t.Fatalf("MkdirTemp() error = %v", err)
		}
		defer os.RemoveAll(tmpDir)
		oldAppDataDir := os.Getenv("APP_DATA_DIR")
		if err := os.Setenv("APP_DATA_DIR", tmpDir); err != nil {
			t.Fatalf("Setenv(APP_DATA_DIR) error = %v", err)
		}
		defer func() {
			if oldAppDataDir == "" {
				_ = os.Unsetenv("APP_DATA_DIR")
			} else {
				_ = os.Setenv("APP_DATA_DIR", oldAppDataDir)
			}
		}()

		db := newTaskControllerTestDB(t)
		ctrl := newTaskControllerForDB(db, &taskControllerRunSvcMock{})
		created, err := db.AddTask(store.Task{
			Name:         "already-running",
			Mode:         "copy",
			SourceRemote: "src",
			SourcePath:   "/from",
			TargetRemote: "dst",
			TargetPath:   "/to",
		})
		if err != nil {
			t.Fatalf("AddTask() error = %v", err)
		}
		if _, err := db.AddRun(store.Run{TaskID: created.ID, Status: "running", Trigger: "manual", Summary: map[string]any{}}); err != nil {
			t.Fatalf("AddRun() error = %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/"+jsonNumber(created.ID)+"/run", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleTaskActions(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
		body := decodeJSONBody(t, rec)
		if started, _ := body["started"].(bool); started {
			t.Fatalf("expected started=false, got %#v", body)
		}
		if body["reason"] != "already_running" {
			t.Fatalf("expected reason already_running, got %#v", body["reason"])
		}
	})

	t.Run("POST run task not found returns 500", func(t *testing.T) {
		db := newTaskControllerTestDB(t)
		ctrl := newTaskControllerForDB(db, &taskControllerRunSvcMock{})
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/999/run", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleTaskActions(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want 500 body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("non run path returns 404", func(t *testing.T) {
		db := newTaskControllerTestDB(t)
		ctrl := newTaskControllerForDB(db, &taskControllerRunSvcMock{})
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/123/not-run", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleTaskActions(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status=%d, want 404", rec.Code)
		}
	})
}

func jsonNumber(v int64) string {
	b, _ := json.Marshal(v)
	return string(b)
}
