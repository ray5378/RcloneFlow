package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"rcloneflow/internal/active_transfer"
	"rcloneflow/internal/service"
)

type mockActiveTransferRunSvcDB struct {
	run service.RunRecord
}

func (m *mockActiveTransferRunSvcDB) ListRuns(page, pageSize int) ([]service.RunRecord, int, error) { return nil, 0, nil }
func (m *mockActiveTransferRunSvcDB) ListRunsByTask(taskId int64) ([]service.RunRecord, error)      { return nil, nil }
func (m *mockActiveTransferRunSvcDB) ListActiveRuns() ([]service.RunRecord, error)                   { return []service.RunRecord{m.run}, nil }
func (m *mockActiveTransferRunSvcDB) GetActiveRunByTaskID(taskID int64) (service.RunRecord, error)   { return m.run, nil }
func (m *mockActiveTransferRunSvcDB) GetRun(id int64) (service.RunRecord, error)                     { return service.RunRecord{}, nil }
func (m *mockActiveTransferRunSvcDB) UpdateRun(id int64, updateFn func(*service.RunRecord))          {}
func (m *mockActiveTransferRunSvcDB) DeleteRun(id int64) error                                       { return nil }
func (m *mockActiveTransferRunSvcDB) DeleteAllRuns() error                                           { return nil }
func (m *mockActiveTransferRunSvcDB) DeleteRunsByTask(taskId int64) error                            { return nil }
func (m *mockActiveTransferRunSvcDB) CleanOldRuns(days int) (int64, error)                           { return 0, nil }

func TestActiveTransferController_OverviewAndLists(t *testing.T) {
	mgr := active_transfer.NewManager()
	mgr.InitState(11, 7, active_transfer.TrackingModeCAS, []active_transfer.TransferCandidateFile{
		{Path: "a/file1.mkv", Name: "file1.mkv", SizeBytes: 123},
		{Path: "a/file2.mkv", Name: "file2.mkv", SizeBytes: 456},
	})
	pct := 32.5
	mgr.UpdateCurrentFile(11, "a/file1.mkv", 12, 36, 4, &pct)
	mgr.MarkCompleted(11, "a/file2.mkv", active_transfer.FileStatusCASMatched, "")

	summary := map[string]any{
		"progress": map[string]any{
			"bytes":      float64(12),
			"totalBytes": float64(36),
			"speed":      float64(4),
			"eta":        float64(6),
			"percentage": float64(33.3),
		},
	}
	bs, _ := json.Marshal(summary)
	ctrl := NewActiveTransferController(mgr, service.NewRunService(&mockActiveTransferRunSvcDB{run: service.RunRecord{ID: 11, TaskID: 7, Status: "running", Summary: string(bs)}}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/7/active-transfer", nil)
	ctrl.HandleOverview(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("overview status=%d", w.Code)
	}
	var overview map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &overview); err != nil {
		t.Fatalf("overview unmarshal: %v", err)
	}
	if overview["trackingMode"] != "cas" {
		t.Fatalf("trackingMode=%v", overview["trackingMode"])
	}
	sum, _ := overview["summary"].(map[string]any)
	if int(sum["completedCount"].(float64)) != 1 {
		t.Fatalf("completedCount=%v", sum["completedCount"])
	}
	if got, _ := sum["preflightFinished"].(bool); !got {
		t.Fatalf("preflightFinished=%v, want true", sum["preflightFinished"])
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/tasks/7/active-transfer/completed?offset=0&limit=10", nil)
	ctrl.HandleCompleted(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("completed status=%d", w.Code)
	}
	var completed map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &completed)
	if int(completed["total"].(float64)) != 1 {
		t.Fatalf("completed total=%v", completed["total"])
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/tasks/7/active-transfer/pending?offset=0&limit=10", nil)
	ctrl.HandlePending(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("pending status=%d", w.Code)
	}
	var pending map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &pending)
	if int(pending["total"].(float64)) != 1 {
		t.Fatalf("pending total=%v", pending["total"])
	}
}

func TestActiveTransferController_OverviewShowsPendingPreflightFlag(t *testing.T) {
	mgr := active_transfer.NewManager()
	mgr.InitState(12, 8, active_transfer.TrackingModeNormal, nil)

	summary := map[string]any{
		"progress": map[string]any{
			"bytes":      float64(0),
			"totalBytes": float64(0),
			"speed":      float64(0),
			"eta":        float64(0),
			"percentage": float64(0),
		},
	}
	bs, _ := json.Marshal(summary)
	ctrl := NewActiveTransferController(mgr, service.NewRunService(&mockActiveTransferRunSvcDB{run: service.RunRecord{ID: 12, TaskID: 8, Status: "running", Summary: string(bs)}}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/8/active-transfer", nil)
	ctrl.HandleOverview(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("overview status=%d", w.Code)
	}
	var overview map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &overview)
	sum, _ := overview["summary"].(map[string]any)
	if got, _ := sum["preflightPending"].(bool); !got {
		t.Fatalf("preflightPending=%v, want true", sum["preflightPending"])
	}
	if got, _ := sum["preflightFinished"].(bool); got {
		t.Fatalf("preflightFinished=%v, want false", sum["preflightFinished"])
	}
}

func TestActiveTransferController_RestoreFromSummarySnapshot(t *testing.T) {
	mgr := active_transfer.NewManager()
	pct := 32.5
	snap := active_transfer.ActiveTransferSnapshot{
		RunID:        11,
		TaskID:       7,
		TrackingMode: active_transfer.TrackingModeCAS,
		TotalCount:   2,
		CurrentFile:  &active_transfer.TransferCurrentFile{Path: "a/file1.mkv", Name: "file1.mkv", Bytes: 12, TotalBytes: 36, Speed: 4, Percentage: &pct, Status: active_transfer.FileStatusInProgress},
		CurrentFiles: []active_transfer.TransferCurrentFile{{Path: "a/file1.mkv", Name: "file1.mkv", Bytes: 12, TotalBytes: 36, Speed: 4, Percentage: &pct, Status: active_transfer.FileStatusInProgress}},
		Completed: []active_transfer.TransferCompletedFile{
			{Path: "a/file2.mkv", Name: "file2.mkv", SizeBytes: 456, At: "2026-05-09T14:00:00Z", Status: active_transfer.FileStatusCASMatched},
		},
		Pending: []active_transfer.TransferPendingFile{
			{Path: "a/file1.mkv", Name: "file1.mkv", SizeBytes: 123, Status: active_transfer.FileStatusInProgress},
		},
	}
	bs, _ := json.Marshal(map[string]any{
		"progress": map[string]any{
			"bytes":      float64(12),
			"totalBytes": float64(36),
			"speed":      float64(4),
			"eta":        float64(6),
			"percentage": float64(33.3),
		},
		"activeTransfer": active_transfer.SnapshotEnvelope(snap)["activeTransfer"],
	})
	ctrl := NewActiveTransferController(mgr, service.NewRunService(&mockActiveTransferRunSvcDB{run: service.RunRecord{ID: 11, TaskID: 7, Status: "running", Summary: string(bs)}}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/7/active-transfer", nil)
	ctrl.HandleOverview(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("overview status=%d", w.Code)
	}
	var overview map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &overview)
	sum, _ := overview["summary"].(map[string]any)
	if int(sum["completedCount"].(float64)) != 1 || int(sum["pendingCount"].(float64)) != 1 {
		t.Fatalf("summary counts=%v", sum)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/tasks/7/active-transfer/completed?offset=0&limit=10", nil)
	ctrl.HandleCompleted(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("completed status=%d", w.Code)
	}
	var completed map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &completed)
	if int(completed["total"].(float64)) != 1 {
		t.Fatalf("completed total=%v", completed["total"])
	}
}
