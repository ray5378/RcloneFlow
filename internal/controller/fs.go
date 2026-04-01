package controller

import (
	"net/http"

	"rcloneflow/internal/rclone"
)

// FsController 文件系统操作控制器
type FsController struct {
	rc *rclone.Client
}

// NewFsController 创建文件系统控制器
func NewFsController(rc *rclone.Client) *FsController {
	return &FsController{rc: rc}
}

// HandleMkdir 创建目录
func (c *FsController) HandleMkdir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	var req struct {
		Fs     string `json:"fs"`
		Remote string `json:"remote"`
	}
	if err := DecodeRequest(r, &req); err != nil {
		WriteJSON(w, 400, map[string]any{"error": err.Error()})
		return
	}
	if err := c.rc.Mkdir(r.Context(), req.Fs, req.Remote); err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	WriteJSON(w, 200, map[string]any{"ok": true})
}

// HandleDeleteFile 删除文件
func (c *FsController) HandleDeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	var req struct {
		Fs     string `json:"fs"`
		Remote string `json:"remote"`
	}
	if err := DecodeRequest(r, &req); err != nil {
		WriteJSON(w, 400, map[string]any{"error": err.Error()})
		return
	}
	if err := c.rc.DeleteFile(r.Context(), req.Fs, req.Remote); err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	WriteJSON(w, 200, map[string]any{"ok": true})
}

// HandlePurge 删除目录
func (c *FsController) HandlePurge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	var req struct {
		Fs     string `json:"fs"`
		Remote string `json:"remote"`
	}
	if err := DecodeRequest(r, &req); err != nil {
		WriteJSON(w, 400, map[string]any{"error": err.Error()})
		return
	}
	if err := c.rc.Purge(r.Context(), req.Fs, req.Remote); err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	WriteJSON(w, 200, map[string]any{"ok": true})
}

// HandleMove 移动/重命名文件
func (c *FsController) HandleMove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	var req struct {
		SrcFs     string `json:"srcFs"`
		SrcRemote string `json:"srcRemote"`
		DstFs     string `json:"dstFs"`
		DstRemote string `json:"dstRemote"`
	}
	if err := DecodeRequest(r, &req); err != nil {
		WriteJSON(w, 400, map[string]any{"error": err.Error()})
		return
	}
	if err := c.rc.MoveFile(r.Context(), req.SrcFs, req.SrcRemote, req.DstFs, req.DstRemote); err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	WriteJSON(w, 200, map[string]any{"ok": true})
}

// HandleCopy 复制文件
func (c *FsController) HandleCopy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	var req struct {
		SrcFs     string `json:"srcFs"`
		SrcRemote string `json:"srcRemote"`
		DstFs     string `json:"dstFs"`
		DstRemote string `json:"dstRemote"`
	}
	if err := DecodeRequest(r, &req); err != nil {
		WriteJSON(w, 400, map[string]any{"error": err.Error()})
		return
	}
	if err := c.rc.CopyFile(r.Context(), req.SrcFs, req.SrcRemote, req.DstFs, req.DstRemote); err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	WriteJSON(w, 200, map[string]any{"ok": true})
}

// HandleCopyDir 复制目录
func (c *FsController) HandleCopyDir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	var req struct {
		SrcFs string `json:"srcFs"`
		DstFs string `json:"dstFs"`
	}
	if err := DecodeRequest(r, &req); err != nil {
		WriteJSON(w, 400, map[string]any{"error": err.Error()})
		return
	}
	if err := c.rc.CopyDir(r.Context(), req.SrcFs, req.DstFs); err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	WriteJSON(w, 200, map[string]any{"ok": true})
}

// HandleMoveDir 移动目录
func (c *FsController) HandleMoveDir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	var req struct {
		SrcFs string `json:"srcFs"`
		DstFs string `json:"dstFs"`
	}
	if err := DecodeRequest(r, &req); err != nil {
		WriteJSON(w, 400, map[string]any{"error": err.Error()})
		return
	}
	if err := c.rc.MoveDir(r.Context(), req.SrcFs, req.DstFs); err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	WriteJSON(w, 200, map[string]any{"ok": true})
}

// HandlePublicLink 生成分享链接
func (c *FsController) HandlePublicLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	var req struct {
		Fs     string `json:"fs"`
		Remote string `json:"remote"`
	}
	if err := DecodeRequest(r, &req); err != nil {
		WriteJSON(w, 400, map[string]any{"error": err.Error()})
		return
	}
	url, err := c.rc.PublicLink(r.Context(), req.Fs, req.Remote)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	WriteJSON(w, 200, map[string]any{"url": url})
}
