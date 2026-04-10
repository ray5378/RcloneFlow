package controller

import (
	"encoding/json"
	"net/http"

	"rcloneflow/internal/settings"
)

type SettingsController struct {}

func NewSettingsController() *SettingsController { return &SettingsController{} }

func (c *SettingsController) HandleTransfer(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		ts, err := settings.Load()
		if err != nil { WriteJSON(w, 500, map[string]any{"error": err.Error()}); return }
		WriteJSON(w, 200, ts); return
	}
	if r.Method == http.MethodPut {
		var ts settings.TransferSettings
		if err := json.NewDecoder(r.Body).Decode(&ts); err != nil { WriteJSON(w, 400, map[string]any{"error":"invalid body"}); return }
		if err := settings.Save(ts); err != nil { WriteJSON(w, 500, map[string]any{"error": err.Error()}); return }
		WriteJSON(w, 200, map[string]any{"ok": true}); return
	}
	w.WriteHeader(405)
}
