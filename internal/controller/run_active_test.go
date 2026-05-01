package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"rcloneflow/internal/service"
)

type mockRunSvcDB struct {
	runs []service.RunRecord
}

func (m *mockRunSvcDB) ListRuns(page, pageSize int) ([]service.RunRecord, int, error) {
	return nil, 0, nil
}
func (m *mockRunSvcDB) ListRunsByTask(taskId int64) ([]service.RunRecord, error) {
	return nil, nil
}
func (m *mockRunSvcDB) ListActiveRuns() ([]service.RunRecord, error) {
	return m.runs, nil
}
func (m *mockRunSvcDB) GetActiveRunByTaskID(taskID int64) (service.RunRecord, error) {
	return service.RunRecord{}, nil
}
func (m *mockRunSvcDB) GetRun(id int64) (service.RunRecord, error)                        { return service.RunRecord{}, nil }
func (m *mockRunSvcDB) UpdateRun(id int64, updateFn func(*service.RunRecord))             {}
func (m *mockRunSvcDB) DeleteRun(id int64) error                                          { return nil }
func (m *mockRunSvcDB) DeleteAllRuns() error                                              { return nil }
func (m *mockRunSvcDB) DeleteRunsByTask(taskId int64) error                               { return nil }
func (m *mockRunSvcDB) CleanOldRuns(days int) (int64, error)                              { return 0, nil }
func (m *mockRunSvcDB) UpdateRunStatusByJobId(jobId int64, status, errorMsg string) error { return nil }

func TestHandleActiveRuns_UsesProgressAndExposesDebugFields(t *testing.T) {
	summary := map[string]any{
		"progress": map[string]any{
			"bytes":          121.377 * 1024 * 1024,
			"totalBytes":     335.968 * 1024 * 1024,
			"speed":          2.474 * 1024 * 1024,
			"eta":            float64(86),
			"percentage":     float64(36),
			"completedFiles": float64(18),
			"plannedFiles":   float64(53),
		},
		"progressLine": `2026/04/17 14:34:08 INFO : 121.377 MiB / 335.968 MiB, 36%, 2.474 MiB/s, ETA 1m26s (xfr#18/53)`,
	}
	bs, _ := json.Marshal(summary)
	ctrl := &RunController{runSvc: service.NewRunService(&mockRunSvcDB{runs: []service.RunRecord{{
		ID:        1,
		TaskID:    100,
		Status:    "running",
		StartedAt: "2026-04-17T14:30:00+08:00",
		Summary:   string(bs),
	}}})}

	req := httptest.NewRequest(http.MethodGet, "/api/runs/active", nil)
	w := httptest.NewRecorder()
	ctrl.HandleActiveRuns(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", w.Code)
	}
	var items []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items)=%d, want 1", len(items))
	}
	it := items[0]
	prog, _ := it["progress"].(map[string]any)
	if prog == nil {
		t.Fatalf("missing progress")
	}
	if got := int(prog["completedFiles"].(float64)); got != 18 {
		t.Fatalf("completedFiles=%d, want 18", got)
	}
	if got := int(prog["totalCount"].(float64)); got != 53 {
		t.Fatalf("totalCount=%d, want 53", got)
	}
	if got := it["progressLine"].(string); got == "" {
		t.Fatalf("missing progressLine")
	}
	if got := it["progressSource"].(string); got != "summary.progress" {
		t.Fatalf("progressSource=%q, want summary.progress", got)
	}
	if _, ok := it["stableProgress"]; ok {
		t.Fatalf("stableProgress should not be exposed in active run response")
	}
	if got := it["progressMismatch"].(bool); got {
		t.Fatalf("progressMismatch=%v, want false", got)
	}
	check, _ := it["progressCheck"].(map[string]any)
	if check == nil || !check["ok"].(bool) {
		t.Fatalf("progressCheck.ok want true, got %#v", check)
	}
}

func TestHandleActiveRuns_DoesNotExposeFinalSummary(t *testing.T) {
	summary := map[string]any{
		"progress": map[string]any{
			"bytes":          float64(100),
			"totalBytes":     float64(300),
			"speed":          float64(10),
			"eta":            float64(20),
			"percentage":     float64(33.3),
			"completedFiles": float64(3),
			"plannedFiles":   float64(9),
		},
		"finalSummary": map[string]any{
			"result":           "success",
			"transferredBytes": float64(9999),
			"totalBytes":       float64(9999),
		},
	}
	bs, _ := json.Marshal(summary)
	ctrl := &RunController{runSvc: service.NewRunService(&mockRunSvcDB{runs: []service.RunRecord{{
		ID:        4,
		TaskID:    103,
		Status:    "running",
		StartedAt: "2026-04-17T14:30:00+08:00",
		Summary:   string(bs),
	}}})}

	req := httptest.NewRequest(http.MethodGet, "/api/runs/active", nil)
	w := httptest.NewRecorder()
	ctrl.HandleActiveRuns(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", w.Code)
	}
	var items []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items)=%d, want 1", len(items))
	}
	it := items[0]
	if _, ok := it["finalSummary"]; ok {
		t.Fatalf("finalSummary should not be exposed in active run response")
	}
	prog, _ := it["progress"].(map[string]any)
	if prog == nil {
		t.Fatalf("missing progress")
	}
	if got := int(prog["totalBytes"].(float64)); got != 300 {
		t.Fatalf("totalBytes=%d, want 300", got)
	}
	if got := int(prog["completedFiles"].(float64)); got != 3 {
		t.Fatalf("completedFiles=%d, want 3", got)
	}
}

func TestHandleActiveRuns_UsesProgressTotalsOnly(t *testing.T) {
	summary := map[string]any{
		"progress": map[string]any{
			"bytes":          float64(100),
			"totalBytes":     float64(300),
			"speed":          float64(10),
			"eta":            float64(20),
			"percentage":     float64(10),
			"completedFiles": float64(3),
			"plannedFiles":   float64(33),
		},
		"preflight": map[string]any{
			"totalBytes": float64(1024),
			"totalCount": float64(166),
		},
	}
	bs, _ := json.Marshal(summary)
	ctrl := &RunController{runSvc: service.NewRunService(&mockRunSvcDB{runs: []service.RunRecord{{
		ID:        3,
		TaskID:    102,
		Status:    "running",
		StartedAt: "2026-04-17T14:30:00+08:00",
		Summary:   string(bs),
	}}})}

	req := httptest.NewRequest(http.MethodGet, "/api/runs/active", nil)
	w := httptest.NewRecorder()
	ctrl.HandleActiveRuns(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", w.Code)
	}
	var items []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	it := items[0]
	prog, _ := it["progress"].(map[string]any)
	if prog == nil {
		t.Fatalf("missing progress")
	}
	if got := int(prog["totalBytes"].(float64)); got != 300 {
		t.Fatalf("totalBytes=%d, want 300", got)
	}
	if got := int(prog["totalCount"].(float64)); got != 33 {
		t.Fatalf("totalCount=%d, want 33", got)
	}
	if got := prog["percentage"].(float64); got < 9.9 || got > 10.1 {
		t.Fatalf("percentage=%v, want about 10", got)
	}
}

func TestHandleActiveRuns_FallsBackToPreflightTotalCountWhenPlannedFilesMissing(t *testing.T) {
	summary := map[string]any{
		"progress": map[string]any{
			"bytes":          float64(100),
			"totalBytes":     float64(300),
			"speed":          float64(10),
			"eta":            float64(20),
			"percentage":     float64(10),
			"completedFiles": float64(3),
		},
		"preflight": map[string]any{
			"totalBytes": float64(1024),
			"totalCount": float64(166),
		},
	}
	bs, _ := json.Marshal(summary)
	ctrl := &RunController{runSvc: service.NewRunService(&mockRunSvcDB{runs: []service.RunRecord{{
		ID:        5,
		TaskID:    104,
		Status:    "running",
		StartedAt: "2026-04-17T14:30:00+08:00",
		Summary:   string(bs),
	}}})}

	req := httptest.NewRequest(http.MethodGet, "/api/runs/active", nil)
	w := httptest.NewRecorder()
	ctrl.HandleActiveRuns(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", w.Code)
	}
	var items []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	it := items[0]
	prog, _ := it["progress"].(map[string]any)
	if prog == nil {
		t.Fatalf("missing progress")
	}
	if got := int(prog["totalCount"].(float64)); got != 166 {
		t.Fatalf("totalCount=%d, want 166 from preflight fallback", got)
	}
	if got := int(prog["completedFiles"].(float64)); got != 3 {
		t.Fatalf("completedFiles=%d, want 3", got)
	}
}

func TestHandleActiveRuns_BackfillsCompletedFilesFromLogCASNotice(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "rcloneflow-active-log-*.log")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	logText := "2026/05/01 14:01:20 NOTICE: 电视剧/国产剧/人间惊鸿客 (2026)/Season 1/人间惊鸿客 - S01E18 - 第 18 集.mkv: CAS compatible match after source cleanup (Failed to copy: object not found)\n"
	if _, err := tmpFile.WriteString(logText); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}

	summary := map[string]any{
		"progress": map[string]any{
			"bytes":          float64(1024),
			"totalBytes":     float64(2048),
			"speed":          float64(1),
			"eta":            float64(60),
			"percentage":     float64(0),
			"completedFiles": float64(0),
			"plannedFiles":   float64(5),
		},
		"stderrFile": tmpFile.Name(),
	}
	bs, _ := json.Marshal(summary)
	ctrl := &RunController{runSvc: service.NewRunService(&mockRunSvcDB{runs: []service.RunRecord{{
		ID:        6,
		TaskID:    105,
		Status:    "running",
		StartedAt: "2026-05-01T14:00:00+08:00",
		Summary:   string(bs),
	}}})}

	req := httptest.NewRequest(http.MethodGet, "/api/runs/active", nil)
	w := httptest.NewRecorder()
	ctrl.HandleActiveRuns(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", w.Code)
	}
	var items []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	prog, _ := items[0]["progress"].(map[string]any)
	if prog == nil {
		t.Fatalf("missing progress")
	}
	if got := int(prog["completedFiles"].(float64)); got != 1 {
		t.Fatalf("completedFiles=%d, want 1 from log fallback", got)
	}
}

func TestHandleActiveRuns_FlagsProgressMismatch(t *testing.T) {
	summary := map[string]any{
		"progress": map[string]any{
			"bytes":          float64(1024),
			"totalBytes":     float64(2048),
			"speed":          float64(1),
			"eta":            float64(60),
			"percentage":     float64(99),
			"completedFiles": float64(10),
			"plannedFiles":   float64(5),
		},
		"progressLine": `bad aggregate line`,
	}
	bs, _ := json.Marshal(summary)
	ctrl := &RunController{runSvc: service.NewRunService(&mockRunSvcDB{runs: []service.RunRecord{{
		ID:        2,
		TaskID:    101,
		Status:    "running",
		StartedAt: "2026-04-17T14:30:00+08:00",
		Summary:   string(bs),
	}}})}

	req := httptest.NewRequest(http.MethodGet, "/api/runs/active", nil)
	w := httptest.NewRecorder()
	ctrl.HandleActiveRuns(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", w.Code)
	}
	var items []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	it := items[0]
	if got := it["progressMismatch"].(bool); !got {
		t.Fatalf("progressMismatch=%v, want true", got)
	}
	check, _ := it["progressCheck"].(map[string]any)
	if check == nil {
		t.Fatalf("missing progressCheck")
	}
	if !check["pctMismatch"].(bool) {
		t.Fatalf("pctMismatch want true")
	}
	if !check["etaMismatch"].(bool) {
		t.Fatalf("etaMismatch want true")
	}
}
