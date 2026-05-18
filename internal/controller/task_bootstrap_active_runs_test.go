package controller

import (
	"encoding/json"
	"testing"

	"rcloneflow/internal/service"
)

func TestBuildActiveRunItems_UsesLogicalTotalsAndKeepsPlannedFiles(t *testing.T) {
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
	ctrl := &TaskController{runSvc: service.NewRunService(&mockRunSvcDB{runs: []service.RunRecord{{
		ID:        3,
		TaskID:    102,
		Status:    "running",
		StartedAt: "2026-04-17T14:30:00+08:00",
		Summary:   string(bs),
	}}})}

	items, err := ctrl.buildActiveRunItems()
	if err != nil {
		t.Fatalf("buildActiveRunItems err: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items)=%d, want 1", len(items))
	}
	prog, _ := items[0]["progress"].(map[string]any)
	if prog == nil {
		t.Fatalf("missing progress")
	}
	if _, ok := prog["plannedFiles"]; !ok {
		t.Fatalf("missing plannedFiles: %#v", prog)
	}
	if _, ok := prog["logicalTotalCount"]; !ok {
		t.Fatalf("missing logicalTotalCount: %#v", prog)
	}
	if _, ok := prog["totalCount"]; !ok {
		t.Fatalf("missing totalCount: %#v", prog)
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
}

func TestBuildActiveRunItems_UsesPreflightTotalCountFallback(t *testing.T) {
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
	ctrl := &TaskController{runSvc: service.NewRunService(&mockRunSvcDB{runs: []service.RunRecord{{
		ID:        5,
		TaskID:    104,
		Status:    "running",
		StartedAt: "2026-04-17T14:30:00+08:00",
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
	if _, ok := prog["totalCount"]; !ok {
		t.Fatalf("missing totalCount: %#v", prog)
	}
	if _, ok := prog["logicalTotalCount"]; !ok {
		t.Fatalf("missing logicalTotalCount: %#v", prog)
	}
	if got := int(prog["totalCount"].(float64)); got != 166 {
		t.Fatalf("totalCount=%d, want 166", got)
	}
	if got := int(prog["logicalTotalCount"].(float64)); got != 166 {
		t.Fatalf("logicalTotalCount=%d, want 166", got)
	}
}
