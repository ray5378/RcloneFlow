package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"rcloneflow/internal/logutil"
	"rcloneflow/internal/service"
)

func TestSplitHistoricalLogSegments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", "", 0},
		{"whitespace only", "   \n  ", 0},
		{"single line no timestamp", "some log line", 1},
		{"single line with timestamp", "2026/04/17 14:34:08 INFO : test", 1},
		{"two segments", "2026/04/17 14:34:08 INFO : first2026/04/17 14:34:09 INFO : second", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segs := logutil.SplitLogSegments(tt.input)
			assert.Len(t, segs, tt.want)
		})
	}
}

func TestClassifyHistoricalLogRow(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		path     string
		msg      string
		wantType string
		wantOk   bool
	}{
		{"copied new", "INFO", "file.txt", "Copied (new)", "copied", true},
		{"copied replaced", "INFO", "file.txt", "Copied (replaced existing)", "copied", true},
		{"deleted", "INFO", "old.txt", "Deleted", "deleted", true},
		{"error", "ERROR", "file.txt", "transfer failed", "failed", true},
		{"other", "INFO", "file.txt", "something", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row, cls, ok := classifyHistoricalLogRow(tt.level, tt.path, tt.msg)
			assert.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				assert.Equal(t, tt.wantType, cls)
				assert.NotNil(t, row)
			} else {
				assert.Nil(t, row)
			}
		})
	}
}

func TestIsCASObjectNotFoundFailureRow(t *testing.T) {
	tests := []struct {
		path string
		msg  string
		want bool
	}{
		{"file.txt", "failed to copy: object not found", true},
		{"file.txt", "failed to copy: no such file", true},
		{"file.txt", "something else", false},
		{"", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.want, isCASObjectNotFoundFailureRow(tt.path, tt.msg))
		})
	}
}

func TestIsCASAttemptObjectNotFoundSummaryRow(t *testing.T) {
	tests := []struct {
		path string
		msg  string
		want bool
	}{
		{"CAS check", "attempt 1: object not found", true},
		{"something", "error", false},
		{"", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.want, isCASAttemptObjectNotFoundSummaryRow(tt.path, tt.msg))
		})
	}
}

func TestIsCASRunObjectNotFoundSummaryRow(t *testing.T) {
	tests := []struct {
		path string
		msg  string
		want bool
	}{
		{"failed to copy", "object not found", true},
		{"failed to copy with retry", "last error was: object not found", true},
		{"summary", "all good", false},
		{"", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.want, isCASRunObjectNotFoundSummaryRow(tt.path, tt.msg))
		})
	}
}

func TestIsCASCompatibleRunSummary(t *testing.T) {
	tests := []struct {
		name    string
		summary map[string]any
		want    bool
	}{
		{"cas mode", map[string]any{"trackingMode": "cas"}, true},
		{"normal mode", map[string]any{"trackingMode": "normal"}, false},
		{"empty", map[string]any{}, false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isCASCompatibleRunSummary(tt.summary))
		})
	}
}

func TestHistoricalSegmentSizeBytes(t *testing.T) {
	tests := []struct {
		seg  string
		want int64
	}{
		{`{"size": 1048576}`, 1048576},
		{`{"size": 500}`, 500},
		{"no size here", 0},
		{"", 0},
		{`{"other": "field"}`, 0},
	}
	for _, tt := range tests {
		t.Run(tt.seg, func(t *testing.T) {
			got := historicalSegmentSizeBytes(tt.seg)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuildLightRunObject(t *testing.T) {
	run := service.RunRecord{
		ID:       1,
		TaskID:   2,
		Status:   "finished",
		Trigger:  "manual",
		TaskName: "test",
	}
	sum := map[string]any{"durationSec": 10}
	obj := buildLightRunObject(run, sum)
	assert.Equal(t, int64(1), obj["id"])
	assert.Equal(t, "finished", obj["status"])
}

func TestEnsureHistoricalFinalSummary(t *testing.T) {
	run := service.RunRecord{
		ID:     1,
		Status: "finished",
	}
	sum := map[string]any{}
	result := ensureHistoricalFinalSummary(run, sum)
	assert.NotNil(t, result)
}

func TestFilterCASHistoricalDetailRows(t *testing.T) {
	rows := []map[string]any{
		{"type": "copied", "path": "file1.txt"},
		{"type": "error", "path": "file2.txt"},
		{"type": "deleted", "path": "file3.txt"},
	}
	filtered := filterCASHistoricalDetailRows(rows)
	assert.Len(t, filtered, 3)
}
