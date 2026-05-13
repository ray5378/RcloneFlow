package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"rcloneflow/internal/service"
)

type runFilesSvcMock struct {
	runs    []service.RunRecord
	listErr error
}

func (m *runFilesSvcMock) ListRuns(page, pageSize int) ([]service.RunRecord, int, error) {
	return m.runs, len(m.runs), m.listErr
}
func (m *runFilesSvcMock) ListRunsByTask(taskId int64) ([]service.RunRecord, error) {
	return nil, m.listErr
}
func (m *runFilesSvcMock) ListActiveRuns() ([]service.RunRecord, error) { return nil, m.listErr }
func (m *runFilesSvcMock) GetActiveRunByTaskID(taskID int64) (service.RunRecord, error) {
	return service.RunRecord{}, m.listErr
}
func (m *runFilesSvcMock) GetRun(id int64) (service.RunRecord, error) { return service.RunRecord{}, m.listErr }
func (m *runFilesSvcMock) UpdateRun(id int64, updateFn func(*service.RunRecord)) {}
func (m *runFilesSvcMock) DeleteRun(id int64) error                     { return nil }
func (m *runFilesSvcMock) DeleteAllRuns() error                         { return nil }
func (m *runFilesSvcMock) DeleteRunsByTask(taskId int64) error          { return nil }
func (m *runFilesSvcMock) CleanOldRuns(days int) (int64, error)         { return 0, nil }

func TestRunController_HandleRunFiles_RawModeAndErrors(t *testing.T) {
	t.Run("raw mode serves log file", func(t *testing.T) {
		logPath := filepath.Join(t.TempDir(), "raw.log")
		if err := os.WriteFile(logPath, []byte("raw-line-1\nraw-line-2\n"), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
		summary, _ := json.Marshal(map[string]any{"stderrFile": logPath})
		ctrl := NewRunController(service.NewRunService(&runFilesSvcMock{runs: []service.RunRecord{{ID: 1, Summary: string(summary)}}}), nil)
		req := httptest.NewRequest(http.MethodGet, "/api/runs/1/files?mode=raw", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRunFiles(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
		if rec.Body.String() != "raw-line-1\nraw-line-2\n" {
			t.Fatalf("unexpected body=%q", rec.Body.String())
		}
	})

	t.Run("list error returns 500", func(t *testing.T) {
		ctrl := NewRunController(service.NewRunService(&runFilesSvcMock{listErr: errors.New("boom")}), nil)
		req := httptest.NewRequest(http.MethodGet, "/api/runs/1/files", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRunFiles(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want 500 body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("missing log read returns 500", func(t *testing.T) {
		summary, _ := json.Marshal(map[string]any{"stderrFile": filepath.Join(t.TempDir(), "missing.log")})
		ctrl := NewRunController(service.NewRunService(&runFilesSvcMock{runs: []service.RunRecord{{ID: 2, Summary: string(summary)}}}), nil)
		req := httptest.NewRequest(http.MethodGet, "/api/runs/2/files", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleRunFiles(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want 500 body=%s", rec.Code, rec.Body.String())
		}
	})
}

func TestRunController_HandleRunFiles_ParsesPaginationAndMoveMerging(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "move.log")
	logText := "" +
		"2026/05/13 21:00:00 INFO  : a.txt: Copied (new)\n" +
		"2026/05/13 21:00:01 INFO  : a.txt: Deleted\n" +
		"2026/05/13 21:00:02 INFO  : b.txt: Copied (new)\n" +
		"2026/05/13 21:00:03 INFO  : c.txt: Skipped\n"
	if err := os.WriteFile(logPath, []byte(logText), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	summary, _ := json.Marshal(map[string]any{"stderrFile": logPath})
	ctrl := NewRunController(service.NewRunService(&runFilesSvcMock{runs: []service.RunRecord{{
		ID:       3,
		TaskMode: "move",
		Summary:  string(summary),
	}}}), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/runs/3/files?offset=1&limit=2", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleRunFiles(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if int(body["total"].(float64)) != 3 {
		t.Fatalf("total=%v, want 3 after move merge", body["total"])
	}
	items, _ := body["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("len(items)=%d, want 2", len(items))
	}
	first, _ := items[0].(map[string]any)
	second, _ := items[1].(map[string]any)
	if first["name"] != "b.txt" || first["action"] != "Copied" {
		t.Fatalf("unexpected first item=%#v", first)
	}
	if second["name"] != "c.txt" || second["action"] != "Skipped" {
		t.Fatalf("unexpected second item=%#v", second)
	}
	info, _ := body["info"].(map[string]any)
	if info == nil || info["logPath"] != logPath {
		t.Fatalf("unexpected info=%#v", info)
	}
}

func TestRunController_HandleRunFiles_FiltersCASNoiseAndHonorsLimitBounds(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "cas.log")
	logText := "" +
		"2026/05/13 21:10:00 ERROR : x.mkv: Failed to copy: object not found\n" +
		"2026/05/13 21:10:01 NOTICE: x.mkv: CAS compatible match after source cleanup\n" +
		"2026/05/13 21:10:02 ERROR : Attempt 1/3 failed with 1 errors and: object not found\n" +
		"2026/05/13 21:10:03 ERROR : Failed to copy with 1 errors: last error was: object not found\n" +
		"2026/05/13 21:10:04 INFO  : y.mkv: Copied (new)\n"
	if err := os.WriteFile(logPath, []byte(logText), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	summary, _ := json.Marshal(map[string]any{
		"stderrFile":    logPath,
		"transferDefaults": map[string]any{"openlistCasCompatible": true},
	})
	ctrl := NewRunController(service.NewRunService(&runFilesSvcMock{runs: []service.RunRecord{{ID: 4, Summary: string(summary)}}}), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/runs/4/files?offset=0&limit=5000", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleRunFiles(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if int(body["total"].(float64)) != 2 {
		t.Fatalf("total=%v, want 2 after CAS filtering", body["total"])
	}
	items, _ := body["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("len(items)=%d, want 2", len(items))
	}
	first, _ := items[0].(map[string]any)
	second, _ := items[1].(map[string]any)
	if first["action"] != "CAS Matched" || second["action"] != "Copied" {
		t.Fatalf("unexpected items=%#v", items)
	}
}
