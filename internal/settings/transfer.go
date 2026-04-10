package settings

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type TransferSettings struct {
	PostVerifyEnabled   bool   `json:"postVerifyEnabled"`
	PostVerifyMode      string `json:"postVerifyMode"`      // mount | remote (先用 mount)
	PostVerifyInterval  string `json:"postVerifyInterval"`  // e.g. "5s"
	PostVerifyTimeout   string `json:"postVerifyTimeout"`   // e.g. "30m"
	PostVerifyMtimeGrace string `json:"postVerifyMtimeGrace"` // e.g. "60s"
	PostVerifyMatch     string `json:"postVerifyMatch"`     // size
	MinRerunInterval    string `json:"minRerunInterval"`    // e.g. "30m"
}

func defaults() TransferSettings {
	return TransferSettings{
		PostVerifyEnabled: true,
		PostVerifyMode: "mount",
		PostVerifyInterval: "5s",
		PostVerifyTimeout: "30m",
		PostVerifyMtimeGrace: "60s",
		PostVerifyMatch: "size",
		MinRerunInterval: "30m",
	}
}

func dataDir() string {
	d := os.Getenv("APP_DATA_DIR")
	if d == "" { d = "./data" }
	_ = os.MkdirAll(d, 0o755)
	return d
}

func path() string { return filepath.Join(dataDir(), "transfer_settings.json") }

func Load() (TransferSettings, error) {
	p := path()
	b, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) { return defaults(), nil }
		return defaults(), err
	}
	var ts TransferSettings
	if err := json.Unmarshal(b, &ts); err != nil { return defaults(), err }
	// 兜底默认
	def := defaults()
	if ts.PostVerifyMode == "" { ts.PostVerifyMode = def.PostVerifyMode }
	if ts.PostVerifyInterval == "" { ts.PostVerifyInterval = def.PostVerifyInterval }
	if ts.PostVerifyTimeout == "" { ts.PostVerifyTimeout = def.PostVerifyTimeout }
	if ts.PostVerifyMtimeGrace == "" { ts.PostVerifyMtimeGrace = def.PostVerifyMtimeGrace }
	if ts.PostVerifyMatch == "" { ts.PostVerifyMatch = def.PostVerifyMatch }
	if ts.MinRerunInterval == "" { ts.MinRerunInterval = def.MinRerunInterval }
	return ts, nil
}

func Save(ts TransferSettings) error {
	b, err := json.MarshalIndent(ts, "", "  ")
	if err != nil { return err }
	return os.WriteFile(path(), b, 0o644)
}
