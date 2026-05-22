package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"rcloneflow/internal/adapter"
)

func (c *TaskController) HandleExportTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(405)
		return
	}
	data, err := c.taskSvc.ExportTasks()
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	if cfg, err := c.rc.DumpConfig(r.Context()); err == nil {
		data["rcloneConfig"] = cfg
	}
	WriteJSON(w, 200, data)
}

func (c *TaskController) HandleImportTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	var req struct {
		Tasks            []any          `json:"tasks"`
		Schedules        []any          `json:"schedules"`
		ConflictStrategy string         `json:"conflictStrategy"`
		RawData          map[string]any `json:"-"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req.RawData); err != nil {
		WriteJSON(w, 400, map[string]any{"error": "无效的请求体"})
		return
	}
	strategy, _ := req.RawData["conflictStrategy"].(string)
	if strategy != "skip" && strategy != "overwrite" {
		strategy = "skip"
	}
	imported, skipped, overwritten, err := c.taskSvc.ImportTasks(req.RawData, strategy)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}

	remotesAdded := 0
	remotesSkipped := 0
	remotesOverwritten := 0
	var remoteErrors []string
	if rcloneCfg, ok := req.RawData["rcloneConfig"].(map[string]any); ok {
		existingRemotes, err := c.rc.ListRemotes(r.Context())
		if err != nil {
			existingRemotes = []string{}
		}
		existingSet := make(map[string]bool)
		for _, name := range existingRemotes {
			existingSet[strings.ToLower(name)] = true
		}
		remoteStrategy, _ := req.RawData["remoteConflictStrategy"].(string)
		if remoteStrategy != "skip" && remoteStrategy != "overwrite" {
			remoteStrategy = "skip"
		}
		for name, v := range rcloneCfg {
			paramMap, ok := v.(map[string]any)
			if !ok {
				remotesSkipped++
				continue
			}
			typ, _ := paramMap["type"].(string)
			if typ == "" {
				remotesSkipped++
				continue
			}
			params := make(map[string]any)
			for k, val := range paramMap {
				if k != "type" {
					params[k] = val
				}
			}
			if existingSet[strings.ToLower(name)] {
				if remoteStrategy == "overwrite" {
					if err := c.rc.DeleteRemote(r.Context(), name); err != nil {
						remoteErrors = append(remoteErrors, fmt.Sprintf("%s(%s): delete failed: %s", name, typ, err.Error()))
						remotesSkipped++
						continue
					}
					if err := c.rc.CreateRemote(r.Context(), &adapter.CreateRemoteRequest{
						Name:       name,
						Type:       typ,
						Parameters: params,
					}); err != nil {
						remoteErrors = append(remoteErrors, fmt.Sprintf("%s(%s): %s", name, typ, err.Error()))
						remotesSkipped++
					} else {
						remotesOverwritten++
					}
				} else {
					remotesSkipped++
				}
				continue
			}
			if err := c.rc.CreateRemote(r.Context(), &adapter.CreateRemoteRequest{
				Name:       name,
				Type:       typ,
				Parameters: params,
			}); err != nil {
				remoteErrors = append(remoteErrors, fmt.Sprintf("%s(%s): %s", name, typ, err.Error()))
				remotesSkipped++
			} else {
				remotesAdded++
			}
		}
	}

	resp := map[string]any{
		"imported":           imported,
		"skipped":            skipped,
		"overwritten":        overwritten,
		"remotesAdded":       remotesAdded,
		"remotesSkipped":     remotesSkipped,
		"remotesOverwritten": remotesOverwritten,
	}
	if len(remoteErrors) > 0 {
		resp["remoteErrors"] = remoteErrors
	}
	WriteJSON(w, 200, resp)
}