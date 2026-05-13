package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"rcloneflow/internal/service"
)

type runStatusLogSvcMock struct {
	runs          []service.RunRecord
	listErr       error
	deleteErr     error
	deleteCalled  []int64
	deleteAllErr  error
	deleteTaskErr error
}

func (m *runStatusLogSvcMock) ListRuns(page, pageSize int) ([]service.RunRecord, int, error) {
	return m.runs, len(m.runs), m.listErr
}
func (m *runStatusLogSvcMock) ListRunsByTask(taskId int64) ([]service.RunRecord, error) {
	out := make([]service.RunRecord, 0, len(m.runs))
	for _, r := range m.runs {
		if r.TaskID == taskId {
			out = append(out, r)
		}
	}
	return out, m.listErr
}
func (m *runStatusLogSvcMock) ListActiveRuns() ([]service.RunRecord, error) {
	return m.runs, m.listErr
}
func (m *runStatusLogSvcMock) GetActiveRunByTaskID(taskID int64) (service.RunRecord, error) {
	return service.RunRecord{}, m.listErr
}
func (m *runStatusLogSvcMock) GetRun(id int64) (service.RunRecord, error) { return service.RunRecord{}, m.listErr }
func (m *runStatusLogSvcMock) UpdateRun(id int64, updateFn func(*service.RunRecord)) {}
func (m *runStatusLogSvcMock) DeleteRun(id int64) error {
	m.deleteCalled = append(m.deleteCalled, id)
	return m.deleteErr
}
func (m *runStatusLogSvcMock) DeleteAllRuns() error                 { return m.deleteAllErr }
func (m *runStatusLogSvcMock) DeleteRunsByTask(taskId int64) error  { return m.deleteTaskErr }
func (m *runStatusLogSvcMock) CleanOldRuns(days int) (int64, error) { return 0, nil }

func TestRunController_HandleRunStatus_GetDeleteAndErrors(t *testing.T) {
	t.Run("GET returns matching run", func(t *testing.T) {
		summary, _ := json.Marshal(map[string]any{
			"finishedAt": "2026-05-13T20:05:00+08:00",
			"finalSummary": map[string]any{
				"durationSec":  float64(300),
				"durationText": "5分",
				"counts": map[string]any{
					"copied": float64(2),
					"total":  float64(2),
				},
			},
		})
		mock := &runStatusLogSvcMock{runs: []service.RunRecord{{
			ID:        42,
			TaskID:    7,
			Status:    "finished",
			Trigger:   "manual",
			StartedAt: "2026-05-13T20:00:00+08:00",
			Summary:   string(summary),
		}}}
		ctrl := NewRunController(service.NewRunService(mock), nil)
		req := httptest.NewRequest(http.MethodGet, "/api/runs/42", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRunStatus(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}
		if int(body["id"].(float64)) != 42 {
			t.Fatalf("id=%v, want 42", body["id"])
		}
		sum, _ := body["summary"].(map[string]any)
		if sum == nil {
			t.Fatalf("expected summary, got %#v", body["summary"])
		}
		fs, _ := sum["finalSummary"].(map[string]any)
		if fs == nil || fs["durationText"] != "5分" {
			t.Fatalf("expected finalSummary.durationText=5分, got %#v", fs)
		}
	})

	t.Run("GET run not found", func(t *testing.T) {
		ctrl := NewRunController(service.NewRunService(&runStatusLogSvcMock{}), nil)
		req := httptest.NewRequest(http.MethodGet, "/api/runs/404", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRunStatus(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status=%d, want 404 body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("GET list error", func(t *testing.T) {
		ctrl := NewRunController(service.NewRunService(&runStatusLogSvcMock{listErr: errors.New("boom")}), nil)
		req := httptest.NewRequest(http.MethodGet, "/api/runs/1", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRunStatus(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want 500 body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("DELETE success", func(t *testing.T) {
		mock := &runStatusLogSvcMock{}
		ctrl := NewRunController(service.NewRunService(mock), nil)
		req := httptest.NewRequest(http.MethodDelete, "/api/runs/55", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRunStatus(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
		if len(mock.deleteCalled) != 1 || mock.deleteCalled[0] != 55 {
			t.Fatalf("deleteCalled=%v, want [55]", mock.deleteCalled)
		}
	})

	t.Run("DELETE error", func(t *testing.T) {
		mock := &runStatusLogSvcMock{deleteErr: errors.New("delete failed")}
		ctrl := NewRunController(service.NewRunService(mock), nil)
		req := httptest.NewRequest(http.MethodDelete, "/api/runs/66", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRunStatus(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want 500 body=%s", rec.Code, rec.Body.String())
		}
	})
}

func TestRunController_HandleRunLog_SummaryFallbackAndErrors(t *testing.T) {
	t.Run("serves stderrFile from summary string", func(t *testing.T) {
		logPath := filepath.Join(t.TempDir(), "run-1.log")
		if err := os.WriteFile(logPath, []byte("hello-log\n"), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
		summary, _ := json.Marshal(map[string]any{"stderrFile": logPath})
		mock := &runStatusLogSvcMock{runs: []service.RunRecord{{ID: 1, Summary: string(summary)}}}
		ctrl := NewRunController(service.NewRunService(mock), nil)
		req := httptest.NewRequest(http.MethodGet, "/api/runs/1/log", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRunLog(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
		if rec.Body.String() != "hello-log\n" {
			t.Fatalf("unexpected body=%q", rec.Body.String())
		}
	})

	t.Run("serves stderrFile from summary map", func(t *testing.T) {
		logPath := filepath.Join(t.TempDir(), "run-2.log")
		if err := os.WriteFile(logPath, []byte("map-log\n"), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
		mock := &runStatusLogSvcMock{runs: []service.RunRecord{{ID: 2, Summary: "", TaskName: "unused"}}}
		mock.runs[0].Summary = string(mustJSON(map[string]any{"stderrFile": logPath}))
		ctrl := NewRunController(service.NewRunService(mock), nil)
		req := httptest.NewRequest(http.MethodGet, "/api/runs/2/log", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRunLog(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
		if rec.Body.String() != "map-log\n" {
			t.Fatalf("unexpected body=%q", rec.Body.String())
		}
	})

	t.Run("falls back to latest task log directory", func(t *testing.T) {
		base := "/app/data/logs"
		taskDir := filepath.Join(base, "测试任务-0513")
		if err := os.MkdirAll(taskDir, 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		logPath := filepath.Join(taskDir, "2141.log")
		if err := os.WriteFile(logPath, []byte("fallback-log\n"), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
		defer os.Remove(logPath)
		defer os.Remove(taskDir)

		mock := &runStatusLogSvcMock{runs: []service.RunRecord{{
			ID:        3,
			TaskName:  "测试任务",
			StartedAt: time.Now().Format(time.RFC3339),
		}}}
		ctrl := NewRunController(service.NewRunService(mock), nil)
		req := httptest.NewRequest(http.MethodGet, "/api/runs/3/log?auth=token", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRunLog(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
		if rec.Body.String() != "fallback-log\n" {
			t.Fatalf("unexpected body=%q", rec.Body.String())
		}
	})

	t.Run("list error returns 500", func(t *testing.T) {
		ctrl := NewRunController(service.NewRunService(&runStatusLogSvcMock{listErr: errors.New("boom")}), nil)
		req := httptest.NewRequest(http.MethodGet, "/api/runs/9/log", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRunLog(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want 500 body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("run not found returns 404", func(t *testing.T) {
		ctrl := NewRunController(service.NewRunService(&runStatusLogSvcMock{}), nil)
		req := httptest.NewRequest(http.MethodGet, "/api/runs/999/log", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRunLog(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status=%d, want 404 body=%s", rec.Code, rec.Body.String())
		}
	})
}

func TestRunController_HandleRuns_ListDeletePagingAndErrors(t *testing.T) {
	t.Run("GET uses paging and returns runs", func(t *testing.T) {
		summary, _ := json.Marshal(map[string]any{
			"finalSummary": map[string]any{
				"durationSec":  float64(61),
				"durationText": "1分1秒",
				"counts": map[string]any{
					"copied": float64(1),
					"total":  float64(1),
				},
			},
		})
		mock := &runStatusLogSvcMock{runs: []service.RunRecord{{
			ID:        10,
			TaskID:    2,
			Status:    "finished",
			StartedAt: "2026-05-13T20:00:00+08:00",
			Summary:   string(summary),
		}}}
		ctrl := NewRunController(service.NewRunService(mock), nil)
		req := httptest.NewRequest(http.MethodGet, "/api/runs?page=2&pageSize=20", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRuns(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}
		if int(body["page"].(float64)) != 2 || int(body["pageSize"].(float64)) != 20 {
			t.Fatalf("unexpected paging body=%#v", body)
		}
		runs, _ := body["runs"].([]any)
		if len(runs) != 1 {
			t.Fatalf("len(runs)=%d, want 1", len(runs))
		}
	})

	t.Run("GET falls back to default paging on invalid query", func(t *testing.T) {
		mock := &runStatusLogSvcMock{}
		ctrl := NewRunController(service.NewRunService(mock), nil)
		req := httptest.NewRequest(http.MethodGet, "/api/runs?page=0&pageSize=999", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRuns(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}
		if int(body["page"].(float64)) != 1 || int(body["pageSize"].(float64)) != 50 {
			t.Fatalf("expected default paging, got %#v", body)
		}
	})

	t.Run("GET list error", func(t *testing.T) {
		ctrl := NewRunController(service.NewRunService(&runStatusLogSvcMock{listErr: errors.New("boom")}), nil)
		req := httptest.NewRequest(http.MethodGet, "/api/runs", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRuns(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want 500 body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("DELETE success", func(t *testing.T) {
		mock := &runStatusLogSvcMock{}
		ctrl := NewRunController(service.NewRunService(mock), nil)
		req := httptest.NewRequest(http.MethodDelete, "/api/runs", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRuns(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("DELETE error", func(t *testing.T) {
		mock := &runStatusLogSvcMock{deleteAllErr: errors.New("delete all failed")}
		ctrl := NewRunController(service.NewRunService(mock), nil)
		req := httptest.NewRequest(http.MethodDelete, "/api/runs", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRuns(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want 500 body=%s", rec.Code, rec.Body.String())
		}
	})
}

func TestRunController_HandleRunsByTask_GetDeleteAndErrors(t *testing.T) {
	t.Run("GET returns filtered task runs", func(t *testing.T) {
		summary, _ := json.Marshal(map[string]any{
			"finalSummary": map[string]any{
				"durationSec":  float64(12),
				"durationText": "12秒",
				"counts": map[string]any{
					"copied": float64(2),
					"total":  float64(2),
				},
			},
		})
		mock := &runStatusLogSvcMock{runs: []service.RunRecord{{ID: 21, TaskID: 7, Status: "finished", Summary: string(summary)}, {ID: 22, TaskID: 8, Status: "failed"}}}
		ctrl := NewRunController(service.NewRunService(mock), nil)
		req := httptest.NewRequest(http.MethodGet, "/api/runs/task/7", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRunsByTask(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
		var body []map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}
		if len(body) != 1 || int(body[0]["id"].(float64)) != 21 {
			t.Fatalf("unexpected body=%#v", body)
		}
	})

	t.Run("GET error", func(t *testing.T) {
		ctrl := NewRunController(service.NewRunService(&runStatusLogSvcMock{listErr: errors.New("boom")}), nil)
		req := httptest.NewRequest(http.MethodGet, "/api/runs/task/7", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRunsByTask(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want 500 body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("DELETE success", func(t *testing.T) {
		mock := &runStatusLogSvcMock{}
		ctrl := NewRunController(service.NewRunService(mock), nil)
		req := httptest.NewRequest(http.MethodDelete, "/api/runs/task/7", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRunsByTask(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("DELETE error", func(t *testing.T) {
		mock := &runStatusLogSvcMock{deleteTaskErr: errors.New("delete task failed")}
		ctrl := NewRunController(service.NewRunService(mock), nil)
		req := httptest.NewRequest(http.MethodDelete, "/api/runs/task/7", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRunsByTask(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want 500 body=%s", rec.Code, rec.Body.String())
		}
	})
}

func TestRunController_HandleRunKillCLI_Branches(t *testing.T) {
	t.Run("method not allowed", func(t *testing.T) {
		ctrl := NewRunController(service.NewRunService(&runStatusLogSvcMock{}), nil)
		req := httptest.NewRequest(http.MethodGet, "/api/runs/1/kill", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRunKillCLI(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status=%d, want 405", rec.Code)
		}
	})

	t.Run("list error", func(t *testing.T) {
		ctrl := NewRunController(service.NewRunService(&runStatusLogSvcMock{listErr: errors.New("boom")}), nil)
		req := httptest.NewRequest(http.MethodPost, "/api/runs/1/kill", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRunKillCLI(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want 500 body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("not found or no pid", func(t *testing.T) {
		mock := &runStatusLogSvcMock{runs: []service.RunRecord{{ID: 2, Summary: string(mustJSON(map[string]any{"foo": "bar"}))}}}
		ctrl := NewRunController(service.NewRunService(mock), nil)
		req := httptest.NewRequest(http.MethodPost, "/api/runs/2/kill", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRunKillCLI(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status=%d, want 404 body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("success", func(t *testing.T) {
		cmd := exec.Command("sleep", "30")
		if err := cmd.Start(); err != nil {
			t.Fatalf("start sleep: %v", err)
		}
		defer func() { _ = cmd.Process.Kill(); _, _ = cmd.Process.Wait() }()
		mock := &runStatusLogSvcMock{runs: []service.RunRecord{{ID: 3, Summary: string(mustJSON(map[string]any{"pid": cmd.Process.Pid}))}}}
		ctrl := NewRunController(service.NewRunService(mock), nil)
		req := httptest.NewRequest(http.MethodPost, "/api/runs/3/kill", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRunKillCLI(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
	})
}

func TestRunController_HandleTaskKill_Branches(t *testing.T) {
	t.Run("method not allowed", func(t *testing.T) {
		ctrl := NewRunController(service.NewRunService(&runStatusLogSvcMock{}), nil)
		req := httptest.NewRequest(http.MethodGet, "/api/tasks/7/kill", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleTaskKill(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status=%d, want 405", rec.Code)
		}
	})

	t.Run("invalid task id", func(t *testing.T) {
		ctrl := NewRunController(service.NewRunService(&runStatusLogSvcMock{}), nil)
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/not-a-number/kill", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleTaskKill(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status=%d, want 400 body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("list by task error", func(t *testing.T) {
		ctrl := NewRunController(service.NewRunService(&runStatusLogSvcMock{listErr: errors.New("boom")}), nil)
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/7/kill", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleTaskKill(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want 500 body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("no runs for task", func(t *testing.T) {
		ctrl := NewRunController(service.NewRunService(&runStatusLogSvcMock{}), nil)
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/7/kill", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleTaskKill(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status=%d, want 404 body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("pid not found", func(t *testing.T) {
		mock := &runStatusLogSvcMock{runs: []service.RunRecord{{ID: 11, TaskID: 7, Status: "running", Summary: string(mustJSON(map[string]any{"x": 1}))}}}
		ctrl := NewRunController(service.NewRunService(mock), nil)
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/7/kill", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleTaskKill(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status=%d, want 404 body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("success prefers running candidate", func(t *testing.T) {
		cmd := exec.Command("sleep", "30")
		if err := cmd.Start(); err != nil {
			t.Fatalf("start sleep: %v", err)
		}
		defer func() { _ = cmd.Process.Kill(); _, _ = cmd.Process.Wait() }()
		mock := &runStatusLogSvcMock{runs: []service.RunRecord{
			{ID: 20, TaskID: 9, Status: "finished", Summary: string(mustJSON(map[string]any{"pid": 0}))},
			{ID: 21, TaskID: 9, Status: "running", Summary: string(mustJSON(map[string]any{"pid": cmd.Process.Pid}))},
		}}
		ctrl := NewRunController(service.NewRunService(mock), nil)
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/9/kill", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleTaskKill(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}
		if int(body["runId"].(float64)) != 21 {
			t.Fatalf("expected runId=21, got %#v", body)
		}
	})
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
