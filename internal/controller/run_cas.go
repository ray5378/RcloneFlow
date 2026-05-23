package controller

import (
	"fmt"
	"strings"
)

func isCASCompatibleRunSummary(sum map[string]any) bool {
	if sum == nil {
		return false
	}
	if td, ok := sum["transferDefaults"].(map[string]any); ok && td != nil {
		switch v := td["openlistCasCompatible"].(type) {
		case bool:
			return v
		case string:
			return strings.EqualFold(strings.TrimSpace(v), "true")
		}
	}
	if mode, ok := sum["trackingMode"].(string); ok && strings.EqualFold(strings.TrimSpace(mode), "cas") {
		return true
	}
	if at, ok := sum["activeTransfer"].(map[string]any); ok && at != nil {
		if mode, ok := at["trackingMode"].(string); ok && strings.EqualFold(strings.TrimSpace(mode), "cas") {
			return true
		}
	}
	return false
}

func enrichRowMapSizesFromFinalSummary(rowMaps []map[string]any, summary map[string]any) {
	if summary == nil {
		return
	}
	sizeMap := make(map[string]int64)

	at, ok := summary["activeTransfer"].(map[string]any)
	if ok && at != nil {
		completed, ok := at["completed"].(any)
		if ok && completed != nil {
			var completedList []map[string]any
			switch v := completed.(type) {
			case []map[string]any:
				completedList = v
			case []any:
				for _, f := range v {
					if fm, ok := f.(map[string]any); ok {
						completedList = append(completedList, fm)
					}
				}
			}
			for _, f := range completedList {
				path := strings.ReplaceAll(fmt.Sprint(f["path"]), "\\", "/")
				if path == "" {
					continue
				}
				sz := anyToInt64(f["sizeBytes"])
				if sz > 0 {
					sizeMap[path] = sz
				}
			}
		}
	}

	if len(sizeMap) == 0 {
		fs, ok := summary["finalSummary"].(map[string]any)
		if !ok || fs == nil {
			return
		}
		files, ok := fs["files"].(any)
		if !ok || files == nil {
			return
		}
		var fileList []map[string]any
		switch v := files.(type) {
		case []map[string]any:
			fileList = v
		case []any:
			for _, f := range v {
				if fm, ok := f.(map[string]any); ok {
					fileList = append(fileList, fm)
				}
			}
		default:
			return
		}
		for _, f := range fileList {
			path := strings.ReplaceAll(fmt.Sprint(f["path"]), "\\", "/")
			if path == "" {
				continue
			}
			sz := anyToInt64(f["sizeBytes"])
			if sz > 0 {
				sizeMap[path] = sz
			}
		}
	}

	if len(sizeMap) == 0 {
		return
	}
	for _, rm := range rowMaps {
		if anyToInt64(rm["sizeBytes"]) > 0 {
			continue
		}
		path := strings.ReplaceAll(fmt.Sprint(rm["path"]), "\\", "/")
		if enriched, ok := sizeMap[path]; ok && enriched > 0 {
			rm["sizeBytes"] = enriched
		}
	}
}

func isCASObjectNotFoundFailureRow(path, msg string) bool {
	path = strings.TrimSpace(path)
	msg = strings.ToLower(strings.TrimSpace(msg))
	if path == "" || msg == "" {
		return false
	}
	if !strings.Contains(msg, "failed to copy") {
		return false
	}
	return strings.Contains(msg, "not found") || strings.Contains(msg, "no such file")
}

func isCASAttemptObjectNotFoundSummaryRow(path, msg string) bool {
	path = strings.TrimSpace(path)
	msg = strings.TrimSpace(msg)
	lowPath := strings.ToLower(path)
	lowMsg := strings.ToLower(msg)
	if strings.HasPrefix(lowPath, "attempt ") && strings.Contains(lowPath, " failed with ") {
		return true
	}
	if strings.HasPrefix(lowMsg, "attempt ") && strings.Contains(lowMsg, " failed with ") {
		return true
	}
	if strings.ToLower(strings.TrimSpace(path)) == "<nil>" && strings.Contains(lowMsg, "attempt ") && strings.Contains(lowMsg, " failed with ") {
		return true
	}
	return false
}

func isCASRunObjectNotFoundSummaryRow(path, msg string) bool {
	p := strings.TrimSpace(path)
	lowP := strings.ToLower(p)
	if lowP == "failed to copy" || strings.HasPrefix(lowP, "failed to copy with ") {
		return true
	}
	return false
}

func filterCASHistoricalDetailRows(rows []map[string]any) []map[string]any {
	if len(rows) == 0 {
		return rows
	}
	casMatched := map[string]struct{}{}
	for _, row := range rows {
		action := strings.ToLower(strings.TrimSpace(fmt.Sprint(row["action"])))
		msg := strings.ToLower(strings.TrimSpace(fmt.Sprint(row["message"])))
		path := strings.TrimSpace(fmt.Sprint(row["path"]))
		if action == "cas matched" || strings.Contains(msg, "cas compatible match after source cleanup") {
			if path != "" {
				casMatched[path] = struct{}{}
			}
		}
	}
	filtered := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		path := strings.TrimSpace(fmt.Sprint(row["path"]))
		msg := strings.TrimSpace(fmt.Sprint(row["message"]))
		action := strings.ToLower(strings.TrimSpace(fmt.Sprint(row["action"])))
		if path == "<nil>" {
			continue
		}
		if isCASAttemptObjectNotFoundSummaryRow(path, msg) {
			continue
		}
		if isCASRunObjectNotFoundSummaryRow(path, msg) {
			continue
		}
		if action == "error" {
			if _, ok := casMatched[path]; ok && isCASObjectNotFoundFailureRow(path, msg) {
				continue
			}
		}
		filtered = append(filtered, row)
	}
	return filtered
}