package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"rcloneflow/internal/active_transfer"
	"rcloneflow/internal/service"
)

func TestActiveTransferController_DegradedListTotalsUseRealCounts(t *testing.T) {
	mgr := active_transfer.NewManager()
	candidates := make([]active_transfer.TransferCandidateFile, 0, 2050)
	for i := 0; i < 2050; i++ {
		name := fmt.Sprintf("file-%04d.bin", i)
		candidates = append(candidates, active_transfer.TransferCandidateFile{Path: name, Name: name})
	}
	st := mgr.InitState(21, 9, active_transfer.TrackingModeNormal, candidates)
	if !st.Degraded {
		t.Fatalf("expected degraded mode for large candidate set")
	}
	for _, c := range candidates {
		mgr.MarkCompleted(21, c.Path, active_transfer.FileStatusCopied, "")
	}

	summary := map[string]any{
		"progress": map[string]any{
			"bytes":      float64(100),
			"totalBytes": float64(100),
			"speed":      float64(0),
			"eta":        float64(0),
			"percentage": float64(100),
		},
	}
	bs, _ := json.Marshal(summary)
	ctrl := NewActiveTransferController(mgr, service.NewRunService(&mockActiveTransferRunSvcDB{run: service.RunRecord{ID: 21, TaskID: 9, Status: "running", Summary: string(bs)}}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/9/active-transfer/completed?offset=0&limit=10", nil)
	ctrl.HandleCompleted(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("completed status=%d", w.Code)
	}
	var completed struct {
		Total int                                  `json:"total"`
		Items []active_transfer.TransferCompletedFile `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &completed); err != nil {
		t.Fatalf("unmarshal completed: %v", err)
	}
	if completed.Total != 2050 {
		t.Fatalf("completed total=%d want=2050", completed.Total)
	}
	if len(completed.Items) != 10 {
		t.Fatalf("completed page size=%d want=10", len(completed.Items))
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/tasks/9/active-transfer", nil)
	ctrl.HandleOverview(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("overview status=%d", w.Code)
	}
	var overview map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &overview); err != nil {
		t.Fatalf("unmarshal overview: %v", err)
	}
	sum, _ := overview["summary"].(map[string]any)
	if int(sum["completedCount"].(float64)) != 2050 {
		t.Fatalf("overview completedCount=%v want=2050", sum["completedCount"])
	}
	if got, _ := overview["degraded"].(bool); !got {
		t.Fatalf("overview degraded=%v want true", overview["degraded"])
	}
}
