package runnercli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"rcloneflow/internal/store"
)

func TestHasValidBisyncLstFiles_NonExistentDir(t *testing.T) {
	assert.False(t, hasValidBisyncLstFiles("/tmp/nonexistent-bisync-dir-99999"))
}

func TestHasValidBisyncLstFiles_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	assert.False(t, hasValidBisyncLstFiles(dir))
}

func TestHasValidBisyncLstFiles_OnlyPath1Lst(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bisync.path1.lst"), []byte("content"), 0644)
	assert.False(t, hasValidBisyncLstFiles(dir))
}

func TestHasValidBisyncLstFiles_OnlyPath2Lst(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bisync.path2.lst"), []byte("content"), 0644)
	assert.False(t, hasValidBisyncLstFiles(dir))
}

func TestHasValidBisyncLstFiles_BothLstFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bisync.path1.lst"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(dir, "bisync.path2.lst"), []byte("content"), 0644)
	assert.True(t, hasValidBisyncLstFiles(dir))
}

func TestHasValidBisyncLstFiles_WithOtherFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bisync.path1.lst"), []byte("c"), 0644)
	os.WriteFile(filepath.Join(dir, "bisync.path2.lst"), []byte("c"), 0644)
	os.WriteFile(filepath.Join(dir, "other.log"), []byte("c"), 0644)
	os.WriteFile(filepath.Join(dir, "data.txt"), []byte("c"), 0644)
	assert.True(t, hasValidBisyncLstFiles(dir))
}

func TestHasValidBisyncLstFiles_OnlyNonLstFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.log"), []byte("c"), 0644)
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("c"), 0644)
	assert.False(t, hasValidBisyncLstFiles(dir))
}

func TestHasValidBisyncLstFiles_SubdirNotCounted(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "subdir")
	os.Mkdir(subdir, 0755)
	os.WriteFile(filepath.Join(subdir, "bisync.path1.lst"), []byte("c"), 0644)
	os.WriteFile(filepath.Join(subdir, "bisync.path2.lst"), []byte("c"), 0644)
	assert.False(t, hasValidBisyncLstFiles(dir))
}

func TestHasValidBisyncLstFiles_LstLikeButWrongName(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "random.lst"), []byte("c"), 0644)
	assert.False(t, hasValidBisyncLstFiles(dir))
}

func TestSanitizeFilenameForBisync_Normal(t *testing.T) {
	result := sanitizeFilenameForBisync("my-backup-job", 42)
	assert.Equal(t, "my-backup-job", result)
}

func TestSanitizeFilenameForBisync_Empty(t *testing.T) {
	result := sanitizeFilenameForBisync("", 7)
	assert.Equal(t, "task-7", result)
}

func TestSanitizeFilenameForBisync_Whitespace(t *testing.T) {
	result := sanitizeFilenameForBisync("   ", 99)
	assert.Equal(t, "task-99", result)
}

func TestSanitizeFilenameForBisync_SpecialChars(t *testing.T) {
	result := sanitizeFilenameForBisync("hello!@#$%^&*()world", 1)
	assert.Equal(t, "hello_world", result)
}

func TestSanitizeFilenameForBisync_HanChars(t *testing.T) {
	result := sanitizeFilenameForBisync("备份任务", 5)
	assert.Equal(t, "备份任务", result)
}

func TestSanitizeFilenameForBisync_MixedHanAndSpecial(t *testing.T) {
	result := sanitizeFilenameForBisync("备份/任务:测试*", 3)
	assert.Equal(t, "备份_任务_测试_", result)
}

func TestSanitizeFilenameForBisync_LeadingTrailingSpaces(t *testing.T) {
	result := sanitizeFilenameForBisync("  my-job  ", 2)
	assert.Equal(t, "my-job", result)
}

func TestSanitizeFilenameForBisync_AlreadyClean(t *testing.T) {
	result := sanitizeFilenameForBisync("abc123_-", 0)
	assert.Equal(t, "abc123_-", result)
}

func TestSanitizeFilenameForBisync_VeryLong(t *testing.T) {
	long := strings.Repeat("a", 100)
	result := sanitizeFilenameForBisync(long, 10)
	assert.Equal(t, 60, len([]rune(result)))
	assert.Equal(t, strings.Repeat("a", 60), result)
}

func TestSanitizeFilenameForBisync_VeryLongWithHan(t *testing.T) {
	long := strings.Repeat("汉", 100)
	result := sanitizeFilenameForBisync(long, 10)
	assert.Equal(t, 60, len([]rune(result)))
	assert.Equal(t, strings.Repeat("汉", 60), result)
}

func TestSanitizeFilenameForBisync_AllInvalid(t *testing.T) {
	result := sanitizeFilenameForBisync("!!!@@@###", 88)
	assert.Equal(t, "_", result)
}

func TestSanitizeFilenameForBisync_SingleChar(t *testing.T) {
	result := sanitizeFilenameForBisync("a", 1)
	assert.Equal(t, "a", result)
}

func TestSanitizeFilenameForBisync_Exactly60Chars(t *testing.T) {
	input := strings.Repeat("a", 60)
	result := sanitizeFilenameForBisync(input, 10)
	assert.Equal(t, input, result)
}

func TestSanitizeFilenameForBisync_HyphensAndUnderscores(t *testing.T) {
	result := sanitizeFilenameForBisync("my-bisync_job-v2", 1)
	assert.Equal(t, "my-bisync_job-v2", result)
}

func TestSanitizeFilenameForBisync_OnlyHan(t *testing.T) {
	result := sanitizeFilenameForBisync("同步任务", 200)
	assert.Equal(t, "同步任务", result)
}

func TestBuildBisyncCommand_EmptyOptionsNoLstFiles(t *testing.T) {
	dir := t.TempDir()
	args := buildBisyncCommand("remote:src", "remote:dst", dir, json.RawMessage("{}"))
	assert.Contains(t, args, "bisync")
	assert.Contains(t, args, "remote:src")
	assert.Contains(t, args, "remote:dst")
	assert.Contains(t, args, "--workdir")
	assert.Contains(t, args, dir)
	assert.Contains(t, args, "--resync")
}

func TestBuildBisyncCommand_EmptyOptionsWithLstFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bisync.path1.lst"), []byte("c"), 0644)
	os.WriteFile(filepath.Join(dir, "bisync.path2.lst"), []byte("c"), 0644)
	args := buildBisyncCommand("remote:src", "remote:dst", dir, json.RawMessage("{}"))
	assert.Contains(t, args, "bisync")
	assert.NotContains(t, args, "--resync")
}

func TestBuildBisyncCommand_NilOptions(t *testing.T) {
	dir := t.TempDir()
	args := buildBisyncCommand("remote:src", "remote:dst", dir, nil)
	assert.Contains(t, args, "bisync")
	assert.Contains(t, args, "--resync")
}

func TestBuildBisyncCommand_ResyncExplicitlyTrue(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bisync.path1.lst"), []byte("c"), 0644)
	os.WriteFile(filepath.Join(dir, "bisync.path2.lst"), []byte("c"), 0644)
	opts := json.RawMessage(`{"resync":true}`)
	args := buildBisyncCommand("remote:src", "remote:dst", dir, opts)
	assert.Contains(t, args, "--resync")
}

func TestBuildBisyncCommand_ResyncExplicitlyFalseWithLstFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bisync.path1.lst"), []byte("c"), 0644)
	os.WriteFile(filepath.Join(dir, "bisync.path2.lst"), []byte("c"), 0644)
	opts := json.RawMessage(`{"resync":false}`)
	args := buildBisyncCommand("remote:src", "remote:dst", dir, opts)
	assert.NotContains(t, args, "--resync")
}

func TestBuildBisyncCommand_ResyncExplicitlyFalseNoLstFiles(t *testing.T) {
	dir := t.TempDir()
	opts := json.RawMessage(`{"resync":false}`)
	args := buildBisyncCommand("remote:src", "remote:dst", dir, opts)
	assert.Contains(t, args, "--resync")
}

func TestBuildBisyncCommand_AutoResyncWhenNoLstFiles(t *testing.T) {
	dir := t.TempDir()
	opts := json.RawMessage(`{"compare":"size"}`)
	args := buildBisyncCommand("remote:src", "remote:dst", dir, opts)
	assert.Contains(t, args, "--resync")
}

func TestBuildBisyncCommand_AllOptions(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bisync.path1.lst"), []byte("c"), 0644)
	os.WriteFile(filepath.Join(dir, "bisync.path2.lst"), []byte("c"), 0644)
	opts := json.RawMessage(`{
		"resync": true,
		"compare": "size,modtime,checksum",
		"maxDelete": "50",
		"checkAccess": true,
		"checkFilename": "test.txt",
		"conflictResolve": "newer",
		"conflictLoser": "delete",
		"conflictSuffix": "conflict",
		"backupDir1": "remote:backup1",
		"backupDir2": "remote:backup2",
		"createEmptySrcDirs": true,
		"removeEmptyDirs": true,
		"recover": true
	}`)
	args := buildBisyncCommand("remote:src", "remote:dst", dir, opts)
	assert.Contains(t, args, "--resync")
	assert.Contains(t, args, "--compare")
	assert.Contains(t, args, "size,modtime,checksum")
	assert.Contains(t, args, "--max-delete")
	assert.Contains(t, args, "50")
	assert.Contains(t, args, "--check-access")
	assert.Contains(t, args, "--check-filename")
	assert.Contains(t, args, "test.txt")
	assert.Contains(t, args, "--conflict-resolve")
	assert.Contains(t, args, "newer")
	assert.Contains(t, args, "--conflict-loser")
	assert.Contains(t, args, "delete")
	assert.Contains(t, args, "--conflict-suffix")
	assert.Contains(t, args, "conflict")
	assert.Contains(t, args, "--backup-dir1")
	assert.Contains(t, args, "remote:backup1")
	assert.Contains(t, args, "--backup-dir2")
	assert.Contains(t, args, "remote:backup2")
	assert.Contains(t, args, "--create-empty-src-dirs")
	assert.Contains(t, args, "--remove-empty-dirs")
	assert.Contains(t, args, "--recover")
}

func TestBuildBisyncCommand_EmptyStringOptionsOmitted(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bisync.path1.lst"), []byte("c"), 0644)
	os.WriteFile(filepath.Join(dir, "bisync.path2.lst"), []byte("c"), 0644)
	opts := json.RawMessage(`{"compare":"","maxDelete":"","checkFilename":"","conflictResolve":"","conflictLoser":"","conflictSuffix":"","backupDir1":"","backupDir2":""}`)
	args := buildBisyncCommand("remote:src", "remote:dst", dir, opts)
	assert.NotContains(t, args, "--compare")
	assert.NotContains(t, args, "--max-delete")
	assert.NotContains(t, args, "--check-filename")
	assert.NotContains(t, args, "--conflict-resolve")
	assert.NotContains(t, args, "--conflict-loser")
	assert.NotContains(t, args, "--conflict-suffix")
	assert.NotContains(t, args, "--backup-dir1")
	assert.NotContains(t, args, "--backup-dir2")
}

func TestBuildBisyncCommand_BoolOptionsFalseOmitted(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bisync.path1.lst"), []byte("c"), 0644)
	os.WriteFile(filepath.Join(dir, "bisync.path2.lst"), []byte("c"), 0644)
	opts := json.RawMessage(`{"checkAccess":false,"createEmptySrcDirs":false,"removeEmptyDirs":false,"recover":false}`)
	args := buildBisyncCommand("remote:src", "remote:dst", dir, opts)
	assert.NotContains(t, args, "--check-access")
	assert.NotContains(t, args, "--create-empty-src-dirs")
	assert.NotContains(t, args, "--remove-empty-dirs")
	assert.NotContains(t, args, "--recover")
}

func TestBuildBisyncCommand_OptionsWithoutResyncWithLstFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bisync.path1.lst"), []byte("c"), 0644)
	os.WriteFile(filepath.Join(dir, "bisync.path2.lst"), []byte("c"), 0644)
	opts := json.RawMessage(`{"compare":"size"}`)
	args := buildBisyncCommand("remote:src", "remote:dst", dir, opts)
	assert.NotContains(t, args, "--resync")
}

func TestClassifyBisyncLogRow_Path1Copied(t *testing.T) {
	row, status, ok := classifyBisyncLogRow("INFO", "file.txt", "file.txt: Copied (new)", nil)
	assert.True(t, ok)
	assert.Equal(t, "success", status)
	assert.Equal(t, "copied", row["action"])
	assert.Equal(t, string(BisyncDirectionPath1ToPath2), row["direction"])
	assert.Equal(t, "file.txt", row["path"])
	assert.Equal(t, "file.txt: Copied (new)", row["message"])
	assert.Equal(t, "INFO", row["level"])
}

func TestClassifyBisyncLogRow_Path2Deleted(t *testing.T) {
	row, status, ok := classifyBisyncLogRow("INFO", "old.txt", "old.txt: Deleted Path2", nil)
	assert.True(t, ok)
	assert.Equal(t, "success", status)
	assert.Equal(t, "deleted", row["action"])
	assert.Equal(t, string(BisyncDirectionPath2ToPath1), row["direction"])
}

func TestClassifyBisyncLogRow_Skipped(t *testing.T) {
	row, status, ok := classifyBisyncLogRow("INFO", "skip.txt", "file skipped", nil)
	assert.True(t, ok)
	assert.Equal(t, "skipped", status)
	assert.Equal(t, "skipped", row["action"])
}

func TestClassifyBisyncLogRow_ErrorKeyword(t *testing.T) {
	row, status, ok := classifyBisyncLogRow("ERROR", "bad.txt", "something error occurred", nil)
	assert.True(t, ok)
	assert.Equal(t, "failed", status)
	assert.Equal(t, "error", row["action"])
}

func TestClassifyBisyncLogRow_FailedKeyword(t *testing.T) {
	row, status, ok := classifyBisyncLogRow("ERROR", "bad.txt", "failed to copy", nil)
	assert.True(t, ok)
	assert.Equal(t, "failed", status)
	assert.Equal(t, "error", row["action"])
}

func TestClassifyBisyncLogRow_Unclassifiable(t *testing.T) {
	row, status, ok := classifyBisyncLogRow("INFO", "", "some random unrelated message", nil)
	assert.False(t, ok)
	assert.Nil(t, row)
	assert.Equal(t, "", status)
}

func TestClassifyBisyncLogRow_SourceDirection(t *testing.T) {
	row, _, ok := classifyBisyncLogRow("INFO", "f.txt", "source: Copied (new)", nil)
	assert.True(t, ok)
	assert.Equal(t, string(BisyncDirectionPath1ToPath2), row["direction"])
}

func TestClassifyBisyncLogRow_DestDirection(t *testing.T) {
	row, _, ok := classifyBisyncLogRow("INFO", "f.txt", "dest: Deleted", nil)
	assert.True(t, ok)
	assert.Equal(t, string(BisyncDirectionPath2ToPath1), row["direction"])
}

func TestClassifyBisyncLogRow_DefaultDirection(t *testing.T) {
	row, _, ok := classifyBisyncLogRow("INFO", "f.txt", "Copied (new)", nil)
	assert.True(t, ok)
	assert.Equal(t, string(BisyncDirectionPath1ToPath2), row["direction"])
}

func TestClassifyBisyncLogRow_WithSizeMatch(t *testing.T) {
	sizes := map[string]int64{"bigfile.dat": 1048576}
	row, _, ok := classifyBisyncLogRow("INFO", "bigfile.dat", "Copied (new)", sizes)
	assert.True(t, ok)
	assert.Equal(t, int64(1048576), row["sizeBytes"])
}

func TestClassifyBisyncLogRow_WithSizeZeroNotIncluded(t *testing.T) {
	sizes := map[string]int64{"empty.dat": 0}
	row, _, ok := classifyBisyncLogRow("INFO", "empty.dat", "Copied (new)", sizes)
	assert.True(t, ok)
	_, hasSize := row["sizeBytes"]
	assert.False(t, hasSize)
}

func TestClassifyBisyncLogRow_NoMatchingSize(t *testing.T) {
	sizes := map[string]int64{"other.dat": 512}
	row, _, ok := classifyBisyncLogRow("INFO", "nope.txt", "Copied (new)", sizes)
	assert.True(t, ok)
	_, hasSize := row["sizeBytes"]
	assert.False(t, hasSize)
}

func TestClassifyBisyncLogRow_Path1Path2InMessage(t *testing.T) {
	row, _, ok := classifyBisyncLogRow("INFO", "f.txt", "path1 Copied -> path2", nil)
	assert.True(t, ok)
	assert.Equal(t, string(BisyncDirectionPath1ToPath2), row["direction"])
}

func TestClassifyBisyncLogRow_LevelAndMessagePreserved(t *testing.T) {
	row, _, ok := classifyBisyncLogRow("ERROR", "/path/to/file", "something failed miserably", nil)
	assert.True(t, ok)
	assert.Equal(t, "/path/to/file", row["path"])
	assert.Equal(t, "something failed miserably", row["message"])
	assert.Equal(t, "ERROR", row["level"])
}

func TestClassifyBisyncLogRow_StatusField(t *testing.T) {
	row, status, ok := classifyBisyncLogRow("INFO", "f.txt", "Copied (new)", nil)
	assert.True(t, ok)
	assert.Equal(t, "success", status)
	assert.Equal(t, "success", row["status"])
}

// --- buildBisyncArgs ---

func TestBuildBisyncArgs_CreatesBisyncDir(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APP_DATA_DIR", tmpDir)

	run := store.Run{
		ID:       1,
		TaskID:   42,
		TaskName: "test-task",
		Summary: map[string]any{
			"bisyncOptions": map[string]any{
				"maxDelete": "100",
			},
		},
	}

	args, err := buildBisyncArgs("remoteA:path", "remoteB:path", run)
	assert.NoError(t, err)
	assert.NotEmpty(t, args)

	// 验证目录已创建
	bisyncDir := filepath.Join(tmpDir, "bisync", "test-task")
	info, err := os.Stat(bisyncDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())

	// 验证参数包含 max-delete
	joined := strings.Join(args, " ")
	assert.Contains(t, joined, "--max-delete")
	assert.Contains(t, joined, "100")
	assert.Contains(t, joined, "--workdir")
	assert.Contains(t, joined, bisyncDir)
}

func TestBuildBisyncArgs_NoOptions(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APP_DATA_DIR", tmpDir)

	run := store.Run{
		ID:       2,
		TaskID:   99,
		TaskName: "simple-task",
	}

	args, err := buildBisyncArgs("remoteA:path", "remoteB:path", run)
	assert.NoError(t, err)
	assert.NotEmpty(t, args)
}

func TestBuildBisyncArgs_WithResyncOption(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APP_DATA_DIR", tmpDir)

	run := store.Run{
		ID:       3,
		TaskID:   77,
		TaskName: "resync-task",
		Summary: map[string]any{
			"bisyncOptions": map[string]any{
				"resync": true,
			},
		},
	}

	args, err := buildBisyncArgs("remoteA:path", "remoteB:path", run)
	assert.NoError(t, err)

	joined := strings.Join(args, " ")
	assert.Contains(t, joined, "--resync")
}

func TestBuildBisyncArgs_AutoResyncWithoutLstFiles(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APP_DATA_DIR", tmpDir)

	// 没有 lst 文件 → 自动 resync
	run := store.Run{
		ID:       4,
		TaskID:   55,
		TaskName: "auto-resync",
	}

	args, err := buildBisyncArgs("remoteA:path", "remoteB:path", run)
	assert.NoError(t, err)

	joined := strings.Join(args, " ")
	assert.Contains(t, joined, "--resync")
}

func TestBuildBisyncArgs_WithExistingLstFilesSkipsResync(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APP_DATA_DIR", tmpDir)

	// 提前创建 lst 文件
	bisyncDir := filepath.Join(tmpDir, "bisync", "no-resync")
	os.MkdirAll(bisyncDir, 0755)
	os.WriteFile(filepath.Join(bisyncDir, "bisync.path1.lst"), []byte("data"), 0644)
	os.WriteFile(filepath.Join(bisyncDir, "bisync.path2.lst"), []byte("data"), 0644)

	run := store.Run{
		ID:       5,
		TaskID:   33,
		TaskName: "no-resync",
	}

	args, err := buildBisyncArgs("remoteA:path", "remoteB:path", run)
	assert.NoError(t, err)

	joined := strings.Join(args, " ")
	assert.NotContains(t, joined, "--resync")
}

// --- ClearBisyncResyncFlag ---

type mockTaskUpdater struct {
	tasks map[int64]*store.Task
}

func (m *mockTaskUpdater) UpdateRun(id int64, fn func(*store.Run)) error { return nil }
func (m *mockTaskUpdater) GetRun(id int64) (store.Run, error)           { return store.Run{}, nil }
func (m *mockTaskUpdater) GetTask(id int64) (store.Task, bool) {
	t, ok := m.tasks[id]
	if !ok {
		return store.Task{}, false
	}
	return *t, true
}
func (m *mockTaskUpdater) UpdateTask(id int64, task store.Task) error {
	m.tasks[id] = &task
	return nil
}

func TestClearBisyncResyncFlag_RemovesResync(t *testing.T) {
	opts := map[string]any{"resync": true, "maxDelete": "50"}
	optsBytes, _ := json.Marshal(opts)
	mockUpdater := &mockTaskUpdater{
		tasks: map[int64]*store.Task{
			1: {ID: 1, Name: "bisync-task", BisyncOptions: optsBytes},
		},
	}

	ClearBisyncResyncFlag(mockUpdater, 1)

	updated, ok := mockUpdater.GetTask(1)
	assert.True(t, ok)
	var result map[string]any
	err := json.Unmarshal(updated.BisyncOptions, &result)
	assert.NoError(t, err)
	_, hasResync := result["resync"]
	assert.False(t, hasResync, "resync should be removed")
}

func TestClearBisyncResyncFlag_NoResync(t *testing.T) {
	opts := map[string]any{"maxDelete": "50"}
	optsBytes, _ := json.Marshal(opts)
	mockUpdater := &mockTaskUpdater{
		tasks: map[int64]*store.Task{
			2: {ID: 2, Name: "normal-task", BisyncOptions: optsBytes},
		},
	}

	ClearBisyncResyncFlag(mockUpdater, 2)

	updated, ok := mockUpdater.GetTask(2)
	assert.True(t, ok)
	var result map[string]any
	err := json.Unmarshal(updated.BisyncOptions, &result)
	assert.NoError(t, err)
	_, hasResync := result["resync"]
	assert.False(t, hasResync)
}

func TestClearBisyncResyncFlag_TaskNotFound(t *testing.T) {
	mockUpdater := &mockTaskUpdater{tasks: map[int64]*store.Task{}}
	ClearBisyncResyncFlag(mockUpdater, 999) // should not panic
}

func TestClearBisyncResyncFlag_EmptyBisyncOptions(t *testing.T) {
	mockUpdater := &mockTaskUpdater{
		tasks: map[int64]*store.Task{
			3: {ID: 3, Name: "no-options-task"},
		},
	}
	ClearBisyncResyncFlag(mockUpdater, 3) // should not panic
}

func TestClearBisyncResyncFlag_KeepsOtherOptions(t *testing.T) {
	opts := map[string]any{"resync": true, "maxDelete": "100", "checkAccess": true}
	optsBytes, _ := json.Marshal(opts)
	mockUpdater := &mockTaskUpdater{
		tasks: map[int64]*store.Task{
			4: {ID: 4, Name: "multi-opt", BisyncOptions: optsBytes},
		},
	}

	ClearBisyncResyncFlag(mockUpdater, 4)

	updated, ok := mockUpdater.GetTask(4)
	assert.True(t, ok)
	var result map[string]any
	err := json.Unmarshal(updated.BisyncOptions, &result)
	assert.NoError(t, err)
	_, hasResync := result["resync"]
	assert.False(t, hasResync)
	assert.Equal(t, "100", result["maxDelete"])
	assert.Equal(t, true, result["checkAccess"])
}
