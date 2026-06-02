package webdavserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"rcloneflow/internal/logger"

	"go.uber.org/zap"
	"golang.org/x/net/webdav"
)

const (
	httpPort   = "127.0.0.1:17872" // rclone serve http
	streamPort = "127.0.0.1:17873" // Go webdav handler
)

type StreamServer struct {
	mu         sync.Mutex
	cmd        *exec.Cmd
	cancel     context.CancelFunc
	running    bool
	dataDir    string
	configFile string

	webdavSrv *http.Server
}

func NewStreamServer(dataDir, configFile string) *StreamServer {
	return &StreamServer{
		dataDir:    dataDir,
		configFile: configFile,
	}
}

func (s *StreamServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("直连 WebDAV 服务已在运行中")
	}

	// 1. 启动 rclone serve http（不走 VFS，直接透传 Range 请求）
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	args := []string{
		"serve", "http",
		combineRemoteName + ":",
		"--addr", httpPort,
		"--config", s.configFile,
		"--no-checksum",
	}

	s.cmd = exec.CommandContext(ctx, "rclone", args...)
	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stderr

	if err := s.cmd.Start(); err != nil {
		cancel()
		s.cmd = nil
		s.cancel = nil
		return fmt.Errorf("启动 rclone HTTP 服务失败: %w", err)
	}

	// 2. 启动 Go 原生 WebDAV 服务（包装 rclone HTTP）
	fs := &streamFileSystem{
		dataDir:    s.dataDir,
		configFile: s.configFile,
	}
	wHandler := &webdav.Handler{
		FileSystem: fs,
		LockSystem: webdav.NewMemLS(),
	}
	// Wrap to add CORS headers
	httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setStreamCORSHeaders(w.Header())
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == "PROPFIND" && r.Header.Get("Depth") == "infinity" {
			r.Header.Set("Depth", "1")
		}
		// 使用 ResponseWriter 包装器抑制 webdav 库内部重复 WriteHeader
		wHandler.ServeHTTP(&suppressWriteHeaderWriter{ResponseWriter: w}, r)
	})

	s.webdavSrv = &http.Server{
		Addr:         streamPort,
		Handler:      httpHandler,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 30 * time.Minute,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		if err := s.webdavSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("直连 WebDAV 启动失败", zap.Error(err))
		}
	}()

	s.running = true
	logger.Info("直连 WebDAV 服务已启动（只读）", zap.String("http_addr", httpPort), zap.String("dav_addr", streamPort))

	return nil
}

func (s *StreamServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	if s.cancel != nil {
		s.cancel()
	}

	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Signal(os.Interrupt)
		done := make(chan struct{})
		go func() {
			s.cmd.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			s.cmd.Process.Kill()
		}
	}

	if s.webdavSrv != nil {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		s.webdavSrv.Shutdown(shutdownCtx)
	}

	s.running = false
	s.cmd = nil
	s.cancel = nil
	s.webdavSrv = nil

	logger.Info("直连 WebDAV 服务已停止")
	return nil
}

func (s *StreamServer) Restart() error {
	if s.IsRunning() {
		_ = s.Stop()
	}
	return s.Start()
}

func (s *StreamServer) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// streamFileSystem 实现 webdav.FileSystem，只读透传
type streamFileSystem struct {
	dataDir    string
	configFile string
}

func (fs *streamFileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return os.ErrPermission
}

func (fs *streamFileSystem) RemoveAll(ctx context.Context, name string) error {
	return os.ErrPermission
}

func (fs *streamFileSystem) Rename(ctx context.Context, oldName, newName string) error {
	return os.ErrPermission
}

func (fs *streamFileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	name = cleanPath(name)
	// 根目录直接返回目录信息，不走 rclone lsjson
	if name == "" || name == "." {
		return &streamFileInfo{name: "/", size: 0, isDir: true}, nil
	}
	info, err := fs.statRemote(ctx, name)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (fs *streamFileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	// 只读模式不允许写入
	if flag&(os.O_WRONLY|os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC) != 0 {
		return nil, os.ErrPermission
	}

	name = cleanPath(name)

	// 如果 name 为空，返回合成目录（虚拟根目录，列出 combine 的 upstream 目录名）
	if name == "" || name == "." {
		return &streamRootDir{fs: fs}, nil
	}

	info, err := fs.statRemote(ctx, name)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return &streamDir{fs: fs, path: name}, nil
	}

	return &streamFile{fs: fs, path: name, size: info.Size()}, nil
}

type lsjsonEntry struct {
	Path    string `json:"Path"`
	Name    string `json:"Name"`
	Size    int64  `json:"Size"`
	IsDir   bool   `json:"IsDir"`
	ModTime string `json:"ModTime"`
}

func (fs *streamFileSystem) statRemote(ctx context.Context, remotePath string) (os.FileInfo, error) {
	// rclone path 格式: remote:path，combine remote 需要加冒号
	rclonePath := combineRemoteName + ":" + remotePath

	execCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "rclone", "lsjson", rclonePath, "--config", fs.configFile, "--no-modtime")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 3 {
			// Exit code 3 means directory not found - return not found
			return nil, os.ErrNotExist
		}
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	var entries []lsjsonEntry
	if err := json.Unmarshal(out, &entries); err != nil {
		return nil, fmt.Errorf("解析文件信息失败: %w", err)
	}

	if len(entries) == 0 {
		return &streamFileInfo{name: filepath.Base(remotePath), size: 0, isDir: false}, nil
	}

	// 找匹配的条目
	for _, e := range entries {
		if e.Path == remotePath || e.Name == filepath.Base(remotePath) {
			return &streamFileInfo{
				name:  e.Name,
				size:  e.Size,
				isDir: e.IsDir,
			}, nil
		}
	}

	// lsjson 对于单个文件也返回包含该文件信息的数组，取第一个
	e := entries[0]
	return &streamFileInfo{
		name:  e.Name,
		size:  e.Size,
		isDir: e.IsDir,
	}, nil
}

func (fs *streamFileSystem) listRemote(ctx context.Context, remotePath string) ([]lsjsonEntry, error) {
	rclonePath := combineRemoteName + ":" + remotePath

	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "rclone", "lsjson", rclonePath, "--config", fs.configFile, "--no-modtime")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取目录列表失败: %w", err)
	}

	var entries []lsjsonEntry
	if err := json.Unmarshal(out, &entries); err != nil {
		return nil, fmt.Errorf("解析目录列表失败: %w", err)
	}
	return entries, nil
}

// 通过 rclone serve http 透传文件内容
func (fs *streamFileSystem) fetchFile(ctx context.Context, remotePath string) (io.ReadCloser, error) {
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		httpReq, _ := http.NewRequestWithContext(ctx, "GET", "http://"+httpPort+"/"+remotePath, nil)
		client := &http.Client{Timeout: 30 * time.Minute}
		httpResp, err := client.Do(httpReq)
		if err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		defer httpResp.Body.Close()

		if httpResp.StatusCode >= 400 {
			_ = pw.CloseWithError(fmt.Errorf("HTTP %d", httpResp.StatusCode))
			return
		}

		_, err = io.Copy(pw, httpResp.Body)
		if err != nil {
			_ = pw.CloseWithError(err)
		}
	}()

	return pr, nil
}

// streamFileInfo 实现 os.FileInfo
type streamFileInfo struct {
	name  string
	size  int64
	isDir bool
}

func (fi *streamFileInfo) Name() string       { return fi.name }
func (fi *streamFileInfo) Size() int64        { return fi.size }
func (fi *streamFileInfo) Mode() os.FileMode  { return 0644 }
func (fi *streamFileInfo) ModTime() time.Time { return time.Now() }
func (fi *streamFileInfo) IsDir() bool        { return fi.isDir }
func (fi *streamFileInfo) Sys() any           { return nil }

// streamFile 实现 webdav.File（文件）
type streamFile struct {
	fs     *streamFileSystem
	path   string
	size   int64
	reader io.ReadCloser
}

func (f *streamFile) Close() error {
	if f.reader != nil {
		return f.reader.Close()
	}
	return nil
}

func (f *streamFile) Read(p []byte) (int, error) {
	if f.reader == nil {
		var err error
		f.reader, err = f.fs.fetchFile(context.Background(), f.path)
		if err != nil {
			return 0, err
		}
	}
	return f.reader.Read(p)
}

func (f *streamFile) Seek(offset int64, whence int) (int64, error) {
	// Seek 不被 webdav.Handler 直接调用，它用于 Range 请求时会重新 OpenFile
	return 0, fmt.Errorf("not supported")
}

func (f *streamFile) Write(p []byte) (int, error) {
	return 0, os.ErrPermission
}

func (f *streamFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("not a directory")
}

func (f *streamFile) Stat() (os.FileInfo, error) {
	return &streamFileInfo{
		name: filepath.Base(f.path),
		size: f.size,
	}, nil
}

// streamDir 实现 webdav.File（目录）
type streamDir struct {
	fs      *streamFileSystem
	path    string
	entries []os.FileInfo
	pos     int
	listed  bool
}

func (d *streamDir) Close() error { return nil }

func (d *streamDir) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("is a directory")
}

func (d *streamDir) Seek(offset int64, whence int) (int64, error) {
	return 0, fmt.Errorf("not supported")
}

func (d *streamDir) Write(p []byte) (int, error) {
	return 0, os.ErrPermission
}

func (d *streamDir) Stat() (os.FileInfo, error) {
	return &streamFileInfo{
		name:  filepath.Base(d.path),
		isDir: true,
	}, nil
}

func (d *streamDir) Readdir(count int) ([]os.FileInfo, error) {
	if !d.listed {
		entries, err := d.fs.listRemote(context.Background(), d.path)
		if err != nil {
			return nil, err
		}
		d.entries = make([]os.FileInfo, 0, len(entries))
		for _, e := range entries {
			d.entries = append(d.entries, &streamFileInfo{
				name:  e.Name,
				size:  e.Size,
				isDir: e.IsDir,
			})
		}
		d.listed = true
		d.pos = 0
	}

	if d.pos >= len(d.entries) {
		return nil, io.EOF
	}

	if count <= 0 {
		count = len(d.entries) - d.pos
	}
	end := d.pos + count
	if end > len(d.entries) {
		end = len(d.entries)
	}

	result := d.entries[d.pos:end]
	d.pos = end
	return result, nil
}

// streamRootDir 虚拟根目录，列出 combine 上游的顶级目录
type streamRootDir struct {
	fs      *streamFileSystem
	entries []os.FileInfo
	pos     int
	listed  bool
}

func (d *streamRootDir) Close() error { return nil }
func (d *streamRootDir) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("is a directory")
}
func (d *streamRootDir) Seek(offset int64, whence int) (int64, error) {
	return 0, fmt.Errorf("not supported")
}
func (d *streamRootDir) Write(p []byte) (int, error) {
	return 0, os.ErrPermission
}
func (d *streamRootDir) Stat() (os.FileInfo, error) {
	return &streamFileInfo{name: "/", isDir: true}, nil
}

func (d *streamRootDir) Readdir(count int) ([]os.FileInfo, error) {
	if !d.listed {
		entries, err := d.fs.listRemote(context.Background(), "")
		if err != nil {
			return nil, err
		}
		d.entries = make([]os.FileInfo, 0, len(entries))
		for _, e := range entries {
			d.entries = append(d.entries, &streamFileInfo{
				name:  e.Name,
				size:  e.Size,
				isDir: e.IsDir,
			})
		}
		d.listed = true
		d.pos = 0
	}

	if d.pos >= len(d.entries) {
		return nil, io.EOF
	}

	if count <= 0 {
		count = len(d.entries) - d.pos
	}
	end := d.pos + count
	if end > len(d.entries) {
		end = len(d.entries)
	}

	result := d.entries[d.pos:end]
	d.pos = end
	return result, nil
}

func cleanPath(name string) string {
	name = filepath.Clean("/" + strings.TrimPrefix(name, "/"))
	name = strings.TrimLeft(name, "/")
	return name
}

func setStreamCORSHeaders(header http.Header) {
	header.Set("Access-Control-Allow-Origin", "*")
	header.Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS, PROPFIND")
	header.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Depth, Range")
	header.Set("Access-Control-Expose-Headers", "Content-Range, Accept-Ranges, Content-Length, ETag, Content-Type")
	header.Set("Access-Control-Max-Age", "86400")
}

// suppressWriteHeaderWriter 包装 http.ResponseWriter，抑制重复的 WriteHeader 调用
type suppressWriteHeaderWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (w *suppressWriteHeaderWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *suppressWriteHeaderWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}
