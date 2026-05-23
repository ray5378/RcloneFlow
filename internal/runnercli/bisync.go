package runnercli

import (
	"os"
	"regexp"
	"strings"
)

func classifyBisyncRow(level, path, msg string, sizes map[string]int64) (map[string]any, string, bool) {
	row := map[string]any{"path": path, "at": "", "status": "", "action": "", "sizeBytes": 0}
	if sz, ok := sizes[path]; ok {
		row["sizeBytes"] = sz
	}
	low := strings.ToLower(strings.TrimSpace(msg))
	lowPath := strings.ToLower(path)

	var direction string
	var pathOrigin string
	if strings.Contains(lowPath, "path1") || strings.Contains(low, "path1") {
		pathOrigin = "Path1"
		direction = "path1-to-path2"
	} else if strings.Contains(lowPath, "path2") || strings.Contains(low, "path2") {
		pathOrigin = "Path2"
		direction = "path2-to-path1"
	}

	if pathOrigin != "" {
		row["pathOrigin"] = pathOrigin
		row["direction"] = direction
	}

	switch {
	case strings.Contains(low, "file is new") || strings.Contains(low, "file is newer") || strings.Contains(low, "file is older"):
		row["status"] = "success"
		row["action"] = "Copied"
		return row, "copied", true
	case strings.Contains(low, "file was deleted"):
		row["status"] = "success"
		row["action"] = "Deleted"
		return row, "deleted", true
	case strings.Contains(low, "new or changed in both paths") || strings.Contains(low, "conflict"):
		row["status"] = "warning"
		row["action"] = "Conflict"
		return row, "conflicts", true
	default:
		return nil, "", false
	}
}

func buildBisyncSummaryFilesFromLog(logPath string) ([]map[string]any, map[string]int) {
	files := []map[string]any{}
	counts := map[string]int{
		"copied":   0,
		"deleted":  0,
		"skipped":  0,
		"failed":   0,
		"conflicts": 0,
		"total":    0,
	}
	if logPath == "" {
		return files, counts
	}
	b, err := os.ReadFile(logPath)
	if err != nil {
		return files, counts
	}
	lines := strings.Split(string(b), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		level, path, msg := extractPartsFromLogLine(line)
		if strings.TrimSpace(path) == "" {
			continue
		}
		row, category, ok := classifyBisyncRow(level, path, msg, nil)
		if ok {
			files = append(files, row)
			if category == "copied" {
				counts["copied"]++
			} else if category == "deleted" {
				counts["deleted"]++
			} else if category == "conflicts" {
				counts["conflicts"]++
			} else if category == "skipped" {
				counts["skipped"]++
			} else if category == "failed" {
				counts["failed"]++
			}
		}
	}
	counts["total"] = counts["copied"] + counts["deleted"] + counts["failed"] + counts["conflicts"] + counts["skipped"]
	return files, counts
}

func extractPartsFromLogLine(line string) (string, string, string) {
	m := regexp.MustCompile(`(?:(\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2})\s+)?(INFO|NOTICE|ERROR|WARNING)\s*:\s*(.+?):\s*(.+)$`).FindStringSubmatch(strings.TrimSpace(line))
	if len(m) == 0 {
		m2 := regexp.MustCompile(`(?:(\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2})\s+)?(INFO|NOTICE|ERROR|WARNING)\s*:\s*(.+)$`).FindStringSubmatch(strings.TrimSpace(line))
		if len(m2) == 0 {
			return "", "", line
		}
		return m2[2], "", m2[3]
	}
	return m[2], strings.TrimSpace(m[3]), strings.TrimSpace(m[4])
}

func buildBisyncFlagsFromOptions(opt map[string]any) []string {
	flags := []string{}
	push := func(k string, vs ...string) { flags = append(flags, k); flags = append(flags, vs...) }
	asBool := func(v any) (bool, bool) { b, ok := v.(bool); return b, ok }
	asStr := func(v any) (string, bool) {
		s, ok := v.(string)
		if !ok {
			return "", false
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return "", false
		}
		return s, true
	}

	// bisync 特定参数
	if s, ok := asStr(opt["compare"]); ok {
		push("--compare", s)
	}
	if s, ok := asStr(opt["maxDelete"]); ok {
		push("--max-delete", s)
	}
	if b, ok := asBool(opt["checkAccess"]); ok && b {
		push("--check-access")
	}
	if s, ok := asStr(opt["checkFilename"]); ok {
		push("--check-filename", s)
	}
	if s, ok := asStr(opt["conflictResolve"]); ok {
		push("--conflict-resolve", s)
	}
	if s, ok := asStr(opt["conflictLoser"]); ok {
		push("--conflict-loser", s)
	}
	if s, ok := asStr(opt["conflictSuffix"]); ok {
		push("--conflict-suffix", s)
	}
	if s, ok := asStr(opt["backupDir1"]); ok {
		push("--backup-dir1", s)
	}
	if s, ok := asStr(opt["backupDir2"]); ok {
		push("--backup-dir2", s)
	}
	if b, ok := asBool(opt["createEmptySrcDirs"]); ok && b {
		push("--create-empty-src-dirs")
	}
	if b, ok := asBool(opt["removeEmptyDirs"]); ok && b {
		push("--remove-empty-dirs")
	}
	if b, ok := asBool(opt["recover"]); ok && b {
		push("--recover")
	}
	if b, ok := asBool(opt["resync"]); ok && b {
		push("--resync")
	}

	return flags
}
