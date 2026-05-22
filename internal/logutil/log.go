package logutil

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

func SplitLogSegments(line string) []string {
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

func ParseLogSegment(seg string) (at, level, path, msg string, ok bool) {
	re := regexp.MustCompile(`(?:(\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2})\s+)?(INFO|NOTICE|ERROR)\s*:\s*(.+?):\s*(.+)$`)
	if m := re.FindStringSubmatch(strings.TrimSpace(seg)); len(m) > 0 {
		return strings.TrimSpace(m[1]), strings.ToUpper(strings.TrimSpace(m[2])), strings.TrimSpace(m[3]), strings.TrimSpace(m[4]), true
	}
	var rec map[string]any
	if json.Unmarshal([]byte(strings.TrimSpace(seg)), &rec) == nil {
		level = upperOrEmpty(rec, "level")
		msg = strOrEmpty(rec, "msg")
		path = strOrEmpty(rec, "object")
		at = strOrEmpty(rec, "time")
		if at == "" {
			at = strOrEmpty(rec, "timestamp")
		}
		if msg != "" {
			return at, level, path, msg, true
		}
	}
	return "", "", "", "", false
}

func strOrEmpty(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

func upperOrEmpty(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	return strings.ToUpper(strings.TrimSpace(fmt.Sprint(v)))
}