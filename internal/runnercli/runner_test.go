package runnercli

import (
	"os"
	"strings"
	"testing"

	"rcloneflow/internal/active_transfer"
	"rcloneflow/internal/store"
)

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
	fp := &fileProgress{m: map[string]*fileProg{}}
	outFile, err := os.CreateTemp(tmpDir, "consume-log-*.log")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	defer outFile.Close()

	line := "2026/05/01 14:01:20 NOTICE: 电视剧/国产剧/人间惊鸿客 (2026)/Season 1/人间惊鸿客 - S01E18 - 第 18 集.mkv: CAS compatible match after source cleanup (Failed to copy: object not found)\n"
	r.consume(run.ID, strings.NewReader(line), outFile, true, fp, true, false)

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
	r.consume(run.ID, strings.NewReader(line), outFile, true, fp, false, false)

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
	r.consume(run.ID, strings.NewReader(line), outFile, true, fp, false, true)

	completed := mgr.ListCompleted(task.ID, 0, 10)
	if completed.Total != 0 {
		t.Fatalf("completed total=%d, want 0", completed.Total)
	}
	pending := mgr.ListPending(task.ID, 0, 10)
	if pending.Total != 1 {
		t.Fatalf("pending total=%d, want 1", pending.Total)
	}
}
