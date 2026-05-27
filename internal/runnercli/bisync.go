package runnercli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"rcloneflow/internal/config"
	"rcloneflow/internal/store"
)

type BisyncFileDirection string

const (
	BisyncDirectionPath1ToPath2 BisyncFileDirection = "path1_to_path2"
	BisyncDirectionPath2ToPath1 BisyncFileDirection = "path2_to_path1"
)

type BisyncSummary struct {
	Path1ToPath2 map[string]int `json:"path1ToPath2"`
	Path2ToPath1 map[string]int `json:"path2ToPath1"`
}

// hasValidBisyncLstFiles 检查工作目录中是否有有效的 bisync lst 文件
func hasValidBisyncLstFiles(workDir string) bool {
	entries, err := os.ReadDir(workDir)
	if err != nil {
		return false
	}
	
	// 检查是否有 .lst 文件
	hasPath1Lst := false
	hasPath2Lst := false
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".lst") {
			if strings.Contains(entry.Name(), ".path1.lst") {
				hasPath1Lst = true
			}
			if strings.Contains(entry.Name(), ".path2.lst") {
				hasPath2Lst = true
			}
		}
	}
	return hasPath1Lst && hasPath2Lst
}

// buildBisyncCommand 构建 rclone bisync 命令行参数
func buildBisyncCommand(src, dst, workDir string, options json.RawMessage) []string {
	args := []string{"bisync", src, dst, "--workdir", workDir}

	needResync := false
	if len(options) > 0 {
		var opts map[string]any
		if err := json.Unmarshal(options, &opts); err == nil {
			// 处理各种 bisync 选项
			if resync, ok := opts["resync"].(bool); ok && resync {
				needResync = true
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

	// 如果用户没有指定 resync 且没有有效的 lst 文件，则自动添加 resync
	if !needResync && !hasValidBisyncLstFiles(workDir) {
		needResync = true
	}

	if needResync {
		args = append(args, "--resync")
	}

	return args
}

// classifyBisyncLogRow 解析 bisync 的日志行，判断方向和类型
func classifyBisyncLogRow(level, path, msg string, sizes map[string]int64) (map[string]any, string, bool) {
	row := make(map[string]any)
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
func buildBisyncFinalSummary(logPath string) (BisyncSummary, []map[string]any, []map[string]any) {
	summary := BisyncSummary{
		Path1ToPath2: map[string]int{"copied": 0, "deleted": 0, "skipped": 0, "failed": 0, "total": 0},
		Path2ToPath1: map[string]int{"copied": 0, "deleted": 0, "skipped": 0, "failed": 0, "total": 0},
	}

	files1To2 := []map[string]any{}
	files2To1 := []map[string]any{}

	// 简单实现：统计不同方向的文件
	// 真实的实现需要完整解析日志
	// 这里先做一个占位实现，逻辑类似 buildFinalSummaryFilesFromLog

	return summary, files1To2, files2To1
}

// sanitizeFilenameForBisync 安全文件名（复用 runner.go 中已有的逻辑）
func sanitizeFilenameForBisync(s string, taskID int64) string {
	s = strings.TrimSpace(s)
	if s == "" {
		s = fmt.Sprintf("task-%d", taskID)
	}
	invalid := regexp.MustCompile(`[^a-zA-Z0-9\p{Han}_.-]+`)
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

// buildBisyncArgs 构建 bisync 启动参数（含工作目录创建和选项读取）
func buildBisyncArgs(src, dst string, run store.Run) ([]string, error) {
	dataDir := config.DataDir()
	bisyncDir := filepath.Join(dataDir, "bisync", sanitizeFilenameForBisync(run.TaskName, run.TaskID))
	if err := os.MkdirAll(bisyncDir, 0o755); err != nil {
		return nil, err
	}
	var bisyncOptions json.RawMessage
	if run.Summary != nil {
		if opts, ok := run.Summary["bisyncOptions"]; ok {
			if optsBytes, err := json.Marshal(opts); err == nil {
				bisyncOptions = optsBytes
			}
		}
	}
	return buildBisyncCommand(src, dst, bisyncDir, bisyncOptions), nil
}

// ClearBisyncResyncFlag 清除 bisync 任务的 resync 标志
func ClearBisyncResyncFlag(updater RunUpdater, taskID int64) {
	if t, ok := updater.GetTask(taskID); ok {
		var opts map[string]any
		if len(t.BisyncOptions) > 0 {
			if json.Unmarshal(t.BisyncOptions, &opts) == nil {
				delete(opts, "resync")
				b, _ := json.Marshal(opts)
				t.BisyncOptions = b
				_ = updater.UpdateTask(taskID, t)
			}
		}
	}
}
