package store

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUserJSON_OmitsPassword(t *testing.T) {
	u := User{ID: 1, Username: "admin", Password: "secret123"}
	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	if _, ok := m["password"]; ok {
		t.Error("password field should be omitted from JSON")
	}
}

func TestUserJSON_RoundTrip(t *testing.T) {
	u := User{ID: 1, Username: "admin", Password: "secret", PasswordChanged: true}
	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	var u2 User
	if err := json.Unmarshal(data, &u2); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	if u2.ID != 1 || u2.Username != "admin" || !u2.PasswordChanged {
		t.Errorf("round-trip mismatch: %+v", u2)
	}
	if u2.Password != "" {
		t.Error("password should not be unmarshalled")
	}
}

func TestTagJSON_RoundTrip(t *testing.T) {
	tag := Tag{ID: 5, Tag: "sync", Type: "action", Selected: true}
	data, err := json.Marshal(tag)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	var tag2 Tag
	if err := json.Unmarshal(data, &tag2); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	if tag2.ID != 5 || tag2.Tag != "sync" || tag2.Type != "action" || !tag2.Selected {
		t.Errorf("round-trip mismatch: %+v", tag2)
	}
}

func TestTaskJSON_RoundTrip(t *testing.T) {
	task := Task{
		ID: 10, Name: "my-task", Mode: "copy",
		SourceRemote: "src:", SourcePath: "/data",
		TargetRemote: "dst:", TargetPath: "/backup",
		SortOrder: 1,
	}
	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	var task2 Task
	if err := json.Unmarshal(data, &task2); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	if task2.Name != "my-task" || task2.Mode != "copy" || task2.SortOrder != 1 {
		t.Errorf("round-trip mismatch: %+v", task2)
	}
}

func TestTaskJSON_OptionsNull(t *testing.T) {
	task := Task{Name: "test", Mode: "sync", Options: nil, BisyncOptions: nil}
	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	var task2 Task
	if err := json.Unmarshal(data, &task2); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	if task2.Options != nil {
		t.Error("expected nil Options from null")
	}
}

func TestTaskJSON_OptionsRoundTrip(t *testing.T) {
	task := Task{
		Name: "test", Mode: "copy",
		Options:       json.RawMessage(`{"transfers": 4}`),
		BisyncOptions: json.RawMessage(`{"resync": true}`),
	}
	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	var task2 Task
	if err := json.Unmarshal(data, &task2); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	var opts map[string]any
	if err := json.Unmarshal(task2.Options, &opts); err != nil {
		t.Fatalf("unmarshal Options: %v", err)
	}
	if opts["transfers"] != float64(4) {
		t.Errorf("expected transfers=4, got %v", opts["transfers"])
	}
	var bisync map[string]any
	if err := json.Unmarshal(task2.BisyncOptions, &bisync); err != nil {
		t.Fatalf("unmarshal BisyncOptions: %v", err)
	}
	if bisync["resync"] != true {
		t.Errorf("expected resync=true, got %v", bisync["resync"])
	}
}

func TestBisyncOptionsJSON_RoundTrip(t *testing.T) {
	bo := BisyncOptions{
		Resync:             true,
		Compare:            "checksum",
		MaxDelete:          "10",
		CheckAccess:        true,
		CheckFilename:      ".rclone-check",
		ConflictResolve:    "source",
		ConflictLoser:      "destination",
		ConflictSuffix:     "-conflict",
		BackupDir1:         "backup1:",
		BackupDir2:         "backup2:",
		CreateEmptySrcDirs: true,
		RemoveEmptyDirs:    true,
		Recover:            true,
		LstBackupCount:     5,
	}
	data, err := json.Marshal(bo)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	var bo2 BisyncOptions
	if err := json.Unmarshal(data, &bo2); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	if bo2.Resync != true || bo2.Compare != "checksum" || bo2.LstBackupCount != 5 {
		t.Errorf("round-trip mismatch: %+v", bo2)
	}
}

func TestBisyncOptionsJSON_ZeroValuesOmitted(t *testing.T) {
	bo := BisyncOptions{}
	data, err := json.Marshal(bo)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	if len(data) > 3 {
		t.Errorf("expected empty JSON object, got %s", string(data))
	}
}

func TestScheduleJSON_RoundTrip(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	sched := Schedule{
		ID: 1, TaskID: 10, Spec: "0 6 * * *", Enabled: true,
		NextRunTime: &now,
	}
	data, err := json.Marshal(sched)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	var sched2 Schedule
	if err := json.Unmarshal(data, &sched2); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	if sched2.Spec != "0 6 * * *" || !sched2.Enabled {
		t.Errorf("round-trip mismatch: %+v", sched2)
	}
	if sched2.NextRunTime == nil || !sched2.NextRunTime.Equal(now) {
		t.Errorf("NextRunTime mismatch: got %v", sched2.NextRunTime)
	}
}

func TestScheduleJSON_NilNextRunTime(t *testing.T) {
	sched := Schedule{ID: 2, TaskID: 20, Spec: "0 12 * * *", Enabled: false}
	data, err := json.Marshal(sched)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	var sched2 Schedule
	if err := json.Unmarshal(data, &sched2); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	if sched2.NextRunTime != nil {
		t.Error("expected nil NextRunTime")
	}
}

func TestRunJSON_RoundTrip(t *testing.T) {
	run := Run{
		ID: 100, TaskID: 10, Status: "finished", Trigger: "manual",
		Summary:         map[string]any{"files": float64(50), "bytes": float64(1024)},
		TaskName:        "test-task",
		TaskMode:        "copy",
		BytesTransferred: 1024,
		Speed:           "5MB/s",
	}
	data, err := json.Marshal(run)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	var run2 Run
	if err := json.Unmarshal(data, &run2); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	if run2.Status != "finished" || run2.Speed != "5MB/s" {
		t.Errorf("round-trip mismatch: %+v", run2)
	}
	if run2.Summary["files"] != float64(50) {
		t.Errorf("expected files=50, got %v", run2.Summary["files"])
	}
}

func TestRunJSON_NilSummary(t *testing.T) {
	run := Run{ID: 1, TaskID: 1, Status: "running", Trigger: "manual"}
	data, err := json.Marshal(run)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	var run2 Run
	if err := json.Unmarshal(data, &run2); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	if run2.Summary != nil {
		t.Error("expected nil Summary")
	}
}

func TestRunJSON_EmptySummary(t *testing.T) {
	run := Run{ID: 2, TaskID: 1, Status: "running", Trigger: "manual", Summary: map[string]any{}}
	data, err := json.Marshal(run)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	var run2 Run
	if err := json.Unmarshal(data, &run2); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	if run2.Summary != nil && len(run2.Summary) != 0 {
		t.Errorf("expected nil or empty Summary, got %v", run2.Summary)
	}
}

func TestRunJSON_NilFinishedAt(t *testing.T) {
	run := Run{ID: 3, TaskID: 1, Status: "running", Trigger: "manual"}
	data, err := json.Marshal(run)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	var run2 Run
	if err := json.Unmarshal(data, &run2); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	if run2.FinishedAt != nil {
		t.Error("expected nil FinishedAt")
	}
}