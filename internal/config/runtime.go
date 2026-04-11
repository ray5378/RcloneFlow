package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// settings.json overrides (used by /api/settings)
func overridesPath() string {
	dir := os.Getenv("APP_DATA_DIR")
	if dir == "" { dir = "." }
	return filepath.Join(dir, "settings.json")
}

func readOverrides() map[string]string {
	b, err := os.ReadFile(overridesPath())
	if err != nil || len(b) == 0 { return map[string]string{} }
	var m map[string]string
	if json.Unmarshal(b, &m) != nil || m == nil { return map[string]string{} }
	return m
}

// GetPrecheckMode returns preflight mode: env > settings.json > default("none")
func GetPrecheckMode() string {
	if v := os.Getenv("PRECHECK_MODE"); v != "" { return v }
	if v := readOverrides()["PRECHECK_MODE"]; v != "" { return v }
	return "none"
}

func GetProgressFlushInterval() time.Duration {
	if v := os.Getenv("PROGRESS_FLUSH_INTERVAL"); v != "" { if d,err := time.ParseDuration(v); err == nil { return d } }
	if v := readOverrides()["PROGRESS_FLUSH_INTERVAL"]; v != "" { if d,err := time.ParseDuration(v); err == nil { return d } }
	return 5 * time.Second
}

func GetProgressFlushDeltaPct() float64 {
	if v := os.Getenv("PROGRESS_FLUSH_MIN_DELTA_PCT"); v != "" { if n,err := strconv.ParseFloat(v,64); err == nil { return n } }
	if v := readOverrides()["PROGRESS_FLUSH_MIN_DELTA_PCT"]; v != "" { if n,err := strconv.ParseFloat(v,64); err == nil { return n } }
	return 1
}

func GetProgressFlushDeltaBytes() int64 {
	if v := os.Getenv("PROGRESS_FLUSH_MIN_DELTA_BYTES"); v != "" { if n,err := strconv.ParseInt(v,10,64); err == nil { return n } }
	if v := readOverrides()["PROGRESS_FLUSH_MIN_DELTA_BYTES"]; v != "" { if n,err := strconv.ParseInt(v,10,64); err == nil { return n } }
	return 50 * 1024 * 1024
}

func GetFinishWaitInterval() time.Duration {
	if v := os.Getenv("FINISH_WAIT_INTERVAL"); v != "" { if d,err := time.ParseDuration(v); err == nil { return d } }
	if v := readOverrides()["FINISH_WAIT_INTERVAL"]; v != "" { if d,err := time.ParseDuration(v); err == nil { return d } }
	return 5 * time.Second
}

func GetFinishWaitTimeout() time.Duration {
	if v := os.Getenv("FINISH_WAIT_TIMEOUT"); v != "" { if d,err := time.ParseDuration(v); err == nil { return d } }
	if v := readOverrides()["FINISH_WAIT_TIMEOUT"]; v != "" { if d,err := time.ParseDuration(v); err == nil { return d } }
	return 5 * time.Hour
}
