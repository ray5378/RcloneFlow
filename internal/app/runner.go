package app

// exported constructor for other packages
func NewCLIRunner(db *store.DB) *Runner { return NewRunner(db) }


import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"rcloneflow/internal/adapter"
	"rcloneflow/internal/logger"
	"rcloneflow/internal/store"
	"go.uber.org/zap"
)

// Runner manages CLI transfers with progress/logs and stop control.
type Runner struct {
	mu  sync.Mutex
	procs map[int64]*exec.Cmd
	db  *store.DB
}

func NewRunner(db *store.DB) *Runner { return &Runner{procs: map[int64]*exec.Cmd{}, db: db} }

// NewCLIRunner exported for cross-package use
func NewCLIRunner(db *store.DB) *Runner { return NewRunner(db) }

func (r *Runner) Start(ctx context.Context, run store.Run, mode, srcRemote, srcPath, dstRemote, dstPath string) error {
	r.mu.Lock()
	if _, ok := r.procs[run.ID]; ok { r.mu.Unlock(); return errors.New("run already exists") }
	r.mu.Unlock()

	runner := &adapter.CmdRunner{}
	src := srcRemote + ":" + strings.TrimPrefix(srcPath, "/")
	dst := dstRemote + ":" + strings.TrimPrefix(dstPath, "/")
	cmdName := strings.ToLower(mode)
	if cmdName != "copy" && cmdName != "sync" && cmdName != "move" { cmdName = "copy" }
	args := []string{cmdName, src, dst, "--use-json-log", "--log-format", "json", "--stats", "5s"}
	// attach advanced task options if present in run.Summary.effectiveOptions
	var eff map[string]any
	if run.Summary != nil {
		if v, ok := run.Summary["effectiveOptions"]; ok {
			if m, ok := v.(map[string]any); ok { eff = m }
		}
	}
	if len(eff) > 0 {
		args = append(args, buildFlagsFromOptions(eff)...)
	}

	// logs dir and files
	dataDir := os.Getenv("APP_DATA_DIR"); if dataDir == "" { dataDir = "./data" }
	logsDir := filepath.Join(dataDir, "logs"); _ = os.MkdirAll(logsDir, 0o755)
	stdoutPath := filepath.Join(logsDir, fmt.Sprintf("run-%d-stdout.log", run.ID))
	stderrPath := filepath.Join(logsDir, fmt.Sprintf("run-%d-stderr.log", run.ID))
	stdoutFile, _ := os.OpenFile(stdoutPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	stderrFile, _ := os.OpenFile(stderrPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)

	cmd := runner.CmdContext(ctx, args...)
	outPipe, _ := cmd.StdoutPipe()
	errPipe, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil { return err }

	r.mu.Lock(); r.procs[run.ID] = cmd; r.mu.Unlock()
	_ = r.db.UpdateRun(run.ID, func(rr *store.Run){
		if rr.Summary == nil { rr.Summary = map[string]any{} }
		rr.Summary["stdoutFile"] = stdoutPath
		rr.Summary["stderrFile"] = stderrPath
	})

	go r.consume(run.ID, outPipe, stdoutFile)
	go r.consume(run.ID, errPipe, stderrFile)
	go func(){
		_ = cmd.Wait()
		stdoutFile.Close(); stderrFile.Close()
		status := "finished"
		if !cmd.ProcessState.Success() { status = "failed" }
		_ = r.db.UpdateRun(run.ID, func(rr *store.Run){ rr.Status = status })
		r.mu.Lock(); delete(r.procs, run.ID); r.mu.Unlock()
	}()
	return nil
}

func (r *Runner) Stop(runID int64) error {
	r.mu.Lock(); cmd, ok := r.procs[runID]; r.mu.Unlock()
	if !ok || cmd == nil || cmd.Process == nil { return errors.New("not running") }
	_ = cmd.Process.Signal(syscall.SIGINT)
	if wait(cmd, 10*time.Second) { return nil }
	if runtime.GOOS != "windows" { _ = cmd.Process.Signal(syscall.SIGTERM) }
	if wait(cmd, 10*time.Second) { return nil }
	_ = cmd.Process.Kill()
	return nil
}

func wait(cmd *exec.Cmd, d time.Duration) bool {
	ch := make(chan struct{}, 1)
	go func(){ _ = cmd.Wait(); close(ch) }()
	select{
	case <-ch: return true
	case <-time.After(d): return false
	}
}

func (r *Runner) consume(runID int64, rd io.Reader, out *os.File){
	s := bufio.NewScanner(rd)
	s.Buffer(make([]byte, 0, 128*1024), 2*1024*1024)
	for s.Scan(){
		line := s.Text()
		if len(line) > 0 { _, _ = out.WriteString(line+"\n") }
		var rec map[string]any
		if json.Unmarshal([]byte(line), &rec) == nil {
			prog := map[string]any{}
			if v, ok := rec["bytes"].(float64); ok { prog["bytes"] = v }
			if v, ok := rec["totalBytes"].(float64); ok { prog["totalBytes"] = v }
			if v, ok := rec["speed"].(float64); ok { prog["speed"] = v }
			if v, ok := rec["eta"].(float64); ok { prog["eta"] = v }
			if len(prog) > 0 {
				_ = r.db.UpdateRun(runID, func(rr *store.Run){
					if rr.Summary == nil { rr.Summary = map[string]any{} }
					rr.Summary["progress"] = prog
					if b, ok := prog["bytes"].(float64); ok { rr.BytesTransferred = int64(b) }
					if sp, ok := prog["speed"].(float64); ok { rr.Speed = fmt.Sprintf("%d B/s", int64(sp)) }
				})
			}
		}
	}
	if err := s.Err(); err != nil {
		logger.Debug("progress scanner error", zap.Error(err))
	}
}
