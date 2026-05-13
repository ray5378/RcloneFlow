package runnercli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsCASCompatibleNotFound(t *testing.T) {
	if !isCASCompatibleNotFound("dir/movie.mkv", "object not found", true) {
		t.Fatalf("expected cas-compatible logical path not found to be tolerated")
	}
	if !isCASCompatibleNotFound("dir/movie.mkv", "No such file or directory", true) {
		t.Fatalf("expected no such file to be tolerated in cas mode")
	}
	if isCASCompatibleNotFound("dir/movie.mkv.cas", "object not found", true) {
		t.Fatalf("did not expect .cas path itself to be treated as logical fallback")
	}
	if isCASCompatibleNotFound("dir/movie.mkv", "object not found", false) {
		t.Fatalf("did not expect tolerance outside cas mode")
	}
}

func TestClassifyRunLogRow_CASCompatibleNotFoundBecomesFailedUntilConfirmed(t *testing.T) {
	row, bucket, ok := classifyRunLogRow("ERROR", "dir/movie.mkv", "object not found", map[string]int64{"dir/movie.mkv": 123}, true)
	if !ok {
		t.Fatalf("expected row to be classified")
	}
	if bucket != "failed" {
		t.Fatalf("bucket=%q, want failed", bucket)
	}
	if got := row["status"]; got != "failed" {
		t.Fatalf("status=%v, want failed", got)
	}
	if got := row["action"]; got != "Error" {
		t.Fatalf("action=%v, want Error", got)
	}
	if got := row["sizeBytes"]; got != int64(123) {
		t.Fatalf("sizeBytes=%v, want 123", got)
	}
}

func TestClassifyRunLogRow_NormalErrorStillFails(t *testing.T) {
	row, bucket, ok := classifyRunLogRow("ERROR", "dir/movie.mkv", "permission denied", nil, true)
	if !ok {
		t.Fatalf("expected normal error row to be classified")
	}
	if bucket != "failed" {
		t.Fatalf("bucket=%q, want failed", bucket)
	}
	if got := row["status"]; got != "failed" {
		t.Fatalf("status=%v, want failed", got)
	}
}

func TestClassifyRunLogRow_AttemptObjectNotFoundSummaryIgnored(t *testing.T) {
	if row, bucket, ok := classifyRunLogRow("ERROR", "Attempt 1/3 failed with 5 errors and", "object not found", nil, true); ok || row != nil || bucket != "" {
		t.Fatalf("expected attempt summary to be ignored, got row=%v bucket=%q ok=%v", row, bucket, ok)
	}
}

func TestSanitizeRunLogLine_CASCompatibleNotFound(t *testing.T) {
	line := `2026/05/01 11:04:20 ERROR : dir/movie.mkv: Failed to copy: object not found`
	got := sanitizeRunLogLine(line, true)
	if got != line {
		t.Fatalf("expected line unchanged before CAS confirmation, got %q", got)
	}
}

func TestSanitizeRunLogLine_NonCASOrRealErrorUnchanged(t *testing.T) {
	line := `2026/05/01 11:04:20 ERROR : dir/movie.mkv: Failed to copy: permission denied`
	if got := sanitizeRunLogLine(line, true); got != line {
		t.Fatalf("expected real error line unchanged, got %q", got)
	}
	line2 := `2026/05/01 11:04:20 ERROR : dir/movie.mkv: Failed to copy: object not found`
	if got := sanitizeRunLogLine(line2, false); got != line2 {
		t.Fatalf("expected non-cas line unchanged, got %q", got)
	}
}

func TestIsCASPath(t *testing.T) {
	if !isCASPath("a/b/movie.mkv.cas") {
		t.Fatalf("expected .cas path to be recognized")
	}
	if !isCASPath("A/B/MOVIE.MKV.CAS") {
		t.Fatalf("expected case-insensitive .cas path to be recognized")
	}
	if isCASPath("a/b/movie.mkv") {
		t.Fatalf("expected normal file not to be treated as .cas")
	}
}

func TestTrimCASSuffix(t *testing.T) {
	got := trimCASSuffix("dir/movie.mkv.cas")
	if got != "dir/movie.mkv" {
		t.Fatalf("trimCASSuffix=%q, want %q", got, "dir/movie.mkv")
	}
	got = trimCASSuffix("dir/movie.mkv")
	if got != "dir/movie.mkv" {
		t.Fatalf("trimCASSuffix should keep non-cas path unchanged, got %q", got)
	}
}

func TestNormalizeVisibleTargetPaths_OpenlistCASCompatible(t *testing.T) {
	arr := []map[string]any{
		{"Path": "dir/a.mp4.cas"},
		{"Path": "dir/b.mp4"},
	}
	got := normalizeVisibleTargetPaths(arr, true)
	if _, ok := got["dir/a.mp4.cas"]; !ok {
		t.Fatalf("expected original cas path to remain visible")
	}
	if _, ok := got["dir/a.mp4"]; !ok {
		t.Fatalf("expected cas-compatible logical path to be added")
	}
	if _, ok := got["dir/b.mp4"]; !ok {
		t.Fatalf("expected normal target path to remain visible")
	}
}

func TestNormalizeVisibleTargetPaths_NormalMode(t *testing.T) {
	arr := []map[string]any{
		{"Path": "dir/a.mp4.cas"},
	}
	got := normalizeVisibleTargetPaths(arr, false)
	if _, ok := got["dir/a.mp4.cas"]; !ok {
		t.Fatalf("expected original cas path to remain visible")
	}
	if _, ok := got["dir/a.mp4"]; ok {
		t.Fatalf("did not expect logical non-cas path in normal mode")
	}
}

func TestBuildOpenlistCASCompatPlanFromPaths_Copy(t *testing.T) {
	plan := buildOpenlistCASCompatPlanFromPaths([]string{"dir/a.mp4", "dir/b.mp4.cas"}, []string{"dir/a.mp4.cas", "dir/b.mp4.cas"}, "copy")
	if len(plan.MatchedSource) != 2 {
		t.Fatalf("matched=%v, want 2 items", plan.MatchedSource)
	}
	if len(plan.DestinationExtras) != 0 {
		t.Fatalf("copy should not produce destination extras, got %v", plan.DestinationExtras)
	}
}

func TestBuildOpenlistCASCompatPlanFromPaths_SyncKeepsEquivalentCAS(t *testing.T) {
	plan := buildOpenlistCASCompatPlanFromPaths([]string{"dir/a.mp4"}, []string{"dir/a.mp4.cas", "dir/extra.txt"}, "sync")
	if len(plan.MatchedSource) != 1 || plan.MatchedSource[0] != "dir/a.mp4" {
		t.Fatalf("matched=%v, want [dir/a.mp4]", plan.MatchedSource)
	}
	if len(plan.DestinationExtras) != 1 || plan.DestinationExtras[0] != "dir/extra.txt" {
		t.Fatalf("extras=%v, want [dir/extra.txt]", plan.DestinationExtras)
	}
}

func TestBuildOpenlistCASCompatPlanFromPaths_SourceCASStillNormal(t *testing.T) {
	plan := buildOpenlistCASCompatPlanFromPaths([]string{"dir/a.mp4.cas"}, []string{"dir/a.mp4"}, "copy")
	if len(plan.MatchedSource) != 0 {
		t.Fatalf("source .cas should not map backward to normal file, got matched=%v", plan.MatchedSource)
	}
}

func TestExpectedVisibleDestinationPaths(t *testing.T) {
	plan := &openlistCASCompatPlan{SourceFiles: []string{"dir/a.mp4", "dir/b.mp4"}, MatchedSource: []string{"dir/a.mp4"}}
	copyExpected := expectedVisibleDestinationPaths(plan, "copy")
	if len(copyExpected) != 2 {
		t.Fatalf("copy expected=%v, want all source files", copyExpected)
	}
	moveExpected := expectedVisibleDestinationPaths(plan, "move")
	if len(moveExpected) != 1 || moveExpected[0] != "dir/a.mp4" {
		t.Fatalf("move expected=%v, want matched files only", moveExpected)
	}
}

func TestAreAllExpectedPathsVisible(t *testing.T) {
	visible := map[string]struct{}{"dir/a.mp4": {}, "dir/b.mp4": {}}
	if !areAllExpectedPathsVisible([]string{"dir/a.mp4"}, visible) {
		t.Fatalf("expected visible check to pass")
	}
	if areAllExpectedPathsVisible([]string{"dir/c.mp4"}, visible) {
		t.Fatalf("expected visible check to fail for missing path")
	}
}

func TestApplyPostActions_MoveDeletesMatchedSource(t *testing.T) {
	calls := runApplyPostActionsWithFakeRclone(t, &openlistCASCompatPlan{MatchedSource: []string{"dir/a.mp4", "dir/b.mp4"}}, "cfg.conf", "src:", "dst:", "move")
	if len(calls) != 2 {
		t.Fatalf("expected 2 delete calls, got %d: %v", len(calls), calls)
	}
	if !strings.Contains(calls[0], "deletefile src:dir/a.mp4 --config cfg.conf") {
		t.Fatalf("unexpected first call: %q", calls[0])
	}
	if !strings.Contains(calls[1], "deletefile src:dir/b.mp4 --config cfg.conf") {
		t.Fatalf("unexpected second call: %q", calls[1])
	}
}

func TestApplyPostActions_SyncDeletesDestinationExtras(t *testing.T) {
	calls := runApplyPostActionsWithFakeRclone(t, &openlistCASCompatPlan{DestinationExtras: []string{"dir/extra.txt", "dir/old.cas"}}, "cfg.conf", "src:", "dst:", "sync")
	if len(calls) != 2 {
		t.Fatalf("expected 2 delete calls, got %d: %v", len(calls), calls)
	}
	if !strings.Contains(calls[0], "deletefile dst:dir/extra.txt --config cfg.conf") {
		t.Fatalf("unexpected first call: %q", calls[0])
	}
	if !strings.Contains(calls[1], "deletefile dst:dir/old.cas --config cfg.conf") {
		t.Fatalf("unexpected second call: %q", calls[1])
	}
}

func TestApplyPostActions_CopyDoesNothing(t *testing.T) {
	calls := runApplyPostActionsWithFakeRclone(t, &openlistCASCompatPlan{MatchedSource: []string{"dir/a.mp4"}, DestinationExtras: []string{"dir/extra.txt"}}, "cfg.conf", "src:", "dst:", "copy")
	if len(calls) != 0 {
		t.Fatalf("copy mode should not delete anything, got %v", calls)
	}
}

func TestAnalyzeCASAttemptLogSegment_AllCASMatched_NoRealFailures(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_cas_attempt_allcas_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logPath := filepath.Join(tmpDir, "attempt.log")
	content := strings.Join([]string{
		`{"time":"2026-05-13T14:37:37+08:00","level":"error","msg":"Failed to copy: object not found","object":"a/file1.mkv"}`,
		`NOTICE : a/file1.mkv: CAS compatible match after source cleanup (Failed to copy: object not found)`,
		`{"time":"2026-05-13T14:39:41+08:00","level":"error","msg":"Failed to copy: object not found","object":"a/file2.mkv"}`,
		`NOTICE : a/file2.mkv: CAS compatible match after source cleanup (Failed to copy: object not found)`,
		`{"time":"2026-05-13T14:41:53+08:00","level":"error","msg":"Attempt 1/3 failed with 2 errors and: object not found"}`,
	}, "\n") + "\n"
	if err := os.WriteFile(logPath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got := analyzeCASAttemptLogSegment(logPath, 0, true)
	if len(got.CASMatchedPaths) != 2 {
		t.Fatalf("CASMatchedPaths=%v, want 2 items", got.CASMatchedPaths)
	}
	if len(got.RealFailures) != 0 {
		t.Fatalf("RealFailures=%v, want empty", got.RealFailures)
	}
}

func TestBuildFinalSummaryFilesFromLog_AllCASMatchedSuppressesObjectNotFoundFailures(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_finalsummary_allcas_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logPath := filepath.Join(tmpDir, "1526.log")
	content := strings.Join([]string{
		`{"time":"2026-05-13T15:27:58.58768172+08:00","level":"error","msg":"Failed to copy: object not found","object":"电视剧/国产剧/罪无可逃 (2026)/Season 1/罪无可逃 - S01E01 - 第 1 集.mkv"}`,
		`NOTICE : 电视剧/国产剧/罪无可逃 (2026)/Season 1/罪无可逃 - S01E01 - 第 1 集.mkv: CAS compatible match after source cleanup (Failed to copy: object not found)`,
		`{"time":"2026-05-13T15:30:01.348575953+08:00","level":"error","msg":"Failed to copy: object not found","object":"电视剧/国产剧/罪无可逃 (2026)/Season 1/罪无可逃 - S01E02 - 第 2 集.mkv"}`,
		`NOTICE : 电视剧/国产剧/罪无可逃 (2026)/Season 1/罪无可逃 - S01E02 - 第 2 集.mkv: CAS compatible match after source cleanup (Failed to copy: object not found)`,
		`{"time":"2026-05-13T15:30:01.348816288+08:00","level":"error","msg":"Attempt 1/1 failed with 2 errors and: object not found"}`,
		`ERROR : Attempt 1/1 failed with 2 errors and: object not found`,
		`{"time":"2026-05-13T15:30:01.363364561+08:00","level":"notice","msg":"Failed to copy with 2 errors: last error was: object not found"}`,
		`ERROR : Failed to copy with 2 errors: last error was: object not found`,
	}, "\n") + "\n"
	if err := os.WriteFile(logPath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	files, counts := buildFinalSummaryFilesFromLog(logPath, true, false)
	if got := counts["copied"]; got != 2 {
		t.Fatalf("copied=%d, want 2; counts=%v files=%v", got, counts, files)
	}
	if got := counts["failed"]; got != 0 {
		t.Fatalf("failed=%d, want 0; counts=%v files=%v", got, counts, files)
	}
	if got := counts["total"]; got != 2 {
		t.Fatalf("total=%d, want 2; counts=%v files=%v", got, counts, files)
	}
}

func TestAnalyzeCASAttemptLogSegment_RealFailureRemains(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_cas_attempt_realfail_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logPath := filepath.Join(tmpDir, "attempt.log")
	content := strings.Join([]string{
		`{"time":"2026-05-13T14:37:37+08:00","level":"error","msg":"Failed to copy: object not found","object":"a/file1.mkv"}`,
		`NOTICE : a/file1.mkv: CAS compatible match after source cleanup (Failed to copy: object not found)`,
		`{"time":"2026-05-13T14:43:23+08:00","level":"error","msg":"Failed to copy: unchunked simple update failed: Method Not Allowed: 405 Method Not Allowed","object":"a/file1.mkv"}`,
		`{"time":"2026-05-13T14:44:00+08:00","level":"error","msg":"Failed to copy: permission denied","object":"a/file2.mkv"}`,
	}, "\n") + "\n"
	if err := os.WriteFile(logPath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got := analyzeCASAttemptLogSegment(logPath, 0, true)
	if len(got.CASMatchedPaths) != 1 {
		t.Fatalf("CASMatchedPaths=%v, want 1 item", got.CASMatchedPaths)
	}
	if len(got.RealFailures) != 2 {
		t.Fatalf("RealFailures=%v, want 2 items", got.RealFailures)
	}
	if _, ok := got.RealFailures["a/file1.mkv"]; !ok {
		t.Fatalf("expected file1 real failure to remain, got %v", got.RealFailures)
	}
	if _, ok := got.RealFailures["a/file2.mkv"]; !ok {
		t.Fatalf("expected file2 real failure to remain, got %v", got.RealFailures)
	}
}

func TestBuildOpenlistCASCompatPlanFromPaths_SyncNestedDirectories(t *testing.T) {
	plan := buildOpenlistCASCompatPlanFromPaths(
		[]string{"top/sub/a.mp4", "top/sub/keep.txt"},
		[]string{"top/sub/a.mp4.cas", "top/sub/keep.txt", "top/sub/old.bin", "top/other/ghost.txt"},
		"sync",
	)
	if len(plan.MatchedSource) != 1 || plan.MatchedSource[0] != "top/sub/a.mp4" {
		t.Fatalf("matched=%v, want [top/sub/a.mp4]", plan.MatchedSource)
	}
	wantExtras := map[string]struct{}{"top/sub/old.bin": {}, "top/other/ghost.txt": {}}
	if len(plan.DestinationExtras) != len(wantExtras) {
		t.Fatalf("extras=%v, want %d extras", plan.DestinationExtras, len(wantExtras))
	}
	for _, p := range plan.DestinationExtras {
		if _, ok := wantExtras[p]; !ok {
			t.Fatalf("unexpected extra path %q in %v", p, plan.DestinationExtras)
		}
	}
}

func TestBuildOpenlistCASCompatPlanFromPaths_TargetHasBothOriginalAndCAS(t *testing.T) {
	plan := buildOpenlistCASCompatPlanFromPaths(
		[]string{"dir/a.mp4"},
		[]string{"dir/a.mp4", "dir/a.mp4.cas"},
		"copy",
	)
	if len(plan.MatchedSource) != 1 || plan.MatchedSource[0] != "dir/a.mp4" {
		t.Fatalf("matched=%v, want [dir/a.mp4] when target also has .cas", plan.MatchedSource)
	}
}

func TestBuildOpenlistCASCompatPlanFromPaths_SpecialCharacterPaths(t *testing.T) {
	plan := buildOpenlistCASCompatPlanFromPaths(
		[]string{"目录/with space/电影 01.mp4", "emoji/📦.txt"},
		[]string{"目录/with space/电影 01.mp4.cas", "emoji/📦.txt.cas"},
		"move",
	)
	wantMatched := map[string]struct{}{"目录/with space/电影 01.mp4": {}, "emoji/📦.txt": {}}
	if len(plan.MatchedSource) != len(wantMatched) {
		t.Fatalf("matched=%v, want %d items", plan.MatchedSource, len(wantMatched))
	}
	for _, p := range plan.MatchedSource {
		if _, ok := wantMatched[p]; !ok {
			t.Fatalf("unexpected matched path %q in %v", p, plan.MatchedSource)
		}
	}
}

func TestNormalizeVisibleTargetPaths_TargetHasBothOriginalAndCAS(t *testing.T) {
	arr := []map[string]any{{"Path": "dir/a.mp4"}, {"Path": "dir/a.mp4.cas"}}
	got := normalizeVisibleTargetPaths(arr, true)
	if _, ok := got["dir/a.mp4"]; !ok {
		t.Fatalf("expected original path to remain visible")
	}
	if _, ok := got["dir/a.mp4.cas"]; !ok {
		t.Fatalf("expected cas path to remain visible")
	}
	if len(got) != 2 {
		t.Fatalf("expected exactly 2 logical keys, got %v", got)
	}
}

func TestApplyPostActions_MoveDeletesNestedAndSpecialPaths(t *testing.T) {
	calls := runApplyPostActionsWithFakeRclone(t, &openlistCASCompatPlan{MatchedSource: []string{"top/sub/a.mp4", "目录/with space/电影 01.mp4"}}, "cfg.conf", "src:", "dst:", "move")
	if len(calls) != 2 {
		t.Fatalf("expected 2 delete calls, got %d: %v", len(calls), calls)
	}
	if !strings.Contains(calls[0], "deletefile src:top/sub/a.mp4 --config cfg.conf") {
		t.Fatalf("unexpected first call: %q", calls[0])
	}
	if !strings.Contains(calls[1], "deletefile src:目录/with space/电影 01.mp4 --config cfg.conf") {
		t.Fatalf("unexpected second call: %q", calls[1])
	}
}

func runApplyPostActionsWithFakeRclone(t *testing.T, plan *openlistCASCompatPlan, cfg, src, dst, mode string) []string {
	t.Helper()
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "calls.log")
	binPath := filepath.Join(tmpDir, "rclone")
	script := "#!/bin/sh\n" +
		"printf '%s\\n' \"$*\" >> \"" + logFile + "\"\n"
	if err := os.WriteFile(binPath, []byte(script), 0o755); err != nil {
		t.Fatalf("WriteFile(fake rclone) error = %v", err)
	}
	oldBin := os.Getenv("RCLONE_BIN")
	if err := os.Setenv("RCLONE_BIN", binPath); err != nil {
		t.Fatalf("Setenv(RCLONE_BIN) error = %v", err)
	}
	defer func() {
		if oldBin == "" {
			_ = os.Unsetenv("RCLONE_BIN")
		} else {
			_ = os.Setenv("RCLONE_BIN", oldBin)
		}
	}()
	if err := plan.ApplyPostActions(cfg, src, dst, mode); err != nil {
		t.Fatalf("ApplyPostActions() error = %v", err)
	}
	data, err := os.ReadFile(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("ReadFile(calls.log) error = %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil
	}
	return lines
}
