package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"rcloneflow/internal/service"
)

type runGlobalStatsSvcMock struct {
	runs []service.RunRecord
	err  error
}

func (m *runGlobalStatsSvcMock) ListRuns(page, pageSize int) ([]service.RunRecord, int, error) {
	return nil, 0, m.err
}
func (m *runGlobalStatsSvcMock) ListRunsByTask(taskId int64) ([]service.RunRecord, error) {
	return nil, m.err
}
func (m *runGlobalStatsSvcMock) ListActiveRuns() ([]service.RunRecord, error) {
	return m.runs, m.err
}
func (m *runGlobalStatsSvcMock) GetActiveRunByTaskID(taskID int64) (service.RunRecord, error) {
	return service.RunRecord{}, m.err
}
func (m *runGlobalStatsSvcMock) GetRun(id int64) (service.RunRecord, error) { return service.RunRecord{}, m.err }
func (m *runGlobalStatsSvcMock) UpdateRun(id int64, updateFn func(*service.RunRecord)) {}
func (m *runGlobalStatsSvcMock) DeleteRun(id int64) error                     { return nil }
func (m *runGlobalStatsSvcMock) DeleteAllRuns() error                         { return nil }
func (m *runGlobalStatsSvcMock) DeleteRunsByTask(taskId int64) error          { return nil }
func (m *runGlobalStatsSvcMock) CleanOldRuns(days int) (int64, error)         { return 0, nil }

func TestRunController_HandleGlobalStats_Branches(t *testing.T) {
	t.Run("list active runs error", func(t *testing.T) {
		ctrl := NewRunController(service.NewRunService(&runGlobalStatsSvcMock{err: errors.New("boom")}), nil)
		req := httptest.NewRequest(http.MethodGet, "/api/runs/global-stats", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleGlobalStats(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d, want 500 body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("aggregates map and string summaries and ignores bad data", func(t *testing.T) {
		stringSummary, _ := json.Marshal(map[string]any{
			"progress": map[string]any{
				"bytes":      float64(100),
				"totalBytes": float64(250),
				"speed":      float64(10),
			},
		})
		mapSummary, _ := json.Marshal(map[string]any{"progress": map[string]any{"bytes": float64(200), "totalBytes": float64(500), "speed": float64(20)}})
		ignoredSummary, _ := json.Marshal(map[string]any{"other": "ignored"})
		ctrl := NewRunController(service.NewRunService(&runGlobalStatsSvcMock{runs: []service.RunRecord{
			{ID: 1, Summary: string(mapSummary)},
			{ID: 2, Summary: string(stringSummary)},
			{ID: 3, Summary: "{bad json}"},
			{ID: 4, Summary: string(ignoredSummary)},
		}}), nil)
		req := httptest.NewRequest(http.MethodGet, "/api/runs/global-stats", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleGlobalStats(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}
		if body["bytes"].(float64) != 300 {
			t.Fatalf("bytes=%v, want 300", body["bytes"])
		}
		if body["totalBytes"].(float64) != 750 {
			t.Fatalf("totalBytes=%v, want 750", body["totalBytes"])
		}
		if body["speed"].(float64) != 30 {
			t.Fatalf("speed=%v, want 30", body["speed"])
		}
		if body["speedAvg"].(float64) != 30 {
			t.Fatalf("speedAvg=%v, want 30", body["speedAvg"])
		}
		if body["eta"] != nil {
			t.Fatalf("eta=%v, want nil", body["eta"])
		}
		pct := body["percentage"].(float64)
		if pct < 39.9 || pct > 40.1 {
			t.Fatalf("percentage=%v, want about 40", pct)
		}
	})

	t.Run("zero total keeps percentage zero", func(t *testing.T) {
		zeroTotalSummary, _ := json.Marshal(map[string]any{"progress": map[string]any{"bytes": float64(50), "speed": float64(5)}})
		ctrl := NewRunController(service.NewRunService(&runGlobalStatsSvcMock{runs: []service.RunRecord{
			{ID: 1, Summary: string(zeroTotalSummary)},
		}}), nil)
		req := httptest.NewRequest(http.MethodGet, "/api/runs/global-stats", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleGlobalStats(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}
		if body["percentage"].(float64) != 0 {
			t.Fatalf("percentage=%v, want 0", body["percentage"])
		}
	})
}
