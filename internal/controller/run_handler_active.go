package controller

import (
	"encoding/json"
	"math"
	"net/http"
	"time"

	"rcloneflow/internal/util"
)

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
		if logicalTotalCount <= 0 && summary != nil {
			if pf, ok := summary["preflight"].(map[string]any); ok {
				if v, ok2 := pf["totalCount"].(float64); ok2 {
					logicalTotalCount = v
				}
			}
		}
		if logicalTotalCount <= 0 && summary != nil && casCompatible {
			if at, ok := summary["activeTransfer"].(map[string]any); ok {
				if v, ok2 := at["totalCount"].(float64); ok2 {
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
			"logicalTotalCount": logicalTotalCount,
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
				rr["durationText"] = util.HumanDuration(dur)
			}
			it["runRecord"] = rr
		}
		items[i] = it
	}
	WriteJSON(w, 200, items)
}

func (c *RunController) HandleGlobalStats(w http.ResponseWriter, r *http.Request) {
	runs, e2 := c.runSvc.ListActiveRuns()
	if e2 != nil {
		WriteJSON(w, 500, map[string]any{"error": e2.Error()})
		return
	}
	var bytesSum, totalSum, speedSum float64
	for _, run := range runs {
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
		"speedAvg":   speedSum,
		"eta":        nil,
		"percentage": percentage,
	})
}
