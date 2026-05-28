package runnercli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
	"rcloneflow/internal/active_transfer"
	"rcloneflow/internal/adapter"
	"rcloneflow/internal/config"
	"rcloneflow/internal/logger"
	"rcloneflow/internal/store"
	"rcloneflow/internal/util"
)

// Runner manages CLI transfers with progress/logs and stop control.
type Runner struct {
	mu              sync.Mutex
	procs           map[int64]*exec.Cmd
	cancelFns       map[int64]context.CancelFunc
	updater         RunUpdater
	broadcaster     EventBroadcaster
	activeMgr       *active_transfer.Manager
	casVerifier     func(cfg, dst, rel string) (bool, error)
	casVerifyDelays []time.Duration
	casExcludeMu    sync.Mutex
}

func New(updater RunUpdater, broadcaster EventBroadcaster, activeMgr ...*active_transfer.Manager) *Runner {
	var mgr *active_transfer.Manager
	if len(activeMgr) > 0 {
		mgr = activeMgr[0]
	}
	return &Runner{
		procs:           map[int64]*exec.Cmd{},
		cancelFns:       map[int64]context.CancelFunc{},
		updater:         updater,
		broadcaster:     broadcaster,
		activeMgr:       mgr,
		casVerifier:     defaultCASFileExists,
		casVerifyDelays: []time.Duration{0, 2 * time.Second, 3 * time.Second, 5 * time.Second, 8 * time.Second},
	}
}

// buildTransferArgs builds the command-line arguments for a transfer
func (r *Runner) buildTransferArgs(mode, src, dst, cfg string, run store.Run) ([]string, *openlistCASCompatPlan, bool, int, map[string]any, string, error) {
	var args []string
	var err error
	var effOpt map[string]any
	var casCompat *openlistCASCompatPlan
	casManagedRetries := false
	maxCASAttempts := 1
	cmdName := strings.ToLower(mode)
	isBisyncMode := false
	if cmdName != "copy" && cmdName != "sync" && cmdName != "move" && cmdName != "bisync" {
		cmdName = "copy"
	} else if cmdName == "bisync" {
		isBisyncMode = true
	}
	originalCmdName := cmdName
	
	if isOpenlistCASCompatible(run) {
		plan, err := buildOpenlistCASCompatPlan(cfg, src, dst, cmdName)
		if err != nil {
			return nil, nil, false, 0, nil, "", fmt.Errorf("prepare openlist-cas compatibility failed: %w", err)
		}
		casCompat = plan
		if cmdName == "sync" {
			cmdName = "copy"
		}
	}
	
	if isBisyncMode {
		args, err = buildBisyncArgs(src, dst, run)
		if err != nil {
			return nil, nil, false, 0, nil, "", fmt.Errorf("build bisync args: %w", err)
		}
		args = append(args, "--stats", "1s", "--stats-one-line", "--config", cfg, "--use-json-log", "--log-level", "INFO", "--stats-log-level", "INFO")
	} else {
		args = []string{cmdName, src, dst, "--stats", "1s", "--stats-one-line", "--config", cfg}
		if casCompat != nil && casCompat.ExcludeFrom != "" {
			args = append(args, "--exclude-from", casCompat.ExcludeFrom)
		}
		
		if run.Summary != nil {
			var merged = map[string]any{}
			var effm map[string]any
			if v, ok := run.Summary["transferDefaults"]; ok {
				if m, ok := v.(map[string]any); ok {
					for k, val := range m {
						merged[k] = val
					}
				}
			}
			if v, ok := run.Summary["effectiveOptions"]; ok {
				if m, ok := v.(map[string]any); ok {
					effm = m
					for k, val := range m {
						merged[k] = val
					}
				}
			}
			
			if isWebDAVUnderlying(cfg, dstRemote(src)) {
				injectIfMissing := func(k string, v any) {
					if effm == nil {
						if _, ok := merged[k]; !ok {
							merged[k] = v
						}
						return
					}
					if _, ok := effm[k]; !ok {
						if _, ok2 := merged[k]; !ok2 {
							merged[k] = v
						}
					}
				}
				injectIfMissing("timeout", 24*3600)
				injectIfMissing("connTimeout", 60)
				injectIfMissing("expectContinueTimeout", 30)
				injectIfMissing("retries", 5)
				injectIfMissing("lowLevelRetries", 20)
				injectIfMissing("disableHttp2", true)
				injectIfMissing("transfers", 1)
				injectIfMissing("multiThreadStreams", 1)
			}
			
			if len(merged) > 0 {
				effOpt = merged
				args = append(args, buildFlagsFromOptions(merged)...)
			}
		}
		
		if casCompat != nil {
			casManagedRetries = true
			maxCASAttempts = configuredRetryCount(effOpt)
			if maxCASAttempts < 1 {
				maxCASAttempts = 1
			}
			args = forceFlagValue(args, "--retries", "0")
			args = forceFlagValue(args, "--low-level-retries", "0")
		}
		
		args = append(args, "--use-json-log", "--log-level", "INFO", "--stats-log-level", "INFO")
	}
	
	args = normalizeArgUnits(args)
	args = deduplicateArgs(args)
	
	return args, casCompat, casManagedRetries, maxCASAttempts, effOpt, originalCmdName, nil
}

// setupLogFiles creates log files and returns them with metadata
func (r *Runner) setupLogFiles(run store.Run, cfg string) (*os.File, string, string) {
	logsBase := config.DataDir()
	logsDir := filepath.Join(logsBase, "logs")
	_ = os.MkdirAll(logsDir, 0o755)
	
	sanitizeFilename := func(s string) string {
		s = strings.TrimSpace(s)
		if s == "" {
			s = fmt.Sprintf("task-%d", run.TaskID)
		}
		invalid := regexp.MustCompile(`[^a-zA-Z0-9\p{Han}_-]+`)
		s = invalid.ReplaceAllString(s, "_")
		runes := []rune(s)
		if len(runes) > 60 {
			s = string(runes[:60])
		}
		if s == "" {
			s = fmt.Sprintf("task-%d", run.TaskID)
		}
		return s
	}
	safeTask := sanitizeFilename(run.TaskName)
	localNow := time.Now().Local()
	datePart := localNow.Format("0102")
	timePart := localNow.Format("1504")
	subDir := filepath.Join(logsDir, fmt.Sprintf("%s-%s", safeTask, datePart))
	_ = os.MkdirAll(subDir, 0o755)
	stderrPath := filepath.Join(subDir, fmt.Sprintf("%s.log", timePart))
	stderrFile, _ := os.OpenFile(stderrPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	
	startLine := "[runner] rclone ...\n"
	missingCfg := ""
	if _, err := os.Stat(cfg); err != nil {
		missingCfg = "[runner] warn: config not found: " + cfg + "\n"
	}
	
	return stderrFile, stderrPath, startLine + missingCfg
}

// runPreflight runs preflight checks and updates stats
func (r *Runner) runPreflight(run store.Run, cfg, src string, effOpt map[string]any) {
	if b, c, e := sizeOfPaged(&adapter.CmdRunner{}, cfg, src, effOpt); e == nil {
		_ = r.updater.UpdateRun(run.ID, func(rr *store.Run) {
			if rr.Summary == nil {
				rr.Summary = map[string]any{}
			}
			rr.Summary["preflight"] = map[string]any{"totalCount": c, "totalBytes": b}
		})
	}
}

// startProcess starts the rclone process and sets up the pipes
func (r *Runner) startProcess(runCtx context.Context, args []string) (*exec.Cmd, io.ReadCloser, io.ReadCloser, io.WriteCloser, io.WriteCloser, error) {
	runner := &adapter.CmdRunner{}
	cmd := runner.CmdContext(runCtx, args...)
	
	outR, outW := io.Pipe()
	errR, errW := io.Pipe()
	
	cmd.Stdout = outW
	cmd.Stderr = errW
	
	if err := cmd.Start(); err != nil {
		return nil, nil, nil, nil, nil, err
	}
	
	return cmd, outR, errR, outW, errW, nil
}

// initializeRunState updates the run with initial state
func (r *Runner) initializeRunState(run store.Run, cmd *exec.Cmd, stderrPath string) {
	r.mu.Lock()
	r.procs[run.ID] = cmd
	r.mu.Unlock()
	
	_ = r.updater.UpdateRun(run.ID, func(rr *store.Run) {
		if rr.Summary == nil {
			rr.Summary = map[string]any{}
		}
		rr.Summary["stderrFile"] = stderrPath
		if rr.Summary["startedAt"] == nil {
			rr.Summary["startedAt"] = time.Now().Local().Format(time.RFC3339)
		}
		if cmd.Process != nil {
			rr.Summary["pid"] = cmd.Process.Pid
		}
		if p, ok := rr.Summary["progress"].(map[string]any); ok {
			if _, ok2 := p["completedFiles"]; !ok2 {
				p["completedFiles"] = float64(0)
			}
		} else {
			rr.Summary["progress"] = map[string]any{"completedFiles": float64(0)}
		}
	})
}

// waitForCompletion waits for the process to complete and handles the outcome
func (r *Runner) waitForCompletion(ctx context.Context, run store.Run, cmd *exec.Cmd, outR, errR io.ReadCloser, outW, errW io.WriteCloser, stderrFile *os.File, stderrPath string, args []string, casCompat *openlistCASCompatPlan, casManagedRetries bool, maxCASAttempts int, effOpt map[string]any, cfg, src, dst, originalCmdName, cmdName string, attemptLogOffset int64) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("goroutine panic", zap.Any("panic", r))
			}
		}()
		defer func() {
			r.mu.Lock()
			delete(r.cancelFns, run.ID)
			delete(r.procs, run.ID)
			r.mu.Unlock()
			if casCompat != nil && casCompat.ExcludeFrom != "" {
				_ = os.Remove(casCompat.ExcludeFrom)
			}
		}()
		
		attempt := 1
		fileStats := &fileProgress{m: map[string]*fileProg{}, recentCap: 100}
		var consumeWG sync.WaitGroup
		casMode := isOpenlistCASCompatible(run)
		excludeFrom := ""
		if casCompat != nil {
			excludeFrom = casCompat.ExcludeFrom
		}
		
		for {
			var newCmd *exec.Cmd
			var newOutR, newErrR io.ReadCloser
			var newOutW, newErrW io.WriteCloser
			
			if attempt == 1 {
				newCmd = cmd
				newOutR = outR
				newErrR = errR
			} else {
				newOutR, newOutW = io.Pipe()
				newErrR, newErrW = io.Pipe()
				stderrFile, _ = os.OpenFile(stderrPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
				attemptLogOffset, _ = stderrFile.Seek(0, io.SeekCurrent)
				newCmd = (&adapter.CmdRunner{}).CmdContext(ctx, args...)
				newCmd.Stdout = newOutW
				newCmd.Stderr = newErrW
				if startErr := newCmd.Start(); startErr != nil {
					break
				}
				
				r.mu.Lock()
				r.procs[run.ID] = newCmd
				r.mu.Unlock()
			}
			
			consumeWG = sync.WaitGroup{}
			consumeWG.Add(2)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Error("goroutine panic", zap.Any("panic", r))
					}
				}()
				defer consumeWG.Done()
				r.consume(run.ID, newOutR, stderrFile, false, fileStats, casMode, originalCmdName == "move", cfg, dst, excludeFrom)
			}()
			go func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Error("goroutine panic", zap.Any("panic", r))
					}
				}()
				defer consumeWG.Done()
				r.consume(run.ID, newErrR, stderrFile, true, fileStats, casMode, originalCmdName == "move", cfg, dst, excludeFrom)
			}()
			
			err := newCmd.Wait()
			if attempt == 1 {
				outW.Close()
				errW.Close()
			} else {
				newOutW.Close()
				newErrW.Close()
			}
			consumeWG.Wait()
			_ = stderrFile.Sync()
			stderrFile.Close()
			
			if err == nil && (newCmd.ProcessState == nil || newCmd.ProcessState.Success()) {
				break
			}
			
			if ctx.Err() != nil {
				break
			}
			
			if casManagedRetries && attempt <= maxCASAttempts {
				analysis := analyzeCASAttemptLogSegment(stderrPath, attemptLogOffset, casMode)
				if len(analysis.RealFailures) == 0 && len(analysis.CASMatchedPaths) > 0 {
					if fileStats != nil {
						for path := range analysis.CASMatchedPaths {
							fileStats.update(path, -1, -1, -1, 100)
							fileStats.markCopied(path)
							if r.activeMgr != nil {
								r.activeMgr.OnFileCASMatched(run.ID, path)
							}
						}
					}
					_ = r.updater.UpdateRun(run.ID, func(rr *store.Run) {
						if rr.Summary == nil {
							rr.Summary = map[string]any{}
						}
						prog, _ := rr.Summary["progress"].(map[string]any)
						if prog == nil {
							prog = map[string]any{}
						}
						if fileStats != nil {
							if lst := fileStats.copiedList(); len(lst) > 0 {
								prog["completedFiles"] = float64(len(lst))
								rr.Summary["files"] = fileStats.snapshot(100)
							}
						}
						rr.Summary["progress"] = prog
					})
					err = nil
					break
				}
				
				if len(analysis.RealFailures) > 0 && attempt < maxCASAttempts && ctx.Err() == nil {
					attempt++
					continue
				}
			}
			
			if err != nil || (newCmd.ProcessState != nil && !newCmd.ProcessState.Success()) {
				r.handleRunFailure(run, newCmd, stderrPath, cfg, dst, originalCmdName, cmdName)
				return
			}
		}
		
		if ctx.Err() != nil {
			r.handleRunStopped(run, stderrPath, cfg, dst, originalCmdName, cmdName)
			return
		}
		
		if casCompat != nil {
			if postErr := casCompat.ApplyPostActions(cfg, src, dst, originalCmdName); postErr != nil {
				_ = r.updater.UpdateRun(run.ID, func(rr *store.Run) {
					rr.Status = "failed"
					if rr.Summary == nil {
						rr.Summary = map[string]any{}
					}
					rr.Summary["finished"] = true
					rr.Summary["success"] = false
					fin := time.Now().Local()
					rr.Summary["finishedAt"] = fin.Format(time.RFC3339)
					rr.Error = postErr.Error()
				})
				r.broadcaster.Broadcast("run_status", map[string]any{
					"run_id": run.ID,
					"status": "failed",
				})
				if r.activeMgr != nil {
					r.activeMgr.RemoveState(run.ID)
				}
				go func() {
					defer func() {
						if r := recover(); r != nil {
							logger.Error("goroutine panic", zap.Any("panic", r))
						}
					}()
					r.postWebhookIfNeeded(run.ID)
				}()
				r.mu.Lock()
				delete(r.procs, run.ID)
				r.mu.Unlock()
				unregisterCancelFn(run.ID)
				return
			}
		}
		
		if isWebDAVUnderlying(cfg, dstRemote(dst)) {
			r.waitForWebDAVFiles(cfg, dst, casCompat, originalCmdName)
		}
		
		r.handleRunSuccess(run, stderrPath, cfg, dst, originalCmdName, cmdName)
	}()
}

func dstRemote(dst string) string {
	parts := strings.SplitN(dst, ":", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return dst
}

func (r *Runner) Start(ctx context.Context, run store.Run, mode, srcRemote, srcPath, dstRemote, dstPath string) error {
	r.mu.Lock()
	if _, ok := r.procs[run.ID]; ok {
		r.mu.Unlock()
		return errors.New("run already exists")
	}
	r.mu.Unlock()

	runCtx, runCancel := context.WithCancel(ctx)
	r.mu.Lock()
	r.cancelFns[run.ID] = runCancel
	r.mu.Unlock()
	registerCancelFn(run.ID, runCancel)

	src := srcRemote + ":" + strings.TrimPrefix(srcPath, "/")
	dst := dstRemote + ":" + strings.TrimPrefix(dstPath, "/")
	cmdName := strings.ToLower(mode)
	if cmdName != "copy" && cmdName != "sync" && cmdName != "move" && cmdName != "bisync" {
		cmdName = "copy"
	}

	dataDir := config.DataDir()
	cfg := os.Getenv("RCLONE_CONFIG")
	if cfg == "" {
		cfg = filepath.Join(dataDir, "rclone.conf")
	}

	args, casCompat, casManagedRetries, maxCASAttempts, effOpt, originalCmdName, err := r.buildTransferArgs(mode, src, dst, cfg, run)
	if err != nil {
		return err
	}

	stderrFile, stderrPath, initialLog := r.setupLogFiles(run, cfg)
	stderrFile.WriteString(initialLog)
	if effOpt != nil {
		if b, _ := json.Marshal(effOpt); len(b) > 0 {
			stderrFile.WriteString("[runner] effectiveOptions " + string(b) + "\n")
		}
	}
	attemptLogOffset, _ := stderrFile.Seek(0, io.SeekCurrent)

	r.runPreflight(run, cfg, src, effOpt)

	cmd, outR, errR, outW, errW, err := r.startProcess(runCtx, args)
	if err != nil {
		return err
	}

	r.initializeRunState(run, cmd, stderrPath)
	r.waitForCompletion(runCtx, run, cmd, outR, errR, outW, errW, stderrFile, stderrPath, args, casCompat, casManagedRetries, maxCASAttempts, effOpt, cfg, src, dst, originalCmdName, cmdName, attemptLogOffset)

	return nil
}

func (r *Runner) Stop(runID int64) error {
	r.mu.Lock()
	if cf, ok := r.cancelFns[runID]; ok {
		cf()
	}
	cmd, ok := r.procs[runID]
	r.mu.Unlock()
	if !ok || cmd == nil || cmd.Process == nil {
		return nil
	}
	_ = cmd.Process.Signal(syscall.SIGINT)
	if wait(cmd, 10*time.Second) {
		return nil
	}
	if runtime.GOOS != "windows" {
		_ = cmd.Process.Signal(syscall.SIGTERM)
	}
	if wait(cmd, 10*time.Second) {
		return nil
	}
	_ = cmd.Process.Kill()
	return nil
}

func wait(cmd *exec.Cmd, d time.Duration) bool {
	ch := make(chan struct{}, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("goroutine panic", zap.Any("panic", r))
			}
		}()
		_ = cmd.Wait()
		close(ch)
	}()
	select {
	case <-ch:
		return true
	case <-time.After(d):
		return false
	}
}

// normalizeArgUnits ensures flags have proper units and converts separators
func normalizeArgUnits(args []string) []string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--buffer-size" || args[i] == "--bwlimit" {
			n := strings.TrimSpace(args[i+1])
			if args[i] == "--bwlimit" {
				if strings.Contains(n, ";") {
					n = strings.ReplaceAll(n, ";", " ")
				}
				args[i+1] = n
			}
			pureNum := n != ""
			for _, ch := range n {
				if ch < '0' || ch > '9' {
					pureNum = false
					break
				}
			}
			if pureNum {
				args[i+1] = n + "M"
			}
		}
	}
	return args
}

// deduplicateArgs removes duplicate flags like --bwlimit
func deduplicateArgs(args []string) []string {
	last := -1
	for i := 0; i < len(args); i++ {
		if args[i] == "--bwlimit" {
			last = i
		}
	}
	if last >= 0 {
		newArgs := make([]string, 0, len(args))
		for i := 0; i < len(args); {
			if args[i] == "--bwlimit" && i != last {
				i += 2
				continue
			}
			newArgs = append(newArgs, args[i])
			i++
		}
		args = newArgs
	}
	return args
}

// handleRunFailure handles a failed run
func (r *Runner) handleRunFailure(run store.Run, cmd *exec.Cmd, stderrPath, cfg, dst, originalCmdName, cmdName string) {
	_ = r.updater.UpdateRun(run.ID, func(rr *store.Run) {
		rr.Status = "failed"
		if rr.Summary == nil {
			rr.Summary = map[string]any{}
		}
		rr.Summary["finished"] = true
		rr.Summary["success"] = false
		fin := time.Now().Local()
		rr.Summary["finishedAt"] = fin.Format(time.RFC3339)
		finalSummary := map[string]any{}
		var start time.Time
		if s, ok := rr.Summary["startedAt"].(string); ok {
			if t, e := time.Parse(time.RFC3339, s); e == nil {
				start = t
			}
		}
		if !start.IsZero() {
			finalSummary["startAt"] = start.Format(time.RFC3339)
		}
		finalSummary["finishedAt"] = fin.Format(time.RFC3339)
		durSec := int64(0)
		if !start.IsZero() {
			durSec = int64(fin.Sub(start).Seconds())
		}
		if durSec < 0 {
			durSec = 0
		}
		finalSummary["durationSec"] = durSec
		finalSummary["durationText"] = util.HumanDuration(durSec)
		finalSummary["result"] = "failed"
		
		var prog map[string]any
		if p, ok := rr.Summary["progress"].(map[string]any); ok {
			prog = p
		}
		var bytes, total int64
		if prog != nil {
			if v, ok := prog["bytes"].(float64); ok {
				bytes = int64(v)
			}
			if v, ok := prog["totalBytes"].(float64); ok {
				total = int64(v)
			}
		}
		finalSummary["transferredBytes"] = bytes
		finalSummary["totalBytes"] = total
		avg := int64(0)
		if durSec > 0 {
			avg = bytes / durSec
		}
		finalSummary["avgSpeedBps"] = avg
		
		files := []map[string]any{}
		counts := map[string]int{"copied": 0, "deleted": 0, "skipped": 0, "failed": 0, "total": 0}
		if p, ok := rr.Summary["stderrFile"].(string); ok && p != "" {
			files, counts = buildFinalSummaryFilesFromLog(p, isOpenlistCASCompatible(run), strings.ToLower(cmdName) == "move")
		}
		r.enrichFilesSizesAsync(run.ID, files, dst, cfg, isOpenlistCASCompatible(run))
		rr.Summary["finalSummary"] = map[string]any{"counts": counts, "files": files, "startAt": finalSummary["startAt"], "finishedAt": finalSummary["finishedAt"], "durationSec": durSec, "durationText": util.HumanDuration(durSec), "result": "failed", "transferredBytes": bytes, "totalBytes": total, "avgSpeedBps": avg}
	})
	r.broadcaster.Broadcast("run_status", map[string]any{
		"run_id": run.ID,
		"status": "failed",
	})
	if r.activeMgr != nil {
		r.activeMgr.RemoveState(run.ID)
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("goroutine panic", zap.Any("panic", r))
			}
		}()
		r.postWebhookIfNeeded(run.ID)
	}()
	r.mu.Lock()
	delete(r.procs, run.ID)
	r.mu.Unlock()
	unregisterCancelFn(run.ID)
}

// handleRunStopped handles a stopped run
func (r *Runner) handleRunStopped(run store.Run, stderrPath, cfg, dst, originalCmdName, cmdName string) {
	_ = r.updater.UpdateRun(run.ID, func(rr *store.Run) {
		rr.Status = "stopped"
		if rr.Summary == nil {
			rr.Summary = map[string]any{}
		}
		rr.Summary["finished"] = true
		rr.Summary["success"] = false
		fin := time.Now().Local()
		rr.Summary["finishedAt"] = fin.Format(time.RFC3339)
		finalSummary := map[string]any{}
		var start time.Time
		if s, ok := rr.Summary["startedAt"].(string); ok {
			if t, e := time.Parse(time.RFC3339, s); e == nil {
				start = t
			}
		}
		if !start.IsZero() {
			finalSummary["startAt"] = start.Format(time.RFC3339)
		}
		finalSummary["finishedAt"] = fin.Format(time.RFC3339)
		durSec := int64(0)
		if !start.IsZero() {
			durSec = int64(fin.Sub(start).Seconds())
		}
		if durSec < 0 {
			durSec = 0
		}
		finalSummary["durationSec"] = durSec
		finalSummary["durationText"] = util.HumanDuration(durSec)
		finalSummary["result"] = "stopped"
		
		var prog map[string]any
		if p, ok := rr.Summary["progress"].(map[string]any); ok {
			prog = p
		}
		var bytes, total int64
		if prog != nil {
			if v, ok := prog["bytes"].(float64); ok {
				bytes = int64(v)
			}
			if v, ok := prog["totalBytes"].(float64); ok {
				total = int64(v)
			}
		}
		finalSummary["transferredBytes"] = bytes
		finalSummary["totalBytes"] = total
		avg := int64(0)
		if durSec > 0 {
			avg = bytes / durSec
		}
		finalSummary["avgSpeedBps"] = avg
		
		files := []map[string]any{}
		counts := map[string]int{"copied": 0, "deleted": 0, "skipped": 0, "failed": 0, "total": 0}
		if p, ok := rr.Summary["stderrFile"].(string); ok && p != "" {
			files, counts = buildFinalSummaryFilesFromLog(p, isOpenlistCASCompatible(run), strings.ToLower(cmdName) == "move")
		}
		r.enrichFilesSizesAsync(run.ID, files, dst, cfg, isOpenlistCASCompatible(run))
		finalSummary["counts"] = counts
		finalSummary["files"] = files
		rr.Summary["finalSummary"] = finalSummary
	})
	r.broadcaster.Broadcast("run_status", map[string]any{
		"run_id": run.ID,
		"status": "stopped",
	})
	if r.activeMgr != nil {
		r.activeMgr.RemoveState(run.ID)
	}
	r.mu.Lock()
	delete(r.procs, run.ID)
	r.mu.Unlock()
	unregisterCancelFn(run.ID)
}

// handleRunSuccess handles a successful run
func (r *Runner) handleRunSuccess(run store.Run, stderrPath, cfg, dst, originalCmdName, cmdName string) {
	_ = r.updater.UpdateRun(run.ID, func(rr *store.Run) {
		rr.Status = "finished"
		if rr.Summary == nil {
			rr.Summary = map[string]any{}
		}
		rr.Summary["finished"] = true
		rr.Summary["success"] = true
		fin := time.Now().Local()
		rr.Summary["finishedAt"] = fin.Format(time.RFC3339)
		finalSummary := map[string]any{}
		var start time.Time
		if s, ok := rr.Summary["startedAt"].(string); ok {
			if t, e := time.Parse(time.RFC3339, s); e == nil {
				start = t
			}
		}
		if !start.IsZero() {
			finalSummary["startAt"] = start.Format(time.RFC3339)
		}
		finalSummary["finishedAt"] = fin.Format(time.RFC3339)
		durSec := int64(0)
		if !start.IsZero() {
			durSec = int64(fin.Sub(start).Seconds())
		}
		if durSec < 0 {
			durSec = 0
		}
		finalSummary["durationSec"] = durSec
		finalSummary["durationText"] = util.HumanDuration(durSec)
		finalSummary["result"] = "success"
		
		var prog map[string]any
		if p, ok := rr.Summary["progress"].(map[string]any); ok {
			prog = p
		}
		var bytes, total int64
		if prog != nil {
			if v, ok := prog["bytes"].(float64); ok {
				bytes = int64(v)
			}
			if v, ok := prog["totalBytes"].(float64); ok {
				total = int64(v)
			}
		}
		finalSummary["transferredBytes"] = bytes
		finalSummary["totalBytes"] = total
		avg := int64(0)
		if durSec > 0 {
			avg = bytes / durSec
		}
		finalSummary["avgSpeedBps"] = avg
		
		files := []map[string]any{}
		counts := map[string]int{"copied": 0, "deleted": 0, "skipped": 0, "failed": 0, "total": 0}
		if p, ok := rr.Summary["stderrFile"].(string); ok && p != "" {
			files, counts = buildFinalSummaryFilesFromLog(p, isOpenlistCASCompatible(run), strings.ToLower(cmdName) == "move")
		}
		r.enrichFilesSizesAsync(run.ID, files, dst, cfg, isOpenlistCASCompatible(run))
		finalSummary["counts"] = counts
		finalSummary["files"] = files
		rr.Summary["finalSummary"] = finalSummary
	})
	r.broadcaster.Broadcast("run_status", map[string]any{
		"run_id": run.ID,
		"status": "finished",
	})
	if r.activeMgr != nil {
		r.activeMgr.RemoveState(run.ID)
	}
	if run.TaskMode == "bisync" {
		ClearBisyncResyncFlag(r.updater, run.TaskID)
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("goroutine panic", zap.Any("panic", r))
			}
		}()
		r.postWebhookIfNeeded(run.ID)
	}()
	r.mu.Lock()
	delete(r.procs, run.ID)
	r.mu.Unlock()
}

// waitForWebDAVFiles waits for WebDAV files to become visible
func (r *Runner) waitForWebDAVFiles(cfg, dst string, casCompat *openlistCASCompatPlan, originalCmdName string) {
	interval := config.GetFinishWaitInterval()
	timeout := config.GetFinishWaitTimeout()
	if timeout > 0 {
		vr := &adapter.CmdRunner{}
		deadline := time.Now().Add(timeout)
		expected := expectedVisibleDestinationPaths(casCompat, originalCmdName)
		for time.Now().Before(deadline) {
			allOk := true
			args := []string{"lsjson", dst, "--config", cfg, "--files-only", "--recursive"}
			out, _, e := vr.Run(context.Background(), args...)
			if e != nil {
				allOk = false
			}
			var arr []map[string]any
			if json.Unmarshal([]byte(out), &arr) != nil {
				allOk = false
			}
			if allOk && len(expected) > 0 {
				visible := normalizeVisibleTargetPaths(arr, casCompat != nil)
				allOk = areAllExpectedPathsVisible(expected, visible)
			}
			if allOk {
				break
			}
			time.Sleep(interval)
		}
	}
}

