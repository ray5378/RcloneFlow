package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"rcloneflow/internal/service"
)

type mockRunSvcDB struct {
	runs []service.RunRecord
}

func (m *mockRunSvcDB) ListRuns(page, pageSize int) ([]service.RunRecord, int, error) {
	return m.runs, len(m.runs), nil
}
func (m *mockRunSvcDB) ListRunsByTask(taskId int64) ([]service.RunRecord, error) {
	out := make([]service.RunRecord, 0, len(m.runs))
	for _, r := range m.runs {
		if r.TaskID == taskId {
			out = append(out, r)
		}
	}
	return out, nil
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

func TestHandleActiveRuns_UsesLogicalTotalsAndKeepsPlannedFiles(t *testing.T) {
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
		"effectiveOptions": map[string]any{
			"openlistCasCompatible": true,
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
	if got := int(prog["plannedFiles"].(float64)); got != 33 {
		t.Fatalf("plannedFiles=%d, want 33", got)
	}
	if got := int(prog["logicalTotalCount"].(float64)); got != 33 {
		t.Fatalf("logicalTotalCount=%d, want 33", got)
	}
	if got := int(prog["totalCount"].(float64)); got != 33 {
		t.Fatalf("totalCount=%d, want 33", got)
	}
	if got := prog["percentage"].(float64); got < 9.9 || got > 10.1 {
		t.Fatalf("percentage=%v, want about 10", got)
	}
}

func TestHandleActiveRuns_UsesPreflightTotalCountForCASCompatibleRuns(t *testing.T) {
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

func TestHandleRuns_DoesNotExposeHistoricalFilesArray(t *testing.T) {
	summary := map[string]any{
		"startedAt": "2026-04-17T14:30:00+08:00",
		"finishedAt": "2026-04-17T14:40:00+08:00",
		"finalSummary": map[string]any{
			"result":           "success",
			"transferredBytes": float64(1234),
			"totalBytes":       float64(5678),
			"counts": map[string]any{
				"total":   float64(2),
				"copied":  float64(1),
				"deleted": float64(1),
			},
		},
	}
	bs, _ := json.Marshal(summary)
	ctrl := &RunController{runSvc: service.NewRunService(&mockRunSvcDB{runs: []service.RunRecord{{
		ID:        8,
		TaskID:    108,
		Status:    "finished",
		StartedAt: "2026-04-17T14:30:00+08:00",
		FinishedAt: "2026-04-17T14:40:00+08:00",
		Summary:   string(bs),
	}}})}

	req := httptest.NewRequest(http.MethodGet, "/api/runs?page=1&pageSize=50", nil)
	w := httptest.NewRecorder()
	ctrl.HandleRuns(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	runs, _ := resp["runs"].([]any)
	if len(runs) != 1 {
		t.Fatalf("len(runs)=%d, want 1", len(runs))
	}
	item, _ := runs[0].(map[string]any)
	if item == nil {
		t.Fatalf("missing run item")
	}
	summaryObj, _ := item["summary"].(map[string]any)
	if summaryObj == nil {
		t.Fatalf("missing summary")
	}
	if _, ok := summaryObj["files"]; ok {
		t.Fatalf("summary.files should not be exposed in run list response")
	}
	fs, _ := summaryObj["finalSummary"].(map[string]any)
	if fs == nil {
		t.Fatalf("missing finalSummary")
	}
	if _, ok := fs["files"]; ok {
		t.Fatalf("finalSummary.files should not be exposed in run list response")
	}
	if got := fs["totalCount"]; got == nil {
		t.Fatalf("missing finalSummary.totalCount")
	}
	if got := item["durationText"]; got == nil {
		t.Fatalf("missing durationText")
	}
}

func TestHandleActiveRuns_CountsCASMatchAsCompletedFiles(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "rcloneflow-active-log-*.log")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	logText := "2026/05/08 18:45:46 NOTICE: 电视剧/国产剧/低智商犯罪 (2026)/Season 1/低智商犯罪 - S01E12 - 第 12 集.mkv: CAS compatible match after source cleanup (Failed to copy: object not found)\n"
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
		t.Fatalf("completedFiles=%d, want 1 for CAS match notice", got)
	}
}

func TestHandleRunsByTask_BackfillsHistoricalFinalSummaryFromCASLog(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "rcloneflow-history-log-*.log")
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
		"stderrFile": tmpFile.Name(),
		"finalSummary": map[string]any{
			"counts": map[string]any{"copied": float64(0), "deleted": float64(0), "failed": float64(0), "skipped": float64(0), "total": float64(0)},
			"files":  []any{},
		},
	}
	bs, _ := json.Marshal(summary)
	ctrl := &RunController{runSvc: service.NewRunService(&mockRunSvcDB{runs: []service.RunRecord{{
		ID:         7,
		TaskID:     106,
		Status:     "failed",
		StartedAt:  "2026-05-01T14:00:00+08:00",
		FinishedAt: "2026-05-01T14:10:00+08:00",
		Summary:    string(bs),
	}}})}

	req := httptest.NewRequest(http.MethodGet, "/api/runs/task/106", nil)
	w := httptest.NewRecorder()
	ctrl.HandleRunsByTask(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", w.Code)
	}
	var items []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	sum, _ := items[0]["summary"].(map[string]any)
	if sum == nil {
		t.Fatalf("missing summary")
	}
	fs, _ := sum["finalSummary"].(map[string]any)
	if fs == nil {
		t.Fatalf("missing finalSummary")
	}
	counts, _ := fs["counts"].(map[string]any)
	if counts == nil {
		t.Fatalf("missing counts")
	}
	if _, ok := fs["files"]; ok {
		t.Fatalf("finalSummary.files should not be exposed in task history response")
	}
	if got := int(fs["totalCount"].(float64)); got != 1 {
		t.Fatalf("totalCount=%d, want 1 after CAS-log backfill", got)
	}
	if got := int(counts["copied"].(float64)); got != 1 {
		t.Fatalf("copied=%d, want 1 after CAS-log backfill", got)
	}
}

func TestHandleRunFiles_ParsesJSONLogRows(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "rcloneflow-json-log-*.log")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	logText := "{\"level\":\"info\",\"msg\":\"Copied (new)\",\"object\":\"20260510/a.mp4\",\"time\":\"2026-05-10T14:49:03+08:00\"}\n" +
		"{\"level\":\"error\",\"msg\":\"Failed to copy: boom\",\"object\":\"20260510/b.mp4\",\"time\":\"2026-05-10T14:49:04+08:00\"}\n"
	if _, err := tmpFile.WriteString(logText); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}
	summary := map[string]any{"stderrFile": tmpFile.Name()}
	bs, _ := json.Marshal(summary)
	ctrl := &RunController{runSvc: service.NewRunService(&mockRunSvcDB{runs: []service.RunRecord{{
		ID:        9,
		TaskID:    109,
		Status:    "finished",
		StartedAt: "2026-05-10T14:48:00+08:00",
		Summary:   string(bs),
	}}})}

	req := httptest.NewRequest(http.MethodGet, "/api/runs/9/files?offset=0&limit=50", nil)
	w := httptest.NewRecorder()
	ctrl.HandleRunFiles(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got := int(resp["total"].(float64)); got != 2 {
		t.Fatalf("total=%d, want 2", got)
	}
	items, _ := resp["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("len(items)=%d, want 2", len(items))
	}
	first, _ := items[0].(map[string]any)
	second, _ := items[1].(map[string]any)
	if got := first["action"].(string); got != "Copied" {
		t.Fatalf("first action=%q, want Copied", got)
	}
	if got := int(first["sizeBytes"].(float64)); got != 0 {
		t.Fatalf("first sizeBytes=%d, want 0 when json log has no size", got)
	}
	if got := second["action"].(string); got != "Error" {
		t.Fatalf("second action=%q, want Error", got)
	}
}

func TestHandleRunFiles_ParsesJSONLogSizeBytes(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "rcloneflow-json-size-log-*.log")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	logText := "{\"level\":\"info\",\"msg\":\"Copied (new)\",\"object\":\"20260510/a.mp4\",\"time\":\"2026-05-10T14:49:03+08:00\",\"size\":12345}\n"
	if _, err := tmpFile.WriteString(logText); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}
	summary := map[string]any{"stderrFile": tmpFile.Name()}
	bs, _ := json.Marshal(summary)
	ctrl := &RunController{runSvc: service.NewRunService(&mockRunSvcDB{runs: []service.RunRecord{{
		ID:        11,
		TaskID:    111,
		Status:    "finished",
		StartedAt: "2026-05-10T14:48:00+08:00",
		Summary:   string(bs),
	}}})}

	req := httptest.NewRequest(http.MethodGet, "/api/runs/11/files?offset=0&limit=50", nil)
	w := httptest.NewRecorder()
	ctrl.HandleRunFiles(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	items, _ := resp["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("len(items)=%d, want 1", len(items))
	}
	first, _ := items[0].(map[string]any)
	if got := int(first["sizeBytes"].(float64)); got != 12345 {
		t.Fatalf("sizeBytes=%d, want 12345", got)
	}
}

func TestHandleRunFiles_MergeMoveCopiedAndDeletedRows(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "rcloneflow-json-move-log-*.log")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	logText := "{\"level\":\"info\",\"msg\":\"Copied (new)\",\"object\":\"20260510/a.mp4\",\"time\":\"2026-05-10T14:49:03+08:00\"}\n" +
		"{\"level\":\"info\",\"msg\":\"Deleted\",\"object\":\"20260510/a.mp4\",\"time\":\"2026-05-10T14:49:04+08:00\"}\n" +
		"{\"level\":\"info\",\"msg\":\"Copied (new)\",\"object\":\"20260510/b.mp4\",\"time\":\"2026-05-10T14:49:05+08:00\"}\n"
	if _, err := tmpFile.WriteString(logText); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}
	summary := map[string]any{"stderrFile": tmpFile.Name()}
	bs, _ := json.Marshal(summary)
	ctrl := &RunController{runSvc: service.NewRunService(&mockRunSvcDB{runs: []service.RunRecord{{
		ID:        10,
		TaskID:    110,
		TaskMode:  "move",
		Status:    "finished",
		StartedAt: "2026-05-10T14:48:00+08:00",
		Summary:   string(bs),
	}}})}

	req := httptest.NewRequest(http.MethodGet, "/api/runs/10/files?offset=0&limit=50", nil)
	w := httptest.NewRecorder()
	ctrl.HandleRunFiles(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got := int(resp["total"].(float64)); got != 2 {
		t.Fatalf("total=%d, want 2", got)
	}
	items, _ := resp["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("len(items)=%d, want 2", len(items))
	}
	actions := []string{}
	for _, raw := range items {
		it, _ := raw.(map[string]any)
		actions = append(actions, it["action"].(string))
	}
	joined := strings.Join(actions, ",")
	if !strings.Contains(joined, "Moved") || !strings.Contains(joined, "Copied") {
		t.Fatalf("unexpected actions=%v", actions)
	}
	if strings.Contains(joined, "Deleted") {
		t.Fatalf("unexpected deleted action remains in merged list: %v", actions)
	}
}

func TestHandleRunFiles_CASHistorySuppressesObjectNotFoundNoise(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "rcloneflow-cas-history-log-*.log")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	logText := "2026/05/13 15:46:09 ERROR : 电视剧/国产剧/罪无可逃 (2026)/Season 1/罪无可逃 - S01E06 - 第 6 集.mkv: Failed to copy: object not found\n" +
		"2026/05/13 15:48:30 ERROR : 电视剧/国产剧/罪无可逃 (2026)/Season 1/罪无可逃 - S01E07 - 第 7 集.mkv: Failed to copy: object not found\n" +
		"2026/05/13 15:48:30 ERROR : <nil>: Attempt 1/1 failed with 2 errors and: object not found\n" +
		"2026/05/13 15:48:31 ERROR : Failed to copy with 2 errors: last error was: object not found\n" +
		"2026/05/13 15:49:07 NOTICE : 电视剧/国产剧/罪无可逃 (2026)/Season 1/罪无可逃 - S01E06 - 第 6 集.mkv: CAS compatible match after source cleanup (Failed to copy: object not found)\n" +
		"2026/05/13 15:49:07 NOTICE : 电视剧/国产剧/罪无可逃 (2026)/Season 1/罪无可逃 - S01E07 - 第 7 集.mkv: CAS compatible match after source cleanup (Failed to copy: object not found)\n"
	if _, err := tmpFile.WriteString(logText); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}
	summary := map[string]any{
		"stderrFile": tmpFile.Name(),
		"transferDefaults": map[string]any{"openlistCasCompatible": true},
	}
	bs, _ := json.Marshal(summary)
	ctrl := &RunController{runSvc: service.NewRunService(&mockRunSvcDB{runs: []service.RunRecord{{
		ID:        12,
		TaskID:    112,
		Status:    "finished",
		StartedAt: "2026-05-13T15:43:32+08:00",
		Summary:   string(bs),
	}}})}

	req := httptest.NewRequest(http.MethodGet, "/api/runs/12/files?offset=0&limit=50", nil)
	w := httptest.NewRecorder()
	ctrl.HandleRunFiles(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got := int(resp["total"].(float64)); got != 2 {
		t.Fatalf("total=%d, want 2", got)
	}
	items, _ := resp["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("len(items)=%d, want 2", len(items))
	}
	for _, raw := range items {
		it, _ := raw.(map[string]any)
		if got := it["status"].(string); got != "success" {
			t.Fatalf("status=%q, want success", got)
		}
		if got := it["action"].(string); got != "CAS Matched" {
			t.Fatalf("action=%q, want CAS Matched", got)
		}
		msg := strings.ToLower(it["message"].(string))
		if !strings.Contains(msg, "cas compatible match after source cleanup") {
			t.Fatalf("unexpected message=%q", it["message"])
		}
	}
}

func TestHandleRunFiles_NonCASHistoryKeepsObjectNotFoundFailures(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "rcloneflow-noncas-history-log-*.log")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	logText := "2026/05/13 15:46:09 ERROR : a/file1.mkv: Failed to copy: object not found\n" +
		"2026/05/13 15:48:30 ERROR : <nil>: Attempt 1/1 failed with 2 errors and: object not found\n"
	if _, err := tmpFile.WriteString(logText); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}
	summary := map[string]any{"stderrFile": tmpFile.Name()}
	bs, _ := json.Marshal(summary)
	ctrl := &RunController{runSvc: service.NewRunService(&mockRunSvcDB{runs: []service.RunRecord{{
		ID:        13,
		TaskID:    113,
		Status:    "failed",
		StartedAt: "2026-05-13T15:43:32+08:00",
		Summary:   string(bs),
	}}})}

	req := httptest.NewRequest(http.MethodGet, "/api/runs/13/files?offset=0&limit=50", nil)
	w := httptest.NewRecorder()
	ctrl.HandleRunFiles(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got := int(resp["total"].(float64)); got != 2 {
		t.Fatalf("total=%d, want 2", got)
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
