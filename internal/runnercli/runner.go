package runnercli

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
	"regexp"

	"rcloneflow/internal/adapter"
	"rcloneflow/internal/logger"
	"rcloneflow/internal/store"
	"go.uber.org/zap"
)

// Runner manages CLI transfers with progress/logs and stop control.
type Runner struct {
	mu    sync.Mutex
	procs map[int64]*exec.Cmd
	db    *store.DB
}

func New(db *store.DB) *Runner { return &Runner{procs: map[int64]*exec.Cmd{}, db: db} }

func (r *Runner) Start(ctx context.Context, run store.Run, mode, srcRemote, srcPath, dstRemote, dstPath string) error {
	r.mu.Lock()
	if _, ok := r.procs[run.ID]; ok { r.mu.Unlock(); return errors.New("run already exists") }
	r.mu.Unlock()

	runner := &adapter.CmdRunner{}
	src := srcRemote + ":" + strings.TrimPrefix(srcPath, "/")
	dst := dstRemote + ":" + strings.TrimPrefix(dstPath, "/")
	cmdName := strings.ToLower(mode)
	if cmdName != "copy" && cmdName != "sync" && cmdName != "move" { cmdName = "copy" }
	// Resolve config path
	dataDir := os.Getenv("APP_DATA_DIR"); if dataDir == "" { dataDir = "./data" }
	cfg := os.Getenv("RCLONE_CONFIG"); if cfg == "" { cfg = filepath.Join(dataDir, "rclone.conf") }
	// One-line JSON stats + JSON log records
	args := []string{cmdName, src, dst, "-vv", "--progress", "--stats", "5s", "--stats-one-line", "--config", cfg}
	// attach advanced options if present
	var effOpt map[string]any
	if run.Summary != nil {
		if v, ok := run.Summary["effectiveOptions"]; ok {
			if m, ok := v.(map[string]any); ok {
				effOpt = m
				args = append(args, buildFlagsFromOptions(m)...)
			}
		}
	}
	// header will be written after files are opened below
	startLine := "[runner] rclone " + strings.Join(args, " ") + "\n"
	missingCfg := ""
	if _, err := os.Stat(cfg); err != nil { missingCfg = "[runner] warn: config not found: " + cfg + "\n" }

	logsBase := os.Getenv("APP_DATA_DIR"); if logsBase == "" { logsBase = "./data" }
	logsDir := filepath.Join(logsBase, "logs"); _ = os.MkdirAll(logsDir, 0o755)
	stdoutPath := filepath.Join(logsDir, fmt.Sprintf("run-%d-stdout.log", run.ID))
	stderrPath := filepath.Join(logsDir, fmt.Sprintf("run-%d-stderr.log", run.ID))
	stdoutFile, _ := os.OpenFile(stdoutPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	stderrFile, _ := os.OpenFile(stderrPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)

	cmd := runner.CmdContext(ctx, args...)
	// fan-out: write to parser via io.Pipe（再由 consumer 单点写入文件，避免重复行）
	outR, outW := io.Pipe()
	errR, errW := io.Pipe()
	// write headers once files are open（同时写入 stdout/stderr 便于查看）
	_, _ = stdoutFile.WriteString(startLine)
	_, _ = stderrFile.WriteString(startLine)
	if effOpt != nil {
		if b, _ := json.Marshal(effOpt); len(b) > 0 {
			optsLine := "[runner] effectiveOptions " + string(b) + "\n"
			_, _ = stdoutFile.WriteString(optsLine)
			_, _ = stderrFile.WriteString(optsLine)
		}
	}
	if missingCfg != "" { _, _ = stdoutFile.WriteString(missingCfg); _, _ = stderrFile.WriteString(missingCfg) }
	cmd.Stdout = outW
	cmd.Stderr = errW
	if err := cmd.Start(); err != nil { return err }

	r.mu.Lock(); r.procs[run.ID] = cmd; r.mu.Unlock()
	_ = r.db.UpdateRun(run.ID, func(rr *store.Run){
		if rr.Summary == nil { rr.Summary = map[string]any{} }
		rr.Summary["stdoutFile"] = stdoutPath
		rr.Summary["stderrFile"] = stderrPath
	})

	go r.consume(run.ID, outR, stdoutFile)
	go r.consume(run.ID, errR, stderrFile)
	// 额外：解析 --stats-one-line 文本行，提取 bytes/total/speed/eta
	go r.parseOneLine(run.ID, errR)
	go func(){
		err := cmd.Wait()
		outW.Close(); errW.Close()
		stdoutFile.Close(); stderrFile.Close()
		if err != nil || (cmd.ProcessState != nil && !cmd.ProcessState.Success()) {
			_ = r.db.UpdateRun(run.ID, func(rr *store.Run){ rr.Status = "failed" })
			r.mu.Lock(); delete(r.procs, run.ID); r.mu.Unlock()
			return
		}
		// Post-Verify（读取任务级/全局设置）
		pvEnabled := true
		pvMode := "mount"
		pvMatch := "size"
		pvInterval := 5 * time.Second
		pvTimeout := 30 * time.Minute
		// 从 run.Summary.effectiveOptions 读取任务覆盖
		_ = r.db.UpdateRun(run.ID, func(rr *store.Run){
			// 仅为了读 summary，不改变状态
		})
		cur, _ := r.db.GetRun(run.ID)
		if cur.Summary != nil {
			if effm, ok := cur.Summary["effectiveOptions"].(map[string]any); ok {
				if v, ok := effm["postVerify.enabled"].(bool); ok { pvEnabled = v }
				if v, ok := effm["postVerify.mode"].(string); ok && v != "" { pvMode = v }
				if v, ok := effm["postVerify.match"].(string); ok && v != "" { pvMatch = v }
				if v, ok := effm["postVerify.interval"].(string); ok { if d, e := time.ParseDuration(v); e == nil { pvInterval = d } }
				if v, ok := effm["postVerify.timeout"].(string); ok { if d, e := time.ParseDuration(v); e == nil { pvTimeout = d } }
			}
			// 再从 transferDefaults 读取全局默认（仅在任务未覆盖时）
			if def, ok := cur.Summary["transferDefaults"].(map[string]any); ok {
				if effm := eff(cur.Summary); effm != nil {
					if !existsBool(effm, "postVerify.enabled") { if b, ok := def["postVerifyEnabled"].(bool); ok { pvEnabled = b } }
					if !existsStr(effm, "postVerify.mode") { if s, ok := def["postVerifyMode"].(string); ok && s != "" { pvMode = s } }
					if !existsStr(effm, "postVerify.match") { if s, ok := def["postVerifyMatch"].(string); ok && s != "" { pvMatch = s } }
					if !existsStr(effm, "postVerify.interval") { if s, ok := def["postVerifyInterval"].(string); ok { if d, e := time.ParseDuration(s); e == nil { pvInterval = d } } }
					if !existsStr(effm, "postVerify.timeout") { if s, ok := def["postVerifyTimeout"].(string); ok { if d, e := time.ParseDuration(s); e == nil { pvTimeout = d } } }
				}
			}
		}
		if pvEnabled && pvMode == "mount" {
			_ = r.db.UpdateRun(run.ID, func(rr *store.Run){ rr.Status = "finalizing" })
			deadline := time.Now().Add(pvTimeout)
			vr := &adapter.CmdRunner{}
			ok := false
			var lastSrcBytes, lastDstBytes int64
			for time.Now().Before(deadline) {
				if pvMatch == "size" {
					sb, sc, sErr := sizeOf(vr, cfg, src)
					db2, dc, dErr := sizeOf(vr, cfg, dst)
					if sErr == nil && dErr == nil {
						lastSrcBytes, lastDstBytes = sb, db2
						if sb == db2 && sc == dc { ok = true; break }
					}
				}
				time.Sleep(pvInterval)
			}
			if ok {
				_ = r.db.UpdateRun(run.ID, func(rr *store.Run){ rr.Status = "finished" })
			} else {
				_ = r.db.UpdateRun(run.ID, func(rr *store.Run){
					rr.Status = "finalizing_timeout"
					if rr.Summary == nil { rr.Summary = map[string]any{} }
					rr.Summary["postVerify"] = map[string]any{"match": pvMatch, "timeout": pvTimeout.String(), "srcBytes": lastSrcBytes, "dstBytes": lastDstBytes}
				})
			}
		} else {
			_ = r.db.UpdateRun(run.ID, func(rr *store.Run){ rr.Status = "finished" })
		}
		r.mu.Lock(); delete(r.procs, run.ID); r.mu.Unlock()
	}()
	return nil
}

func sizeOf(r *adapter.CmdRunner, cfg, target string) (bytes int64, count int64, err error) {
	// Prefer JSON when available. Fallback to parsing text if needed.
	args := []string{"size", target, "--config", cfg, "--json"}
	out, _, e := r.Run(context.Background(), args...)
	if e == nil {
		var m map[string]any
		if json.Unmarshal([]byte(out), &m) == nil {
			if v, ok := m["bytes"].(float64); ok { bytes = int64(v) }
			if v, ok := m["count"].(float64); ok { count = int64(v) }
			return bytes, count, nil
		}
	}
	// Fallback: text parse
	args = []string{"size", target, "--config", cfg}
	out, _, e = r.Run(context.Background(), args...)
	if e != nil { return 0, 0, e }
	// look for lines: "Total objects: X" and "Total size: Y"
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Total objects:") {
			fmt.Sscanf(line, "Total objects: %d", &count)
		}
		if strings.HasPrefix(line, "Total size:") {
			var human string
			fmt.Sscanf(line, "Total size: %s (%d)", &human, &bytes)
		}
	}
	return bytes, count, nil
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

var oneLineRe = regexp.MustCompile(`(?i)(?:(\d+(?:\.\d+)?)([KMGTP]?)[b]?)\/\s*(\d+(?:\.\d+)?)([KMGTP]?)\s*[, ]+([\d\.]+)\s*([KMGTP]?)[b]?/s.*?ETA\s*(\d+h)?(?::?(\d+)m)?(?::?(\d+)s)?`)
func humanToBytes(num, unit string) int64 {
	f := 0.0; fmt.Sscanf(num, "%f", &f)
	switch strings.ToUpper(unit) {
	case "K": return int64(f * 1024)
	case "M": return int64(f * 1024 * 1024)
	case "G": return int64(f * 1024 * 1024 * 1024)
	case "T": return int64(f * 1024 * 1024 * 1024 * 1024)
	case "P": return int64(f * 1024 * 1024 * 1024 * 1024 * 1024)
	default: return int64(f)
	}
}

func (r *Runner) parseOneLine(runID int64, rd io.Reader){
	s := bufio.NewScanner(rd)
	s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for s.Scan(){
		line := s.Text()
		m := oneLineRe.FindStringSubmatch(line)
		if len(m) > 0 {
			bytes := humanToBytes(m[1], m[2])
			total := humanToBytes(m[3], m[4])
			speed := humanToBytes(m[5], m[6])
			etaSec := 0
			if m[7] != "" { var h int; fmt.Sscanf(m[7], "%dh", &h); etaSec += h*3600 }
			if m[8] != "" { var mm int; fmt.Sscanf(m[8], "%dm", &mm); etaSec += mm*60 }
			if m[9] != "" { var ss int; fmt.Sscanf(m[9], "%ds", &ss); etaSec += ss }
			prog := map[string]any{"bytes": float64(bytes), "totalBytes": float64(total), "speed": float64(speed), "eta": float64(etaSec)}
			_ = r.db.UpdateRun(runID, func(rr *store.Run){
				if rr.Summary == nil { rr.Summary = map[string]any{} }
				rr.Summary["progress"] = prog
				rr.BytesTransferred = bytes
				rr.Speed = fmt.Sprintf("%d B/s", speed)
			})
		}
	}
}
