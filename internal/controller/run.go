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
	"strconv"
	stdiostrconv "strconv"
	"strings"
	"time"

	"rcloneflow/internal/rclone"
	"rcloneflow/internal/service"
)

// resolveLogPath returns absolute log file path for a run, sharing logic for files+download
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
	// attach durationSeconds/durationText (freeze when finished) and passthrough finalSummary
	out := make([]map[string]any, 0, len(runs))
	for _, r := range runs {
		b, _ := json.Marshal(r)
		var obj map[string]any
		_ = json.Unmarshal(b, &obj)
		// parse summary JSON for times/finalSummary
		var sum map[string]any
		switch v := any(r.Summary).(type) {
		case string:
			if v != "" {
				_ = json.Unmarshal([]byte(v), &sum)
			}
		case map[string]any:
			sum = v
		}
		// prefer finalSummary.duration* if exists
		if fs, ok := sum["finalSummary"].(map[string]any); ok {
			if ds, ok2 := fs["durationSec"].(float64); ok2 {
				obj["durationSeconds"] = int64(ds)
			}
			if dt, ok2 := fs["durationText"].(string); ok2 {
				obj["durationText"] = dt
			}
		} else {
			// compute from started/finished
			var start, fin time.Time
			if r.StartedAt != "" {
				if t, e := time.Parse(time.RFC3339, r.StartedAt); e == nil {
					start = t
				}
			}
			if r.FinishedAt != "" {
				if t, e := time.Parse(time.RFC3339, r.FinishedAt); e == nil {
					fin = t
				}
			}
			if start.IsZero() {
				if sum != nil {
					if s, ok := sum["startedAt"].(string); ok {
						if t, e := time.Parse(time.RFC3339, s); e == nil {
							start = t
						}
					}
				}
			}
			if fin.IsZero() {
				if sum != nil {
					if s, ok := sum["finishedAt"].(string); ok {
						if t, e := time.Parse(time.RFC3339, s); e == nil {
							fin = t
						}
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
		out = append(out, obj)
	}
	WriteJSON(w, 200, map[string]any{
		"runs": out,
		"total": total,
		"page": page,
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
	WriteJSON(w, 200, runs)
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
		if run.Status == "running" && run.RcJobID > 0 {
			st, err := c.rc.JobStatus(r.Context(), run.RcJobID)
			if err == nil {
				c.runSvc.UpdateRunStatus(run.ID, st)
				WriteJSON(w, 200, st)
				return
			}
		}
		// attach duration* & passthrough finalSummary
		b, _ := json.Marshal(run)
		var obj map[string]any
		_ = json.Unmarshal(b, &obj)
		var sum map[string]any
		switch v := any(run.Summary).(type) {
		case string:
			if v != "" {
				_ = json.Unmarshal([]byte(v), &sum)
			}
		case map[string]any:
			sum = v
		}
		if fs, ok := sum["finalSummary"].(map[string]any); ok {
			if ds, ok2 := fs["durationSec"].(float64); ok2 {
				obj["durationSeconds"] = int64(ds)
			}
			if dt, ok2 := fs["durationText"].(string); ok2 {
				obj["durationText"] = dt
			}
		} else {
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
			if start.IsZero() {
				if sum != nil {
					if s, ok := sum["startedAt"].(string); ok {
						if t, e := time.Parse(time.RFC3339, s); e == nil {
							start = t
						}
					}
				}
			}
			if fin.IsZero() {
				if sum != nil {
					if s, ok := sum["finishedAt"].(string); ok {
						if t, e := time.Parse(time.RFC3339, s); e == nil {
							fin = t
						}
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
		WriteJSON(w, 200, obj)
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
	for _, run := range runs {
		if run.ID != id {
			continue
		}
		// Summary 里优先取 stderrFile
		if s, ok := any(run.Summary).(string); ok && s != "" {
			var m map[string]any
			if json.Unmarshal([]byte(s), &m) == nil {
				if p, ok := m["stderrFile"].(string); ok && p != "" {
					logPath = p
				}
			}
		}
		if logPath == "" {
			if m, ok := any(run.Summary).(map[string]any); ok {
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
	rows := make([]Row, 0, 200)
	re := regexp.MustCompile(`(?:(\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2})\s+)?(INFO|NOTICE|ERROR)\s*:\s*(.+?):\s*(.+)$`)
	tsRe := regexp.MustCompile(`\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2}\s+(?:INFO|NOTICE|ERROR)\s*:`)
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
			at := strings.TrimSpace(m[1])
			level := strings.ToUpper(strings.TrimSpace(m[2]))
			path := strings.TrimSpace(m[3])
			msg := strings.TrimSpace(m[4])
			row := Row{Name: path, At: at, Size: 0, Message: msg}
			low := strings.ToLower(msg)
			if level == "ERROR" {
				row.Status = "failed"
				row.Action = "Error"
			} else {
				switch {
				case strings.Contains(low, "copied"):
					row.Status = "success"
					row.Action = "Copied"
				case strings.Contains(low, "deleted") || strings.Contains(low, "removed"):
					row.Status = "success"
					row.Action = "Deleted"
				case strings.Contains(low, "skipped"):
					row.Status = "skipped"
					row.Action = "Skipped"
				case strings.Contains(low, "renamed"):
					row.Status = "success"
					row.Action = "Renamed"
				default:
					continue
				}
			}
			rows = append(rows, row)
		}
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
		preflightTotal := int64(0)
		if summary != nil {
			if pf, ok := summary["preflight"].(map[string]any); ok {
				if v, ok2 := pf["totalBytes"].(float64); ok2 {
					preflightTotal = int64(v)
				}
			}
		}
		progressTotal := int64(0)
		if v, ok := progress["totalBytes"].(float64); ok {
			progressTotal = int64(v)
		}
		total := progressTotal
		usePreflightTotal := preflightTotal > 0 && (total == 0 || total < preflightTotal)
		if usePreflightTotal {
			total = preflightTotal
		}
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
		if !usePreflightTotal {
			if v, ok := progress["percentage"].(float64); ok {
				pct = v
			} else if total > 0 {
				pct = float64(bytes) / float64(total) * 100
			}
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
		preflightTotalCount := float64(0)
		if summary != nil {
			if pf, ok := summary["preflight"].(map[string]any); ok {
				if v, ok2 := pf["totalCount"].(float64); ok2 {
					preflightTotalCount = v
				}
			}
		}
		totalCount := float64(0)
		if v, ok := progress["plannedFiles"].(float64); ok {
			totalCount = v
		}
		if preflightTotalCount > 0 && (totalCount == 0 || totalCount < preflightTotalCount) {
			totalCount = preflightTotalCount
		}
		if totalCount > 0 && completedFiles > totalCount {
			completedFiles = totalCount
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
			"bytes":          bytes,
			"totalBytes":     total,
			"speed":          speed,
			"eta":            eta,
			"percentage":     pct,
			"phase":          phase,
			"lastUpdatedAt":  time.Now().Format(time.RFC3339),
			"completedFiles": completedFiles,
			"totalCount":     totalCount,
		}

		calcPct := 0.0
		if total > 0 {
			calcPct = float64(bytes) / float64(total) * 100
		}
		pctMismatch := total > 0 && math.Abs(calcPct-pct) > 1.5
		countMismatch := totalCount > 0 && completedFiles > totalCount
		etaMismatch := eta > 0 && speed > 0 && total > bytes && math.Abs((float64(total-bytes)/float64(speed))-eta) > 300
		progressMismatch := pctMismatch || countMismatch || etaMismatch
		progressCheck := map[string]any{
			"ok":            !progressMismatch,
			"pctMismatch":   pctMismatch,
			"countMismatch": countMismatch,
			"etaMismatch":   etaMismatch,
			"calcPct":       calcPct,
		}

		// 让 activeRuns/stableProgress 直接对齐当前日志解析结果
		c.runSvc.UpdateRunStatus(run.ID, map[string]any{"stableProgress": stable})

		item := map[string]any{
			"runRecord": map[string]any{
				"id":               run.ID,
				"taskId":           run.TaskID,
				"status":           run.Status,
				"rcJobId":          0,
				"bytesTransferred": run.BytesTransferred,
				"error":            run.Error,
				"startedAt":        run.StartedAt,
				"finishedAt":       run.FinishedAt,
			},
			"progress":           stable,
			"stableProgress":     stable,
			"progressLine":       progressLine,
			"progressSource":     "summary.progress",
			"progressMismatch":   progressMismatch,
			"progressCheck":      progressCheck,
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

// HandleJobStatus 处理获取指定 Job 的状态
func (c *RunController) HandleJobStatus(w http.ResponseWriter, r *http.Request) {
	jobIdStr := r.PathValue("jobId")
	jobId, err := strconv.ParseInt(jobIdStr, 10, 64)
	if err != nil {
		WriteJSON(w, 400, map[string]any{"error": "invalid job id"})
		return
	}

	status, err := c.rc.JobStatus(r.Context(), jobId)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	WriteJSON(w, 200, status)
}

// HandleJobStop 处理停止指定的 Job
func (c *RunController) HandleJobStop(w http.ResponseWriter, r *http.Request) {
	jobIdStr := r.PathValue("jobId")
	jobId, err := strconv.ParseInt(jobIdStr, 10, 64)
	if err != nil {
		WriteJSON(w, 400, map[string]any{"error": "invalid job id"})
		return
	}

	if err := c.rc.JobStop(r.Context(), jobId); err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}

	// 更新数据库中该任务的状态为 stopped
	c.runSvc.UpdateRunStatusByJobId(jobId, "stopped", "用户手动停止")

	WriteJSON(w, 200, map[string]any{"success": true})
}
