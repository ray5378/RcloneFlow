package runnercli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"rcloneflow/internal/adapter"
	"rcloneflow/internal/logger"
	"rcloneflow/internal/store"
)

func mergeMoveRows(files []map[string]any) ([]map[string]any, map[string]int) {
	copiedMap := map[string]map[string]any{}
	deletedMap := map[string]map[string]any{}
	others := []map[string]any{}
	for _, f := range files {
		a := strings.ToLower(fmt.Sprint(f["action"]))
		p := fmt.Sprint(f["path"])
		switch a {
		case "copied":
			copiedMap[p] = f
		case "deleted":
			deletedMap[p] = f
		default:
			others = append(others, f)
		}
	}
	moved := []map[string]any{}
	remainingCopied := []map[string]any{}
	remainingDeleted := []map[string]any{}
	for p, c := range copiedMap {
		if _, ok := deletedMap[p]; ok {
			moved = append(moved, map[string]any{"path": p, "at": c["at"], "status": "success", "action": "Moved", "sizeBytes": c["sizeBytes"]})
			delete(deletedMap, p)
		} else {
			remainingCopied = append(remainingCopied, c)
		}
	}
	for _, d := range deletedMap {
		remainingDeleted = append(remainingDeleted, d)
	}
	merged := append([]map[string]any{}, moved...)
	merged = append(merged, remainingCopied...)
	merged = append(merged, others...)
	merged = append(merged, remainingDeleted...)
	counts := map[string]int{"copied": len(moved) + len(remainingCopied), "deleted": len(remainingDeleted), "failed": 0, "skipped": 0, "total": 0}
	for _, f := range merged {
		a := strings.ToLower(fmt.Sprint(f["action"]))
		s := strings.ToLower(fmt.Sprint(f["status"]))
		if a == "error" || s == "failed" {
			counts["failed"]++
		} else if s == "skipped" || a == "skipped" {
			counts["skipped"]++
		}
	}
	counts["total"] = counts["copied"] + counts["deleted"] + counts["failed"] + counts["skipped"]
	return merged, counts
}

func buildFinalSummaryFilesFromLog(logPath string, openlistCASCompatible bool, moveMode bool) ([]map[string]any, map[string]int) {
	files := []map[string]any{}
	counts := map[string]int{"copied": 0, "deleted": 0, "skipped": 0, "failed": 0, "total": 0}
	if logPath == "" {
		return files, counts
	}
	b, e := os.ReadFile(logPath)
	if e != nil {
		return files, counts
	}
	lines := strings.Split(string(b), "\n")
	sizes := map[string]int64{}
	for _, ln := range lines {
		if m := fileLineRe.FindStringSubmatch(ln); len(m) > 0 {
			name := strings.TrimSpace(m[1])
			var tb float64
			fmt.Sscanf(m[4], "%f", &tb)
			total := int64(tb * unitToMul(m[5]))
			if total > 0 {
				sizes[name] = total
			}
		}
	}
	casMatchedPaths := map[string]struct{}{}
	for _, ln := range lines {
		for _, seg := range splitRunLogSegments(ln) {
			at, level, path, msg, ok := parseRunLogSegment(seg)
			if !ok {
				continue
			}
			row, bucket, ok := classifyRunLogRow(level, path, msg, sizes, openlistCASCompatible)
			if !ok {
				continue
			}
			row["at"] = at
			if bucket == "copied" && strings.EqualFold(anyString(row["action"]), "CAS Matched") {
				casMatchedPaths[strings.TrimSpace(anyString(row["path"]))] = struct{}{}
			}
			files = append(files, row)
		}
	}
	filtered := make([]map[string]any, 0, len(files))
	for _, row := range files {
		action := strings.ToLower(strings.TrimSpace(anyString(row["action"])))
		status := strings.ToLower(strings.TrimSpace(anyString(row["status"])))
		path := strings.TrimSpace(anyString(row["path"]))
		msg := strings.TrimSpace(anyString(row["message"]))
		if action == "error" {
			if isAttemptObjectNotFoundSummary(path, msg) || isRunObjectNotFoundSummary(path, msg) {
				continue
			}
			if _, ok := casMatchedPaths[path]; ok && isCASCompatibleNotFound(path, msg, openlistCASCompatible) {
				continue
			}
		}
		filtered = append(filtered, row)
		switch {
		case action == "error" || status == "failed":
			counts["failed"]++
		case status == "success" && (action == "copied" || action == "cas matched"):
			counts["copied"]++
		case action == "deleted":
			counts["deleted"]++
		case status == "skipped" || action == "skipped":
			counts["skipped"]++
		}
	}
	counts["total"] = counts["copied"] + counts["deleted"] + counts["failed"] + counts["skipped"]
	if moveMode {
		return mergeMoveRows(filtered)
	}
	return filtered, counts
}

func (r *Runner) enrichFilesSizesAsync(runID int64, files []map[string]any, dst, cfg string, openlistCASCompatible bool) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("goroutine panic", zap.Any("panic", r))
			}
		}()
		if len(files) == 0 {
			return
		}
		if len(files) > 5000 {
			return
		}
		cr := &adapter.CmdRunner{}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		out, _, e2 := cr.Run(ctx, []string{"lsjson", dst, "--config", cfg, "--files-only", "--recursive"}...)
		if e2 != nil {
			return
		}
		var arr []map[string]any
		if json.Unmarshal([]byte(out), &arr) != nil {
			return
		}
		m := map[string]int64{}
		for _, it := range arr {
			p, _ := it["Path"].(string)
			if p == "" {
				p, _ = it["path"].(string)
			}
			var sz int64
			switch vv := it["Size"].(type) {
			case float64:
				sz = int64(vv)
			}
			if p != "" {
				p = strings.ReplaceAll(p, "\\", "/")
				m[p] = sz
				if openlistCASCompatible && isCASPath(p) {
					m[trimCASSuffix(p)] = sz
				}
			}
		}
		if len(m) == 0 {
			return
		}
		for i := range files {
			p := strings.ReplaceAll(fmt.Sprint(files[i]["path"]), "\\", "/")
			if sz, ok := m[p]; ok && sz > 0 {
				files[i]["sizeBytes"] = sz
			}
		}
		_ = r.updater.UpdateRun(runID, func(rr *store.Run) {
			if rr == nil || rr.Summary == nil {
				return
			}
			if fs, ok := rr.Summary["finalSummary"].(map[string]any); ok {
				fs["files"] = files
			}
		})
	}()
}