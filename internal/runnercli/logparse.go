package runnercli

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

func sanitizeRunLogLine(line string, openlistCASCompatible bool) string {
	return line
}

func splitRunLogSegments(line string) []string {
	l := strings.TrimSpace(line)
	if l == "" {
		return nil
	}
	tsRe := regexp.MustCompile(`\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2}\s+(?:INFO|NOTICE|ERROR)\s*:`)
	idx := tsRe.FindAllStringIndex(l, -1)
	if len(idx) <= 1 {
		return []string{l}
	}
	segments := make([]string, 0, len(idx))
	for i := 0; i < len(idx); i++ {
		start := idx[i][0]
		end := len(l)
		if i+1 < len(idx) {
			end = idx[i+1][0]
		}
		segments = append(segments, strings.TrimSpace(l[start:end]))
	}
	return segments
}

func parseRunLogSegment(seg string) (at, level, path, msg string, ok bool) {
	re := regexp.MustCompile(`(?:(\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2})\s+)?(INFO|NOTICE|ERROR)\s*:\s*(.+?):\s*(.+)$`)
	if m := re.FindStringSubmatch(strings.TrimSpace(seg)); len(m) > 0 {
		return strings.TrimSpace(m[1]), strings.ToUpper(strings.TrimSpace(m[2])), strings.TrimSpace(m[3]), strings.TrimSpace(m[4]), true
	}
	var rec map[string]any
	if json.Unmarshal([]byte(strings.TrimSpace(seg)), &rec) == nil {
		level = strings.ToUpper(strings.TrimSpace(fmt.Sprint(rec["level"])))
		msg = strings.TrimSpace(fmt.Sprint(rec["msg"]))
		path = strings.TrimSpace(fmt.Sprint(rec["object"]))
		at = strings.TrimSpace(fmt.Sprint(rec["time"]))
		if at == "" {
			at = strings.TrimSpace(fmt.Sprint(rec["timestamp"]))
		}
		if msg != "" {
			return at, level, path, msg, true
		}
	}
	return "", "", "", "", false
}

func extractPathFromLogLine(line string) string {
	m := regexp.MustCompile(`(?:(\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2})\s+)?(INFO|NOTICE|ERROR)\s*:\s*(.+?):\s*(.+)$`).FindStringSubmatch(strings.TrimSpace(line))
	if len(m) == 0 {
		return ""
	}
	return strings.TrimSpace(m[3])
}

func extractMsgFromLogLine(line string) string {
	m := regexp.MustCompile(`(?:(\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2})\s+)?(INFO|NOTICE|ERROR)\s*:\s*(.+?):\s*(.+)$`).FindStringSubmatch(strings.TrimSpace(line))
	if len(m) == 0 {
		return ""
	}
	return strings.TrimSpace(m[4])
}

func classifyRunLogRow(level, path, msg string, sizes map[string]int64, openlistCASCompatible bool) (map[string]any, string, bool) {
	at := ""
	row := map[string]any{"path": path, "at": at, "status": "", "action": "", "sizeBytes": 0}
	if sz, ok := sizes[path]; ok {
		row["sizeBytes"] = sz
	}
	low := strings.ToLower(strings.TrimSpace(msg))
	upperLevel := strings.ToUpper(strings.TrimSpace(level))
	if isAttemptObjectNotFoundSummary(path, msg) {
		return nil, "", false
	}
	if isRunObjectNotFoundSummary(path, msg) {
		return nil, "", false
	}
	if strings.TrimSpace(path) == "<nil>" {
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

func isAttemptObjectNotFoundSummary(path, msg string) bool {
	pathTrim := strings.TrimSpace(path)
	msgTrim := strings.TrimSpace(msg)
	lowMsg := strings.ToLower(msgTrim)
	if strings.HasPrefix(pathTrim, "Attempt ") && strings.Contains(lowMsg, "object not found") {
		return true
	}
	if strings.HasPrefix(msgTrim, "Attempt ") && strings.Contains(lowMsg, "object not found") {
		return true
	}
	return false
}

func isRunObjectNotFoundSummary(path, msg string) bool {
	p := strings.TrimSpace(path)
	if !strings.EqualFold(p, "Failed to copy with 2 errors") && !strings.HasPrefix(p, "Failed to copy with ") && !strings.EqualFold(p, "Failed to copy") {
		return false
	}
	low := strings.ToLower(strings.TrimSpace(msg))
	return strings.Contains(low, "object not found")
}