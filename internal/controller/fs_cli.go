package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strings"
	"time"

	"rcloneflow/internal/config"
)

// FsController 文件系统操作控制器（CLI 实现）
// 浏览仍走 RC（在 BrowserController）；本控制器仅负责 /api/fs/* 变更类操作。

type FsController struct{}

func NewFsController(_ any) *FsController { return &FsController{} }

func sanitizePath(p string) string {
	p = strings.TrimSpace(p)
	for strings.HasPrefix(p, "-") {
		p = strings.TrimPrefix(p, "-")
	}
	p = filepath.ToSlash(p)
	parts := strings.Split(p, "/")
	for i, s := range parts {
		for strings.HasSuffix(s, ":") { s = strings.TrimSuffix(s, ":") }
		parts[i] = s
	}
	return strings.Join(parts, "/")
}

func sanitizeFsRemote(fs string) string {
	fs = strings.TrimSpace(fs)
	for strings.HasPrefix(fs, "-") {
		fs = strings.TrimPrefix(fs, "-")
	}
	return fs
}

// ---------- Request models ----------

type fileOpReq struct {
	Fs     string `json:"fs"`
	Remote string `json:"remote"`
}

type copyMoveFileReq struct {
	SrcFs     string `json:"srcFs"`
	SrcRemote string `json:"srcRemote"`
	DstFs     string `json:"dstFs"`
	DstRemote string `json:"dstRemote"`
}

// ---------- HTTP handlers ----------

func (c *FsController) HandlePublicLink(w http.ResponseWriter, r *http.Request) { c.wrap(w, r, c.doPublicLink) }

// ---------- Core ----------

func (c *FsController) wrap(w http.ResponseWriter, r *http.Request, fn func(context.Context, []byte) (any, error)) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	body, _ := io.ReadAll(r.Body)
	resp, err := fn(r.Context(), body)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	if resp == nil {
		WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}
	WriteJSON(w, http.StatusOK, resp)
}

// normalize path for CLI: ensure fs like "remote:" and remote path relative (no leading "/")
func normalize(fs, remote string) (string, string) {
	fs = strings.TrimSpace(fs)
	remote = strings.TrimSpace(remote)
	if fs != "" {
		name := strings.TrimSuffix(fs, ":")
		fs = canonicalRemoteName(name) + ":"
	}
	remote = strings.TrimPrefix(remote, "/")
	remote = filepath.ToSlash(remote)
	return fs, remote
}

// splitFsRemote: accept either (fs, remote) pair or fs-with-path in fs argument
func splitFsRemote(fs, remote string) (string, string) {
	fs = strings.TrimSpace(fs)
	remote = strings.TrimSpace(remote)
	if remote != "" { return normalize(fs, remote) }
	// if fs already contains a path, split at the first ':'
	i := strings.Index(fs, ":")
	if i < 0 { return normalize(fs, remote) }
	base := fs[:i+1] // include colon
	path := fs[i+1:]
	path = strings.TrimPrefix(path, "/")
	path = filepath.ToSlash(path)
	return base, path
}

// smb share duplicate guard: if error indicates share issue, try strip first path segment and retry once
func tryStripFirstSegment(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 1 { return strings.Join(parts[1:], "/") }
	return path
}

func rcConfigPath() string {
	// Use APP_DATA_DIR/rclone.conf if present
	dir := os.Getenv("APP_DATA_DIR")
	if dir == "" { dir = "." }
	return filepath.Join(dir, "rclone.conf")
}

// map lowercased remote names to canonical names from rclone.conf
func remoteNameMap() map[string]string {
	m := map[string]string{}
	cfg := rcConfigPath()
	b, err := os.ReadFile(cfg)
	if err != nil { return m }
	lines := strings.Split(string(b), "\n")
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if len(ln) > 2 && ln[0] == '[' && ln[len(ln)-1] == ']' {
			name := strings.TrimSpace(ln[1:len(ln)-1])
			if name != "" { m[strings.ToLower(name)] = name }
		}
	}
	return m
}

func canonicalRemoteName(name string) string {
	if name == "" { return name }
	m := remoteNameMap()
	if v, ok := m[strings.ToLower(name)]; ok { return v }
	return name
}

var runRclone = func(ctx context.Context, args ...string) (string, error) {
	// attach config
	cfg := rcConfigPath()
	if _, err := os.Stat(cfg); err == nil {
		args = append([]string{"--config", cfg}, args...)
	}
	cmd := osexec.CommandContext(ctx, "rclone", args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	out := buf.String()
	if err != nil { return out, fmt.Errorf("rclone %v: %w\n%s", args, err, out) }
	return out, nil
}

// WebDAV fallback: if move/moveto fails with DirMove/MOVE errors, do copy(+dir) + delete(+purge)
func isWebdavMoveError(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "dirmove") || strings.Contains(msg, "move call failed") || strings.Contains(msg, "internal server error")
}

// visibility wait using lsjson polling
func waitVisible(ctx context.Context, fs, remote string) {
	interval := 2 * time.Second
	if v := config.GetFinishWaitTimeout(); v > 0 { interval = v / 60 } // coarse: ~60 steps
	deadline := time.Now().Add(config.GetFinishWaitTimeout())
	for time.Now().Before(deadline) {
		_, err := runRclone(ctx, "lsjson", fs+remote)
		if err == nil { return }
		time.Sleep(interval)
	}
}

// deepCleanDir: aggressive removal for stubborn backends (purge → delete -r --rmdirs → rmdir → purge) with path variants
func deepCleanDir(ctx context.Context, fs, remote string) {
	candidates := []string{remote}
	if s := sanitizePath(remote); s != remote { candidates = append(candidates, s) }
	if t := tryStripFirstSegment(remote); t != remote { candidates = append(candidates, t) }
	if t := tryStripFirstSegment(sanitizePath(remote)); t != remote { candidates = append(candidates, t) }
	for _, r := range candidates {
		_, _ = runRclone(ctx, "purge", fs+r)
		_, _ = runRclone(ctx, "delete", fs+r, "-r", "--rmdirs", "--ignore-errors")
		_, _ = runRclone(ctx, "rmdir", fs+r)
		_, _ = runRclone(ctx, "purge", fs+r)
	}
}

// disappearance wait: ensure source path is gone; inspect parent listing and retry delete/purge until name vanishes
func waitGoneDir(ctx context.Context, fs, remote string) {
	interval := 2 * time.Second
	if v := config.GetFinishWaitTimeout(); v > 0 { interval = v / 60 }
	deadline := time.Now().Add(config.GetFinishWaitTimeout())
	par := parentDir(remote)
	name := filepath.Base(remote)
	for time.Now().Before(deadline) {
		out, err := runRclone(ctx, "lsjson", fs+par)
		if err != nil { return } // if parent not found, consider gone
		if !strings.Contains(out, fmt.Sprintf("\"Name\":\"%s\"", name)) { return }
		// still listed: try stronger cleanup
		_, _ = runRclone(ctx, "delete", fs+remote, "-r", "--rmdirs", "--ignore-errors")
		_, _ = runRclone(ctx, "rmdir", fs+remote)
		_, _ = runRclone(ctx, "purge", fs+remote)
		time.Sleep(interval)
	}
}

func waitGoneFile(ctx context.Context, fs, remote string) {
	interval := 2 * time.Second
	if v := config.GetFinishWaitTimeout(); v > 0 { interval = v / 60 }
	deadline := time.Now().Add(config.GetFinishWaitTimeout())
	for time.Now().Before(deadline) {
		// check parent dir listing contains filename
		par := parentDir(remote)
		name := filepath.Base(remote)
		out, err := runRclone(ctx, "lsjson", fs+par)
		if err != nil { return }
		if !strings.Contains(out, name) { return }
		_, _ = runRclone(ctx, "deletefile", fs+remote)
		time.Sleep(interval)
	}
}

func parentDir(p string) string {
	p = filepath.ToSlash(p)
	if p == "" { return "" }
	p = strings.TrimSuffix(p, "/")
	if p == "" { return "" }
	d := filepath.ToSlash(filepath.Dir(p))
	if d == "." { return "" }
	return d
}

func ensureDir(ctx context.Context, fs, remote string) {
	if remote == "" { return }
	remote = sanitizePath(remote)
	_, _ = runRclone(ctx, "mkdir", fs+remote)
}

func (c *FsController) doPublicLink(ctx context.Context, body []byte) (any, error) {
	var req fileOpReq
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("无效的请求格式: %w", err)
	}
	fs, p := normalize(req.Fs, req.Remote)
	out, err := runRclone(ctx, "link", fs+p)
	if err != nil { return nil, err }
	return map[string]string{"url": strings.TrimSpace(out)}, nil
}
