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
	args := []string{cmdName, src, dst, "-vv", "--progress", "--stats", "5s", "--stats-one-line", "--use-json-log", "--log-level", "INFO", "--stats-log-level", "INFO", "--config", cfg}
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

	// 同时在 stdout/stderr 中解析 one-line 进度，兼容不同输出流
	go r.consume(run.ID, outR, stdoutFile, true)
	go r.consume(run.ID, errR, stderrFile, true)
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

func (r *Runner) consume(runID int64, rd io.Reader, out *os.File, parseStats ...bool){
	wantParse := len(parseStats) > 0 && parseStats[0]
	s := bufio.NewScanner(rd)
	s.Buffer(make([]byte, 0, 128*1024), 2*1024*1024)
	for s.Scan(){
		line := s.Text()
		if len(line) > 0 { _, _ = out.WriteString(line+"\n") }
		// 1) JSON 行（极少数情况下）
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
				continue
			}
		}
		// 2) 文本 one-line 解析（仅在需要时）
		if wantParse {
			if prog, ok := parseOneLineProgress(line); ok {
				_ = r.db.UpdateRun(runID, func(rr *store.Run){
					if rr.Summary == nil { rr.Summary = map[string]any{} }
					rr.Summary["progress"] = prog
					rr.BytesTransferred = int64(prog["bytes"].(float64))
					rr.Speed = fmt.Sprintf("%d B/s", int64(prog["speed"].(float64)))
				})
			}
		}
	}
	if err := s.Err(); err != nil {
		logger.Debug("progress scanner error", zap.Error(err))
	}
}

var oneLineRe = regexp.MustCompile(`(?i)^(?:transferred:)?\s*(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)(?:B)?\s*/\s*(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)(?:B)?\s*,\s*(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)(?:B)?/s\s*,\s*ETA\s*([0-9hms:.-]+)`) 

func unitToMul(u string) float64 {
	u = strings.ToUpper(u)
	switch u {
	case "K", "KI": return 1024
	case "M", "MI": return 1024 * 1024
	case "G", "GI": return 1024 * 1024 * 1024
	case "T", "TI": return 1024 * 1024 * 1024 * 1024
	case "P", "PI": return 1024 * 1024 * 1024 * 1024 * 1024
	default: return 1
	}
}

func parseETA(s string) int {
	s = strings.TrimSpace(s)
	if s == "-" || s == "" { return 0 }
	// formats: 1h2m3s | 2m3s | 45s | 01:23:45 | 12:34
	if strings.Contains(s, ":") {
		parts := strings.Split(s, ":")
		if len(parts) == 3 { // hh:mm:ss
			var h, m, sec int
			fmt.Sscanf(s, "%d:%d:%d", &h, &m, &sec)
			return h*3600 + m*60 + sec
		}
		if len(parts) == 2 { // mm:ss
			var m, sec int
			fmt.Sscanf(s, "%d:%d", &m, &sec)
			return m*60 + sec
		}
	}
	// h/m/s suffix
	sec := 0
	var h, m, ss int
	fmt.Sscanf(s, "%dh%dm%ds", &h, &m, &ss)
	if h == 0 && m == 0 && ss == 0 {
		fmt.Sscanf(s, "%dm%ds", &m, &ss)
		if m == 0 && ss == 0 {
			fmt.Sscanf(s, "%ds", &ss)
		}
	}
	sec += h*3600 + m*60 + ss
	return sec
}

func parseOneLineProgress(line string) (map[string]any, bool) {
	m := oneLineRe.FindStringSubmatch(strings.TrimSpace(line))
	if len(m) == 0 { return nil, false }
	// m[1]=cur, m[2]=curUnit, m[3]=total, m[4]=totalUnit, m[5]=speed, m[6]=speedUnit, m[7]=eta
	var cur, tot, sp float64
	fmt.Sscanf(m[1], "%f", &cur)
	fmt.Sscanf(m[3], "%f", &tot)
	fmt.Sscanf(m[5], "%f", &sp)
	curBytes := cur * unitToMul(m[2])
	totBytes := tot * unitToMul(m[4])
	spBytes := sp * unitToMul(m[6])
	eta := parseETA(m[7])
	prog := map[string]any{"bytes": curBytes, "totalBytes": totBytes, "speed": spBytes, "eta": float64(eta)}
	return prog, true
}
