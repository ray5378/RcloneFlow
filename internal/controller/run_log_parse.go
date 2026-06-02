package controller

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
)

func classifyHistoricalLogRow(level, path, msg string) (map[string]any, string, bool) {
	row := map[string]any{"path": path, "status": "", "action": "", "sizeBytes": 0}
	low := strings.ToLower(strings.TrimSpace(msg))
	upperLevel := strings.ToUpper(strings.TrimSpace(level))
	if isCASAttemptObjectNotFoundSummaryRow(path, msg) {
		return nil, "", false
	}
	if isCASRunObjectNotFoundSummaryRow(path, msg) {
		return nil, "", false
	}
	if strings.TrimSpace(path) == "" || strings.TrimSpace(path) == "<nil>" {
		return nil, "", false
	}
	if upperLevel == "ERROR" {
		row["status"] = "failed"
		row["action"] = "Error"
		row["message"] = msg
		return row, "failed", true
	}
	switch {
	case strings.Contains(low, "cas compatible match after source cleanup"):
		row["status"] = "success"
		row["action"] = "CAS Matched"
		row["message"] = msg
		return row, "copied", true
	case strings.Contains(low, "copied"):
		row["status"] = "success"
		row["action"] = "Copied"
		return row, "copied", true
	case strings.Contains(low, "renamed"):
		row["status"] = "success"
		row["action"] = "Renamed"
		return row, "copied", true
	case strings.Contains(low, "moved"):
		row["status"] = "success"
		row["action"] = "Moved"
		return row, "copied", true
	case strings.Contains(low, "deleted") || strings.Contains(low, "removed"):
		row["status"] = "success"
		row["action"] = "Deleted"
		return row, "deleted", true
	case strings.Contains(low, "skipped"):
		row["status"] = "skipped"
		row["action"] = "Skipped"
		return row, "skipped", true
	default:
		return nil, "", false
	}
}

func historicalSegmentSizeBytes(seg string) int64 {
	var rec map[string]any
	if json.Unmarshal([]byte(strings.TrimSpace(seg)), &rec) != nil {
		return 0
	}
	switch v := rec["size"].(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	default:
		return 0
	}
}

func countCompletedFilesFromLog(logPath string) int {
	data, err := os.ReadFile(logPath)
	if err != nil || len(data) == 0 {
		return 0
	}
	lines := strings.Split(string(data), "\n")
	tsRe := regexp.MustCompile(`\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2}\s+(?:INFO|NOTICE|ERROR)\s*:`)
	re := regexp.MustCompile(`(?:(\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2})\s+)?(INFO|NOTICE|ERROR)\s*:\s*(.+?):\s*(.+)$`)
	fileDoneRe := regexp.MustCompile(`(?i)^(?:copied\s*\(new\)|renamed\b|moved\b|deleted\b|removed\b|purged\b)`)
	seen := map[string]struct{}{}
	for _, ln := range lines {
		l := strings.TrimSpace(ln)
		if l == "" {
			continue
		}
		segments := []string{}
		idx := tsRe.FindAllStringIndex(l, -1)
		if len(idx) > 1 {
			for i := 0; i < len(idx); i++ {
				start := idx[i][0]
				end := len(l)
				if i+1 < len(idx) {
					end = idx[i+1][0]
				}
				segments = append(segments, strings.TrimSpace(l[start:end]))
			}
		} else {
			segments = []string{l}
		}
		for _, seg := range segments {
			m := re.FindStringSubmatch(seg)
			if len(m) == 0 {
				continue
			}
			name := strings.TrimSpace(m[3])
			msg := strings.ToLower(strings.TrimSpace(m[4]))
			if name == "" || name == "<nil>" {
				continue
			}
			if strings.HasPrefix(name, "Attempt ") || strings.HasPrefix(name, "Failed to copy") {
				continue
			}
			if !(fileDoneRe.MatchString(msg) || strings.Contains(msg, "cas compatible match after source cleanup")) {
				continue
			}
			seen[name] = struct{}{}
		}
	}
	return len(seen)
}
