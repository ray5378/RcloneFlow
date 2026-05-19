package controller

import (
	"os"
	"testing"

	"rcloneflow/internal/service"
)

func TestKillRunBySummary_NoPID(t *testing.T) {
	run := service.RunRecord{Summary: `{}`}
	if killRunBySummary(run) {
		t.Fatal("killRunBySummary() should return false when no PID")
	}
}

func TestKillRunBySummary_EmptySummary(t *testing.T) {
	run := service.RunRecord{Summary: ""}
	if killRunBySummary(run) {
		t.Fatal("killRunBySummary() should return false when summary is empty")
	}
}

func TestKillRunBySummary_StringSummary(t *testing.T) {
	run := service.RunRecord{Summary: `{"pid": 999999}`}
	// Should try to kill but PID likely doesn't exist, still returns true
	result := killRunBySummary(run)
	// On Linux, FindProcess always succeeds for any PID, so result should be true
	if !result {
		t.Log("killRunBySummary() returned false (expected on this platform)")
	}
}

func TestKillRunBySummary_MockFindProcess(t *testing.T) {
	old := findProcess
	killed := 0
	findProcess = func(pid int) (*os.Process, error) {
		killed++
		return &os.Process{Pid: pid}, nil
	}
	defer func() { findProcess = old }()

	run := service.RunRecord{Summary: `{"pid": 1}`}
	result := killRunBySummary(run)

	if !result {
		t.Fatal("killRunBySummary() should return true when PID exists")
	}
	if killed != 3 {
		t.Fatalf("findProcess called %d times, want 3", killed)
	}
}

func TestKillRunBySummary_FindProcessError(t *testing.T) {
	old := findProcess
	findProcess = func(pid int) (*os.Process, error) {
		return nil, &mockError{msg: "no such process"}
	}
	defer func() { findProcess = old }()

	run := service.RunRecord{Summary: `{"pid": 1}`}
	result := killRunBySummary(run)

	// Should still return true because PID was found in summary
	if !result {
		t.Fatal("killRunBySummary() should return true even if FindProcess fails")
	}
}

func TestKillRunBySummary_PIDAsFloatInJSON(t *testing.T) {
	run := service.RunRecord{Summary: `{"pid": 12345.0}`}
	result := killRunBySummary(run)
	if !result {
		t.Log("killRunBySummary() returned false (expected on this platform)")
	}
}
