package controller

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	stdiostrconv "strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"rcloneflow/internal/logger"
	"rcloneflow/internal/service"
)

func isSafeLogPath(p string) bool {
	clean := filepath.Clean(p)
	if strings.Contains(clean, "..") {
		return false
	}
	abs, err := filepath.Abs(clean)
	if err != nil {
		return false
	}
	_ = abs
	return true
}

func (c *RunController) resolveLogPath(run service.RunRecord) (string, bool) {
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

var findProcess = os.FindProcess

func killRunBySummary(run service.RunRecord) bool {
	var pid int
	var sum map[string]any
	switch v := any(run.Summary).(type) {
	case map[string]any:
		sum = v
	case string:
		if v != "" {
			if err := json.Unmarshal([]byte(v), &sum); err != nil {
				logger.Error("unmarshal run summary for kill", zap.Error(err))
			}
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
		if p, err := findProcess(pid); err == nil {
			_ = p.Kill()
			time.Sleep(2 * time.Second)
			if p2, err := findProcess(pid); err == nil {
				_ = p2.Kill()
				time.Sleep(2 * time.Second)
				if p3, err := findProcess(pid); err == nil {
					_ = p3.Kill()
				}
			}
		}
		return true
	}
	return false
}
