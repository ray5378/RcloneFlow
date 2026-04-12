package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"rcloneflow/internal/logger"
)

type SettingsController struct{}

func NewSettingsController() *SettingsController { return &SettingsController{} }

// model
type settingsPayload struct {
	Values map[string]string `json:"values"`
	Reset  bool              `json:"reset"`
}

type settingsResponse struct {
	Auth     map[string]map[string]string `json:"auth"`
	Log      map[string]map[string]string `json:"log"`
	History  map[string]map[string]string `json:"history"`
	Precheck map[string]map[string]string `json:"precheck"`
	Progress map[string]map[string]string `json:"progress"`
	Webdav   map[string]map[string]string `json:"webdav"`
	Stored   map[string]string            `json:"stored"` // raw overrides from settings.json
	Meta     map[string]string            `json:"meta"`   // APP_DATA_DIR, settingsPath
}

func defaultsMap() map[string]string {
	return map[string]string{
		"ACCESS_TOKEN_TTL":               "24h",
		"REFRESH_TOKEN_TTL":              "90d",
		"LOG_LEVEL":                      "info",
		"LOG_OUTPUT":                     "stdout",
		"FINAL_SUMMARY_RETENTION_DAYS":   "7",
		"CLEANUP_INTERVAL_HOURS":         "24",
		"PRECHECK_MODE":                  "none",
		"PROGRESS_FLUSH_INTERVAL":        "5s",
		"PROGRESS_FLUSH_MIN_DELTA_PCT":   "1",
		"PROGRESS_FLUSH_MIN_DELTA_BYTES": "52428800",
		"FINISH_WAIT_INTERVAL":           "5s",
		"FINISH_WAIT_TIMEOUT":            "5h",
	}
}

func settingsPath() string {
	dataDir := os.Getenv("APP_DATA_DIR")
	if dataDir == "" {
		dataDir = "."
	}
	_ = os.MkdirAll(dataDir, 0755)
	return filepath.Join(dataDir, "settings.json")
}

func readOverrides() map[string]string {
	fp := settingsPath()
	b, err := os.ReadFile(fp)
	if err != nil || len(b) == 0 {
		return map[string]string{}
	}
	var m map[string]string
	if json.Unmarshal(b, &m) != nil || m == nil {
		return map[string]string{}
	}
	return m
}

func writeOverrides(m map[string]string) error {
	fp := settingsPath()
	b, _ := json.MarshalIndent(m, "", "  ")
	return os.WriteFile(fp, b, 0644)
}

func effectiveValue(key string, overrides map[string]string, defs map[string]string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	if v := overrides[key]; v != "" {
		return v
	}
	return defs[key]
}

// HandleSettings handles GET/PUT /api/settings
func (s *SettingsController) HandleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.handleGet(w, r)
		return
	}
	if r.Method == http.MethodPut {
		s.handlePut(w, r)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (s *SettingsController) handleGet(w http.ResponseWriter, r *http.Request) {
	defs := defaultsMap()
	over := readOverrides()
	eff := func(k string) string { return effectiveValue(k, over, defs) }

	resp := settingsResponse{
		Auth: map[string]map[string]string{
			"ACCESS_TOKEN_TTL":  {"effective": eff("ACCESS_TOKEN_TTL"), "default": defs["ACCESS_TOKEN_TTL"]},
			"REFRESH_TOKEN_TTL": {"effective": eff("REFRESH_TOKEN_TTL"), "default": defs["REFRESH_TOKEN_TTL"]},
		},
		Log: map[string]map[string]string{
			"LOG_LEVEL":  {"effective": eff("LOG_LEVEL"), "default": defs["LOG_LEVEL"]},
			"LOG_OUTPUT": {"effective": eff("LOG_OUTPUT"), "default": defs["LOG_OUTPUT"]},
		},
		History: map[string]map[string]string{
			"FINAL_SUMMARY_RETENTION_DAYS": {"effective": eff("FINAL_SUMMARY_RETENTION_DAYS"), "default": defs["FINAL_SUMMARY_RETENTION_DAYS"]},
			"CLEANUP_INTERVAL_HOURS":       {"effective": eff("CLEANUP_INTERVAL_HOURS"), "default": defs["CLEANUP_INTERVAL_HOURS"]},
		},
		Precheck: map[string]map[string]string{
			"PRECHECK_MODE": {"effective": eff("PRECHECK_MODE"), "default": defs["PRECHECK_MODE"]},
		},
		Progress: map[string]map[string]string{
			"PROGRESS_FLUSH_INTERVAL":        {"effective": eff("PROGRESS_FLUSH_INTERVAL"), "default": defs["PROGRESS_FLUSH_INTERVAL"]},
			"PROGRESS_FLUSH_MIN_DELTA_PCT":   {"effective": eff("PROGRESS_FLUSH_MIN_DELTA_PCT"), "default": defs["PROGRESS_FLUSH_MIN_DELTA_PCT"]},
			"PROGRESS_FLUSH_MIN_DELTA_BYTES": {"effective": eff("PROGRESS_FLUSH_MIN_DELTA_BYTES"), "default": defs["PROGRESS_FLUSH_MIN_DELTA_BYTES"]},
		},
		Webdav: map[string]map[string]string{
			"FINISH_WAIT_INTERVAL": {"effective": eff("FINISH_WAIT_INTERVAL"), "default": defs["FINISH_WAIT_INTERVAL"]},
			"FINISH_WAIT_TIMEOUT":  {"effective": eff("FINISH_WAIT_TIMEOUT"), "default": defs["FINISH_WAIT_TIMEOUT"]},
		},
		Stored: over,
		Meta:   map[string]string{"APP_DATA_DIR": os.Getenv("APP_DATA_DIR"), "settingsPath": settingsPath()},
	}
	WriteJSON(w, http.StatusOK, resp)
}

func (s *SettingsController) handlePut(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var p settingsPayload
	_ = json.Unmarshal(body, &p)
	if p.Reset {
		_ = writeOverrides(map[string]string{})
		WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "reset": true})
		return
	}
	// sanitize: only accept known keys
	defs := defaultsMap()
	cur := readOverrides()
	for k, v := range p.Values {
		k = strings.TrimSpace(k)
		if _, ok := defs[k]; !ok {
			continue
		}
		cur[k] = strings.TrimSpace(v)
	}
	if err := writeOverrides(cur); err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	// 热生效：日志与清理
	if _, ok := cur["LOG_LEVEL"]; ok || cur["LOG_OUTPUT"] != "" {
		lev := cur["LOG_LEVEL"]
		if lev == "" {
			lev = defaultsMap()["LOG_LEVEL"]
		}
		out := cur["LOG_OUTPUT"]
		if out == "" {
			out = defaultsMap()["LOG_OUTPUT"]
		}
		_ = logger.HotSet(lev, out)
	}
	// 清理计划重排：通过一个可选的回调（由 app 层注入）
	if ReplanCleanupHook != nil {
		intervalHours := atoiDefault(cur["CLEANUP_INTERVAL_HOURS"], atoiDefault(defaultsMap()["CLEANUP_INTERVAL_HOURS"], 24))
		retentionDays := atoiDefault(cur["FINAL_SUMMARY_RETENTION_DAYS"], atoiDefault(defaultsMap()["FINAL_SUMMARY_RETENTION_DAYS"], 7))
		ReplanCleanupHook(intervalHours, retentionDays)
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

var ReplanCleanupHook func(intervalHours int, retentionDays int)

func atoiDefault(s string, d int) int {
	if s == "" {
		return d
	}
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return d
}
