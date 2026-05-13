package runnercli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"rcloneflow/internal/active_transfer"
	"rcloneflow/internal/store"
)

func TestBuildArgs_ForceJSONLog(t *testing.T) {
	args := []string{"copy", "src:/a", "dst:/b", "--stats", "1s", "--stats-one-line", "--config", "/tmp/rclone.conf"}
	args = append(args, "--use-json-log", "--log-level", "INFO", "--stats-log-level", "INFO")
	joined := strings.Join(args, " ")
	if strings.Contains(joined, " -v ") || strings.HasSuffix(joined, " -v") || strings.Contains(joined, "-v --use-json-log") {
		t.Fatalf("forced json-log args should not contain -v, got %q", joined)
	}
	if !strings.Contains(joined, "--use-json-log") || !strings.Contains(joined, "--log-level INFO") || !strings.Contains(joined, "--stats-log-level INFO") {
		t.Fatalf("forced json-log args missing expected flags: %q", joined)
	}
}

func TestParseOneLineProgress_AggregateWithXfrAndETA(t *testing.T) {
	line := `2026/04/17 14:34:08 INFO : 121.377 MiB / 335.968 MiB, 36%, 2.474 MiB/s, ETA 1m26s (xfr#18/53)`
	prog, ok := parseOneLineProgress(line)
	if !ok {
		t.Fatalf("expected aggregate line to parse")
	}
	if got := int(prog["completedFiles"].(float64)); got != 18 {
		t.Fatalf("completedFiles=%d, want 18", got)
	}
	if got := int(prog["plannedFiles"].(float64)); got != 53 {
		t.Fatalf("plannedFiles=%d, want 53", got)
	}
	if got := int(prog["eta"].(float64)); got != 86 {
		t.Fatalf("eta=%d, want 86", got)
	}
	if got := int(prog["percentage"].(float64)); got != 36 {
		t.Fatalf("percentage=%d, want 36", got)
	}
}

func TestParseOneLineProgress_AggregateWithETAOnly(t *testing.T) {
	line := `2026/04/17 14:34:09 INFO : 123.377 MiB / 335.968 MiB, 37%, 2.434 MiB/s, ETA 1m27s`
	prog, ok := parseOneLineProgress(line)
	if !ok {
		t.Fatalf("expected aggregate ETA-only line to parse")
	}
	if got := int(prog["eta"].(float64)); got != 87 {
		t.Fatalf("eta=%d, want 87", got)
	}
	if got := int(prog["percentage"].(float64)); got != 37 {
		t.Fatalf("percentage=%d, want 37", got)
	}
	if got := int(prog["bytes"].(float64)); got <= 1024*1024 {
		t.Fatalf("bytes=%d looks wrong; likely matched timestamp instead of MiB payload", got)
	}
	if got := int(prog["totalBytes"].(float64)); got <= 1024*1024 {
		t.Fatalf("totalBytes=%d looks wrong; likely matched timestamp instead of MiB payload", got)
	}
}

func TestParseOneLineProgress_AggregateWithDayETA(t *testing.T) {
	line := `2026/05/01 10:52:47 INFO  :   991.070 MiB / 932.457 GiB, 0%, 2.516 MiB/s, ETA 4d9h18m (xfr#3/777)`
	prog, ok := parseOneLineProgress(line)
	if !ok {
		t.Fatalf("expected aggregate day-eta line to parse")
	}
	if got := int(prog["completedFiles"].(float64)); got != 3 {
		t.Fatalf("completedFiles=%d, want 3", got)
	}
	if got := int(prog["plannedFiles"].(float64)); got != 777 {
		t.Fatalf("plannedFiles=%d, want 777", got)
	}
	if got := int(prog["eta"].(float64)); got != (4*24*3600 + 9*3600 + 18*60) {
		t.Fatalf("eta=%d, want %d", got, 4*24*3600+9*3600+18*60)
	}
}

func TestParseOneLineProgress_IgnoreFileLevelProgress(t *testing.T) {
	line := `2026/04/17 14:34:08 INFO : 20260417/20260417135617-61000.mp4: 10.000 MiB / 100.000 MiB, 10%, 2.474 MiB/s`
	if prog, ok := parseOneLineProgress(line); ok || prog != nil {
		t.Fatalf("expected file-level progress line to be ignored, got %#v", prog)
	}
}

func TestParseOneLineProgress_IgnoreFileCopied(t *testing.T) {
	line := `2026/04/17 14:34:07 INFO : 20260417/20260417135617-61000.mp4: Copied (new)`
	if prog, ok := parseOneLineProgress(line); ok || prog != nil {
		t.Fatalf("expected copied line to be ignored, got %#v", prog)
	}
}

func TestParseOneLineProgress_IgnoreDeleted(t *testing.T) {
	line := `2026/04/17 14:34:07 INFO : 20260417/20260417135617-61000.mp4: Deleted`
	if prog, ok := parseOneLineProgress(line); ok || prog != nil {
		t.Fatalf("expected deleted line to be ignored, got %#v", prog)
	}
}

func TestFileCASMatchedRe_MatchesCASCompatibleNotice(t *testing.T) {
	line := `2026/05/01 12:49:10 NOTICE: 电视剧/国产剧/佳偶天成 (2026)/Season 1/佳偶天成 - S01E16 - 第 16 集.mkv: CAS compatible match after source cleanup (Failed to copy: object not found)`
	m := fileCASMatchedRe.FindStringSubmatch(line)
	if len(m) < 2 {
		t.Fatalf("expected CAS compatible notice to match, got %#v", m)
	}
	if got := m[1]; got != `电视剧/国产剧/佳偶天成 (2026)/Season 1/佳偶天成 - S01E16 - 第 16 集.mkv` {
		t.Fatalf("matched path=%q", got)
	}
}

func TestParseOneLineProgress_IgnoreNonAggregateWithoutETAOrXfr(t *testing.T) {
	line := `2026/04/17 14:34:10 NOTICE: something happened`
	if prog, ok := parseOneLineProgress(line); ok || prog != nil {
		t.Fatalf("expected unrelated line to be ignored, got %#v", prog)
	}
}

func TestClassifyRunLogRow_CASNoticeCountsAsCopied(t *testing.T) {
	row, bucket, ok := classifyRunLogRow(
		"NOTICE",
		"电视剧/国产剧/人间惊鸿客 (2026)/Season 1/人间惊鸿客 - S01E18 - 第 18 集.mkv",
		"CAS compatible match after source cleanup (Failed to copy: object not found)",
		map[string]int64{},
		true,
	)
	if !ok {
		t.Fatalf("expected CAS notice to be classified")
	}
	if bucket != "copied" {
		t.Fatalf("bucket=%q, want copied", bucket)
	}
	if got := row["status"]; got != "success" {
		t.Fatalf("status=%v, want success", got)
	}
	if got := row["action"]; got != "CAS Matched" {
		t.Fatalf("action=%v, want CAS Matched", got)
	}
}

func TestBuildFinalSummaryFilesFromLog_JSONMoveMergesCopiedAndDeleted(t *testing.T) {
	tmp, err := os.CreateTemp("", "rcloneflow-finalsummary-json-*.log")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()
	logText := "{\"level\":\"info\",\"msg\":\"Copied (new)\",\"object\":\"20260510/a.mp4\",\"time\":\"2026-05-10T14:49:03+08:00\",\"size\":123}\n" +
		"{\"level\":\"info\",\"msg\":\"Deleted\",\"object\":\"20260510/a.mp4\",\"time\":\"2026-05-10T14:49:04+08:00\"}\n" +
		"{\"level\":\"info\",\"msg\":\"Copied (new)\",\"object\":\"20260510/b.mp4\",\"time\":\"2026-05-10T14:49:05+08:00\"}\n"
	if _, err := tmp.WriteString(logText); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}
	files, counts := buildFinalSummaryFilesFromLog(tmp.Name(), false, true)
	if got := len(files); got != 2 {
		t.Fatalf("len(files)=%d, want 2; files=%#v", got, files)
	}
	if got := counts["copied"]; got != 2 {
		t.Fatalf("copied=%d, want 2; counts=%#v", got, counts)
	}
	if got := counts["deleted"]; got != 0 {
		t.Fatalf("deleted=%d, want 0; counts=%#v", got, counts)
	}
	actions := []string{anyString(files[0]["action"]), anyString(files[1]["action"])}
	if !(strings.Contains(strings.Join(actions, ","), "Moved") && strings.Contains(strings.Join(actions, ","), "Copied")) {
		t.Fatalf("unexpected actions=%#v", actions)
	}
}

func TestConsume_CASNoticeIncrementsCompletedFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runnercli_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	defer db.Close()

	task, err := db.AddTask(store.Task{
		Name:         "cas-notice-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/a",
		TargetRemote: "dst",
		TargetPath:   "/b",
	})
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}

	run, err := db.AddRun(store.Run{
		TaskID:  task.ID,
		Status:  "running",
		Trigger: "manual",
		Summary: map[string]any{
			"progress": map[string]any{
				"completedFiles": float64(0),
			},
		},
	})
	if err != nil {
		t.Fatalf("AddRun() error = %v", err)
	}

	r := New(db)
	r.casVerifier = func(cfg, dst, rel string) (bool, error) { return true, nil }
	r.casVerifyDelays = nil
	fp := &fileProgress{m: map[string]*fileProg{}}
	outFile, err := os.CreateTemp(tmpDir, "consume-log-*.log")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	defer outFile.Close()

	line := "2026/05/01 14:01:20 NOTICE: 电视剧/国产剧/人间惊鸿客 (2026)/Season 1/人间惊鸿客 - S01E18 - 第 18 集.mkv: CAS compatible match after source cleanup (Failed to copy: object not found)\n"
	r.consume(run.ID, strings.NewReader(line), outFile, true, fp, true, false, "/tmp/rclone.conf", "dst:/b", "")

	gotRun, err := db.GetRun(run.ID)
	if err != nil {
		t.Fatalf("GetRun() error = %v", err)
	}
	prog, _ := gotRun.Summary["progress"].(map[string]any)
	if prog == nil {
		t.Fatalf("expected progress map, got %#v", gotRun.Summary)
	}
	if got := int(prog["completedFiles"].(float64)); got != 1 {
		t.Fatalf("completedFiles=%d, want 1; summary=%#v", got, gotRun.Summary)
	}
	files, _ := gotRun.Summary["files"].([]fileProg)
	_ = files
	if got := len(fp.copiedList()); got != 1 {
		t.Fatalf("fp.copied len=%d, want 1", got)
	}
}

func TestConsume_JSONWrappedFileProgressUpdatesActiveTransfer(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runnercli_json_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	defer db.Close()

	task, err := db.AddTask(store.Task{Name: "json-progress-task", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}
	run, err := db.AddRun(store.Run{TaskID: task.ID, Status: "running", Trigger: "manual", Summary: map[string]any{}})
	if err != nil {
		t.Fatalf("AddRun() error = %v", err)
	}

	mgr := active_transfer.NewManager()
	mgr.InitState(run.ID, task.ID, active_transfer.TrackingModeNormal, []active_transfer.TransferCandidateFile{{Path: "a/file1.mkv", Name: "file1.mkv", SizeBytes: 100}})
	r := New(db, mgr)
	fp := &fileProgress{m: map[string]*fileProg{}}
	outFile, err := os.CreateTemp(tmpDir, "consume-json-log-*.log")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	defer outFile.Close()

	line := `{"level":"info","msg":"12.000 MiB / 36.000 MiB, 33%, 4.000 MiB/s, ETA 6s (xfr#0/1)","object":"a/file1.mkv"}` + "\n"
	r.consume(run.ID, strings.NewReader(line), outFile, true, fp, false, false, "/tmp/rclone.conf", "dst:/b", "")

	st, ok := mgr.GetByTaskID(task.ID)
	if !ok || st.CurrentFile == nil {
		t.Fatalf("expected current file state, got ok=%v st=%#v", ok, st)
	}
	if got := st.CurrentFile.Path; got != "a/file1.mkv" {
		t.Fatalf("current file path=%q, want a/file1.mkv", got)
	}
	if got := st.CurrentFile.TotalBytes; got <= st.CurrentFile.Bytes {
		t.Fatalf("unexpected current file progress: %#v", st.CurrentFile)
	}
	gotRun, err := db.GetRun(run.ID)
	if err != nil {
		t.Fatalf("GetRun() error = %v", err)
	}
	if got := anyString(gotRun.Summary["progressLine"]); !strings.Contains(got, "a/file1.mkv") {
		t.Fatalf("progressLine=%q, want file path included", got)
	}
}

func TestConsume_JSONErrorObjectNotFoundMarksCASMatchedInActiveTransfer(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runnercli_json_error_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	defer db.Close()

	task, err := db.AddTask(store.Task{Name: "json-error-cas-task", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}
	run, err := db.AddRun(store.Run{TaskID: task.ID, Status: "running", Trigger: "manual", Summary: map[string]any{"progress": map[string]any{"completedFiles": float64(0)}}})
	if err != nil {
		t.Fatalf("AddRun() error = %v", err)
	}

	mgr := active_transfer.NewManager()
	mgr.InitState(run.ID, task.ID, active_transfer.TrackingModeCAS, []active_transfer.TransferCandidateFile{{Path: "a/file1.mkv", Name: "file1.mkv", SizeBytes: 100}})
	r := New(db, mgr)
	called := 0
	r.casVerifier = func(cfg, dst, rel string) (bool, error) { called++; return true, nil }
	r.casVerifyDelays = nil
	fp := &fileProgress{m: map[string]*fileProg{}}
	outFile, err := os.CreateTemp(tmpDir, "consume-json-error-log-*.log")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	defer outFile.Close()

	line := `{"time":"2026-05-13T10:08:29.48904232+08:00","level":"error","msg":"Failed to copy: object not found","object":"a/file1.mkv","objectType":"*smb.Object","source":"slog/logger.go:256"}` + "\n"
	r.consume(run.ID, strings.NewReader(line), outFile, true, fp, true, false, "/tmp/rclone.conf", "dst:/b", "")

	completed := mgr.ListCompleted(task.ID, 0, 10)
	if completed.Total != 1 {
		t.Fatalf("completed total=%d, want 1", completed.Total)
	}
	if got := completed.Items[0].Status; got != active_transfer.FileStatusCASMatched {
		t.Fatalf("completed status=%q, want %q", got, active_transfer.FileStatusCASMatched)
	}
	pending := mgr.ListPending(task.ID, 0, 10)
	if pending.Total != 0 {
		t.Fatalf("pending total=%d, want 0", pending.Total)
	}
	gotRun, err := db.GetRun(run.ID)
	if err != nil {
		t.Fatalf("GetRun() error = %v", err)
	}
	prog, _ := gotRun.Summary["progress"].(map[string]any)
	if prog == nil {
		t.Fatalf("expected progress map, got %#v", gotRun.Summary)
	}
	if gotv, ok := prog["completedFiles"].(float64); !ok || int(gotv) != 1 {
		t.Fatalf("completedFiles=%#v, want 1; progress=%#v", prog["completedFiles"], prog)
	}
	if called == 0 {
		t.Fatalf("expected casVerifier to be called")
	}
}

func TestConsume_JSONErrorObjectNotFoundAppendsRuntimeExclude(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runnercli_json_error_exclude_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	defer db.Close()

	task, err := db.AddTask(store.Task{Name: "json-error-exclude-task", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}
	run, err := db.AddRun(store.Run{TaskID: task.ID, Status: "running", Trigger: "manual", Summary: map[string]any{"progress": map[string]any{"completedFiles": float64(0)}}})
	if err != nil {
		t.Fatalf("AddRun() error = %v", err)
	}

	mgr := active_transfer.NewManager()
	mgr.InitState(run.ID, task.ID, active_transfer.TrackingModeCAS, []active_transfer.TransferCandidateFile{{Path: "a/file1.mkv", Name: "file1.mkv", SizeBytes: 100}})
	r := New(db, mgr)
	r.casVerifier = func(cfg, dst, rel string) (bool, error) { return true, nil }
	r.casVerifyDelays = nil
	fp := &fileProgress{m: map[string]*fileProg{}}
	outFile, err := os.CreateTemp(tmpDir, "consume-json-error-exclude-log-*.log")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	defer outFile.Close()
	excludeFile := filepath.Join(tmpDir, "runtime-exclude.txt")

	line := `{"time":"2026-05-13T10:08:29.48904232+08:00","level":"error","msg":"Failed to copy: object not found","object":"a/file1.mkv","objectType":"*smb.Object","source":"slog/logger.go:256"}` + "\n"
	r.consume(run.ID, strings.NewReader(line), outFile, true, fp, true, false, "/tmp/rclone.conf", "dst:/b", excludeFile)

	b, err := os.ReadFile(excludeFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if got := strings.TrimSpace(string(b)); got != "a/file1.mkv" {
		t.Fatalf("runtime exclude=%q, want %q", got, "a/file1.mkv")
	}
}

func TestConsume_JSONErrorObjectNotFoundWithoutCASStaysFailed(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runnercli_json_error_failed_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	defer db.Close()

	task, err := db.AddTask(store.Task{Name: "json-error-failed-task", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}
	run, err := db.AddRun(store.Run{TaskID: task.ID, Status: "running", Trigger: "manual", Summary: map[string]any{"progress": map[string]any{"completedFiles": float64(0)}}})
	if err != nil {
		t.Fatalf("AddRun() error = %v", err)
	}

	mgr := active_transfer.NewManager()
	mgr.InitState(run.ID, task.ID, active_transfer.TrackingModeCAS, []active_transfer.TransferCandidateFile{{Path: "a/file1.mkv", Name: "file1.mkv", SizeBytes: 100}})
	r := New(db, mgr)
	r.casVerifier = func(cfg, dst, rel string) (bool, error) { return false, nil }
	r.casVerifyDelays = nil
	fp := &fileProgress{m: map[string]*fileProg{}}
	outFile, err := os.CreateTemp(tmpDir, "consume-json-error-failed-log-*.log")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	defer outFile.Close()

	line := `{"time":"2026-05-13T10:08:29.48904232+08:00","level":"error","msg":"Failed to copy: object not found","object":"a/file1.mkv","objectType":"*smb.Object","source":"slog/logger.go:256"}` + "\n"
	r.consume(run.ID, strings.NewReader(line), outFile, true, fp, true, false, "/tmp/rclone.conf", "dst:/b", "")

	completed := mgr.ListCompleted(task.ID, 0, 10)
	if completed.Total != 1 {
		t.Fatalf("completed total=%d, want 1", completed.Total)
	}
	if got := completed.Items[0].Status; got != active_transfer.FileStatusFailed {
		t.Fatalf("completed status=%q, want %q", got, active_transfer.FileStatusFailed)
	}
	pending := mgr.ListPending(task.ID, 0, 10)
	if pending.Total != 0 {
		t.Fatalf("pending total=%d, want 0", pending.Total)
	}
	gotRun, err := db.GetRun(run.ID)
	if err != nil {
		t.Fatalf("GetRun() error = %v", err)
	}
	prog, _ := gotRun.Summary["progress"].(map[string]any)
	if prog == nil {
		t.Fatalf("expected progress map, got %#v", gotRun.Summary)
	}
	if gotv, ok := prog["completedFiles"].(float64); !ok || int(gotv) != 0 {
		t.Fatalf("completedFiles=%#v, want 0; progress=%#v", prog["completedFiles"], prog)
	}
}

func TestConsume_JSONStatsTransferringUpdatesCurrentFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runnercli_json_stats_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	defer db.Close()

	task, err := db.AddTask(store.Task{Name: "json-stats-task", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}
	run, err := db.AddRun(store.Run{TaskID: task.ID, Status: "running", Trigger: "manual", Summary: map[string]any{}})
	if err != nil {
		t.Fatalf("AddRun() error = %v", err)
	}

	mgr := active_transfer.NewManager()
	mgr.InitState(run.ID, task.ID, active_transfer.TrackingModeNormal, []active_transfer.TransferCandidateFile{{Path: "电视剧/国产剧/风过留痕 (2026)/Season 1/风过留痕 - S01E01 - 第 1 集.mkv", Name: "风过留痕 - S01E01 - 第 1 集.mkv", SizeBytes: 1545914693}})
	r := New(db, mgr)
	fp := &fileProgress{m: map[string]*fileProg{}}
	outFile, err := os.CreateTemp(tmpDir, "consume-json-stats-log-*.log")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	defer outFile.Close()

	line := `{"time":"2026-05-10T10:48:20+08:00","level":"info","msg":"4.996 MiB / 2.829 GiB, 0%, 1.665 MiB/s, ETA 28m56s (xfr#0/2)\n","stats":{"bytes":5238784,"totalBytes":3037871949,"speed":1746263.2,"eta":1736,"transferring":[{"bytes":5238784,"name":"电视剧/国产剧/风过留痕 (2026)/Season 1/风过留痕 - S01E01 - 第 1 集.mkv","percentage":0,"size":1545914693,"speed":4232516.1}]}}` + "\n"
	r.consume(run.ID, strings.NewReader(line), outFile, true, fp, false, false, "/tmp/rclone.conf", "dst:/b", "")

	st, ok := mgr.GetByTaskID(task.ID)
	if !ok || st.CurrentFile == nil {
		t.Fatalf("expected current file state, got ok=%v st=%#v", ok, st)
	}
	if got := st.CurrentFile.Path; !strings.Contains(got, "风过留痕 - S01E01 - 第 1 集.mkv") {
		t.Fatalf("current file path=%q", got)
	}
	if got := st.CurrentFile.TotalBytes; got != 1545914693 {
		t.Fatalf("totalBytes=%d, want 1545914693", got)
	}
	if got := st.CurrentFile.Bytes; got != 5238784 {
		t.Fatalf("bytes=%d, want 5238784", got)
	}
	gotRun, err := db.GetRun(run.ID)
	if err != nil {
		t.Fatalf("GetRun() error = %v", err)
	}
	cf, _ := gotRun.Summary["currentFile"].(map[string]any)
	if cf == nil {
		t.Fatalf("expected currentFile in summary, got %#v", gotRun.Summary)
	}
	if got := anyString(cf["name"]); !strings.Contains(got, "风过留痕 - S01E01 - 第 1 集.mkv") {
		t.Fatalf("summary currentFile name=%q", got)
	}
	cfs, _ := gotRun.Summary["currentFiles"].([]any)
	if len(cfs) != 1 {
		t.Fatalf("summary currentFiles len=%d, want 1; summary=%#v", len(cfs), gotRun.Summary)
	}
	prog, _ := gotRun.Summary["progress"].(map[string]any)
	if prog == nil {
		t.Fatalf("expected progress in summary, got %#v", gotRun.Summary)
	}
	if gotv, ok := prog["plannedFiles"].(float64); !ok || int(gotv) != 2 {
		t.Fatalf("plannedFiles=%#v, want 2; progress=%#v", prog["plannedFiles"], prog)
	}
	if gotv, ok := prog["completedFiles"].(float64); !ok || int(gotv) != 0 {
		t.Fatalf("completedFiles=%#v, want 0; progress=%#v", prog["completedFiles"], prog)
	}
	if got := int(prog["eta"].(float64)); got != 1736 {
		t.Fatalf("eta=%d, want 1736; progress=%#v", got, prog)
	}
	if got := int(prog["percentage"].(float64)); got != 0 {
		t.Fatalf("percentage=%d, want 0; progress=%#v", got, prog)
	}
}

func TestConsume_MoveDeletedDoesNotMarkActiveTransferDeleted(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runnercli_move_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	defer db.Close()

	task, err := db.AddTask(store.Task{Name: "move-task", Mode: "move", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}
	run, err := db.AddRun(store.Run{TaskID: task.ID, Status: "running", Trigger: "manual", Summary: map[string]any{}})
	if err != nil {
		t.Fatalf("AddRun() error = %v", err)
	}

	mgr := active_transfer.NewManager()
	mgr.InitState(run.ID, task.ID, active_transfer.TrackingModeNormal, []active_transfer.TransferCandidateFile{{Path: "a.mp4", Name: "a.mp4"}})
	r := New(db, mgr)
	fp := &fileProgress{m: map[string]*fileProg{}}
	outFile, err := os.CreateTemp(tmpDir, "consume-log-*.log")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	defer outFile.Close()

	line := "2026/05/10 00:20:00 INFO : a.mp4: Deleted\n"
	r.consume(run.ID, strings.NewReader(line), outFile, true, fp, false, true, "/tmp/rclone.conf", "dst:/b", "")

	completed := mgr.ListCompleted(task.ID, 0, 10)
	if completed.Total != 0 {
		t.Fatalf("completed total=%d, want 0", completed.Total)
	}
	pending := mgr.ListPending(task.ID, 0, 10)
	if pending.Total != 1 {
		t.Fatalf("pending total=%d, want 1", pending.Total)
	}
}
