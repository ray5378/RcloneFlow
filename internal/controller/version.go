package controller

import (
	"net/http"

	"rcloneflow/internal/adapter"
	"rcloneflow/internal/version"
)

type VersionController struct {
	rc *adapter.RcloneClient
}

func NewVersionController(rc *adapter.RcloneClient) *VersionController {
	return &VersionController{rc: rc}
}

// HandleVersion 返回项目版本和 rclone 版本
func (c *VersionController) HandleVersion(w http.ResponseWriter, r *http.Request) {
	rcloneVersion := ""
	if versionResp, err := c.rc.Version(r.Context()); err == nil && versionResp != nil {
		rcloneVersion = versionResp.Version
	}
	WriteJSON(w, 200, map[string]any{
		"commitHash":    version.CommitHash,
		"rcloneVersion": rcloneVersion,
	})
}
