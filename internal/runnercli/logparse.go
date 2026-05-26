package runnercli

import (
	"regexp"
	"strings"
)

func sanitizeRunLogLine(line string, openlistCASCompatible bool) string {
	return line
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

func isAttemptSummary(path, msg string) bool {
	pathTrim := strings.TrimSpace(path)
	msgTrim := strings.TrimSpace(msg)
	if strings.HasPrefix(pathTrim, "Attempt ") && strings.Contains(pathTrim, " failed with ") {
		return true
	}
	if strings.HasPrefix(msgTrim, "Attempt ") && strings.Contains(msgTrim, " failed with ") {
		return true
	}
	return false
}

func isFailedToCopySummary(path, msg string) bool {
	p := strings.TrimSpace(path)
	if strings.EqualFold(p, "Failed to copy") || strings.HasPrefix(p, "Failed to copy with ") {
		return true
	}
	return false
}

func isAttemptObjectNotFoundSummary(path, msg string) bool {
	return isAttemptSummary(path, msg)
}

func isRunObjectNotFoundSummary(path, msg string) bool {
	return isFailedToCopySummary(path, msg)
}