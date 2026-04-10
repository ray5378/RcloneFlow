package controller

import (
	"context"
	"net/http"
	"strings"

	"rcloneflow/internal/adapter"
)

// helpers
func buildPath(fs, remote string) string { return fs + ":" + strings.TrimPrefix(remote, "/") }

// HandleMkdir 创建目录（CLI）
func (c *FsController) HandleMkdir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { w.WriteHeader(405); return }
	var req struct{ Fs, Remote string }
	if err := DecodeRequest(r, &req); err != nil { WriteJSON(w, 400, map[string]any{"error": err.Error()}); return }
	runner := &adapter.CmdRunner{}
	_, stderr, err := runner.Run(context.Background(), "mkdir", buildPath(req.Fs, req.Remote))
	if err != nil { WriteJSON(w, 500, map[string]any{"error": stderr}); return }
	WriteJSON(w, 200, map[string]any{"ok": true})
}

func (c *FsController) HandleDeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { w.WriteHeader(405); return }
	var req struct{ Fs, Remote string }
	if err := DecodeRequest(r, &req); err != nil { WriteJSON(w, 400, map[string]any{"error": err.Error()}); return }
	runner := &adapter.CmdRunner{}
	_, stderr, err := runner.Run(context.Background(), "delete", buildPath(req.Fs, req.Remote))
	if err != nil { WriteJSON(w, 500, map[string]any{"error": stderr}); return }
	WriteJSON(w, 200, map[string]any{"ok": true})
}

func (c *FsController) HandlePurge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { w.WriteHeader(405); return }
	var req struct{ Fs, Remote string }
	if err := DecodeRequest(r, &req); err != nil { WriteJSON(w, 400, map[string]any{"error": err.Error()}); return }
	runner := &adapter.CmdRunner{}
	_, stderr, err := runner.Run(context.Background(), "purge", buildPath(req.Fs, req.Remote))
	if err != nil { WriteJSON(w, 500, map[string]any{"error": stderr}); return }
	WriteJSON(w, 200, map[string]any{"ok": true})
}

func (c *FsController) HandleMove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { w.WriteHeader(405); return }
	var req struct{ SrcFs, SrcRemote, DstFs, DstRemote string }
	if err := DecodeRequest(r, &req); err != nil { WriteJSON(w, 400, map[string]any{"error": err.Error()}); return }
	runner := &adapter.CmdRunner{}
	_, stderr, err := runner.Run(context.Background(), "move", buildPath(req.SrcFs, req.SrcRemote), buildPath(req.DstFs, req.DstRemote))
	if err != nil { WriteJSON(w, 500, map[string]any{"error": stderr}); return }
	WriteJSON(w, 200, map[string]any{"ok": true})
}

func (c *FsController) HandleCopy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { w.WriteHeader(405); return }
	var req struct{ SrcFs, SrcRemote, DstFs, DstRemote string }
	if err := DecodeRequest(r, &req); err != nil { WriteJSON(w, 400, map[string]any{"error": err.Error()}); return }
	runner := &adapter.CmdRunner{}
	_, stderr, err := runner.Run(context.Background(), "copy", buildPath(req.SrcFs, req.SrcRemote), buildPath(req.DstFs, req.DstRemote))
	if err != nil { WriteJSON(w, 500, map[string]any{"error": stderr}); return }
	WriteJSON(w, 200, map[string]any{"ok": true})
}

func (c *FsController) HandleCopyDir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { w.WriteHeader(405); return }
	var req struct{ SrcFs, DstFs string }
	if err := DecodeRequest(r, &req); err != nil { WriteJSON(w, 400, map[string]any{"error": err.Error()}); return }
	runner := &adapter.CmdRunner{}
	_, stderr, err := runner.Run(context.Background(), "copy", req.SrcFs+":", req.DstFs+":")
	if err != nil { WriteJSON(w, 500, map[string]any{"error": stderr}); return }
	WriteJSON(w, 200, map[string]any{"ok": true})
}

func (c *FsController) HandleMoveDir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { w.WriteHeader(405); return }
	var req struct{ SrcFs, DstFs string }
	if err := DecodeRequest(r, &req); err != nil { WriteJSON(w, 400, map[string]any{"error": err.Error()}); return }
	runner := &adapter.CmdRunner{}
	_, stderr, err := runner.Run(context.Background(), "move", req.SrcFs+":", req.DstFs+":")
	if err != nil { WriteJSON(w, 500, map[string]any{"error": stderr}); return }
	WriteJSON(w, 200, map[string]any{"ok": true})
}

func (c *FsController) HandlePublicLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { w.WriteHeader(405); return }
	var req struct{ Fs, Remote string }
	if err := DecodeRequest(r, &req); err != nil { WriteJSON(w, 400, map[string]any{"error": err.Error()}); return }
	runner := &adapter.CmdRunner{}
	stdout, stderr, err := runner.Run(context.Background(), "link", buildPath(req.Fs, req.Remote))
	if err != nil { WriteJSON(w, 500, map[string]any{"error": stderr}); return }
	WriteJSON(w, 200, map[string]any{"url": strings.TrimSpace(stdout)})
}
