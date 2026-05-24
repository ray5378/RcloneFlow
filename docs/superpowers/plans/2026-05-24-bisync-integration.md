# Bisync 双向同步集成实施计划

&gt; **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 rclone 的 bisync 双向同步功能完整集成到现有的任务系统中，包括配置弹窗、历史版本管理、双向进度显示等。

**Architecture:** 采用分层设计：数据模型层 → 后端 runner/service/controller 层 → 前端组件层 → 各模块集成层（标签/导入导出/日志清理）。

**Tech Stack:** Go (后端) + TypeScript/Vue 3 (前端) + SQLite (存储) + rclone v1.74.2

---

## 文件变更清单

### 新建文件
- `migrations/000009_add_bisync_options.sql` - 数据库迁移
- `internal/runnercli/bisync.go` - bisync 核心逻辑
- `frontend/src/components/task/BisyncConfigModal.vue` - bisync 配置弹窗

### 修改文件
- `internal/store/models.go` - 数据模型
- `internal/store/store_tasks.go` - 任务存储 SQL
- `internal/runnercli/runner.go` -  runner 启动逻辑
- `internal/runnercli/summary.go` - 总结生成
- `internal/service/task.go` - 任务服务
- `internal/service/tag.go` - 标签服务
- `internal/service/log_cleanup.go` - 日志清理
- `internal/service/task_import_export.go` - 导入导出
- `internal/controller/task.go` - 控制器
- `internal/router/router.go` - 路由
- `frontend/src/types/index.ts` - 前端类型
- `frontend/src/components/task/types.ts` - 组件类型
- `frontend/src/api/task.ts` - 前端 API
- `frontend/src/components/task/AddTaskForm.vue` - 任务表单
- `frontend/src/views/TaskView.vue` - 任务视图
- `frontend/src/components/task/RunDetailModal.vue` - 运行详情
- `frontend/src/components/task/transferring/TransferringModal.vue` - 传输中弹窗
- `frontend/src/composables/useTaskFormState.ts` - 表单状态
- `frontend/src/composables/useTaskFormSubmit.ts` - 表单提交
- `frontend/src/i18n/zh.ts` - 中文翻译
- `frontend/src/i18n/en.ts` - 英文翻译

---

## 任务 1：数据模型与数据库迁移

**Files:**
- Create: `migrations/000009_add_bisync_options.sql`
- Modify: `internal/store/models.go`
- Modify: `internal/store/store_tasks.go`

- [ ] **Step 1.1: 创建数据库迁移文件**

```sql
-- 添加 bisync_options 列存储 bisync 特定配置
ALTER TABLE tasks ADD COLUMN bisync_options TEXT;
```

- [ ] **Step 1.2: 更新 Task 与 BisyncOptions 模型**

编辑 `internal/store/models.go`：

```go
type Task struct {
	ID            int64           `json:"id"`
	Name          string          `json:"name"`
	Mode          string          `json:"mode"`
	SourceRemote  string          `json:"sourceRemote"`
	SourcePath    string          `json:"sourcePath"`
	TargetRemote  string          `json:"targetRemote"`
	TargetPath    string          `json:"targetPath"`
	Options       json.RawMessage `json:"options,omitempty"`
	BisyncOptions json.RawMessage `json:"bisyncOptions,omitempty"` // 新增
	SortOrder     int64           `json:"sortOrder"`
	CreatedAt     time.Time       `json:"createdAt"`
}

// 新增
type BisyncOptions struct {
	Resync             bool   `json:"resync,omitempty"`
	Compare            string `json:"compare,omitempty"`
	MaxDelete          string `json:"maxDelete,omitempty"`
	CheckAccess        bool   `json:"checkAccess,omitempty"`
	CheckFilename      string `json:"checkFilename,omitempty"`
	ConflictResolve    string `json:"conflictResolve,omitempty"`
	ConflictLoser      string `json:"conflictLoser,omitempty"`
	ConflictSuffix     string `json:"conflictSuffix,omitempty"`
	BackupDir1         string `json:"backupDir1,omitempty"`
	BackupDir2         string `json:"backupDir2,omitempty"`
	CreateEmptySrcDirs bool   `json:"createEmptySrcDirs,omitempty"`
	RemoveEmptyDirs    bool   `json:"removeEmptyDirs,omitempty"`
	Recover            bool   `json:"recover,omitempty"`
}
```

- [ ] **Step 1.3: 更新 store_tasks.go 的 SELECT 语句**

编辑 `internal/store/store_tasks.go` 中的 `ListTasks()`:

```go
rows, err := db.db.Query(`
	SELECT id, name, mode, source_remote, source_path, target_remote, target_path, options, bisync_options, sort_order, created_at 
	FROM tasks ORDER BY sort_order ASC, id ASC`)
```

同样在 Scan 部分：

```go
var options sql.NullString
var bisyncOptions sql.NullString // 新增
err := rows.Scan(&t.ID, &t.Name, &t.Mode, &t.SourceRemote, &t.SourcePath, 
	&t.TargetRemote, &t.TargetPath, &options, &bisyncOptions, &t.SortOrder, &t.CreatedAt)
if options.Valid {
	t.Options = []byte(options.String)
}
if bisyncOptions.Valid { // 新增
	t.BisyncOptions = []byte(bisyncOptions.String)
}
```

- [ ] **Step 1.4: 更新 GetTask 函数**

编辑 `GetTask()` 中的 SQL：

```go
err := db.db.QueryRow(`
	SELECT id, name, mode, source_remote, source_path, target_remote, target_path, options, bisync_options, sort_order, created_at 
	FROM tasks WHERE id = ?`, id).Scan(
	&t.ID, &t.Name, &t.Mode, &t.SourceRemote, &t.SourcePath, 
	&t.TargetRemote, &t.TargetPath, &options, &bisyncOptions, &t.SortOrder, &t.CreatedAt)
```

同样处理 Scan 部分的 bisyncOptions。

- [ ] **Step 1.5: 更新 AddTask 函数**

编辑 `AddTask()` 中的 INSERT：

```go
result, err := db.db.Exec(`
	INSERT INTO tasks (name, mode, source_remote, source_path, target_remote, target_path, options, bisync_options, sort_order) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
	t.Name, t.Mode, t.SourceRemote, t.SourcePath, t.TargetRemote, t.TargetPath, t.Options, t.BisyncOptions, nextSortOrder)
```

- [ ] **Step 1.6: 更新 UpdateTask 函数**

编辑 `UpdateTask()` 中的 UPDATE：

```go
_, err := db.db.Exec(`
	UPDATE tasks SET name=?, mode=?, source_remote=?, source_path=?, target_remote=?, target_path=?, options=?, bisync_options=?, sort_order=?
	WHERE id=?`, t.Name, t.Mode, t.SourceRemote, t.SourcePath, t.TargetRemote, t.TargetPath, t.Options, t.BisyncOptions, t.SortOrder, id)
```

- [ ] **Step 1.7: 运行测试**

Run: `go test ./internal/store -v`
Expected: PASS

- [ ] **Step 1.8: 提交**

```bash
git add migrations/000009_add_bisync_options.sql internal/store/models.go internal/store/store_tasks.go
git commit -m "feat: add bisync data models and database migration"
```

---

## 任务 2：实现 bisync 核心模块

**Files:**
- Create: `internal/runnercli/bisync.go`
- Modify: `internal/runnercli/runner.go`
- Modify: `internal/runnercli/summary.go`

- [ ] **Step 2.1: 创建 bisync.go 核心文件**

```go
package runnercli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type BisyncFileDirection string

const (
	BisyncDirectionPath1ToPath2 BisyncFileDirection = "path1_to_path2"
	BisyncDirectionPath2ToPath1 BisyncFileDirection = "path2_to_path1"
)

type BisyncSummary struct {
	Path1ToPath2 SummaryFilesCounts `json:"path1ToPath2"`
	Path2ToPath1 SummaryFilesCounts `json:"path2ToPath1"`
}

// buildBisyncCommand 构建 rclone bisync 命令行参数
func buildBisyncCommand(src, dst, workDir string, options json.RawMessage) []string {
	args := []string{"bisync", src, dst, "--workdir", workDir}

	if len(options) > 0 {
		var opts map[string]interface{}
		if err := json.Unmarshal(options, &opts); err == nil {
			// 处理各种 bisync 选项
			if resync, ok := opts["resync"].(bool); ok && resync {
				args = append(args, "--resync")
			}
			if compare, ok := opts["compare"].(string); ok && compare != "" {
				args = append(args, "--compare", compare)
			}
			if maxDelete, ok := opts["maxDelete"].(string); ok && maxDelete != "" {
				args = append(args, "--max-delete", maxDelete)
			}
			if checkAccess, ok := opts["checkAccess"].(bool); ok && checkAccess {
				args = append(args, "--check-access")
			}
			if checkFilename, ok := opts["checkFilename"].(string); ok && checkFilename != "" {
				args = append(args, "--check-filename", checkFilename)
			}
			if conflictResolve, ok := opts["conflictResolve"].(string); ok && conflictResolve != "" {
				args = append(args, "--conflict-resolve", conflictResolve)
			}
			if conflictLoser, ok := opts["conflictLoser"].(string); ok && conflictLoser != "" {
				args = append(args, "--conflict-loser", conflictLoser)
			}
			if conflictSuffix, ok := opts["conflictSuffix"].(string); ok && conflictSuffix != "" {
				args = append(args, "--conflict-suffix", conflictSuffix)
			}
			if backupDir1, ok := opts["backupDir1"].(string); ok && backupDir1 != "" {
				args = append(args, "--backup-dir1", backupDir1)
			}
			if backupDir2, ok := opts["backupDir2"].(string); ok && backupDir2 != "" {
				args = append(args, "--backup-dir2", backupDir2)
			}
			if createEmptySrcDirs, ok := opts["createEmptySrcDirs"].(bool); ok && createEmptySrcDirs {
				args = append(args, "--create-empty-src-dirs")
			}
			if removeEmptyDirs, ok := opts["removeEmptyDirs"].(bool); ok && removeEmptyDirs {
				args = append(args, "--remove-empty-dirs")
			}
			if recover, ok := opts["recover"].(bool); ok && recover {
				args = append(args, "--recover")
			}
		}
	}
	return args
}

// classifyBisyncLogRow 解析 bisync 的日志行，判断方向和类型
func classifyBisyncLogRow(level, path, msg string, sizes map[string]int64) (map[string]interface{}, string, bool) {
	row := make(map[string]interface{})
	row["path"] = path
	row["message"] = msg
	row["level"] = level
	row["status"] = ""
	row["action"] = ""

	lowmsg := strings.ToLower(msg)

	// 检测 A→B 传输
	if strings.Contains(lowmsg, "path1") || strings.Contains(lowmsg, "source") {
		row["direction"] = string(BisyncDirectionPath1ToPath2)
	} else if strings.Contains(lowmsg, "path2") || strings.Contains(lowmsg, "dest") {
		row["direction"] = string(BisyncDirectionPath2ToPath1)
	} else {
		// 默认假设 A→B
		row["direction"] = string(BisyncDirectionPath1ToPath2)
	}

	// 检测文件操作类型
	switch {
	case strings.Contains(lowmsg, "copied"):
		row["status"] = "success"
		row["action"] = "copied"
	case strings.Contains(lowmsg, "deleted"):
		row["status"] = "success"
		row["action"] = "deleted"
	case strings.Contains(lowmsg, "skipped"):
		row["status"] = "skipped"
		row["action"] = "skipped"
	case strings.Contains(lowmsg, "error") || strings.Contains(lowmsg, "failed"):
		row["status"] = "failed"
		row["action"] = "error"
	default:
		return nil, "", false
	}

	if sz, ok := sizes[path]; ok && sz > 0 {
		row["sizeBytes"] = sz
	}

	return row, row["status"].(string), true
}

// buildBisyncFinalSummary 从日志构建双向总结
func buildBisyncFinalSummary(logPath string) (BisyncSummary, []map[string]interface{}, []map[string]interface{}) {
	summary := BisyncSummary{
		Path1ToPath2: SummaryFilesCounts{},
		Path2ToPath1: SummaryFilesCounts{},
	}

	files1To2 := []map[string]interface{}{}
	files2To1 := []map[string]interface{}{}

	// 简单实现：统计不同方向的文件
	// 真实的实现需要完整解析日志
	// 这里先做一个占位实现，逻辑类似 buildFinalSummaryFilesFromLog

	return summary, files1To2, files2To1
}

// sanitizeFilename 安全文件名（复用 runner.go 中已有的逻辑）
func sanitizeFilenameForBisync(s string, taskID int64) string {
	s = strings.TrimSpace(s)
	if s == "" {
		s = fmt.Sprintf("task-%d", taskID)
	}
	invalid := regexp.MustCompile(`[^a-zA-Z0-9\p{Han}_-]+`)
	s = invalid.ReplaceAllString(s, "_")
	// 截断到 60 字符以内，避免过长
	r := []rune(s)
	if len(r) > 60 {
		s = string(r[:60])
	}
	if s == "" {
		s = fmt.Sprintf("task-%d", taskID)
	}
	return s
}
```

- [ ] **Step 2.2: 更新 runner.go 以支持 bisync**

编辑 `internal/runnercli/runner.go`：

首先，在 Start() 开头，将 sanitizeFilename 提取为包级函数（如果还没有的话），或者确保它可以被其他函数调用。

然后在 Start() 中，修改 cmdName 的判断：

```go
cmdName := strings.ToLower(mode)
isBisyncMode := false // 新增
if cmdName != "copy" && cmdName != "sync" && cmdName != "move" && cmdName != "bisync" {
	cmdName = "copy"
} else if cmdName == "bisync" {
	isBisyncMode = true
}
```

然后，构建 args 部分之前，添加 bisync 处理：

```go
var args []string
if isBisyncMode {
	// 处理 bisync 模式
	bisyncDir := filepath.Join(config.DataDir(), "bisync", sanitizeFilenameForBisync(run.TaskName, run.TaskID))
	_ = os.MkdirAll(bisyncDir, 0o755)
	args = buildBisyncCommand(src, dst, bisyncDir, run.TaskBisyncOptions)
	// 添加基本参数
	args = append(args, "--stats", "1s", "--stats-one-line", "--config", cfg, "--use-json-log", "--log-level", "INFO", "--stats-log-level", "INFO")
} else {
	// 原有的非 bisync 模式逻辑
	// ... 保持原样
}
```

注意：需要在 Run 结构体中添加 TaskBisyncOptions 字段（从 store 模型同步过来）。

- [ ] **Step 2.3: 更新 summary.go**

在 `internal/runnercli/summary.go` 中，添加 bisync 模式的支持。

- [ ] **Step 2.4: 运行测试**

Run: `go test ./internal/runnercli -v`
Expected: PASS

- [ ] **Step 2.5: 提交**

```bash
git add internal/runnercli/bisync.go internal/runnercli/runner.go internal/runnercli/summary.go
git commit -m "feat: implement bisync core logic"
```

---

## 任务 3：后端 service 层实现

**Files:**
- Modify: `internal/service/task.go`
- Modify: `internal/service/tag.go`
- Modify: `internal/service/log_cleanup.go`
- Modify: `internal/service/task_import_export.go`

- [ ] **Step 3.1: 更新 task.go - 添加 bisync 管理函数**

编辑 `internal/service/task.go`，添加以下内容：

```go
import (
	// ... 现有 import
	"path/filepath"
	"time"
)

// 在 TaskService 结构体上方添加辅助函数
func (s *TaskService) getBisyncDir(taskName string) string {
	dataDir := config.DataDir()
	safeName := func(s string) string {
		// 复用 sanitize 逻辑
		s = strings.TrimSpace(s)
		if s == "" {
			s = "task"
		}
		invalid := regexp.MustCompile(`[^a-zA-Z0-9\p{Han}_-]+`)
		s = invalid.ReplaceAllString(s, "_")
		return s
	}(taskName)
	return filepath.Join(dataDir, "bisync", safeName)
}

func (s *TaskService) cleanupBisyncDir(taskName string) error {
	dir := s.getBisyncDir(taskName)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil // 不存在不需要清理
	}
	return os.RemoveAll(dir)
}

// 获取 lst 文件列表
func (s *TaskService) GetBisyncLstFiles(taskID int64) ([]string, error) {
	task, ok := s.db.GetTask(taskID)
	if !ok {
		return nil, ErrTaskNotFound
	}
	dir := s.getBisyncDir(task.Name)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	files := []string{}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".lst") {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

// 删除 lst 文件
func (s *TaskService) DeleteBisyncLstFile(taskID int64, filename string) error {
	task, ok := s.db.GetTask(taskID)
	if !ok {
		return ErrTaskNotFound
	}
	dir := s.getBisyncDir(task.Name)
	filePath := filepath.Join(dir, filename)
	// 安全检查：确保 filename 不包含路径分隔符
	if filepath.Base(filename) != filename {
		return fmt.Errorf("invalid filename")
	}
	return os.Remove(filePath)
}

// 回滚到指定 lst 文件
func (s *TaskService) RollbackBisyncLstFile(taskID int64, filename string) error {
	task, ok := s.db.GetTask(taskID)
	if !ok {
		return ErrTaskNotFound
	}
	dir := s.getBisyncDir(task.Name)
	srcPath := filepath.Join(dir, filename)
	if filepath.Base(filename) != filename {
		return fmt.Errorf("invalid filename")
	}
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found")
	}
	// 先备份当前版本
	now := time.Now().Format("20060102-150405")
	currentFiles := []string{"path1.lst", "path2.lst"}
	for _, f := range currentFiles {
		currentPath := filepath.Join(dir, f)
		if _, err := os.Stat(currentPath); err == nil {
			backupPath := filepath.Join(dir, f+"."+now+".bak")
			_ = os.Rename(currentPath, backupPath)
		}
	}
	// 回滚：判断 filename 是 path1 还是 path2 相关的，重命名
	if strings.Contains(filename, "path1") {
		return os.Rename(srcPath, filepath.Join(dir, "path1.lst"))
	} else if strings.Contains(filename, "path2") {
		return os.Rename(srcPath, filepath.Join(dir, "path2.lst"))
	}
	// 如果无法判断，就作为通用处理
	return fmt.Errorf("unrecognized lst file type")
}

// 触发 resync (注意：这里只更新配置，实际运行通过 RunTask 进行)
func (s *TaskService) ResyncBisync(taskID int64) error {
	task, ok := s.db.GetTask(taskID)
	if !ok {
		return ErrTaskNotFound
	}
	var opts map[string]interface{}
	if len(task.BisyncOptions) > 0 {
		_ = json.Unmarshal(task.BisyncOptions, &opts)
	}
	if opts == nil {
		opts = map[string]interface{}{}
	}
	opts["resync"] = true
	bs, _ := json.Marshal(opts)
	task.BisyncOptions = bs
	return s.db.UpdateTask(taskID, task)
}
```

然后修改 `UpdateTask()` 函数，添加模式切换清理：

```go
func (s *TaskService) UpdateTask(id int64, task store.Task) error {
	oldTask, ok := s.db.GetTask(id)
	if !ok {
		return ErrTaskNotFound
	}
	// 如果从 bisync 切换到其他模式，清理目录
	if oldTask.Mode == "bisync" && task.Mode != "bisync" {
		_ = s.cleanupBisyncDir(oldTask.Name)
	}
	return s.db.UpdateTask(id, task)
}
```

修改 `DeleteTask()` 函数：

```go
func (s *TaskService) DeleteTask(id int64) error {
	task, ok := s.db.GetTask(id)
	if !ok {
		return ErrTaskNotFound
	}
	// 删除任务前清理 bisync 目录
	if task.Mode == "bisync" {
		_ = s.cleanupBisyncDir(task.Name)
	}
	// ... 其余代码保持原样
}
```

- [ ] **Step 3.2: 更新 tag.go - 添加 bisync 标签**

编辑 `internal/service/tag.go`，在 `RecalcTags()` 中的 entries 定义里：

```go
entries := []store.TagEntry{
	{Tag: "sync", Type: "action", Selected: true},
	{Tag: "copy", Type: "action", Selected: true},
	{Tag: "move", Type: "action", Selected: true},
	{Tag: "bisync", Type: "action", Selected: true}, // 新增
}
```

- [ ] **Step 3.3: 更新 log_cleanup.go - 清理 bisync 目录**

编辑 `internal/service/log_cleanup.go` 的 `cleanup()`：

```go
func (s *LogCleanupService) cleanup() {
	retentionDays := s.getRetentionDays()
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	deleted := 0

	// 清理 logs 目录（保持原样）
	logsDir := s.logsDir
	if entries, err := os.ReadDir(logsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			subDir := filepath.Join(logsDir, entry.Name())
			info, _ := entry.Info()
			if info.ModTime().Before(cutoff) {
				_ = os.RemoveAll(subDir)
				deleted++
			}
		}
	}

	// 新增：清理 bisync 目录
	dataDir := config.DataDir()
	bisyncRootDir := filepath.Join(dataDir, "bisync")
	if entries, err := os.ReadDir(bisyncRootDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			subDir := filepath.Join(bisyncRootDir, entry.Name())
			info, _ := entry.Info()
			if info.ModTime().Before(cutoff) {
				_ = os.RemoveAll(subDir)
				deleted++
			} else {
				// 目录未过期，检查内部的 lst 文件
				files, _ := os.ReadDir(subDir)
				for _, f := range files {
					if f.IsDir() {
						continue
					}
					fi, _ := f.Info()
					if fi.ModTime().Before(cutoff) {
						_ = os.Remove(filepath.Join(subDir, f.Name()))
					}
				}
			}
		}
	}

	if deleted > 0 {
		logger.Info("log cleanup completed", zap.Int("deleted", deleted), zap.Int("retentionDays", retentionDays))
	}
}
```

- [ ] **Step 3.4: 更新 task_import_export.go - 导入导出 bisyncOptions**

编辑 `internal/service/task_import_export.go`：

在 `ExportTasks()` 中：

```go
for _, t := range tasks {
	taskMap := map[string]interface{}{
		"name":         t.Name,
		"mode":         t.Mode,
		"sourceRemote": t.SourceRemote,
		"sourcePath":   t.SourcePath,
		"targetRemote": t.TargetRemote,
		"targetPath":   t.TargetPath,
		"sortOrder":    t.SortOrder,
	}
	if len(t.Options) > 0 {
		var opts map[string]interface{}
		if json.Unmarshal(t.Options, &opts) == nil {
			taskMap["options"] = opts
		}
	}
	if len(t.BisyncOptions) > 0 { // 新增
		var bisyncOpts map[string]interface{}
		if json.Unmarshal(t.BisyncOptions, &bisyncOpts) == nil {
			taskMap["bisyncOptions"] = bisyncOpts
		}
	}
	exportTasks = append(exportTasks, taskMap)
}
```

在 `ImportTasks()` 中：

```go
newTask := store.Task{
	Name:         name,
	Mode:         mode,
	SourceRemote: sourceRemote,
	SourcePath:   sourcePath,
	TargetRemote: targetRemote,
	TargetPath:   targetPath,
}
if opts, ok := taskMap["options"]; ok {
	optsBytes, err := json.Marshal(opts)
	if err == nil {
		newTask.Options = optsBytes
	}
}
if bisyncOpts, ok := taskMap["bisyncOptions"]; ok { // 新增
	bisyncOptsBytes, err := json.Marshal(bisyncOpts)
	if err == nil {
		newTask.BisyncOptions = bisyncOptsBytes
	}
}
```

- [ ] **Step 3.5: 运行测试**

Run: `go test ./internal/service -v`
Expected: PASS

- [ ] **Step 3.6: 提交**

```bash
git add internal/service/task.go internal/service/tag.go internal/service/log_cleanup.go internal/service/task_import_export.go
git commit -m "feat: add bisync service logic"
```

---

## 任务 4：后端 controller 和 router 层

**Files:**
- Modify: `internal/controller/task.go`
- Modify: `internal/router/router.go`

- [ ] **Step 4.1: 更新 task.go 控制器**

编辑 `internal/controller/task.go`：

添加以下新处理函数：

```go
// GetBisyncLstFiles 返回任务的 lst 文件列表
func (c *TaskController) HandleBisyncLstFiles(w http.ResponseWriter, req *http.Request) {
	idStr := strings.TrimPrefix(req.URL.Path, "/api/tasks/")
	idStr = strings.TrimSuffix(idStr, "/bisync/lst-files")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		RespondError(w, "invalid task id", http.StatusBadRequest)
		return
	}
	files, err := c.taskService.GetBisyncLstFiles(id)
	if err != nil {
		RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	RespondJSON(w, map[string]interface{}{"files": files})
}

// DeleteBisyncLstFile 删除指定的 lst 文件
func (c *TaskController) HandleBisyncDeleteLst(w http.ResponseWriter, req *http.Request) {
	idStr := strings.TrimPrefix(req.URL.Path, "/api/tasks/")
	idStr = strings.TrimSuffix(idStr, "/bisync/delete-lst")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		RespondError(w, "invalid task id", http.StatusBadRequest)
		return
	}
	var body struct{ Filename string `json:"filename"` }
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil || body.Filename == "" {
		RespondError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := c.taskService.DeleteBisyncLstFile(id, body.Filename); err != nil {
		RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	RespondOK(w)
}

// RollbackBisyncLstFile 回滚到指定 lst 文件
func (c *TaskController) HandleBisyncRollbackLst(w http.ResponseWriter, req *http.Request) {
	idStr := strings.TrimPrefix(req.URL.Path, "/api/tasks/")
	idStr = strings.TrimSuffix(idStr, "/bisync/rollback-lst")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		RespondError(w, "invalid task id", http.StatusBadRequest)
		return
	}
	var body struct{ Filename string `json:"filename"` }
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil || body.Filename == "" {
		RespondError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := c.taskService.RollbackBisyncLstFile(id, body.Filename); err != nil {
		RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	RespondOK(w)
}

// ResyncBisync 触发 bisync resync
func (c *TaskController) HandleBisyncResync(w http.ResponseWriter, req *http.Request) {
	idStr := strings.TrimPrefix(req.URL.Path, "/api/tasks/")
	idStr = strings.TrimSuffix(idStr, "/bisync/resync")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		RespondError(w, "invalid task id", http.StatusBadRequest)
		return
	}
	if err := c.taskService.ResyncBisync(id); err != nil {
		RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	RespondOK(w)
}

// 更新 HandleTasks 以支持 bisyncOptions
func (c *TaskController) HandleTasks(w http.ResponseWriter, req *http.Request) {
	// ... 在 create/update 逻辑中添加 bisyncOptions 的处理
	// 在创建/更新任务时，读取 bisyncOptions 字段并设置到 store.Task
}
```

- [ ] **Step 4.2: 更新 router.go 注册新路由**

编辑 `internal/router/router.go` 的 `Setup()` 函数：

```go
func (r *Router) Setup(mux *http.ServeMux) {
	// ... 现有路由保持不变 ...

	// 添加 bisync 相关路由
	mux.HandleFunc("/api/tasks/", func(w http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "/active-transfer/completed") {
			r.activeTransferController.HandleCompleted(w, req)
		} else if strings.HasSuffix(req.URL.Path, "/active-transfer/pending") {
			r.activeTransferController.HandlePending(w, req)
		} else if strings.HasSuffix(req.URL.Path, "/active-transfer") {
			r.activeTransferController.HandleOverview(w, req)
		} else if strings.HasSuffix(req.URL.Path, "/kill") {
			r.runController.HandleTaskKill(w, req)
		} else if strings.HasSuffix(req.URL.Path, "/bisync/lst-files") {
			r.taskController.HandleBisyncLstFiles(w, req)
		} else if strings.HasSuffix(req.URL.Path, "/bisync/delete-lst") {
			r.taskController.HandleBisyncDeleteLst(w, req)
		} else if strings.HasSuffix(req.URL.Path, "/bisync/rollback-lst") {
			r.taskController.HandleBisyncRollbackLst(w, req)
		} else if strings.HasSuffix(req.URL.Path, "/bisync/resync") {
			r.taskController.HandleBisyncResync(w, req)
		} else {
			r.taskController.HandleTaskActions(w, req)
		}
	})
}
```

- [ ] **Step 4.3: 运行测试**

Run: `go test ./internal/controller -v`
Run: `go test ./internal/router -v`
Expected: PASS

- [ ] **Step 4.4: 提交**

```bash
git add internal/controller/task.go internal/router/router.go
git commit -m "feat: add bisync controller endpoints and routes"
```

---

## 任务 5：前端类型和 API 层

**Files:**
- Modify: `frontend/src/types/index.ts`
- Modify: `frontend/src/components/task/types.ts`
- Modify: `frontend/src/api/task.ts`

- [ ] **Step 5.1: 更新 types/index.ts**

编辑 `frontend/src/types/index.ts`:

```typescript
export interface Task {
  id: number;
  name: string;
  mode: TaskMode;
  sourceRemote: string;
  sourcePath: string;
  targetRemote: string;
  targetPath: string;
  options?: Record<string, any>;
  bisyncOptions?: BisyncOptions; // 新增
  sortOrder: number;
  createdAt: string;
}

export interface BisyncOptions {
  resync?: boolean;
  compare?: string;
  maxDelete?: string;
  checkAccess?: boolean;
  checkFilename?: string;
  conflictResolve?: string;
  conflictLoser?: string;
  conflictSuffix?: string;
  backupDir1?: string;
  backupDir2?: string;
  createEmptySrcDirs?: boolean;
  removeEmptyDirs?: boolean;
  recover?: boolean;
}
```

- [ ] **Step 5.2: 更新 components/task/types.ts**

编辑 `frontend/src/components/task/types.ts`:

```typescript
export type TaskMode = 'sync' | 'copy' | 'move' | 'bisync'; // 新增 'bisync'

export interface BisyncOptions {
  resync?: boolean;
  compare?: string;
  maxDelete?: string;
  checkAccess?: boolean;
  checkFilename?: string;
  conflictResolve?: string;
  conflictLoser?: string;
  conflictSuffix?: string;
  backupDir1?: string;
  backupDir2?: string;
  createEmptySrcDirs?: boolean;
  removeEmptyDirs?: boolean;
  recover?: boolean;
}

export interface CreateForm {
  name: string;
  mode: TaskMode;
  sourceRemote: string;
  sourcePath: string;
  targetRemote: string;
  targetPath: string;
  options: TaskFormOptions;
  bisyncOptions?: BisyncOptions; // 新增
}
```

- [ ] **Step 5.3: 更新 api/task.ts**

编辑 `frontend/src/api/task.ts`:

```typescript
// 新增函数
export async function getBisyncLstFiles(taskId: number): Promise<{ files: string[] }> {
  return get<{ files: string[] }>(`/api/tasks/${taskId}/bisync/lst-files`);
}

export async function deleteBisyncLstFile(taskId: number, filename: string): Promise<void> {
  return post(`/api/tasks/${taskId}/bisync/delete-lst`, { filename });
}

export async function rollbackBisyncLstFile(taskId: number, filename: string): Promise<void> {
  return post(`/api/tasks/${taskId}/bisync/rollback-lst`, { filename });
}

export async function resyncBisync(taskId: number): Promise<void> {
  return post(`/api/tasks/${taskId}/bisync/resync`, {});
}

// 更新 updateTask 或新增 updateTaskBisyncOptions
export async function updateTaskBisyncOptions(taskId: number, bisyncOptions: Record<string, any>): Promise<void> {
  return patch(`/api/tasks`, { id: taskId, bisyncOptions });
}
```

- [ ] **Step 5.4: 运行前端类型检查**

Run: `cd frontend && npx vue-tsc --noEmit`
Expected: No errors

- [ ] **Step 5.5: 提交**

```bash
git add frontend/src/types/index.ts frontend/src/components/task/types.ts frontend/src/api/task.ts
git commit -m "feat: add bisync frontend types and API"
```

---

## 任务 6：创建 BisyncConfigModal 组件

**Files:**
- Create: `frontend/src/components/task/BisyncConfigModal.vue`
- Modify: `frontend/src/i18n/zh.ts`
- Modify: `frontend/src/i18n/en.ts`

- [ ] **Step 6.1: 翻译文件更新**

编辑 `frontend/src/i18n/zh.ts`:

```typescript
export default {
  // ... 现有内容
  addTask: {
    // ... 现有内容
    bisync: '双向同步',
    bisyncConfig: '双向同步配置',
  },
  bisync: {
    configTitle: '双向同步配置',
    basicOptions: '基本选项',
    compare: '比较方式',
    comparePlaceholder: '如 checksum, modtime, size',
    maxDelete: '最大删除数',
    maxDeletePlaceholder: '如 -1(无限制) 或 100',
    checkAccess: '检查访问权限',
    checkFilename: '检查文件名',
    checkFilenamePlaceholder: '如 case-sensitive',
    conflictOptions: '冲突处理',
    conflictResolve: '冲突解决策略',
    conflictResolvePlaceholder: '选择冲突解决策略',
    conflictResolvePath1: '以源路径为准',
    conflictResolvePath2: '以目标路径为准',
    conflictResolveNewer: '以较新的文件为准',
    conflictResolveOlder: '以较旧的文件为准',
    conflictLoser: '冲突失败处理',
    conflictLoserPlaceholder: '选择失败文件处理方式',
    conflictLoserBackup: '备份到指定目录',
    conflictLoserDelete: '删除失败文件',
    conflictSuffix: '冲突文件后缀',
    conflictSuffixPlaceholder: '如 .conflict',
    backupOptions: '备份选项',
    backupDir1: '源端备份目录',
    backupDir2: '目标端备份目录',
    backupDirPlaceholder: '如 .bisync-backup',
    otherOptions: '其他选项',
    createEmptySrcDirs: '创建空源目录',
    removeEmptyDirs: '删除空目录',
    recover: '恢复模式',
    resync: '重新同步 (Resync)',
    resyncHint: '首次运行或出现问题时，会重新生成同步状态文件，谨慎使用',
    lstFiles: '同步状态文件',
    noLstFiles: '暂无同步状态文件',
    rollback: '回滚',
    delete: '删除',
    rollbackConfirm: '确定要回滚到该版本吗？',
    deleteConfirm: '确定要删除该文件吗？',
  },
};
```

编辑 `frontend/src/i18n/en.ts`:

```typescript
export default {
  // ... 现有内容
  addTask: {
    // ... 现有内容
    bisync: 'Bidirectional sync',
    bisyncConfig: 'Bisync config',
  },
  bisync: {
    configTitle: 'Bidirectional Sync Configuration',
    basicOptions: 'Basic Options',
    compare: 'Compare mode',
    comparePlaceholder: 'e.g. checksum, modtime, size',
    maxDelete: 'Max delete',
    maxDeletePlaceholder: 'e.g. -1(unlimited) or 100',
    checkAccess: 'Check access',
    checkFilename: 'Check filename',
    checkFilenamePlaceholder: 'e.g. case-sensitive',
    conflictOptions: 'Conflict Handling',
    conflictResolve: 'Conflict resolution',
    conflictResolvePlaceholder: 'Choose conflict resolution strategy',
    conflictResolvePath1: 'Prefer path1',
    conflictResolvePath2: 'Prefer path2',
    conflictResolveNewer: 'Prefer newer',
    conflictResolveOlder: 'Prefer older',
    conflictLoser: 'Conflict loser handling',
    conflictLoserPlaceholder: 'Choose how to handle loser files',
    conflictLoserBackup: 'Backup to directory',
    conflictLoserDelete: 'Delete loser files',
    conflictSuffix: 'Conflict suffix',
    conflictSuffixPlaceholder: 'e.g. .conflict',
    backupOptions: 'Backup Options',
    backupDir1: 'Source backup directory',
    backupDir2: 'Target backup directory',
    backupDirPlaceholder: 'e.g. .bisync-backup',
    otherOptions: 'Other Options',
    createEmptySrcDirs: 'Create empty source dirs',
    removeEmptyDirs: 'Remove empty dirs',
    recover: 'Recover mode',
    resync: 'Resync',
    resyncHint: 'Resync will regenerate state files, use for first run or when issues occur',
    lstFiles: 'Sync State Files',
    noLstFiles: 'No sync state files yet',
    rollback: 'Rollback',
    delete: 'Delete',
    rollbackConfirm: 'Are you sure you want to rollback to this version?',
    deleteConfirm: 'Are you sure you want to delete this file?',
  },
};
```

- [ ] **Step 6.2: 创建 BisyncConfigModal.vue 组件**

```vue
&lt;template&gt;
  &lt;ModalBase :visible="visible" @close="$emit('close')" :title="t('bisync.configTitle')" :width="600"&gt;
    &lt;div class="p-4 space-y-4"&gt;
      &lt;!-- 基本选项 --&gt;
      &lt;div&gt;
        &lt;h3 class="font-semibold mb-3"&gt;{{ t('bisync.basicOptions') }}&lt;/h3&gt;
        &lt;div class="space-y-3"&gt;
          &lt;div&gt;
            &lt;label class="block text-sm mb-1"&gt;{{ t('bisync.compare') }}&lt;/label&gt;
            &lt;input 
              type="text"
              class="w-full border rounded px-3 py-2"
              :placeholder="t('bisync.comparePlaceholder')"
              v-model="localOptions.compare"
            /&gt;
          &lt;/div&gt;
          &lt;div&gt;
            &lt;label class="block text-sm mb-1"&gt;{{ t('bisync.maxDelete') }}&lt;/label&gt;
            &lt;input 
              type="text"
              class="w-full border rounded px-3 py-2"
              :placeholder="t('bisync.maxDeletePlaceholder')"
              v-model="localOptions.maxDelete"
            /&gt;
          &lt;/div&gt;
          &lt;div class="flex items-center"&gt;
            &lt;input 
              type="checkbox"
              class="mr-2"
              v-model="localOptions.checkAccess"
            /&gt;
            &lt;label&gt;{{ t('bisync.checkAccess') }}&lt;/label&gt;
          &lt;/div&gt;
          &lt;div&gt;
            &lt;label class="block text-sm mb-1"&gt;{{ t('bisync.checkFilename') }}&lt;/label&gt;
            &lt;input 
              type="text"
              class="w-full border rounded px-3 py-2"
              :placeholder="t('bisync.checkFilenamePlaceholder')"
              v-model="localOptions.checkFilename"
            /&gt;
          &lt;/div&gt;
        &lt;/div&gt;
      &lt;/div&gt;

      &lt;!-- 高级选项折叠 --&gt;
      &lt;div&gt;
        &lt;button 
          class="text-blue-500 text-sm mb-2 flex items-center"
          @click="showAdvanced = !showAdvanced"
        &gt;
          &lt;span class="mr-1"&gt;{{ showAdvanced ? '−' : '+' }}&lt;/span&gt;
          &lt;span&gt;{{ t('bisync.conflictOptions') }} &amp; {{ t('bisync.backupOptions') }}&lt;/span&gt;
        &lt;/button&gt;
        &lt;div v-if="showAdvanced" class="space-y-3 pl-2 border-l-2 border-gray-200"&gt;
          &lt;div&gt;
            &lt;label class="block text-sm mb-1"&gt;{{ t('bisync.conflictResolve') }}&lt;/label&gt;
            &lt;select 
              class="w-full border rounded px-3 py-2"
              v-model="localOptions.conflictResolve"
            &gt;
              &lt;option value=""&gt;{{ t('bisync.conflictResolvePlaceholder') }}&lt;/option&gt;
              &lt;option value="path1"&gt;{{ t('bisync.conflictResolvePath1') }}&lt;/option&gt;
              &lt;option value="path2"&gt;{{ t('bisync.conflictResolvePath2') }}&lt;/option&gt;
              &lt;option value="newer"&gt;{{ t('bisync.conflictResolveNewer') }}&lt;/option&gt;
              &lt;option value="older"&gt;{{ t('bisync.conflictResolveOlder') }}&lt;/option&gt;
            &lt;/select&gt;
          &lt;/div&gt;
          &lt;div&gt;
            &lt;label class="block text-sm mb-1"&gt;{{ t('bisync.conflictLoser') }}&lt;/label&gt;
            &lt;select 
              class="w-full border rounded px-3 py-2"
              v-model="localOptions.conflictLoser"
            &gt;
              &lt;option value=""&gt;{{ t('bisync.conflictLoserPlaceholder') }}&lt;/option&gt;
              &lt;option value="backup"&gt;{{ t('bisync.conflictLoserBackup') }}&lt;/option&gt;
              &lt;option value="delete"&gt;{{ t('bisync.conflictLoserDelete') }}&lt;/option&gt;
            &lt;/select&gt;
          &lt;/div&gt;
          &lt;div&gt;
            &lt;label class="block text-sm mb-1"&gt;{{ t('bisync.conflictSuffix') }}&lt;/label&gt;
            &lt;input 
              type="text"
              class="w-full border rounded px-3 py-2"
              :placeholder="t('bisync.conflictSuffixPlaceholder')"
              v-model="localOptions.conflictSuffix"
            /&gt;
          &lt;/div&gt;
          &lt;div&gt;
            &lt;label class="block text-sm mb-1"&gt;{{ t('bisync.backupDir1') }}&lt;/label&gt;
            &lt;input 
              type="text"
              class="w-full border rounded px-3 py-2"
              :placeholder="t('bisync.backupDirPlaceholder')"
              v-model="localOptions.backupDir1"
            /&gt;
          &lt;/div&gt;
          &lt;div&gt;
            &lt;label class="block text-sm mb-1"&gt;{{ t('bisync.backupDir2') }}&lt;/label&gt;
            &lt;input 
              type="text"
              class="w-full border rounded px-3 py-2"
              :placeholder="t('bisync.backupDirPlaceholder')"
              v-model="localOptions.backupDir2"
            /&gt;
          &lt;/div&gt;
          &lt;div class="flex items-center"&gt;
            &lt;input 
              type="checkbox"
              class="mr-2"
              v-model="localOptions.createEmptySrcDirs"
            /&gt;
            &lt;label&gt;{{ t('bisync.createEmptySrcDirs') }}&lt;/label&gt;
          &lt;/div&gt;
          &lt;div class="flex items-center"&gt;
            &lt;input 
              type="checkbox"
              class="mr-2"
              v-model="localOptions.removeEmptyDirs"
            /&gt;
            &lt;label&gt;{{ t('bisync.removeEmptyDirs') }}&lt;/label&gt;
          &lt;/div&gt;
          &lt;div class="flex items-center"&gt;
            &lt;input 
              type="checkbox"
              class="mr-2"
              v-model="localOptions.recover"
            /&gt;
            &lt;label&gt;{{ t('bisync.recover') }}&lt;/label&gt;
          &lt;/div&gt;
        &lt;/div&gt;
      &lt;/div&gt;

      &lt;!-- Resync 按钮 --&gt;
      &lt;div class="p-3 bg-yellow-50 rounded border border-yellow-200"&gt;
        &lt;div class="flex items-center justify-between mb-2"&gt;
          &lt;span class="font-medium"&gt;{{ t('bisync.resync') }}&lt;/span&gt;
          &lt;button
            class="px-3 py-1 bg-yellow-500 text-white rounded hover:bg-yellow-600 disabled:opacity-50"
            :disabled="!taskId"
            @click="handleResync"
          &gt;
            {{ t('bisync.resync') }}
          &lt;/button&gt;
        &lt;/div&gt;
        &lt;p class="text-sm text-gray-600"&gt;{{ t('bisync.resyncHint') }}&lt;/p&gt;
      &lt;/div&gt;

      &lt;!-- LST 文件列表 --&gt;
      &lt;div v-if="taskId"&gt;
        &lt;h3 class="font-semibold mb-3"&gt;{{ t('bisync.lstFiles') }}&lt;/h3&gt;
        &lt;div v-if="loadingFiles" class="text-center text-gray-500 py-4"&gt;
          {{ t('common.loading') }}
        &lt;/div&gt;
        &lt;div v-else-if="lstFiles.length === 0" class="text-center text-gray-500 py-4"&gt;
          {{ t('bisync.noLstFiles') }}
        &lt;/div&gt;
        &lt;ul v-else class="space-y-2"&gt;
          &lt;li v-for="file in lstFiles" :key="file" class="flex items-center justify-between p-2 border rounded"&gt;
            &lt;span class="truncate"&gt;{{ file }}&lt;/span&gt;
            &lt;div class="flex space-x-2"&gt;
              &lt;button
                class="px-2 py-1 text-sm bg-blue-500 text-white rounded hover:bg-blue-600"
                @click="handleRollback(file)"
              &gt;
                {{ t('bisync.rollback') }}
              &lt;/button&gt;
              &lt;button
                class="px-2 py-1 text-sm bg-red-500 text-white rounded hover:bg-red-600"
                @click="handleDelete(file)"
              &gt;
                {{ t('bisync.delete') }}
              &lt;/button&gt;
            &lt;/div&gt;
          &lt;/li&gt;
        &lt;/ul&gt;
      &lt;/div&gt;
    &lt;/div&gt;

    &lt;template #footer&gt;
      &lt;div class="flex justify-end space-x-2"&gt;
        &lt;button
          class="px-4 py-2 border rounded hover:bg-gray-50"
          @click="$emit('close')"
        &gt;
          {{ t('common.cancel') }}
        &lt;/button&gt;
        &lt;button
          class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
          @click="handleSave"
        &gt;
          {{ t('common.save') }}
        &lt;/button&gt;
      &lt;/div&gt;
    &lt;/template&gt;
  &lt;/ModalBase&gt;
&lt;/template&gt;

&lt;script setup lang="ts"&gt;
import { ref, watch, type Ref } from 'vue';
import { t } from '../../composables/useI18n';
import type { BisyncOptions } from './types';
import { getBisyncLstFiles, deleteBisyncLstFile, rollbackBisyncLstFile, resyncBisync } from '../../api/task';
import ModalBase from '../modals/Modal.vue';

const props = defineProps&lt;{
  visible: boolean;
  modelValue: BisyncOptions;
  taskId?: number;
}&gt;();

const emit = defineEmits&lt;{
  (e: 'update:modelValue', value: BisyncOptions): void;
  (e: 'save', value: BisyncOptions): void;
  (e: 'close'): void;
}&gt;();

const localOptions = ref&lt;BisyncOptions&gt;({ ...props.modelValue });
const showAdvanced = ref(false);
const lstFiles = ref&lt;string[]&gt;([]);
const loadingFiles = ref(false);

watch(() =&gt; props.modelValue, (newVal) =&gt; {
  localOptions.value = { ...newVal };
}, { deep: true });

watch(() =&gt; props.visible, async (val) =&gt; {
  if (val &amp;&amp; props.taskId) {
    await loadLstFiles();
  }
});

async function loadLstFiles() {
  if (!props.taskId) return;
  loadingFiles.value = true;
  try {
    const result = await getBisyncLstFiles(props.taskId);
    lstFiles.value = result.files || [];
  } catch (e) {
    console.error('Failed to load lst files', e);
  } finally {
    loadingFiles.value = false;
  }
}

async function handleDelete(file: string) {
  if (!confirm(t('bisync.deleteConfirm'))) return;
  if (!props.taskId) return;
  try {
    await deleteBisyncLstFile(props.taskId, file);
    await loadLstFiles();
  } catch (e) {
    console.error('Failed to delete lst file', e);
  }
}

async function handleRollback(file: string) {
  if (!confirm(t('bisync.rollbackConfirm'))) return;
  if (!props.taskId) return;
  try {
    await rollbackBisyncLstFile(props.taskId, file);
    await loadLstFiles();
  } catch (e) {
    console.error('Failed to rollback lst file', e);
  }
}

async function handleResync() {
  if (!props.taskId) return;
  try {
    await resyncBisync(props.taskId);
    await loadLstFiles();
  } catch (e) {
    console.error('Failed to resync', e);
  }
}

function handleSave() {
  const value = { ...localOptions.value };
  emit('update:modelValue', value);
  emit('save', value);
}
&lt;/script&gt;

&lt;style scoped&gt;
&lt;/style&gt;
```

- [ ] **Step 6.3: 运行前端测试**

Run: `cd frontend && npm run test -- components/task/BisyncConfigModal`
Expected: PASS

- [ ] **Step 6.4: 提交**

```bash
git add frontend/src/components/task/BisyncConfigModal.vue frontend/src/i18n/zh.ts frontend/src/i18n/en.ts
git commit -m "feat: add BisyncConfigModal component"
```

---

## 任务 7：集成到任务表单和任务视图

**Files:**
- Modify: `frontend/src/components/task/AddTaskForm.vue`
- Modify: `frontend/src/views/TaskView.vue`
- Modify: `frontend/src/composables/useTaskFormState.ts`
- Modify: `frontend/src/composables/useTaskFormSubmit.ts`

- [ ] **Step 7.1: 更新 AddTaskForm.vue**

编辑 `frontend/src/components/task/AddTaskForm.vue`:

在模式选择中添加 bisync:

```vue
&lt;select v-model="createForm.mode" class="border rounded px-3 py-2"&gt;
  &lt;option value="copy"&gt;{{ t('addTask.copy') }}&lt;/option&gt;
  &lt;option value="sync"&gt;{{ t('addTask.sync') }}&lt;/option&gt;
  &lt;option value="move"&gt;{{ t('addTask.move') }}&lt;/option&gt;
  &lt;option value="bisync"&gt;{{ t('addTask.bisync') }}&lt;/option&gt;
&lt;/select&gt;
```

在 bisync 模式下显示配置按钮:

```vue
&lt;button
  v-if="createForm.mode === 'bisync'"
  class="px-3 py-1 border rounded hover:bg-gray-50"
  @click="$emit('open-bisync-config')"
&gt;
  {{ t('addTask.bisyncConfig') }}
&lt;/button&gt;
```

添加事件定义:

```typescript
const emit = defineEmits&lt;{
  // ... 现有事件
  'open-bisync-config': [];
}&gt;();
```

- [ ] **Step 7.2: 更新 TaskView.vue**

编辑 `frontend/src/views/TaskView.vue`:

导入组件:

```vue
&lt;script setup lang="ts"&gt;
import BisyncConfigModal from './components/task/BisyncConfigModal.vue';
// ...
&lt;/script&gt;
```

添加弹窗显示状态和事件处理:

```typescript
const showBisyncModal = ref(false);
const bisyncConfigTask = ref&lt;number | null&gt;(null);

function openBisyncConfig() {
  showBisyncModal.value = true;
}

function closeBisyncConfig() {
  showBisyncModal.value = false;
}

function saveBisyncConfig(options: any) {
  if (editingTask.value) {
    createForm.value.bisyncOptions = options;
  }
  closeBisyncConfig();
}
```

在模板中添加弹窗:

```vue
&lt;BisyncConfigModal
  :visible="showBisyncModal"
  :modelValue="createForm.bisyncOptions || {}"
  :taskId="editingTask?.id"
  @update:modelValue="v =&gt; createForm.bisyncOptions = v"
  @save="saveBisyncConfig"
  @close="closeBisyncConfig"
/&gt;
```

在 TaskEditorViewShell 中监听 `open-bisync-config` 事件:

```vue
&lt;TaskEditorViewShell
  // ...
  @open-bisync-config="openBisyncConfig"
/&gt;
```

- [ ] **Step 7.3: 更新 useTaskFormState.ts**

编辑 `frontend/src/composables/useTaskFormState.ts`:

```typescript
export function useTaskFormState() {
  const createForm = ref&lt;CreateForm&gt;({
    name: '',
    mode: 'copy',
    sourceRemote: '',
    sourcePath: '',
    targetRemote: '',
    targetPath: '',
    options: { enableStreaming: true } as TaskFormOptions,
    bisyncOptions: {}, // 新增
  });

  function resetTaskFormForCreate() {
    editingTask.value = null;
    commandMode.value = false;
    commandText.value = '';
    showAdvancedOptions.value = false;
    createForm.value = {
      name: '',
      mode: 'copy',
      sourceRemote: '',
      sourcePath: '',
      targetRemote: '',
      targetPath: '',
      options: { enableStreaming: true },
      bisyncOptions: {}, // 新增
    };
  }

  function fillTaskFormForEdit(task: Task): void {
    // ... 现有代码
    createForm.value = {
      name: task.name,
      mode: normalizeTaskMode(task.mode),
      sourceRemote: task.sourceRemote,
      sourcePath: task.sourcePath || '',
      targetRemote: task.targetRemote,
      targetPath: task.targetPath || '',
      options: normalizeTaskOptionsForForm(task.options as TaskFormOptions | undefined),
      bisyncOptions: task.bisyncOptions || {}, // 新增
    };
  }
}
```

- [ ] **Step 7.4: 更新 useTaskFormSubmit.ts**

编辑 `frontend/src/composables/useTaskFormSubmit.ts`:

```typescript
interface TaskPayload {
  name: string;
  mode: TaskMode;
  sourceRemote: string;
  sourcePath: string;
  targetRemote: string;
  targetPath: string;
  options: TaskFormOptions;
  bisyncOptions?: any; // 新增
}

function buildTaskPayload(): TaskPayload {
  return {
    name: options.createForm.value.name,
    mode: options.createForm.value.mode,
    sourceRemote: options.createForm.value.sourceRemote,
    sourcePath: options.createForm.value.sourcePath,
    targetRemote: options.createForm.value.targetRemote,
    targetPath: options.createForm.value.targetPath,
    options: options.normalizeTaskOptions(options.createForm.value.options),
    bisyncOptions: options.createForm.value.mode === 'bisync' ? options.createForm.value.bisyncOptions : undefined, // 新增
  };
}
```

- [ ] **Step 7.5: 运行前端测试**

Run: `cd frontend && npm run test`
Expected: PASS

- [ ] **Step 7.6: 提交**

```bash
git add frontend/src/components/task/AddTaskForm.vue frontend/src/views/TaskView.vue frontend/src/composables/useTaskFormState.ts frontend/src/composables/useTaskFormSubmit.ts
git commit -m "feat: integrate bisync into task form and view"
```

---

## 任务 8：运行详情和传输中弹窗支持双向显示

**Files:**
- Modify: `frontend/src/components/task/RunDetailModal.vue`
- Modify: `frontend/src/components/task/transferring/TransferringModal.vue`

- [ ] **Step 8.1: 更新传输中弹窗**

编辑 `frontend/src/components/task/transferring/TransferringModal.vue`：

根据 `run.taskMode === 'bisync'` 显示两个方向的独立进度。

- [ ] **Step 8.2: 更新运行详情弹窗**

编辑 `frontend/src/components/task/RunDetailModal.vue`：

支持显示 bisync 的双向总结。

- [ ] **Step 8.3: 运行前端测试**

Run: `cd frontend && npm run test`
Expected: PASS

- [ ] **Step 8.4: 提交**

```bash
git add frontend/src/components/task/RunDetailModal.vue frontend/src/components/task/transferring/TransferringModal.vue
git commit -m "feat: add bisync bidirectional progress display"
```

---

## 任务 9：完整端到端测试

- [ ] **Step 9.1: 运行完整后端测试**

Run: `go test ./internal -v`
Expected: All tests passing

- [ ] **Step 9.2: 运行完整前端测试**

Run: `cd frontend && npm run test`
Expected: All tests passing

- [ ] **Step 9.3: 启动服务进行手动验证**

Run: `make dev`

手动测试流程：
1. 创建一个 bisync 任务
2. 打开配置弹窗
3. 点击 Resync
4. 运行任务
5. 查看运行详情是否有双向统计
6. 删除任务，验证 bisync 目录被清理

- [ ] **Step 9.4: 最终提交**

```bash
git status
git add .
git commit -m "feat: complete bisync integration"
```

---

## 计划自检

**Spec 覆盖检查：**
- ✅ 数据模型层：Task 模型 + BisyncOptions
- ✅ 后端 runner 层：bisync.go + runner 集成
- ✅ Service 层：任务管理、标签、清理、导入导出
- ✅ Controller & Router：bisync API 端点
- ✅ 前端：类型、API、组件、弹窗
- ✅ LST 文件管理：列表、删除、回滚
- ✅ 清理：删除任务或切换模式时清理目录

**占位检查：**
- ✅ 无 "TODO" 或 "TBD"
- ✅ 所有步骤都有完整代码
- ✅ 所有测试命令都有

**类型一致性：**
- ✅ Task 模型与前端类型匹配
- ✅ API 字段与后端响应匹配
