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

func sanitizeFilename(s string, taskID int64) string {
	s = strings.TrimSpace(s)
	if s == "" {
		s = fmt.Sprintf("task-%d", taskID)
	}
	invalid := regexp.MustCompile(`[^a-zA-Z0-9\p{Han}_-]+`)
	s = invalid.ReplaceAllString(s, "_")
	// 截断到 60 字符以内，避免过长
	r := []rune(s)
	if len(r) > 60 {
		s = string(r[:60])
	}
	if s == "" {
		s = fmt.Sprintf("task-%d", taskID)
	}
	return s
}

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

	runner := &adapter.CmdRunner{}
	src := srcRemote + ":" + strings.TrimPrefix(srcPath, "/")
	dst := dstRemote + ":" + strings.TrimPrefix(dstPath, "/")
	cmdName := strings.ToLower(mode)
	if cmdName != "copy" && cmdName != "sync" && cmdName != "move" && cmdName != "bisync" {
		cmdName = "copy"
	}
	originalCmdName := cmdName
	isBisyncMode := cmdName == "bisync"
	// Resolve config path
	dataDir := config.DataDir()
	cfg := os.Getenv("RCLONE_CONFIG")
	if cfg == "" {
		cfg = filepath.Join(dataDir, "rclone.conf")
	}
	var casCompat *openlistCASCompatPlan
	if isOpenlistCASCompatible(run) && !isBisyncMode {
		plan, err := buildOpenlistCASCompatPlan(cfg, src, dst, cmdName)
		if err != nil {
			return fmt.Errorf("prepare openlist-cas compatibility failed: %w", err)
		}
		casCompat = plan
		if cmdName == "sync" {
			cmdName = "copy"
		}
	}
	// Base args：非交互环境使用 --stats-one-line（不与 --progress 同用）
	// 降低默认日志级别：从 -vv 改为 -v，显著减少日志行数和解析/写库开销
	args := []string{cmdName, src, dst, "--stats", "1s", "--stats-one-line", "--config", cfg}
	// 处理 bisync 特定参数
	if isBisyncMode {
		// 设置 bisync 工作目录：/app/data/bisync/<任务名>
		bisyncDir := filepath.Join(dataDir, "bisync", sanitizeFilename(run.TaskName, run.TaskID))
		_ = os.MkdirAll(bisyncDir, 0o755)
		args = append(args, "--workdir", bisyncDir)
		// 处理 bisync 选项
		if run.Summary != nil {
			if bisyncOpts, ok := run.Summary["bisyncOptions"].(map[string]any); ok && bisyncOpts != nil {
				args = append(args, buildBisyncFlagsFromOptions(bisyncOpts)...)
			}
		}
	}
	if casCompat != nil && casCompat.ExcludeFrom != "" {
		args = append(args, "--exclude-from", casCompat.ExcludeFrom)
	}
	// attach advanced options: merge transferDefaults (global) <- effectiveOptions (task)，并对 WebDAV 目标注入稳态默认（未显式配置时）
	var effOpt map[string]any
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
		// WebDAV 稳态参数（当目标底层是 WebDAV）
		// 显式设置（effectiveOptions）优先：仅在用户未显式设置时注入建议默认；不再做“下限兜底”强制覆盖
		if isWebDAVUnderlying(cfg, dstRemote) {
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
			// 建议默认（保守）：
			injectIfMissing("timeout", 24*3600)
			injectIfMissing("connTimeout", 60)
			injectIfMissing("expectContinueTimeout", 30)
			injectIfMissing("retries", 5)
			injectIfMissing("lowLevelRetries", 20)
			injectIfMissing("disableHttp2", true)
			// 并发/多线程：仅当用户未显式设置时给出建议默认，用户设置优先生效
			injectIfMissing("transfers", 1)
			injectIfMissing("multiThreadStreams", 1)
		}
		if len(merged) > 0 {
			effOpt = merged
			args = append(args, buildFlagsFromOptions(merged)...)
		}
	}
	casManagedRetries := false
	maxCASAttempts := 1
	if casCompat != nil {
		casManagedRetries = true
		maxCASAttempts = configuredRetryCount(effOpt)
		if maxCASAttempts < 1 {
			maxCASAttempts = 1
		}
		args = forceFlagValue(args, "--retries", "0")
		args = forceFlagValue(args, "--low-level-retries", "0")
	}
	// 强制启用 JSON 日志：作为系统默认行为，不再提供任务级开关。
	args = append(args, "--use-json-log", "--log-level", "INFO", "--stats-log-level", "INFO")
	// 二次兜底：如 --buffer-size/--bwlimit 后是纯数字，自动补单位（M）；
	// 同时将 --bwlimit 的分号分隔写法转为空格分隔，保证多时段正确识别
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--buffer-size" || args[i] == "--bwlimit" {
			n := strings.TrimSpace(args[i+1])
			if args[i] == "--bwlimit" {
				// 兼容 07:30,2M;17:40,2M;23:00,3M → 07:30,2M 17:40,2M 23:00,3M
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
	// 去重：--bwlimit 若出现多次，仅保留最后一次（后者覆盖前者）
	{
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
					// 跳过成对参数
					i += 2
					continue
				}
				newArgs = append(newArgs, args[i])
				i++
			}
			args = newArgs
		}
	}
	// header will be written after files are opened below
	startLine := "[runner] rclone " + strings.Join(args, " ") + "\n"
	missingCfg := ""
	if _, err := os.Stat(cfg); err != nil {
		missingCfg = "[runner] warn: config not found: " + cfg + "\n"
	}

	logsBase := config.DataDir()
	logsDir := filepath.Join(logsBase, "logs")
	_ = os.MkdirAll(logsDir, 0o755)
	// 日志目录与文件：logs/<任务名-MMDD>/<HHMM>.log（stdout 也合并写入该文件）
	safeTask := sanitizeFilename(run.TaskName, run.TaskID)
	localNow := time.Now().Local()
	datePart := localNow.Format("0102") // MMDD
	timePart := localNow.Format("1504") // HHMM
	subDir := filepath.Join(logsDir, fmt.Sprintf("%s-%s", safeTask, datePart))
	_ = os.MkdirAll(subDir, 0o755)
	stderrPath := filepath.Join(subDir, fmt.Sprintf("%s.log", timePart))
	stderrFile, _ := os.OpenFile(stderrPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)

	// Mandatory preflight: sequential pagination by top-level dirs to stabilize totals
	if b, c, e := sizeOfPaged(&adapter.CmdRunner{}, cfg, src, effOpt); e == nil {
		_ = r.updater.UpdateRun(run.ID, func(rr *store.Run) {
			if rr.Summary == nil {
				rr.Summary = map[string]any{}
			}
			rr.Summary["preflight"] = map[string]any{"totalCount": c, "totalBytes": b}
		})
	}

	cmd := runner.CmdContext(runCtx, args...)
	// fan-out: write to parser via io.Pipe（由 consumer 单点写入同一文件）
	outR, outW := io.Pipe()
	errR, errW := io.Pipe()
	// 写入头信息到单一日志文件
	_, _ = stderrFile.WriteString(startLine)
	if effOpt != nil {
		if b, _ := json.Marshal(effOpt); len(b) > 0 {
			optsLine := "[runner] effectiveOptions " + string(b) + "\n"
			_, _ = stderrFile.WriteString(optsLine)
		}
	}
	if missingCfg != "" {
		_, _ = stderrFile.WriteString(missingCfg)
	}
	attemptLogOffset, _ := stderrFile.Seek(0, io.SeekCurrent)
	cmd.Stdout = outW
	cmd.Stderr = errW
	if err := cmd.Start(); err != nil {
		return err
	}

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
		// 初始化运行中统计：已完成文件数
		if p, ok := rr.Summary["progress"].(map[string]any); ok {
			if _, ok2 := p["completedFiles"]; !ok2 {
				p["completedFiles"] = float64(0)
			}
		} else {
			rr.Summary["progress"] = map[string]any{"completedFiles": float64(0)}
		}
	})

	// 两路都写入同一日志文件，并启用 one-line 解析 + 按文件统计
	fileStats := &fileProgress{m: map[string]*fileProg{}, recentCap: 100}
	// 仅解析 stderr（rclone 进度通常在 stderr），stdout 只写文件，减少重复解析/写库
	casMode := isOpenlistCASCompatible(run)
	excludeFrom := ""
	if casCompat != nil {
		excludeFrom = casCompat.ExcludeFrom
	}
	var consumeWG sync.WaitGroup
	consumeWG.Add(2)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("goroutine panic", zap.Any("panic", r))
			}
		}()
		defer consumeWG.Done()
		r.consume(run.ID, outR, stderrFile, false, fileStats, casMode, originalCmdName == "move", cfg, dst, excludeFrom)
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("goroutine panic", zap.Any("panic", r))
			}
		}()
		defer consumeWG.Done()
		r.consume(run.ID, errR, stderrFile, true, fileStats, casMode, originalCmdName == "move", cfg, dst, excludeFrom)
	}()
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
		for {
			err := cmd.Wait()
			outW.Close()
			errW.Close()
			consumeWG.Wait()
			_ = stderrFile.Sync()
			stderrFile.Close()
			if err == nil && (cmd.ProcessState == nil || cmd.ProcessState.Success()) {
				break
			}
			// 如果已触发停止，不再重试
			if runCtx.Err() != nil {
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
				if len(analysis.RealFailures) > 0 && attempt < maxCASAttempts && runCtx.Err() == nil {
					attempt++
					outR, outW = io.Pipe()
					errR, errW = io.Pipe()
					stderrFile, _ = os.OpenFile(stderrPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
					attemptLogOffset, _ = stderrFile.Seek(0, io.SeekCurrent)
					cmd = runner.CmdContext(runCtx, args...)
					cmd.Stdout = outW
					cmd.Stderr = errW
					if startErr := cmd.Start(); startErr != nil {
						err = startErr
						break
					}
					r.mu.Lock()
					r.procs[run.ID] = cmd
					r.mu.Unlock()
					consumeWG = sync.WaitGroup{}
					consumeWG.Add(2)
					go func() {
						defer func() {
							if r := recover(); r != nil {
								logger.Error("goroutine panic", zap.Any("panic", r))
							}
						}()
						defer consumeWG.Done()
						r.consume(run.ID, outR, stderrFile, false, fileStats, casMode, originalCmdName == "move", cfg, dst, excludeFrom)
					}()
					go func() {
						defer func() {
							if r := recover(); r != nil {
								logger.Error("goroutine panic", zap.Any("panic", r))
							}
						}()
						defer consumeWG.Done()
						r.consume(run.ID, errR, stderrFile, true, fileStats, casMode, originalCmdName == "move", cfg, dst, excludeFrom)
					}()
					continue
				}
			}
			if err != nil || (cmd.ProcessState != nil && !cmd.ProcessState.Success()) {
				_ = r.updater.UpdateRun(run.ID, func(rr *store.Run) {
					rr.Status = "failed"
					if rr.Summary == nil {
						rr.Summary = map[string]any{}
					}
					rr.Summary["finished"] = true
					rr.Summary["success"] = false
					fin := time.Now().Local()
					rr.Summary["finishedAt"] = fin.Format(time.RFC3339)
					// 冻结最终总结（失败态）
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
					// 体量/均速：失败态 finalSummary 也只从 progress 读取
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
					// 文件明细（从 stderrFile 解析）
					files := []map[string]any{}
					counts := map[string]int{"copied": 0, "deleted": 0, "skipped": 0, "failed": 0, "total": 0, "conflicts": 0}
					if p, ok := rr.Summary["stderrFile"].(string); ok && p != "" {
						if isBisyncMode {
							files, counts = buildBisyncSummaryFilesFromLog(p)
						} else {
							files, counts = buildFinalSummaryFilesFromLog(p, isOpenlistCASCompatible(run), strings.ToLower(cmdName) == "move")
						}
					}
					// 异步补全文件大小，不阻塞状态更新（bisync 模式下不进行大小补全，因为双向同步可能没有明确的目标）
					if !isBisyncMode {
						r.enrichFilesSizesAsync(run.ID, files, dst, cfg, isOpenlistCASCompatible(run))
					}
					rr.Summary["finalSummary"] = map[string]any{"counts": counts, "files": files, "startAt": finalSummary["startAt"], "finishedAt": finalSummary["finishedAt"], "durationSec": durSec, "durationText": util.HumanDuration(durSec), "result": "failed", "transferredBytes": bytes, "totalBytes": total, "avgSpeedBps": avg, "isBisync": true}
				})
				r.broadcaster.Broadcast("run_status", map[string]any{
					"run_id": run.ID,
					"status": "failed",
				})
				if r.activeMgr != nil {
					r.activeMgr.RemoveState(run.ID)
				}
				// fire webhook for failed run
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
		if runCtx.Err() != nil {
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
				counts := map[string]int{"copied": 0, "deleted": 0, "skipped": 0, "failed": 0, "total": 0, "conflicts": 0}
				if p, ok := rr.Summary["stderrFile"].(string); ok && p != "" {
					if isBisyncMode {
						files, counts = buildBisyncSummaryFilesFromLog(p)
					} else {
						files, counts = buildFinalSummaryFilesFromLog(p, isOpenlistCASCompatible(run), strings.ToLower(cmdName) == "move")
					}
				}
				if !isBisyncMode {
					r.enrichFilesSizesAsync(run.ID, files, dst, cfg, isOpenlistCASCompatible(run))
				}
				finalSummary["counts"] = counts
				finalSummary["files"] = files
				if isBisyncMode {
					finalSummary["isBisync"] = true
				}
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
		// WebDAV 完成确认（copy/sync/move 通用）：在目录可读基础上，对预期文件做可见性确认。
		if isWebDAVUnderlying(cfg, dstRemote) {
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
						visible := normalizeVisibleTargetPaths(arr, isOpenlistCASCompatible(run))
						allOk = areAllExpectedPathsVisible(expected, visible)
					}
					if allOk {
						break
					}
					time.Sleep(interval)
				}
			}
		}
		_ = r.updater.UpdateRun(run.ID, func(rr *store.Run) {
			rr.Status = "finished"
			if rr.Summary == nil {
				rr.Summary = map[string]any{}
			}
			rr.Summary["finished"] = true
			rr.Summary["success"] = true
			fin := time.Now().Local()
			rr.Summary["finishedAt"] = fin.Format(time.RFC3339)
			// 生成并冻结最终总结 finalSummary（仅在结束时一次性写入）
			finalSummary := map[string]any{}
			// 时间
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
			// 结果
			finalSummary["result"] = "success"
			// 体量/均速：完成态只从 progress 读取
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
			// 从 stderrFile 解析文件级明细
			files := []map[string]any{}
			counts := map[string]int{"copied": 0, "deleted": 0, "skipped": 0, "failed": 0, "total": 0, "conflicts": 0}
			if p, ok := rr.Summary["stderrFile"].(string); ok && p != "" {
				if isBisyncMode {
					files, counts = buildBisyncSummaryFilesFromLog(p)
				} else {
					files, counts = buildFinalSummaryFilesFromLog(p, isOpenlistCASCompatible(run), strings.ToLower(cmdName) == "move")
				}
			}
			// 异步补全文件大小，不阻塞状态更新。
			// finalSummary 只服务于历史详情 / 最终总结展示；
			// 不要再往回恢复 stableProgress / cardSummary 这类完成态兼容字段，
			// 以免运行中链路与任务卡片完成态再次发生语义混用。
			if !isBisyncMode {
				r.enrichFilesSizesAsync(run.ID, files, dst, cfg, isOpenlistCASCompatible(run))
			}
			finalSummary["counts"] = counts
			finalSummary["files"] = files
			if isBisyncMode {
				finalSummary["isBisync"] = true
			}
			rr.Summary["finalSummary"] = finalSummary

		})
		r.broadcaster.Broadcast("run_status", map[string]any{
			"run_id": run.ID,
			"status": "finished",
		})
		if r.activeMgr != nil {
			r.activeMgr.RemoveState(run.ID)
		}
		// fire webhook for successful run
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
	}()
	return nil
}

func (r *Runner) Stop(runID int64) error {
	// Cancel context first to signal retry loops to abort
	r.mu.Lock()
	if cf, ok := r.cancelFns[runID]; ok {
		cf()
	}
	cmd, ok := r.procs[runID]
	r.mu.Unlock()
	if !ok || cmd == nil || cmd.Process == nil {
		// Process already exited (e.g. retry loop about to start a new one);
		// context cancellation above will abort any pending retry.
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
