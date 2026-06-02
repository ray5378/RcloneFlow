package controller

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"rcloneflow/internal/logutil"
	"sort"
	"strconv"
	stdiostrconv "strconv"
	"strings"
	"time"
)

func (c *RunController) HandleRunFiles(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/runs/"), "/files")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	offset := 0
	limit := 50
	filter := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("filter")))
	if filter == "" {
		filter = "all"
	}
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
	var runSummary map[string]any
	for _, run := range runs {
		if run.ID != id {
			continue
		}
		moveMode = strings.EqualFold(run.TaskMode, "move")
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
		if run.Summary != "" {
			var parsed map[string]any
			if json.Unmarshal([]byte(run.Summary), &parsed) == nil {
				runSummary = parsed
			}
		}

		if logPath == "" {
			base := "/app/data/logs"
			if run.StartedAt != "" {
				if t, e := time.Parse(time.RFC3339, run.StartedAt); e == nil {
					sub := t.Local().Format("0102")
					entries, _ := os.ReadDir(base)
					for _, ent := range entries {
						name := ent.Name()
						if ent.IsDir() && strings.HasSuffix(name, "-"+sub) {
							files, _ := os.ReadDir(filepath.Join(base, name))
							var best string
							var bestDiff int64 = 1 << 62
							for _, f := range files {
								if f.IsDir() || !strings.HasSuffix(f.Name(), ".log") {
									continue
								}
								fn := strings.TrimSuffix(f.Name(), ".log")
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
	if !isSafeLogPath(logPath) {
		WriteJSON(w, 403, map[string]any{"error": "access denied"})
		return
	}
	if strings.EqualFold(r.URL.Query().Get("mode"), "raw") {
		f, e := os.Open(logPath)
		if e != nil {
			WriteJSON(w, 500, map[string]any{"error": "failed to read log"})
			return
		}
		defer f.Close()
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.ServeFile(w, r, logPath)
		return
	}
	data, readErr := os.ReadFile(logPath)
	if readErr != nil {
		WriteJSON(w, 500, map[string]any{"error": "failed to read log"})
		return
	}
	lines := strings.Split(string(data), "\n")
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
		for _, seg := range logutil.SplitLogSegments(ln) {
			at, level, path, msg, ok := logutil.ParseLogSegment(seg)
			if !ok {
				continue
			}
			histRow, _, ok := classifyHistoricalLogRow(level, path, msg)
			if !ok {
				continue
			}
			rowMap := map[string]any{
				"path":      path,
				"name":      path,
				"at":        at,
				"sizeBytes": historicalSegmentSizeBytes(seg),
				"message":   msg,
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
		enrichRowMapSizesFromFinalSummary(rowMaps, runSummary)
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
	if filter != "all" {
		filtered := make([]Row, 0, len(rows))
		for _, row := range rows {
			kind := strings.ToLower(strings.TrimSpace(row.Status))
			switch filter {
			case "success":
				if kind == "success" {
					filtered = append(filtered, row)
				}
			case "failed":
				if kind == "failed" {
					filtered = append(filtered, row)
				}
			case "other":
				if kind == "skipped" {
					filtered = append(filtered, row)
				}
			default:
				filtered = append(filtered, row)
			}
		}
		rows = filtered
	}
	total := len(rows)
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = total
	}
	page := rows[offset:end]
	h := sha1.Sum(data)
	info := map[string]any{"logPath": logPath, "logSize": len(data), "logSha1": fmt.Sprintf("%x", h[:])}
	WriteJSON(w, 200, map[string]any{"total": total, "items": page, "info": info})
}
