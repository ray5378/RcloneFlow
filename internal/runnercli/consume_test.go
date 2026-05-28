package runnercli

import (
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"rcloneflow/internal/store"
)

type mockRunUpdater struct {
	updateRunCalls atomic.Int64
	lastRun       atomic.Int64
}

func (m *mockRunUpdater) UpdateRun(id int64, fn func(*store.Run)) error {
	m.updateRunCalls.Add(1)
	m.lastRun.Store(id)
	return nil
}

func (m *mockRunUpdater) GetRun(id int64) (store.Run, error) {
	return store.Run{}, nil
}

func (m *mockRunUpdater) GetTask(id int64) (store.Task, bool) {
	return store.Task{}, false
}

func (m *mockRunUpdater) UpdateTask(id int64, task store.Task) error {
	return nil
}

type mockEventBroadcaster struct{}

func (m *mockEventBroadcaster) Broadcast(eventType string, data map[string]any) {}

func TestConsume_JsonStatsLine_WithTransferringFiles_CallsUpdateRunOnce(t *testing.T) {
	mock := &mockRunUpdater{}
	runner := New(mock, &mockEventBroadcaster{}, nil)

	outFile, _ := os.CreateTemp(t.TempDir(), "consume_out_*")
	defer outFile.Close()

	line := `{"level":"info","msg":"Transferred","stats":{"bytes":10485760,"totalBytes":104857600,"speed":1048576,"eta":90,"transferring":[{"name":"file1.txt","size":5242880,"bytes":1048576,"percentage":20,"speed":524288},{"name":"file2.txt","size":5242880,"bytes":524288,"percentage":10,"speed":262144}]}}` + "\n"

	runner.consume(123, strings.NewReader(line), outFile, true, nil, false, false, "", "", "")

	calls := mock.updateRunCalls.Load()
	if calls != 1 {
		t.Errorf("UpdateRun called %d times, want 1", calls)
	}
}

func TestConsume_JsonStatsLine_OnlyProgress_NoTransferring_CallsUpdateRunOnce(t *testing.T) {
	mock := &mockRunUpdater{}
	runner := New(mock, &mockEventBroadcaster{}, nil)

	outFile, _ := os.CreateTemp(t.TempDir(), "consume_out_*")
	defer outFile.Close()

	line := `{"level":"info","msg":"Transferred","stats":{"bytes":10485760,"totalBytes":104857600,"speed":1048576}}` + "\n"

	runner.consume(123, strings.NewReader(line), outFile, true, nil, false, false, "", "", "")

	calls := mock.updateRunCalls.Load()
	if calls != 1 {
		t.Errorf("UpdateRun called %d times, want 1", calls)
	}
}

func TestConsume_JsonProgressLine_NoStatsKey_CallsNoUpdateRun(t *testing.T) {
	mock := &mockRunUpdater{}
	runner := New(mock, &mockEventBroadcaster{}, nil)

	outFile, _ := os.CreateTemp(t.TempDir(), "consume_out_*")
	defer outFile.Close()

	line := `{"level":"info","msg":"file1.txt: Copied (new)"}` + "\n"

	runner.consume(123, strings.NewReader(line), outFile, true, nil, false, false, "", "", "")

	calls := mock.updateRunCalls.Load()
	if calls != 0 {
		t.Errorf("UpdateRun called %d times, want 0", calls)
	}
}

func TestConsume_JsonStatsLineWithEmptyTransferringArray_CallsUpdateRunOnce(t *testing.T) {
	mock := &mockRunUpdater{}
	runner := New(mock, &mockEventBroadcaster{}, nil)

	outFile, _ := os.CreateTemp(t.TempDir(), "consume_out_*")
	defer outFile.Close()

	line := `{"level":"info","msg":"Transferred","stats":{"bytes":0,"totalBytes":0,"speed":0,"transferring":[]}}` + "\n"

	runner.consume(123, strings.NewReader(line), outFile, true, nil, false, false, "", "", "")

	calls := mock.updateRunCalls.Load()
	if calls != 1 {
		t.Errorf("UpdateRun called %d times, want 1 (stats with bytes, empty transferring array)", calls)
	}
}

func TestConsume_JsonStatsLineWithNilTransferring_CallsUpdateRunOnce(t *testing.T) {
	mock := &mockRunUpdater{}
	runner := New(mock, &mockEventBroadcaster{}, nil)

	outFile, _ := os.CreateTemp(t.TempDir(), "consume_out_*")
	defer outFile.Close()

	line := `{"level":"info","msg":"Transferred","stats":{"bytes":100,"totalBytes":1000,"speed":100}}` + "\n"

	runner.consume(123, strings.NewReader(line), outFile, true, nil, false, false, "", "", "")

	calls := mock.updateRunCalls.Load()
	if calls != 1 {
		t.Errorf("UpdateRun called %d times, want 1", calls)
	}
}

func TestConsume_MultipleJsonStatsLines_EachCallsUpdateRunOnce(t *testing.T) {
	mock := &mockRunUpdater{}
	runner := New(mock, &mockEventBroadcaster{}, nil)

	outFile, _ := os.CreateTemp(t.TempDir(), "consume_out_*")
	defer outFile.Close()

	lines := `{"level":"info","msg":"Transferred","stats":{"bytes":1048576,"totalBytes":10485760,"speed":1048576,"transferring":[{"name":"file1.txt","size":5242880,"bytes":1048576,"percentage":20,"speed":524288}]}}`
	lines += "\n" + `{"level":"info","msg":"Transferred","stats":{"bytes":2097152,"totalBytes":10485760,"speed":1048576,"transferring":[{"name":"file1.txt","size":5242880,"bytes":2097152,"percentage":40,"speed":524288}]}}`
	lines += "\n" + `{"level":"info","msg":"Transferred","stats":{"bytes":3145728,"totalBytes":10485760,"speed":1048576,"transferring":[{"name":"file1.txt","size":5242880,"bytes":3145728,"percentage":60,"speed":524288}]}}`
	lines += "\n"

	runner.consume(123, strings.NewReader(lines), outFile, true, nil, false, false, "", "", "")

	calls := mock.updateRunCalls.Load()
	if calls != 3 {
		t.Errorf("UpdateRun called %d times, want 3 (one per stats line)", calls)
	}
}

func TestConsume_JsonStatsLineWithSingleTransferringFile_CallsUpdateRunOnce(t *testing.T) {
	mock := &mockRunUpdater{}
	runner := New(mock, &mockEventBroadcaster{}, nil)

	outFile, _ := os.CreateTemp(t.TempDir(), "consume_out_*")
	defer outFile.Close()

	line := `{"level":"info","msg":"Transferred","stats":{"bytes":1048576,"totalBytes":5242880,"speed":524288,"transferring":[{"name":"onlyfile.txt","size":5242880,"bytes":1048576,"percentage":20,"speed":524288}]}}` + "\n"

	runner.consume(123, strings.NewReader(line), outFile, true, nil, false, false, "", "", "")

	calls := mock.updateRunCalls.Load()
	if calls != 1 {
		t.Errorf("UpdateRun called %d times, want 1", calls)
	}
}

func TestConsume_JsonStatsLineWithMultipleTransferringFiles_CallsUpdateRunOnce(t *testing.T) {
	mock := &mockRunUpdater{}
	runner := New(mock, &mockEventBroadcaster{}, nil)

	outFile, _ := os.CreateTemp(t.TempDir(), "consume_out_*")
	defer outFile.Close()

	line := `{"level":"info","msg":"Transferred","stats":{"bytes":2097152,"totalBytes":10485760,"speed":1048576,"transferring":[{"name":"file1.txt","size":5242880,"bytes":1048576,"percentage":20,"speed":524288},{"name":"file2.txt","size":5242880,"bytes":1048576,"percentage":20,"speed":524288},{"name":"file3.txt","size":5242880,"bytes":0,"percentage":0,"speed":0}]}}` + "\n"

	runner.consume(123, strings.NewReader(line), outFile, true, nil, false, false, "", "", "")

	calls := mock.updateRunCalls.Load()
	if calls != 1 {
		t.Errorf("UpdateRun called %d times, want 1", calls)
	}
}
