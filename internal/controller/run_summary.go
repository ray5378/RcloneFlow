package controller

import (
	"os"
	"regexp"
	"strings"
	"time"

	"rcloneflow/internal/logutil"
	"rcloneflow/internal/service"
	"rcloneflow/internal/util"
)

func buildFinalSummaryFromLog(run service.RunRecord, sum map[string]any) map[string]any {
	logPath, ok := func() (string, bool) {
		if sum != nil {
			if p, ok := sum["stderrFile"].(string); ok && p != "" {
				return p, true
			}
		}
		return "", false
	}()
	if !ok || logPath == "" {
		return nil
	}
	data, err := os.ReadFile(logPath)
	if err != nil || len(data) == 0 {
		return nil
	}
	lines := strings.Split(string(data), "\n")
	re := regexp.MustCompile(`(?:(\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2})\s+)?(INFO|NOTICE|ERROR)\s*:\s*(.+?):\s*(.+)$`)
	tsRe := regexp.MustCompile(`\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2}\s+(?:INFO|NOTICE|ERROR)\s*:`)
	casMatched := map[string]struct{}{}
	for _, ln := range lines {
		l := strings.TrimSpace(ln)
		if l == "" {
			continue
		}
		for _, seg := range logutil.SplitLogSegments(l) {
			m := re.FindStringSubmatch(seg)
			if len(m) == 0 {
				continue
			}
			path := strings.TrimSpace(m[3])
			msg := strings.TrimSpace(m[4])
			if low := strings.ToLower(msg); strings.Contains(low, "cas compatible match after source cleanup") && path != "" {
				casMatched[path] = struct{}{}
			}
		}
	}
	counts := map[string]int{"copied": 0, "deleted": 0, "skipped": 0, "failed": 0, "total": 0}
	for _, ln := range lines {
		l := strings.TrimSpace(ln)
		if l == "" {
			continue
		}
		segments := []string{}
		idx := tsRe.FindAllStringIndex(l, -1)
		if len(idx) > 1 {
			for i := 0; i < len(idx); i++ {
				startI := idx[i][0]
				endI := len(l)
				if i+1 < len(idx) {
					endI = idx[i+1][0]
				}
				segments = append(segments, strings.TrimSpace(l[startI:endI]))
			}
		} else {
			segments = []string{l}
		}
		for _, seg := range segments {
			m := re.FindStringSubmatch(seg)
			if len(m) == 0 {
				continue
			}
			level := strings.ToUpper(strings.TrimSpace(m[2]))
			path := strings.TrimSpace(m[3])
			msg := strings.TrimSpace(m[4])
			_, bucket, ok := classifyHistoricalLogRow(level, path, msg)
			if !ok {
				continue
			}
			if _, casOk := casMatched[path]; casOk && strings.EqualFold(level, "ERROR") && isCASObjectNotFoundFailureRow(path, msg) {
				continue
			}
			counts[bucket]++
			counts["total"]++
		}
	}
	fs := map[string]any{}
	if old, ok := sum["finalSummary"].(map[string]any); ok && old != nil {
		for k, v := range old {
			fs[k] = v
		}
	}
	var start, fin time.Time
	if run.StartedAt != "" {
		if t, e := time.Parse(time.RFC3339, run.StartedAt); e == nil {
			start = t
		}
	}
	if run.FinishedAt != "" {
		if t, e := time.Parse(time.RFC3339, run.FinishedAt); e == nil {
			fin = t
		}
	}
	if start.IsZero() && sum != nil {
		if s, ok := sum["startedAt"].(string); ok {
			if t, e := time.Parse(time.RFC3339, s); e == nil {
				start = t
			}
		}
	}
	if fin.IsZero() && sum != nil {
		if s, ok := sum["finishedAt"].(string); ok {
			if t, e := time.Parse(time.RFC3339, s); e == nil {
				fin = t
			}
		}
	}
	if !start.IsZero() {
		fs["startAt"] = start.Format(time.RFC3339)
	}
	if !fin.IsZero() {
		fs["finishedAt"] = fin.Format(time.RFC3339)
		dur := int64(fin.Sub(start).Seconds())
		if dur < 0 {
			dur = 0
		}
		fs["durationSec"] = dur
		fs["durationText"] = util.HumanDuration(dur)
	}
	fs["counts"] = counts
	if _, ok := fs["totalCount"]; !ok {
		fs["totalCount"] = counts["total"]
	}
	if _, ok := fs["filesCount"]; !ok {
		fs["filesCount"] = counts["total"]
	}
	if _, ok := fs["copiedCount"]; !ok {
		fs["copiedCount"] = counts["copied"]
	}
	if _, ok := fs["deletedCount"]; !ok {
		fs["deletedCount"] = counts["deleted"]
	}
	if _, ok := fs["skippedCount"]; !ok {
		fs["skippedCount"] = counts["skipped"]
	}
	if _, ok := fs["failedCount"]; !ok {
		fs["failedCount"] = counts["failed"]
	}
	if _, ok := fs["transferredBytes"]; !ok {
		fs["transferredBytes"] = run.BytesTransferred
	}
	if _, ok := fs["totalBytes"]; !ok {
		if sum != nil {
			if prog, ok := sum["progress"].(map[string]any); ok {
				if v, ok := prog["totalBytes"].(float64); ok {
					fs["totalBytes"] = int64(v)
				}
			}
		}
	}
	if _, ok := fs["result"]; !ok {
		fs["result"] = run.Status
	}
	return fs
}

func buildLightRunObject(run service.RunRecord, sum map[string]any) map[string]any {
	obj := map[string]any{
		"id":               run.ID,
		"taskId":           run.TaskID,
		"status":           run.Status,
		"trigger":          run.Trigger,
		"startedAt":        run.StartedAt,
		"finishedAt":       run.FinishedAt,
		"taskName":         run.TaskName,
		"taskMode":         run.TaskMode,
		"sourceRemote":     run.SourceRemote,
		"sourcePath":       run.SourcePath,
		"targetRemote":     run.TargetRemote,
		"targetPath":       run.TargetPath,
		"bytesTransferred": run.BytesTransferred,
		"speed":            run.Speed,
		"error":            run.Error,
	}
	if sum != nil {
		light := map[string]any{}
		for _, key := range []string{"startedAt", "finishedAt", "progress", "stderrFile", "pid", "transferDefaults"} {
			if v, ok := sum[key]; ok {
				light[key] = v
			}
		}
		if fs, ok := sum["finalSummary"].(map[string]any); ok {
			if _, ok := fs["totalCount"]; !ok {
				if counts, ok := fs["counts"].(map[string]any); ok {
					if total, ok2 := counts["total"]; ok2 {
						fs["totalCount"] = total
					}
				}
			}
			fsLight := map[string]any{}
			for _, key := range []string{"startAt", "finishedAt", "durationSec", "durationText", "result", "transferredBytes", "totalBytes", "avgSpeedBps", "counts", "totalCount"} {
				if v, ok := fs[key]; ok {
					fsLight[key] = v
				}
			}
			if len(fsLight) > 0 {
				light["finalSummary"] = fsLight
			}
		}
		if len(light) > 0 {
			obj["summary"] = light
		}
	}
	if fs, ok := sum["finalSummary"].(map[string]any); ok {
		if ds, ok2 := fs["durationSec"].(float64); ok2 {
			obj["durationSeconds"] = int64(ds)
		}
		if dt, ok2 := fs["durationText"].(string); ok2 {
			obj["durationText"] = dt
		}
	}
	if _, ok := obj["durationText"]; !ok {
		var start, fin time.Time
		if run.StartedAt != "" {
			if t, e := time.Parse(time.RFC3339, run.StartedAt); e == nil {
				start = t
			}
		}
		if run.FinishedAt != "" {
			if t, e := time.Parse(time.RFC3339, run.FinishedAt); e == nil {
				fin = t
			}
		}
		if start.IsZero() && sum != nil {
			if s, ok := sum["startedAt"].(string); ok {
				if t, e := time.Parse(time.RFC3339, s); e == nil {
					start = t
				}
			}
		}
		if fin.IsZero() && sum != nil {
			if s, ok := sum["finishedAt"].(string); ok {
				if t, e := time.Parse(time.RFC3339, s); e == nil {
					fin = t
				}
			}
		}
		dur := int64(0)
		if !start.IsZero() {
			if !fin.IsZero() {
				dur = int64(fin.Sub(start).Seconds())
			} else {
				dur = int64(time.Since(start).Seconds())
			}
			if dur < 0 {
				dur = 0
			}
		}
		obj["durationSeconds"] = dur
		obj["durationText"] = util.HumanDuration(dur)
	}
	return obj
}

func ensureHistoricalFinalSummary(run service.RunRecord, sum map[string]any) map[string]any {
	if sum == nil {
		return nil
	}
	fs, hasFS := sum["finalSummary"].(map[string]any)
	need := !hasFS || fs == nil
	if !need {
		if counts, ok := fs["counts"].(map[string]any); ok {
			copied, _ := counts["copied"].(float64)
			failed, _ := counts["failed"].(float64)
			total, _ := counts["total"].(float64)
			if total > 0 && copied > 0 && failed > 0 {
				need = true
			}
		}
		if !need {
			counts, _ := fs["counts"].(map[string]any)
			copied := 0.0
			if counts != nil {
				if v, ok := counts["copied"].(float64); ok {
					copied = v
				}
			}
			if copied <= 0 {
				if total, ok := counts["total"].(float64); !ok || total <= 0 {
					need = true
				}
			}
		}
	}
	if !need {
		return sum
	}
	if rebuilt := buildFinalSummaryFromLog(run, sum); rebuilt != nil {
		sum["finalSummary"] = rebuilt
	}
	return sum
}
