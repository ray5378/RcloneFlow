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

func (c *FsController) HandleMkdir(w http.ResponseWriter, r *http.Request)    { c.wrap(w, r, c.doMkdir) }
func (c *FsController) HandleDeleteFile(w http.ResponseWriter, r *http.Request){ c.wrap(w, r, c.doDeleteFile) }
func (c *FsController) HandlePurge(w http.ResponseWriter, r *http.Request)    { c.wrap(w, r, c.doPurge) }
func (c *FsController) HandleMove(w http.ResponseWriter, r *http.Request)     { c.wrap(w, r, c.doMoveFile) }
func (c *FsController) HandleCopy(w http.ResponseWriter, r *http.Request)     { c.wrap(w, r, c.doCopyFile) }
func (c *FsController) HandleCopyDir(w http.ResponseWriter, r *http.Request)  { c.wrap(w, r, c.doCopyDir) }
func (c *FsController) HandleMoveDir(w http.ResponseWriter, r *http.Request)  { c.wrap(w, r, c.doMoveDir) }
func (c *FsController) HandlePublicLink(w http.ResponseWriter, r *http.Request){ c.wrap(w, r, c.doPublicLink) }

// ---------- Core ----------

func (c *FsController) wrap(w http.ResponseWriter, r *http.Request, fn func(context.Context, []byte) (any, error)) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	body, _ := io.ReadAll(r.Body)
	resp, err := fn(r.Context(), body)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	if resp == nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
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

func runRclone(ctx context.Context, args ...string) (string, error) {
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

// ---------- Operations ----------

func (c *FsController) doMkdir(ctx context.Context, body []byte) (any, error) {
	var req fileOpReq
	_ = json.Unmarshal(body, &req)
	fs, p := normalize(req.Fs, req.Remote)
	_, err := runRclone(ctx, "mkdir", fs+p)
	return nil, err
}

func (c *FsController) doDeleteFile(ctx context.Context, body []byte) (any, error) {
	var req fileOpReq
	_ = json.Unmarshal(body, &req)
	fs, p := normalize(req.Fs, req.Remote)
	// file delete
	_, err := runRclone(ctx, "deletefile", fs+p)
	if err != nil {
		// treat 404-ish as success
		if strings.Contains(strings.ToLower(err.Error()), "not found") { return nil, nil }
	}
	return nil, err
}

func (c *FsController) doPurge(ctx context.Context, body []byte) (any, error) {
	var req fileOpReq
	_ = json.Unmarshal(body, &req)
	fs, p := normalize(req.Fs, req.Remote)
	_, err := runRclone(ctx, "purge", fs+p)
	return nil, err
}

func parentDir(p string) string {
	p = filepath.ToSlash(p)
	if p == "" { return "" }
	// remove trailing slash to avoid Dir giving parent of empty
	p = strings.TrimSuffix(p, "/")
	if p == "" { return "" }
	d := filepath.ToSlash(filepath.Dir(p))
	if d == "." { return "" }
	return d
}

func sanitizePath(p string) string {
	p = filepath.ToSlash(p)
	parts := strings.Split(p, "/")
	for i, s := range parts {
		// strip trailing ASCII colon which is invalid on SMB/Windows
		for strings.HasSuffix(s, ":") { s = strings.TrimSuffix(s, ":") }
		parts[i] = s
	}
	return strings.Join(parts, "/")
}

func ensureDir(ctx context.Context, fs, remote string) {
	if remote == "" { return }
	remote = sanitizePath(remote)
	_, _ = runRclone(ctx, "mkdir", fs+remote)
}

func (c *FsController) doCopyFile(ctx context.Context, body []byte) (any, error) {
	var req copyMoveFileReq
	_ = json.Unmarshal(body, &req)
	srcFs, src := normalize(req.SrcFs, req.SrcRemote)
	dstFs, dst := normalize(req.DstFs, req.DstRemote)
	src = sanitizePath(src)
	dst = sanitizePath(dst)
	// ensure parent dir for destination exists
	ensureDir(ctx, dstFs, parentDir(dst))
	_, err := runRclone(ctx, "copyto", srcFs+src, dstFs+dst)
	if err != nil {
		// SMB duplicate share fallback
		if strings.Contains(strings.ToLower(err.Error()), "network name not found") || strings.Contains(strings.ToLower(err.Error()), "create filesystem") {
			s2 := tryStripFirstSegment(src)
			d2 := tryStripFirstSegment(dst)
			if s2 != src || d2 != dst {
				ensureDir(ctx, dstFs, parentDir(d2))
				if _, err2 := runRclone(ctx, "copyto", srcFs+s2, dstFs+d2); err2 == nil { return nil, nil } else { err = err2 }
			}
		}
	}
	return nil, err
}

func (c *FsController) doMoveFile(ctx context.Context, body []byte) (any, error) {
	var req copyMoveFileReq
	_ = json.Unmarshal(body, &req)
	srcFs, src := normalize(req.SrcFs, req.SrcRemote)
	dstFs, dst := normalize(req.DstFs, req.DstRemote)
	src = sanitizePath(src)
	dst = sanitizePath(dst)
	// first try moveto
	_, err := runRclone(ctx, "moveto", srcFs+src, dstFs+dst)
	if err == nil {
		// step2: ensure old file gone
		go waitGoneFile(context.Background(), srcFs, src)
		return nil, nil
	}
	// retry once for smb duplicate share: strip first segment
	if strings.Contains(strings.ToLower(err.Error()), "network name not found") || strings.Contains(strings.ToLower(err.Error()), "create filesystem") {
		s2 := tryStripFirstSegment(src)
		d2 := tryStripFirstSegment(dst)
		if s2 != src || d2 != dst {
			if _, err2 := runRclone(ctx, "moveto", srcFs+s2, dstFs+d2); err2 == nil {
				go waitGoneFile(context.Background(), srcFs, s2)
				return nil, nil
			} else { err = err2 }
		}
	}
	// WebDAV fallback: copy + delete
	if isWebdavMoveError(err) {
		if _, er2 := runRclone(ctx, "copyto", srcFs+src, dstFs+dst); er2 == nil {
			_, _ = runRclone(ctx, "deletefile", srcFs+src)
			go waitVisible(context.Background(), dstFs, dst)
			go waitGoneFile(context.Background(), srcFs, src)
			return nil, nil
		}
	}
	return nil, err
}

func (c *FsController) doCopyDir(ctx context.Context, body []byte) (any, error) {
	var req copyMoveFileReq
	_ = json.Unmarshal(body, &req)
	srcFs, src := splitFsRemote(req.SrcFs, req.SrcRemote)
	dstFs, dst := splitFsRemote(req.DstFs, req.DstRemote)
	src = sanitizePath(src)
	dst = sanitizePath(dst)
	// ensure destination exists
	ensureDir(ctx, dstFs, dst)
	_, err := runRclone(ctx, "copy", srcFs+src, dstFs+dst)
	if err != nil {
		// SMB duplicate share fallback
		low := strings.ToLower(err.Error())
		if strings.Contains(low, "network name not found") || strings.Contains(low, "create filesystem") {
			s2 := tryStripFirstSegment(src)
			d2 := tryStripFirstSegment(dst)
			if s2 != src || d2 != dst {
				ensureDir(ctx, dstFs, d2)
				if _, err2 := runRclone(ctx, "copy", srcFs+s2, dstFs+d2); err2 == nil { return nil, nil } else { err = err2 }
			}
		}
	}
	return nil, err
}

func (c *FsController) doMoveDir(ctx context.Context, body []byte) (any, error) {
	var req copyMoveFileReq
	_ = json.Unmarshal(body, &req)
	srcFs, src := splitFsRemote(req.SrcFs, req.SrcRemote)
	dstFs, dst := splitFsRemote(req.DstFs, req.DstRemote)
	src = sanitizePath(src)
	dst = sanitizePath(dst)
	// ensure destination exists
	ensureDir(ctx, dstFs, dst)
	_, err := runRclone(ctx, "move", srcFs+src, dstFs+dst)
	if err == nil {
		// match CLI-good behavior: after move, aggressively ensure old src disappears
		go func(){ deepCleanDir(context.Background(), srcFs, src); waitGoneDir(context.Background(), srcFs, src) }()
		return nil, nil
	}
	if strings.Contains(strings.ToLower(err.Error()), "network name not found") || strings.Contains(strings.ToLower(err.Error()), "create filesystem") {
		s2 := tryStripFirstSegment(src)
		d2 := tryStripFirstSegment(dst)
		if s2 != src || d2 != dst { ensureDir(ctx, dstFs, d2); if _, err2 := runRclone(ctx, "move", srcFs+s2, dstFs+d2); err2 == nil {
			go func(){ deepCleanDir(context.Background(), srcFs, s2); waitGoneDir(context.Background(), srcFs, s2) }()
			return nil, nil
		} else { err = err2 } }
	}
	if isWebdavMoveError(err) {
		if _, er2 := runRclone(ctx, "copy", srcFs+src, dstFs+dst); er2 == nil {
			_, _ = runRclone(ctx, "purge", srcFs+src)
			go waitVisible(context.Background(), dstFs, dst)
			go func(){ deepCleanDir(context.Background(), srcFs, src); waitGoneDir(context.Background(), srcFs, src) }()
			return nil, nil
		}
	}
	return nil, err
}

func (c *FsController) doPublicLink(ctx context.Context, body []byte) (any, error) {
	var req fileOpReq
	_ = json.Unmarshal(body, &req)
	fs, p := normalize(req.Fs, req.Remote)
	out, err := runRclone(ctx, "link", fs+p)
	if err != nil { return nil, err }
	return map[string]string{"url": strings.TrimSpace(out)}, nil
}
