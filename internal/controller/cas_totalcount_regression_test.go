package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"rcloneflow/internal/service"
)

func TestHandleActiveRuns_CASCompatibleFallsBackToPreflightTotalCount(t *testing.T) {
	summary := map[string]any{
		"progress": map[string]any{
			"bytes":          float64(100),
			"totalBytes":     float64(100),
			"speed":          float64(0),
			"eta":            float64(0),
			"percentage":     float64(100),
			"completedFiles": float64(1),
		},
		"preflight": map[string]any{
			"totalCount": float64(1),
		},
		"effectiveOptions": map[string]any{
			"openlistCasCompatible": true,
		},
	}
	bs, _ := json.Marshal(summary)
	ctrl := &RunController{runSvc: service.NewRunService(&mockRunSvcDB{runs: []service.RunRecord{{
		ID:        61,
		TaskID:    701,
		Status:    "running",
		StartedAt: "2026-05-14T20:00:00+08:00",
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
	if got := int(prog["logicalTotalCount"].(float64)); got != 1 {
		t.Fatalf("logicalTotalCount=%d, want 1", got)
	}
	if got := int(prog["totalCount"].(float64)); got != 1 {
		t.Fatalf("totalCount=%d, want 1", got)
	}
}

func TestBuildActiveRunItems_CASCompatibleFallsBackToPreflightTotalCount(t *testing.T) {
	summary := map[string]any{
		"progress": map[string]any{
			"bytes":          float64(100),
			"totalBytes":     float64(100),
			"speed":          float64(0),
			"eta":            float64(0),
			"percentage":     float64(100),
			"completedFiles": float64(1),
		},
		"preflight": map[string]any{
			"totalCount": float64(1),
		},
		"effectiveOptions": map[string]any{
			"openlistCasCompatible": true,
		},
	}
	bs, _ := json.Marshal(summary)
	ctrl := &TaskController{runSvc: service.NewRunService(&mockRunSvcDB{runs: []service.RunRecord{{
		ID:        62,
		TaskID:    702,
		Status:    "running",
		StartedAt: "2026-05-14T20:00:00+08:00",
		Summary:   string(bs),
	}}})}

	items, err := ctrl.buildActiveRunItems()
	if err != nil {
		t.Fatalf("buildActiveRunItems err: %v", err)
	}
	prog, _ := items[0]["progress"].(map[string]any)
	if prog == nil {
		t.Fatalf("missing progress")
	}
	if got := int(prog["logicalTotalCount"].(float64)); got != 1 {
		t.Fatalf("logicalTotalCount=%d, want 1", got)
	}
	if got := int(prog["totalCount"].(float64)); got != 1 {
		t.Fatalf("totalCount=%d, want 1", got)
	}
}
