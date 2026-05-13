package controller

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	stdiostrconv "strconv"
	"strings"
	"time"

	"rcloneflow/internal/rclone"
	"rcloneflow/internal/service"
)

// resolveLogPath returns absolute log file path for a run, sharing logic for files+download
func classifyHistoricalLogRow(level, path, msg string) (map[string]any, string, bool) {
	row := map[string]any{"path": path, "status": "", "action": "", "sizeBytes": 0}
	low := strings.ToLower(strings.TrimSpace(msg))
	upperLevel := strings.ToUpper(strings.TrimSpace(level))
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

func splitHistoricalLogSegments(line string) []string {
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

func parseHistoricalLogSegment(seg string) (at, level, path, msg string, ok bool) {
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

func isCASObjectNotFoundFailureRow(path, msg string) bool {
	path = strings.TrimSpace(path)
	msg = strings.ToLower(strings.TrimSpace(msg))
	if path == "" || msg == "" {
		return false
	}
	return strings.Contains(msg, "failed to copy") && strings.Contains(msg, "object not found")
}

func isCASAttemptObjectNotFoundSummaryRow(path, msg string) bool {
	path = strings.TrimSpace(path)
	msg = strings.ToLower(strings.TrimSpace(msg))
	if msg == "" {
		return false
	}
	return strings.HasPrefix(msg, "attempt ") && strings.Contains(msg, "object not found") && (path == "" || path == "<nil>" || strings.HasPrefix(strings.ToLower(path), "attempt "))
}

func isCASRunObjectNotFoundSummaryRow(path, msg string) bool {
	path = strings.TrimSpace(path)
	msg = strings.ToLower(strings.TrimSpace(msg))
	if path == "" && msg == "" {
		return false
	}
	return strings.HasPrefix(strings.ToLower(path), "failed to copy with ") && strings.Contains(msg, "last error was: object not found")
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
		fs["durationText"] = humanDuration(dur)
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
		"id": run.ID,
		"taskId": run.TaskID,
		"status": run.Status,
		"trigger": run.Trigger,
		"startedAt": run.StartedAt,
		"finishedAt": run.FinishedAt,
		"taskName": run.TaskName,
		"taskMode": run.TaskMode,
		"sourceRemote": run.SourceRemote,
		"sourcePath": run.SourcePath,
		"targetRemote": run.TargetRemote,
		"targetPath": run.TargetPath,
		"bytesTransferred": run.BytesTransferred,
		"speed": run.Speed,
		"error": run.Error,
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
		obj["durationText"] = humanDuration(dur)
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
	if !need {
		return sum
	}
	if rebuilt := buildFinalSummaryFromLog(run, sum); rebuilt != nil {
		sum["finalSummary"] = rebuilt
	}
	return sum
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
			if name == "" {
				continue
			}
			// 统计明确的单文件完成事件；CAS 命中属于等效已传输，也计入完成数。
			if !(fileDoneRe.MatchString(msg) || strings.Contains(msg, "cas compatible match after source cleanup")) {
				continue
			}
			seen[name] = struct{}{}
		}
	}
	return len(seen)
}

func (c *RunController) resolveLogPath(run service.RunRecord) (string, bool) {
	// 1) summary.stderrFile
	if s, ok := any(run.Summary).(string); ok && s != "" {
		var m map[string]any
		if json.Unmarshal([]byte(s), &m) == nil {
			if p, ok := m["stderrFile"].(string); ok && p != "" {
				return p, true
			}
		}
	}
	if m, ok := any(run.Summary).(map[string]any); ok {
		if p, ok := m["stderrFile"].(string); ok && p != "" {
			return p, true
		}
	}
	// 2) search logs/<task-MMDD>/<HHMM>.log around StartedAt/CreatedAt
	base := "/app/data/logs"
	parseStart := func(s string) (time.Time, bool) {
		layouts := []string{time.RFC3339, "2006-01-02 15:04:05"}
		for _, l := range layouts {
			if t, e := time.ParseInLocation(l, s, time.Local); e == nil {
				return t, true
			}
		}
		return time.Time{}, false
	}
	var t time.Time
	var ok bool
	if run.StartedAt != "" {
		if tt, o := parseStart(run.StartedAt); o {
			t, ok = tt, true
		}
	}
	if !ok && run.FinishedAt != "" {
		if tt, o := parseStart(run.FinishedAt); o {
			t, ok = tt, true
		}
	}
	if ok {
		sub := t.Local().Format("0102")
		sanitize := func(s string) string {
			s = strings.TrimSpace(s)
			if s == "" {
				return s
			}
			inv := regexp.MustCompile(`[^a-zA-Z0-9\p{Han}_-]+`)
			s = inv.ReplaceAllString(s, "_")
			r := []rune(s)
			if len(r) > 60 {
				s = string(r[:60])
			}
			return s
		}
		candDirs := []string{}
		if run.TaskName != "" {
			candDirs = append(candDirs, filepath.Join(base, sanitize(run.TaskName)+"-"+sub))
		}
		entries, _ := os.ReadDir(base)
		for _, ent := range entries {
			if ent.IsDir() && strings.HasSuffix(ent.Name(), "-"+sub) {
				candDirs = append(candDirs, filepath.Join(base, ent.Name()))
			}
		}
		var best string
		var bestDiff int64 = 1 << 62
		for _, dir := range candDirs {
			files, _ := os.ReadDir(dir)
			for _, f := range files {
				if f.IsDir() || !strings.HasSuffix(f.Name(), ".log") {
					continue
				}
				fn := strings.TrimSuffix(f.Name(), ".log")
				if len(fn) == 4 {
					th, _ := stdiostrconv.Atoi(fn[:2])
					tm, _ := stdiostrconv.Atoi(fn[2:])
					cand := time.Date(t.Year(), t.Month(), t.Day(), th, tm, 0, 0, t.Location())
					diff := t.Unix() - cand.Unix()
					if diff < 0 {
						diff = -diff
					}
					if diff < bestDiff {
						bestDiff = diff
						best = filepath.Join(dir, f.Name())
					}
				}
			}
		}
		if best != "" {
			return best, true
		}
	}
	return "", false
}

// RunController 运行记录控制器
type RunController struct {
	runSvc *service.RunService
	rc     *rclone.Client
}

// NewRunController 创建运行记录控制器
func NewRunController(runSvc *service.RunService, rc *rclone.Client) *RunController {
	return &RunController{
		runSvc: runSvc,
		rc:     rc,
	}
}

// HandleRuns 处理运行记录列表
func (c *RunController) HandleRuns(w http.ResponseWriter, r *http.Request) {
	// DELETE /api/runs - 删除所有历史记录
	if r.Method == http.MethodDelete {
		if err := c.runSvc.DeleteAllRuns(); err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, map[string]any{"deleted": true})
		return
	}

	// 解析分页参数
	page := 1
	pageSize := 50
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := r.URL.Query().Get("pageSize"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	runs, total, err := c.runSvc.ListRuns(page, pageSize)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	out := make([]map[string]any, 0, len(runs))
	for _, r := range runs {
		var sum map[string]any
		switch v := any(r.Summary).(type) {
		case string:
			if v != "" {
				_ = json.Unmarshal([]byte(v), &sum)
			}
		case map[string]any:
			sum = v
		}
		sum = ensureHistoricalFinalSummary(r, sum)
		out = append(out, buildLightRunObject(r, sum))
	}
	WriteJSON(w, 200, map[string]any{
		"runs":     out,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// HandleRunsByTask 处理按任务ID删除历史记录
func (c *RunController) HandleRunsByTask(w http.ResponseWriter, r *http.Request) {
	taskId, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/runs/task/"), 10, 64)

	if r.Method == http.MethodDelete {
		if err := c.runSvc.DeleteRunsByTask(taskId); err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, map[string]any{"deleted": true})
		return
	}

	// GET 请求 - 获取该任务的历史记录
	runs, err := c.runSvc.ListRunsByTask(taskId)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	out := make([]map[string]any, 0, len(runs))
	for _, run := range runs {
		var sum map[string]any
		if run.Summary != "" {
			_ = json.Unmarshal([]byte(run.Summary), &sum)
		}
		sum = ensureHistoricalFinalSummary(run, sum)
		out = append(out, buildLightRunObject(run, sum))
	}
	WriteJSON(w, 200, out)
}

// HandleRunStatus 处理运行状态查询
func (c *RunController) HandleRunStatus(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/runs/"), 10, 64)

	// DELETE 请求
	if r.Method == http.MethodDelete {
		if err := c.runSvc.DeleteRun(id); err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, map[string]any{"deleted": true})
		return
	}

	// GET 请求
	runs, _, err := c.runSvc.ListRuns(1, 1000)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}

	for _, run := range runs {
		if run.ID != id {
			continue
		}
		var sum map[string]any
		switch v := any(run.Summary).(type) {
		case string:
			if v != "" {
				_ = json.Unmarshal([]byte(v), &sum)
			}
		case map[string]any:
			sum = v
		}
		sum = ensureHistoricalFinalSummary(run, sum)
		WriteJSON(w, 200, buildLightRunObject(run, sum))
		return
	}
	WriteJSON(w, 404, map[string]any{"error": "run not found"})
}

// HandleRunKillCLI 强制终止指定 run（优先内部 runner；否则按 PID 逐级信号）
func (c *RunController) HandleRunKillCLI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	idStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/runs/"), "/kill")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	// 读出 run，尝试从 summary 取 pid
	runs, _, err := c.runSvc.ListRuns(1, 1000)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	for _, run := range runs {
		if run.ID != id {
			continue
		}
		if killRunBySummary(run) {
			WriteJSON(w, 200, map[string]any{"killed": true})
			return
		}
		break
	}
	WriteJSON(w, 404, map[string]any{"error": "run not found or no pid"})
}

// HandleTaskKill 强制终止某任务的当前 rclone 进程（按最近 run 定位，兼容空窗期）
func (c *RunController) HandleTaskKill(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	idStr = strings.TrimSuffix(idStr, "/kill")
	tid, _ := strconv.ParseInt(strings.Trim(idStr, "/"), 10, 64)
	if tid == 0 {
		WriteJSON(w, 400, map[string]any{"error": "invalid task id"})
		return
	}
	// 找到该任务最近的 run（running/finalizing 优先，找不到就按开始时间最近）
	runs, err := c.runSvc.ListRunsByTask(tid)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	var candidate *service.RunRecord
	for i := range runs {
		r := runs[i]
		if r.Status == "running" || r.Status == "finalizing" {
			candidate = &r
			break
		}
	}
	if candidate == nil {
		// 回退：取最近一条
		if len(runs) > 0 {
			candidate = &runs[0]
		}
	}
	if candidate == nil {
		WriteJSON(w, 404, map[string]any{"error": "no runs for task"})
		return
	}
	if killRunBySummary(*candidate) {
		WriteJSON(w, 200, map[string]any{"killed": true, "runId": candidate.ID})
		return
	}
	WriteJSON(w, 404, map[string]any{"error": "pid not found"})
}

func killRunBySummary(run service.RunRecord) bool {
	var pid int
	var sum map[string]any
	switch v := any(run.Summary).(type) {
	case map[string]any:
		sum = v
	case string:
		if v != "" {
			_ = json.Unmarshal([]byte(v), &sum)
		}
	}
	if sum != nil {
		if p, ok := sum["pid"].(float64); ok {
			pid = int(p)
		}
		if p2, ok := sum["pid"].(int); ok {
			pid = p2
		}
	}
	if pid > 0 {
		if p, err := os.FindProcess(pid); err == nil {
			_ = p.Kill()
			time.Sleep(2 * time.Second)
			if p2, err := os.FindProcess(pid); err == nil {
				_ = p2.Kill()
				time.Sleep(2 * time.Second)
				if p3, err := os.FindProcess(pid); err == nil {
					_ = p3.Kill()
				}
			}
		}
		return true
	}
	return false
}

// HandleRunFiles 返回指定 run 的文件级明细（从 stderr 日志解析）
func (c *RunController) HandleRunFiles(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/runs/"), "/files")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	offset := 0
	limit := 50
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n >= 0 {
			offset = n
		}
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 && n <= 1000 {
			limit = n
		}
	}

	runs, _, err := c.runSvc.ListRuns(1, 1000)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	var logPath string
	moveMode := false
	openlistCASCompatible := false
	for _, run := range runs {
		if run.ID != id {
			continue
		}
		moveMode = strings.EqualFold(run.TaskMode, "move")
		// Summary 里优先取 stderrFile
		if s, ok := any(run.Summary).(string); ok && s != "" {
			var m map[string]any
			if json.Unmarshal([]byte(s), &m) == nil {
				openlistCASCompatible = isCASCompatibleRunSummary(m)
				if p, ok := m["stderrFile"].(string); ok && p != "" {
					logPath = p
				}
			}
		}
		if logPath == "" || !openlistCASCompatible {
			if m, ok := any(run.Summary).(map[string]any); ok {
				if !openlistCASCompatible {
					openlistCASCompatible = isCASCompatibleRunSummary(m)
				}
				if p, ok := m["stderrFile"].(string); ok && p != "" {
					logPath = p
				}
			}
		}
		if logPath == "" {
			// 尝试新目录结构：logs/<任务名-MMDD>/<HHMM>.log
			base := "/app/data/logs"
			// 以 run.StartedAt 推导可能的子目录（如果有）
			if run.StartedAt != "" {
				if t, e := time.Parse(time.RFC3339, run.StartedAt); e == nil {
					sub := t.Local().Format("0102")
					// 任务名可能未知，这里扫描匹配 *-MMDD 目录下最近的 HHMM.log
					entries, _ := os.ReadDir(base)
					for _, ent := range entries {
						name := ent.Name()
						if ent.IsDir() && strings.HasSuffix(name, "-"+sub) {
							// 取该目录内最接近 startedAt 的文件
							files, _ := os.ReadDir(filepath.Join(base, name))
							var best string
							var bestDiff int64 = 1 << 62
							for _, f := range files {
								if f.IsDir() || !strings.HasSuffix(f.Name(), ".log") {
									continue
								}
								fn := strings.TrimSuffix(f.Name(), ".log") // HHMM
								if len(fn) == 4 {
									th, _ := stdiostrconv.Atoi(fn[:2])
									tm, _ := stdiostrconv.Atoi(fn[2:])
									cand := time.Date(t.Year(), t.Month(), t.Day(), th, tm, 0, 0, t.Location())
									diff := abs64(t.Unix() - cand.Unix())
									if diff < bestDiff {
										bestDiff = diff
										best = filepath.Join(base, name, f.Name())
									}
								}
							}
							if best != "" {
								logPath = best
								break
							}
						}
					}
				}
			}
		}
		if logPath == "" {
			logPath = "/app/data/logs/run-" + idStr + "-stderr.log"
		}
		break
	}
	if logPath == "" {
		WriteJSON(w, 404, map[string]any{"error": "log not found"})
		return
	}
	// raw 模式：直接返回日志文本（便于前端/人工对照）
	if strings.EqualFold(r.URL.Query().Get("mode"), "raw") {
		f, e := os.Open(logPath)
		if e != nil {
			WriteJSON(w, 500, map[string]any{"error": e.Error()})
			return
		}
		defer f.Close()
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.ServeFile(w, r, logPath)
		return
	}
	// 读取并解析日志
	data, readErr := os.ReadFile(logPath)
	if readErr != nil {
		WriteJSON(w, 500, map[string]any{"error": readErr.Error()})
		return
	}
	lines := strings.Split(string(data), "\n")
	// 解析更稳：YYYY/MM/DD HH:MM:SS LEVEL : <path>: <msg>
	type Row struct {
		Name    string `json:"name"`
		Status  string `json:"status"`
		Action  string `json:"action"`
		At      string `json:"at"`
		Size    int64  `json:"sizeBytes"`
		Message string `json:"message,omitempty"`
	}
	rowMaps := make([]map[string]any, 0, 200)
	for _, ln := range lines {
		for _, seg := range splitHistoricalLogSegments(ln) {
			at, level, path, msg, ok := parseHistoricalLogSegment(seg)
			if !ok {
				continue
			}
			histRow, _, ok := classifyHistoricalLogRow(level, path, msg)
			if !ok {
				continue
			}
			rowMap := map[string]any{
				"path":     path,
				"name":     path,
				"at":       at,
				"sizeBytes": historicalSegmentSizeBytes(seg),
				"message":  msg,
			}
			if v, ok := histRow["status"].(string); ok {
				rowMap["status"] = v
			}
			if v, ok := histRow["action"].(string); ok {
				rowMap["action"] = v
			}
			rowMaps = append(rowMaps, rowMap)
		}
	}
	if openlistCASCompatible {
		rowMaps = filterCASHistoricalDetailRows(rowMaps)
	}
	rows := make([]Row, 0, len(rowMaps))
	for _, rowMap := range rowMaps {
		rows = append(rows, Row{
			Name:    strings.TrimSpace(fmt.Sprint(rowMap["name"])),
			Status:  strings.TrimSpace(fmt.Sprint(rowMap["status"])),
			Action:  strings.TrimSpace(fmt.Sprint(rowMap["action"])),
			At:      strings.TrimSpace(fmt.Sprint(rowMap["at"])),
			Size:    anyToInt64(rowMap["sizeBytes"]),
			Message: strings.TrimSpace(fmt.Sprint(rowMap["message"])),
		})
	}
	if moveMode {
		copied := map[string]Row{}
		deleted := map[string]Row{}
		others := make([]Row, 0, len(rows))
		for _, row := range rows {
			action := strings.ToLower(strings.TrimSpace(row.Action))
			switch action {
			case "copied":
				copied[row.Name] = row
			case "deleted":
				deleted[row.Name] = row
			default:
				others = append(others, row)
			}
		}
		merged := make([]Row, 0, len(rows))
		for name, row := range copied {
			if _, ok := deleted[name]; ok {
				row.Action = "Moved"
				delete(deleted, name)
			}
			merged = append(merged, row)
		}
		merged = append(merged, others...)
		for _, row := range deleted {
			merged = append(merged, row)
		}
		sort.SliceStable(merged, func(i, j int) bool {
			if merged[i].At != merged[j].At {
				return merged[i].At < merged[j].At
			}
			return merged[i].Name < merged[j].Name
		})
		rows = merged
	}
	// 分页
	total := len(rows)
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = total
	}
	page := rows[offset:end]
	// 附带校验信息，便于核对与“传输日志”一致性
	h := sha1.Sum(data)
	info := map[string]any{"logPath": logPath, "logSize": len(data), "logSha1": fmt.Sprintf("%x", h[:])}
	WriteJSON(w, 200, map[string]any{"total": total, "items": page, "info": info})
}

func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func anyToInt64(v any) int64 {
	switch x := v.(type) {
	case int64:
		return x
	case int:
		return int64(x)
	case float64:
		return int64(x)
	case float32:
		return int64(x)
	case json.Number:
		if n, err := x.Int64(); err == nil {
			return n
		}
	}
	return 0
}

// humanDuration renders seconds to X小时Y分Z秒（省略 0 单位）
func humanDuration(sec int64) string {
	h := sec / 3600
	m := (sec % 3600) / 60
	s := sec % 60
	parts := []string{}
	if h > 0 {
		parts = append(parts, fmt.Sprintf("%d小时", h))
	}
	if m > 0 || (h > 0 && s > 0) {
		parts = append(parts, fmt.Sprintf("%d分", m))
	}
	if s > 0 || (h == 0 && m == 0) {
		parts = append(parts, fmt.Sprintf("%d秒", s))
	}
	return strings.Join(parts, "")
}

// HandleRunLog 统一提供 stderr 单文件下载
func (c *RunController) HandleRunLog(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/runs/"), "/log")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	runs, _, err := c.runSvc.ListRuns(1, 1000)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	for _, run := range runs {
		if run.ID == id {
			// Summary 里优先取 stderrFile
			if s, ok := any(run.Summary).(string); ok && s != "" {
				var m map[string]any
				if json.Unmarshal([]byte(s), &m) == nil {
					if p, ok := m["stderrFile"].(string); ok && p != "" {
						http.ServeFile(w, r, p)
						return
					}
				}
			}
			if m, ok := any(run.Summary).(map[string]any); ok {
				if p, ok := m["stderrFile"].(string); ok && p != "" {
					http.ServeFile(w, r, p)
					return
				}
			}
			// 兼容前端带 auth 查询参数（忽略，仅用于传递 Bearer token 给中间件）
			r.URL.RawQuery = ""
			// 搜索新目录结构：/app/data/logs/<任务名-MMDD>/<HHMM>.log
			base := "/app/data/logs"
			sanitize := func(s string) string {
				s = strings.TrimSpace(s)
				if s == "" {
					return s
				}
				inv := regexp.MustCompile(`[^a-zA-Z0-9\p{Han}_-]+`)
				s = inv.ReplaceAllString(s, "_")
				r := []rune(s)
				if len(r) > 60 {
					s = string(r[:60])
				}
				return s
			}
			if run.TaskName != "" {
				san := sanitize(run.TaskName)
				entries, _ := os.ReadDir(base)
				var best string
				var bestMod int64
				for _, ent := range entries {
					if !ent.IsDir() {
						continue
					}
					name := ent.Name()
					if !strings.HasPrefix(name, san+"-") {
						continue
					}
					sub := filepath.Join(base, name)
					files, _ := os.ReadDir(sub)
					for _, f := range files {
						if f.IsDir() || !strings.HasSuffix(f.Name(), ".log") {
							continue
						}
						fi, _ := f.Info()
						if fi != nil {
							mod := fi.ModTime().Unix()
							if mod > bestMod {
								bestMod = mod
								best = filepath.Join(sub, f.Name())
							}
						}
					}
				}
				if best != "" {
					http.ServeFile(w, r, best)
					return
				}
			}
			// 未找到任何日志文件
			WriteJSON(w, 404, map[string]any{"error": "log not found"})
			return
		}
	}
	WriteJSON(w, 404, map[string]any{"error": "log not found"})
}

// HandleActiveRuns 处理获取所有运行中的任务及其实时状态
func (c *RunController) HandleActiveRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := c.runSvc.ListActiveRuns()
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	items := make([]map[string]any, 0, len(runs))
	for _, run := range runs {
		var summary map[string]any
		var progress map[string]any
		switch v := any(run.Summary).(type) {
		case map[string]any:
			summary = v
			if p, ok := v["progress"].(map[string]any); ok {
				progress = p
			}
		case string:
			if v != "" {
				var m map[string]any
				if json.Unmarshal([]byte(v), &m) == nil {
					summary = m
					if p, ok := m["progress"].(map[string]any); ok {
						progress = p
					}
				}
			}
		}

		bytes := int64(0)
		if v, ok := progress["bytes"].(float64); ok {
			bytes = int64(v)
		}
		progressTotal := int64(0)
		if v, ok := progress["totalBytes"].(float64); ok {
			progressTotal = int64(v)
		}
		total := progressTotal
		if total > 0 && bytes > total {
			bytes = total
		}

		speed := int64(0)
		if v, ok := progress["speed"].(float64); ok {
			speed = int64(v)
		}
		eta := float64(0)
		if v, ok := progress["eta"].(float64); ok {
			eta = v
		}

		pct := 0.0
		if v, ok := progress["percentage"].(float64); ok {
			pct = v
		} else if total > 0 {
			pct = float64(bytes) / float64(total) * 100
		}
		if pct < 0 {
			pct = 0
		}
		if pct > 100 {
			pct = 100
		}

		completedFiles := float64(0)
		if v, ok := progress["completedFiles"].(float64); ok {
			completedFiles = v
		}
		if completedFiles <= 0 {
			if logPath, ok := c.resolveLogPath(run); ok && logPath != "" {
				if n := countCompletedFilesFromLog(logPath); n > 0 {
					completedFiles = float64(n)
				}
			}
		}
		plannedFiles := float64(0)
		if v, ok := progress["plannedFiles"].(float64); ok {
			plannedFiles = v
		}
		logicalTotalCount := plannedFiles
		casCompatible := false
		if summary != nil {
			if opts, ok := summary["effectiveOptions"].(map[string]any); ok {
				casCompatible, _ = opts["openlistCasCompatible"].(bool)
			}
		}
		if logicalTotalCount <= 0 && summary != nil && !casCompatible {
			if pf, ok := summary["preflight"].(map[string]any); ok {
				if v, ok2 := pf["totalCount"].(float64); ok2 {
					logicalTotalCount = v
				}
			}
		}
		if logicalTotalCount > 0 && completedFiles > logicalTotalCount {
			completedFiles = logicalTotalCount
		}

		phase := "transferring"
		if total == 0 && bytes == 0 {
			phase = "preparing"
		}
		if summary != nil {
			if fw, ok := summary["finishWait"].(map[string]any); ok {
				if en, ok2 := fw["enabled"].(bool); ok2 && en {
					if done, ok3 := fw["done"].(bool); !ok3 || !done {
						phase = "finalizing"
					}
				}
			}
		}

		progressLine := ""
		if summary != nil {
			if v, ok := summary["progressLine"].(string); ok {
				progressLine = v
			}
		}

		stable := map[string]any{
			"bytes":             bytes,
			"totalBytes":        total,
			"speed":             speed,
			"eta":               eta,
			"percentage":        pct,
			"phase":             phase,
			"lastUpdatedAt":     time.Now().Format(time.RFC3339),
			"completedFiles":    completedFiles,
			"plannedFiles":      plannedFiles,
			"logicalTotalCount":  logicalTotalCount,
			"totalCount":        logicalTotalCount,
		}

		calcPct := 0.0
		if total > 0 {
			calcPct = float64(bytes) / float64(total) * 100
		}
		pctMismatch := total > 0 && math.Abs(calcPct-pct) > 1.5
		countMismatch := logicalTotalCount > 0 && completedFiles > logicalTotalCount
		etaMismatch := eta > 0 && speed > 0 && total > bytes && math.Abs((float64(total-bytes)/float64(speed))-eta) > 300
		progressMismatch := pctMismatch || countMismatch || etaMismatch
		progressCheck := map[string]any{
			"ok":            !progressMismatch,
			"pctMismatch":   pctMismatch,
			"countMismatch": countMismatch,
			"etaMismatch":   etaMismatch,
			"calcPct":       calcPct,
		}

		item := map[string]any{
			// progress: 运行中 UI 主字段。
			// 这里故意只暴露 live progress + 调试辅助字段；
			// 不要把 finalSummary、completedFreezeByTask 语义或其他完成态摘要重新回流成 active runs 主字段。
			// active runs 的职责只到“当前正在跑的实时帧”为止。
			"runRecord": map[string]any{
				"id":               run.ID,
				"taskId":           run.TaskID,
				"status":           run.Status,
				"bytesTransferred": run.BytesTransferred,
				"error":            run.Error,
				"startedAt":        run.StartedAt,
				"finishedAt":       run.FinishedAt,
			},
			"progress":         stable,
			"progressLine":     progressLine,
			"progressSource":   "summary.progress",
			"progressMismatch": progressMismatch,
			"progressCheck":    progressCheck,
		}
		items = append(items, item)
	}
	// 前端期望返回数组
	// attach frozen durations to active runs too (computed, not stored)
	for i := range items {
		it := items[i]
		if rr, ok := it["runRecord"].(map[string]any); ok {
			var start time.Time
			if s, ok2 := rr["startedAt"].(string); ok2 {
				if t, e := time.Parse(time.RFC3339, s); e == nil {
					start = t
				}
			}
			if !start.IsZero() {
				dur := int64(time.Since(start).Seconds())
				if dur < 0 {
					dur = 0
				}
				rr["durationSeconds"] = dur
				rr["durationText"] = humanDuration(dur)
			}
			it["runRecord"] = rr
		}
		items[i] = it
	}
	WriteJSON(w, 200, items)
}

// HandleGlobalStats 处理获取全局实时统计信息
func (c *RunController) HandleGlobalStats(w http.ResponseWriter, r *http.Request) {
	// 聚合本地 CLI Runner 的活动任务进度（RC 仅用于浏览/管理存储，不涉及进度）
	runs, e2 := c.runSvc.ListActiveRuns()
	if e2 != nil {
		WriteJSON(w, 500, map[string]any{"error": e2.Error()})
		return
	}
	var bytesSum, totalSum, speedSum float64
	for _, run := range runs {
		// 兼容 summary 为 map 或 string
		var p map[string]any
		switch v := any(run.Summary).(type) {
		case map[string]any:
			if pp, ok := v["progress"].(map[string]any); ok {
				p = pp
			}
		case string:
			if v != "" {
				var m map[string]any
				if json.Unmarshal([]byte(v), &m) == nil {
					if pp, ok := m["progress"].(map[string]any); ok {
						p = pp
					}
				}
			}
		}
		if p != nil {
			if v, ok := p["bytes"].(float64); ok {
				bytesSum += v
			}
			if v, ok := p["totalBytes"].(float64); ok {
				totalSum += v
			}
			if v, ok := p["speed"].(float64); ok {
				speedSum += v
			}
		}
	}
	percentage := 0.0
	if totalSum > 0 {
		percentage = (bytesSum / totalSum) * 100
	}
	WriteJSON(w, 200, map[string]any{
		"bytes":      bytesSum,
		"totalBytes": totalSum,
		"speed":      speedSum,
		"speedAvg":   speedSum, // 简化：无历史窗口，先返回当前合计
		"eta":        nil,      // CLI 模式无法可靠聚合 ETA，这里暂置空
		"percentage": percentage,
	})
}
