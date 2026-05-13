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
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"go.uber.org/zap"
	"rcloneflow/internal/active_transfer"
	"rcloneflow/internal/adapter"
	"rcloneflow/internal/config"
	"rcloneflow/internal/logger"
	"rcloneflow/internal/store"
	"rcloneflow/internal/websocket"
)

// Runner manages CLI transfers with progress/logs and stop control.
type Runner struct {
	mu              sync.Mutex
	procs           map[int64]*exec.Cmd
	db              *store.DB
	activeMgr       *active_transfer.Manager
	casVerifier     func(cfg, dst, rel string) (bool, error)
	casVerifyDelays []time.Duration
	casExcludeMu    sync.Mutex
}

func New(db *store.DB, activeMgr ...*active_transfer.Manager) *Runner {
	var mgr *active_transfer.Manager
	if len(activeMgr) > 0 {
		mgr = activeMgr[0]
	}
	return &Runner{
		procs:           map[int64]*exec.Cmd{},
		db:              db,
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

	runner := &adapter.CmdRunner{}
	src := srcRemote + ":" + strings.TrimPrefix(srcPath, "/")
	dst := dstRemote + ":" + strings.TrimPrefix(dstPath, "/")
	cmdName := strings.ToLower(mode)
	if cmdName != "copy" && cmdName != "sync" && cmdName != "move" {
		cmdName = "copy"
	}
	originalCmdName := cmdName
	// Resolve config path
	dataDir := os.Getenv("APP_DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	cfg := os.Getenv("RCLONE_CONFIG")
	if cfg == "" {
		cfg = filepath.Join(dataDir, "rclone.conf")
	}
	var casCompat *openlistCASCompatPlan
	if isOpenlistCASCompatible(run) {
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
		args = forceFlagValue(args, "--retries", "1")
		args = forceFlagValue(args, "--low-level-retries", "1")
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

	logsBase := os.Getenv("APP_DATA_DIR")
	if logsBase == "" {
		logsBase = "./data"
	}
	logsDir := filepath.Join(logsBase, "logs")
	_ = os.MkdirAll(logsDir, 0o755)
	// 日志目录与文件：logs/<任务名-MMDD>/<HHMM>.log（stdout 也合并写入该文件）
	sanitizeFilename := func(s string) string {
		s = strings.TrimSpace(s)
		if s == "" {
			s = fmt.Sprintf("task-%d", run.TaskID)
		}
		invalid := regexp.MustCompile(`[^a-zA-Z0-9\p{Han}_-]+`)
		s = invalid.ReplaceAllString(s, "_")
		// 截断到 60 字符以内，避免过长
		r := []rune(s)
		if len(r) > 60 {
			s = string(r[:60])
		}
		if s == "" {
			s = fmt.Sprintf("task-%d", run.TaskID)
		}
		return s
	}
	safeTask := sanitizeFilename(run.TaskName)
	localNow := time.Now().Local()
	datePart := localNow.Format("0102") // MMDD
	timePart := localNow.Format("1504") // HHMM
	subDir := filepath.Join(logsDir, fmt.Sprintf("%s-%s", safeTask, datePart))
	_ = os.MkdirAll(subDir, 0o755)
	stderrPath := filepath.Join(subDir, fmt.Sprintf("%s.log", timePart))
	stderrFile, _ := os.OpenFile(stderrPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)

	// Mandatory preflight: sequential pagination by top-level dirs to stabilize totals
	if b, c, e := sizeOfPaged(&adapter.CmdRunner{}, cfg, src, effOpt); e == nil {
		_ = r.db.UpdateRun(run.ID, func(rr *store.Run) {
			if rr.Summary == nil {
				rr.Summary = map[string]any{}
			}
			rr.Summary["preflight"] = map[string]any{"totalCount": c, "totalBytes": b}
		})
	}

	// dynamic progress flush thresholds (read on each run start; consumer also re-reads periodically)
	_ = config.GetProgressFlushInterval()
	_ = config.GetProgressFlushDeltaPct()
	_ = config.GetProgressFlushDeltaBytes()

	cmd := runner.CmdContext(ctx, args...)
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
	cmd.Stdout = outW
	cmd.Stderr = errW
	if err := cmd.Start(); err != nil {
		return err
	}

	r.mu.Lock()
	r.procs[run.ID] = cmd
	r.mu.Unlock()
	_ = r.db.UpdateRun(run.ID, func(rr *store.Run) {
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
	fileStats := &fileProgress{m: map[string]*fileProg{}}
	// 仅解析 stderr（rclone 进度通常在 stderr），stdout 只写文件，减少重复解析/写库
	casMode := isOpenlistCASCompatible(run)
	excludeFrom := ""
	if casCompat != nil {
		excludeFrom = casCompat.ExcludeFrom
	}
	var consumeWG sync.WaitGroup
	consumeWG.Add(2)
	go func() {
		defer consumeWG.Done()
		r.consume(run.ID, outR, stderrFile, false, fileStats, casMode, originalCmdName == "move", cfg, dst, excludeFrom)
	}()
	go func() {
		defer consumeWG.Done()
		r.consume(run.ID, errR, stderrFile, true, fileStats, casMode, originalCmdName == "move", cfg, dst, excludeFrom)
	}()
	go func() {
		defer func() {
			if casCompat != nil && casCompat.ExcludeFrom != "" {
				_ = os.Remove(casCompat.ExcludeFrom)
			}
		}()
		attemptLogOffset, _ := stderrFile.Seek(0, io.SeekCurrent)
		attempt := 1
		for {
			err := cmd.Wait()
			outW.Close()
			errW.Close()
			consumeWG.Wait()
			stderrFile.Close()
			if err == nil && (cmd.ProcessState == nil || cmd.ProcessState.Success()) {
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
					_ = r.db.UpdateRun(run.ID, func(rr *store.Run) {
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
				if len(analysis.RealFailures) > 0 && attempt < maxCASAttempts {
					attempt++
					outR, outW = io.Pipe()
					errR, errW = io.Pipe()
					stderrFile, _ = os.OpenFile(stderrPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
					attemptLogOffset, _ = stderrFile.Seek(0, io.SeekCurrent)
					cmd = runner.CmdContext(ctx, args...)
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
						defer consumeWG.Done()
						r.consume(run.ID, outR, stderrFile, false, fileStats, casMode, originalCmdName == "move", cfg, dst, excludeFrom)
					}()
					go func() {
						defer consumeWG.Done()
						r.consume(run.ID, errR, stderrFile, true, fileStats, casMode, originalCmdName == "move", cfg, dst, excludeFrom)
					}()
					continue
				}
			}
			if err != nil || (cmd.ProcessState != nil && !cmd.ProcessState.Success()) {
				_ = r.db.UpdateRun(run.ID, func(rr *store.Run) {
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
				finalSummary["durationText"] = humanDuration(durSec)
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
				counts := map[string]int{"copied": 0, "deleted": 0, "skipped": 0, "failed": 0, "total": 0}
				if p, ok := rr.Summary["stderrFile"].(string); ok && p != "" {
					files, counts = buildFinalSummaryFilesFromLog(p, isOpenlistCASCompatible(run), strings.ToLower(cmdName) == "move")
				}
				// 异步补全文件大小，不阻塞状态更新
				r.enrichFilesSizesAsync(run.ID, files, dst, cfg, isOpenlistCASCompatible(run))
				rr.Summary["finalSummary"] = map[string]any{"counts": counts, "files": files, "startAt": finalSummary["startAt"], "finishedAt": finalSummary["finishedAt"], "durationSec": durSec, "durationText": humanDuration(durSec), "result": "failed", "transferredBytes": bytes, "totalBytes": total, "avgSpeedBps": avg}
			})
			websocket.Broadcast("run_status", map[string]any{
				"run_id": run.ID,
				"status": "failed",
			})
			if r.activeMgr != nil {
				r.activeMgr.RemoveState(run.ID)
			}
			// fire webhook for failed run
			go r.postWebhookIfNeeded(run.ID)
			r.mu.Lock()
			delete(r.procs, run.ID)
			r.mu.Unlock()
			return
			}
		}
		if casCompat != nil {
			if postErr := casCompat.ApplyPostActions(cfg, src, dst, originalCmdName); postErr != nil {
				_ = r.db.UpdateRun(run.ID, func(rr *store.Run) {
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
				websocket.Broadcast("run_status", map[string]any{
					"run_id": run.ID,
					"status": "failed",
				})
				if r.activeMgr != nil {
					r.activeMgr.RemoveState(run.ID)
				}
				go r.postWebhookIfNeeded(run.ID)
				r.mu.Lock()
				delete(r.procs, run.ID)
				r.mu.Unlock()
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
		_ = r.db.UpdateRun(run.ID, func(rr *store.Run) {
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
			finalSummary["durationText"] = humanDuration(durSec)
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
			counts := map[string]int{"copied": 0, "deleted": 0, "skipped": 0, "failed": 0, "total": 0}
			if p, ok := rr.Summary["stderrFile"].(string); ok && p != "" {
				files, counts = buildFinalSummaryFilesFromLog(p, isOpenlistCASCompatible(run), strings.ToLower(cmdName) == "move")
			}
			// 异步补全文件大小，不阻塞状态更新。
			// finalSummary 只服务于历史详情 / 最终总结展示；
			// 不要再往回恢复 stableProgress / cardSummary 这类完成态兼容字段，
			// 以免运行中链路与任务卡片完成态再次发生语义混用。
			r.enrichFilesSizesAsync(run.ID, files, dst, cfg, isOpenlistCASCompatible(run))
			finalSummary["counts"] = counts
			finalSummary["files"] = files
			rr.Summary["finalSummary"] = finalSummary

		})
		websocket.Broadcast("run_status", map[string]any{
			"run_id": run.ID,
			"status": "finished",
		})
		if r.activeMgr != nil {
			r.activeMgr.RemoveState(run.ID)
		}
		// fire webhook for successful run
		go r.postWebhookIfNeeded(run.ID)
		r.mu.Lock()
		delete(r.procs, run.ID)
		r.mu.Unlock()
	}()
	return nil
}

// humanDuration renders seconds to X小时Y分Z秒（省略 0 单位）
func humanDuration(sec int64) string {
	h := sec / 3600
	m := (sec % 3600) / 60
	s := sec % 60
	parts := []string{}
	if h > 0 {
		parts = append(parts, fmt.Sprintf("%d小时", h))
	}
	if m > 0 || (h > 0 && s > 0) {
		parts = append(parts, fmt.Sprintf("%d分", m))
	}
	if s > 0 || (h == 0 && m == 0) {
		parts = append(parts, fmt.Sprintf("%d秒", s))
	}
	return strings.Join(parts, "")
}

func isWebDAVUnderlying(cfgPath, remote string) bool {
	// 调用 `rclone config dump --config cfg` 并解析 remote 链；判断底层是否 webdav
	cr := &adapter.CmdRunner{}
	out, _, err := cr.Run(context.Background(), []string{"config", "dump", "--config", cfgPath}...)
	if err != nil {
		return false
	}
	var dump map[string]any
	if json.Unmarshal([]byte(out), &dump) != nil {
		return false
	}
	name := remote
	// 直接命中远端名
	for depth := 0; depth < 4; depth++ {
		sec, _ := dump[name].(map[string]any)
		if sec == nil {
			break
		}
		// type 命中 webdav
		if t, _ := sec["type"].(string); strings.EqualFold(t, "webdav") {
			return true
		}
		// crypt/alias 等 wrapper：跟随 remote 指向
		if base, _ := sec["remote"].(string); base != "" {
			// remote 形如 "webdav:root" 或 "other:"，取冒号前的 remote 名
			if i := strings.Index(base, ":"); i > 0 {
				name = base[:i]
				continue
			}
		}
		break
	}
	return false
}

func addFilterFlags(args []string, opts map[string]any, fastListFlag bool) []string {
	if opts == nil {
		return args
	}
	pass := []string{"include", "exclude", "filter", "filterFrom", "includeFrom", "excludeFrom", "filesFrom", "minSize", "maxSize", "minAge", "maxAge", "fastList"}
	for _, k := range pass {
		if v, ok := opts[k]; ok {
			flag := ""
			switch k {
			case "include":
				flag = "--include"
			case "exclude":
				flag = "--exclude"
			case "filter":
				flag = "--filter"
			case "filterFrom":
				flag = "--filter-from"
			case "includeFrom":
				flag = "--include-from"
			case "excludeFrom":
				flag = "--exclude-from"
			case "filesFrom":
				flag = "--files-from"
			case "minSize":
				flag = "--min-size"
			case "maxSize":
				flag = "--max-size"
			case "minAge":
				flag = "--min-age"
			case "maxAge":
				flag = "--max-age"
			case "fastList":
				if fastListFlag {
					if s := strings.ToLower(fmt.Sprint(v)); s == "true" || s == "1" {
						args = append(args, "--fast-list")
					}
				}
				continue
			}
			args = append(args, flag, fmt.Sprint(v))
		}
	}
	return args
}

func sizeOf(r *adapter.CmdRunner, cfg, target string, opts map[string]any) (bytes int64, count int64, err error) {
	// Prefer JSON when available. Fallback to parsing text if needed.
	args := []string{"size", target, "--config", cfg, "--json"}
	args = addFilterFlags(args, opts, true)
	out, _, e := r.Run(context.Background(), args...)
	if e == nil {
		var m map[string]any
		if json.Unmarshal([]byte(out), &m) == nil {
			if v, ok := m["bytes"].(float64); ok {
				bytes = int64(v)
			}
			if v, ok := m["count"].(float64); ok {
				count = int64(v)
			}
			return bytes, count, nil
		}
	}
	// Fallback: text parse
	args = []string{"size", target, "--config", cfg}
	args = addFilterFlags(args, opts, false)
	out, _, e = r.Run(context.Background(), args...)
	if e != nil {
		return 0, 0, e
	}
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

// sizeOfPaged: sequential directory pagination — sum sizes per top-level dir + root files
func sizeOfPaged(r *adapter.CmdRunner, cfg, target string, opts map[string]any) (int64, int64, error) {
	// list top-level dirs (depth=1)
	lsArgs := []string{"lsf", target, "--config", cfg, "--dirs-only", "--max-depth", "1"}
	lsArgs = addFilterFlags(lsArgs, opts, true)
	out, _, e := r.Run(context.Background(), lsArgs...)
	if e != nil {
		// fallback: single shot size
		return sizeOf(r, cfg, target, opts)
	}
	collectDirs := func(s string) []string {
		arr := []string{}
		for _, ln := range strings.Split(s, "\n") {
			ln = strings.TrimSpace(ln)
			if ln == "" {
				continue
			}
			arr = append(arr, strings.TrimSuffix(ln, "/"))
		}
		return arr
	}
	lines := collectDirs(out)
	// auto-descend: if no top-level dirs, try second-level dirs (depth=2), select those that contain one '/'
	if len(lines) == 0 {
		ls2 := []string{"lsf", target, "--config", cfg, "--dirs-only", "--max-depth", "2"}
		ls2 = addFilterFlags(ls2, opts, true)
		if out2, _, e2 := r.Run(context.Background(), ls2...); e2 == nil {
			all := collectDirs(out2)
			// keep only second-level like "a/b"
			sec := []string{}
			for _, d := range all {
				if strings.Count(d, "/") == 1 {
					sec = append(sec, d)
				}
			}
			lines = sec
		}
	}
	if len(lines) == 0 {
		// no subdirs even after descend — try single shot
		return sizeOf(r, cfg, target, opts)
	}
	var totalBytes, totalCount int64
	// root files: lsjson --files-only --max-depth 1
	rootArgs := []string{"lsjson", target, "--config", cfg, "--files-only", "--max-depth", "1"}
	rootArgs = addFilterFlags(rootArgs, opts, true)
	if out2, _, e2 := r.Run(context.Background(), rootArgs...); e2 == nil {
		var arr []map[string]any
		if json.Unmarshal([]byte(out2), &arr) == nil {
			for _, it := range arr {
				if sz, ok := it["Size"].(float64); ok {
					totalBytes += int64(sz)
					totalCount++
				}
			}
		}
	}
	// sum each (sub)dir sequentially
	for _, d := range lines {
		child := target
		if !strings.HasSuffix(child, "/") {
			child += "/"
		}
		child += d
		b, c, e3 := sizeOf(r, cfg, child, opts)
		if e3 == nil {
			totalBytes += b
			totalCount += c
		}
	}
	return totalBytes, totalCount, nil
}

func (r *Runner) Stop(runID int64) error {
	r.mu.Lock()
	cmd, ok := r.procs[runID]
	r.mu.Unlock()
	if !ok || cmd == nil || cmd.Process == nil {
		return errors.New("not running")
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
	go func() { _ = cmd.Wait(); close(ch) }()
	select {
	case <-ch:
		return true
	case <-time.After(d):
		return false
	}
}

type fileProg struct {
	Name  string  `json:"name"`
	Bytes float64 `json:"bytes"`
	Total float64 `json:"totalBytes"`
	Pct   float64 `json:"percentage"`
	Speed float64 `json:"speed"`
}

type fileProgress struct {
	m      map[string]*fileProg
	mu     sync.Mutex
	ver    uint64
	copied []string
}

func (fp *fileProgress) update(name string, bytes, total, speed, pct float64) {
	fp.mu.Lock()
	p, ok := fp.m[name]
	if !ok {
		p = &fileProg{Name: name}
		fp.m[name] = p
	}
	if total > 0 {
		p.Total = total
	}
	if bytes >= 0 {
		p.Bytes = bytes
	}
	if speed >= 0 {
		p.Speed = speed
	}
	if pct >= 0 {
		p.Pct = pct
	}
	fp.mu.Unlock()
	atomic.AddUint64(&fp.ver, 1)
}

func (fp *fileProgress) markCopied(name string) {
	fp.mu.Lock()
	fp.copied = append(fp.copied, name)
	fp.mu.Unlock()
}

func (fp *fileProgress) snapshot(limit int) []fileProg {
	fp.mu.Lock()
	defer fp.mu.Unlock()
	out := make([]fileProg, 0, len(fp.m))
	for _, v := range fp.m {
		out = append(out, *v)
	}
	if limit > 0 && len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out
}

func (fp *fileProgress) copiedList() []string {
	fp.mu.Lock()
	defer fp.mu.Unlock()
	out := make([]string, len(fp.copied))
	copy(out, fp.copied)
	return out
}

// statsRe matches rclone progress lines like "2026/04/17 14:33:38 INFO : 46.593 MiB / 335.968 MiB, 14%, 2.234 MiB/s, ETA 2m9s"
var statsRe = regexp.MustCompile(`INFO\s*:\s*([\d.]+\s*[KMGTP]?i?B?)\s*/\s*([\d.]+\s*[KMGTP]?i?B?),\s*([\d.]+)%,\s*([\d.]+\s*[KMGTP]?i?B?)/s`)
var fileLineRe = regexp.MustCompile(`(?i)INFO\s*:\s*([^:]+):\s*(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)B\s*/\s*(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)B,\s*(\d+(?:\.\d+)?)%?,\s*(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)B/s`)
var fileCopiedRe = regexp.MustCompile(`(?i)INFO\s*:\s*([^:]+):\s*Copied\s*\(new\)`)
var fileCASMatchedRe = regexp.MustCompile(`(?i)(?:INFO|NOTICE)\s*:\s*([^:]+):\s*CAS compatible match after source cleanup\b`)

func (r *Runner) consume(runID int64, rd io.Reader, out *os.File, parseStats bool, fp *fileProgress, openlistCASCompatible bool, isMove bool, cfg string, dst string, excludeFrom string) {
	wantParse := parseStats
	s := bufio.NewScanner(rd)
	s.Buffer(make([]byte, 0, 128*1024), 2*1024*1024)
	for s.Scan() {
		line := s.Text()
		parsedLine := line
		jsonLevel := ""
		jsonMsg := ""
		jsonObj := ""
		// 1) JSON 行：既尝试直接提取 machine-readable progress，也把 msg/object 解包给现有文本解析链复用。
		var rec map[string]any
		if json.Unmarshal([]byte(line), &rec) == nil {
			prog := map[string]any{}
			if v, ok := rec["bytes"].(float64); ok {
				prog["bytes"] = v
			}
			if v, ok := rec["totalBytes"].(float64); ok {
				prog["totalBytes"] = v
			}
			if v, ok := rec["speed"].(float64); ok {
				prog["speed"] = v
			}
			if v, ok := rec["eta"].(float64); ok {
				prog["eta"] = v
			}
			if stats, ok := rec["stats"].(map[string]any); ok {
				if v, ok := stats["bytes"].(float64); ok {
					prog["bytes"] = v
				}
				if v, ok := stats["totalBytes"].(float64); ok {
					prog["totalBytes"] = v
				}
				if v, ok := stats["speed"].(float64); ok {
					prog["speed"] = v
				}
				if v, ok := stats["eta"].(float64); ok {
					prog["eta"] = v
				}
				if tr, ok := stats["transferring"].([]any); ok && len(tr) > 0 {
					currentFiles := make([]map[string]any, 0, len(tr))
					for i, rawItem := range tr {
						item, ok := rawItem.(map[string]any)
						if !ok {
							continue
						}
						name := strings.TrimSpace(anyString(item["name"]))
						if name == "" {
							continue
						}
						cb := anyFloat64(item["bytes"])
						tb := anyFloat64(item["size"])
						sp := anyFloat64(item["speed"])
						var pctPtr *float64
						if v, ok := item["percentage"].(float64); ok {
							pct := v
							pctPtr = &pct
						}
						if r.activeMgr != nil {
							r.activeMgr.OnFileProgress(runID, name, int64(cb), int64(tb), int64(sp), pctPtr)
						}
						currentFile := map[string]any{
							"name": name,
							"path": name,
							"bytes": cb,
							"totalBytes": tb,
							"speed": sp,
							"status": "in_progress",
						}
						if pctPtr != nil {
							currentFile["percentage"] = *pctPtr
						}
						currentFiles = append(currentFiles, currentFile)
						if i == 0 {
							_ = r.db.UpdateRun(runID, func(rr *store.Run) {
								if rr.Summary == nil {
									rr.Summary = map[string]any{}
								}
								rr.Summary["currentFile"] = currentFile
							})
						}
					}
					if len(currentFiles) > 0 {
						_ = r.db.UpdateRun(runID, func(rr *store.Run) {
							if rr.Summary == nil {
								rr.Summary = map[string]any{}
							}
							rr.Summary["currentFiles"] = currentFiles
						})
					}
				}
			}
			jsonLevel = strings.ToUpper(anyString(rec["level"]))
			msg := strings.TrimSpace(anyString(rec["msg"]))
			obj := strings.TrimSpace(anyString(rec["object"]))
			jsonMsg = msg
			jsonObj = obj
			if msg != "" {
				prefix := strings.TrimSpace(jsonLevel)
				if prefix == "" {
					prefix = "INFO"
				}
				switch {
				case obj != "" && !strings.Contains(msg, obj):
					parsedLine = fmt.Sprintf("%s : %s: %s", prefix, obj, msg)
				case jsonLevel != "":
					parsedLine = fmt.Sprintf("%s : %s", prefix, msg)
				default:
					parsedLine = msg
				}
			}
			if len(prog) > 0 {
				if parsed, ok := parseOneLineProgress(parsedLine); ok {
					if v, ok2 := parsed["completedFiles"]; ok2 {
						prog["completedFiles"] = v
					} else if _, ok2 := parsed["plannedFiles"]; ok2 {
						prog["completedFiles"] = float64(0)
					}
					if v, ok2 := parsed["plannedFiles"]; ok2 {
						prog["plannedFiles"] = v
					}
					if v, ok2 := parsed["eta"]; ok2 {
						prog["eta"] = v
					}
					if v, ok2 := parsed["percentage"]; ok2 {
						prog["percentage"] = v
					}
				}
				_ = r.db.UpdateRun(runID, func(rr *store.Run) {
					if rr.Summary == nil {
						rr.Summary = map[string]any{}
					}
					if prev, ok := rr.Summary["progress"].(map[string]any); ok {
						if pc, ok2 := prev["completedFiles"].(float64); ok2 {
							if nc, ok3 := prog["completedFiles"].(float64); ok3 {
								if nc < pc {
									prog["completedFiles"] = pc
								}
							} else {
								prog["completedFiles"] = pc
							}
						}
						if pp, ok2 := prev["plannedFiles"].(float64); ok2 {
							if np, ok3 := prog["plannedFiles"].(float64); ok3 {
								if np < pp {
									prog["plannedFiles"] = pp
								}
							} else {
								prog["plannedFiles"] = pp
							}
						}
					}
					if fp != nil {
						if lst := fp.copiedList(); len(lst) > 0 {
							if nc, ok3 := prog["completedFiles"].(float64); !ok3 || float64(len(lst)) > nc {
								prog["completedFiles"] = float64(len(lst))
							}
							rr.Summary["files"] = fp.snapshot(100)
						}
					}
					rr.Summary["progress"] = prog
					rr.Summary["progressLine"] = parsedLine
					if b, ok := prog["bytes"].(float64); ok {
						rr.BytesTransferred = int64(b)
					}
					if sp, ok := prog["speed"].(float64); ok {
						rr.Speed = fmt.Sprintf("%d B/s", int64(sp))
					}
				})
				websocket.Broadcast("run_progress", map[string]any{
					"run_id":         runID,
					"bytes":          prog["bytes"],
					"total":          prog["totalBytes"],
					"speed":          prog["speed"],
					"percent":        prog["percentage"],
					"completedFiles": prog["completedFiles"],
					"plannedFiles":   prog["plannedFiles"],
					"totalCount":     prog["plannedFiles"],
					"eta":            prog["eta"],
				})
				continue
			}
		}
		if len(line) > 0 {
			_, _ = out.WriteString(sanitizeRunLogLine(line, openlistCASCompatible) + "\n")
		}
		line = parsedLine
		// 2) statsRe 文本解析（优先于 parseOneLineProgress）
		if m := statsRe.FindStringSubmatch(line); len(m) > 0 {
			cur := parseUnit(m[1])
			tot := parseUnit(m[2])
			pct, _ := strconv.ParseFloat(m[3], 64)
			spd := parseUnit(m[4])
			prog := map[string]any{
				"bytes":      cur,
				"totalBytes": tot,
				"percentage": pct,
				"speed":      spd,
			}
			if parsed, ok := parseOneLineProgress(line); ok {
				if v, ok2 := parsed["completedFiles"]; ok2 {
					prog["completedFiles"] = v
				}
				if v, ok2 := parsed["plannedFiles"]; ok2 {
					prog["plannedFiles"] = v
				}
				if v, ok2 := parsed["eta"]; ok2 {
					prog["eta"] = v
				}
			}
			_ = r.db.UpdateRun(runID, func(rr *store.Run) {
				if rr.Summary == nil {
					rr.Summary = map[string]any{}
				}
				if prev, ok := rr.Summary["progress"].(map[string]any); ok {
					if pc, ok2 := prev["completedFiles"].(float64); ok2 {
						if nc, ok3 := prog["completedFiles"].(float64); ok3 {
							if nc < pc {
								prog["completedFiles"] = pc
							}
						} else {
							prog["completedFiles"] = pc
						}
					}
					if pp, ok2 := prev["plannedFiles"].(float64); ok2 {
						if np, ok3 := prog["plannedFiles"].(float64); ok3 {
							if np < pp {
								prog["plannedFiles"] = pp
							}
						} else {
							prog["plannedFiles"] = pp
						}
					}
				}
				if fp != nil {
					if lst := fp.copiedList(); len(lst) > 0 {
						if nc, ok3 := prog["completedFiles"].(float64); !ok3 || float64(len(lst)) > nc {
							prog["completedFiles"] = float64(len(lst))
						}
					}
				}
				rr.Summary["progress"] = prog
				rr.Summary["progressLine"] = line
				rr.BytesTransferred = int64(cur)
				rr.Speed = fmt.Sprintf("%d B/s", int64(spd))
				if fp != nil {
					rr.Summary["files"] = fp.snapshot(100)
				}
			})
			websocket.Broadcast("run_progress", map[string]any{
				"run_id":         runID,
				"bytes":          prog["bytes"],
				"total":          prog["totalBytes"],
				"speed":          prog["speed"],
				"percent":        prog["percentage"],
				"completedFiles": prog["completedFiles"],
				"plannedFiles":   prog["plannedFiles"],
				"totalCount":     prog["plannedFiles"],
				"eta":            prog["eta"],
			})
			continue
		}
		// 3) parseOneLineProgress 兜底（仅在需要时）
		if wantParse {
			if prog, ok := parseOneLineProgress(line); ok {
				_ = r.db.UpdateRun(runID, func(rr *store.Run) {
					if rr.Summary == nil {
						rr.Summary = map[string]any{}
					}
					rr.Summary["progressLine"] = line
					// preserve non-decreasing completedFiles; fallback to copied list if needed
					if prev, ok := rr.Summary["progress"].(map[string]any); ok {
						if pc, ok2 := prev["completedFiles"].(float64); ok2 {
							if nc, ok3 := prog["completedFiles"].(float64); ok3 {
								if nc < pc {
									prog["completedFiles"] = pc
								}
							} else {
								prog["completedFiles"] = pc
							}
						}
					}
					if fp != nil {
						if lst := fp.copiedList(); len(lst) > 0 {
							if nc, ok3 := prog["completedFiles"].(float64); !ok3 || float64(len(lst)) > nc {
								prog["completedFiles"] = float64(len(lst))
							}
						}
					}
					rr.Summary["progress"] = prog
					rr.BytesTransferred = int64(prog["bytes"].(float64))
					rr.Speed = fmt.Sprintf("%d B/s", int64(prog["speed"].(float64)))
					// 同步部分文件列表快照（最近 100 条）
					if fp != nil {
						rr.Summary["files"] = fp.snapshot(100)
					}
				})
				websocket.Broadcast("run_progress", map[string]any{
					"run_id":         runID,
					"bytes":          prog["bytes"],
					"total":          prog["totalBytes"],
					"speed":          prog["speed"],
					"percent":        prog["percentage"],
					"completedFiles": prog["completedFiles"],
					"plannedFiles":   prog["plannedFiles"],
					"totalCount":     prog["plannedFiles"],
					"eta":            prog["eta"],
				})
			}
		}
		// 文件级完成识别不依赖 wantParse：即使当前流不做 aggregate 解析，也要累计 completedFiles。
		if fp != nil {
			marked := false
			if m := fileLineRe.FindStringSubmatch(line); len(m) > 0 {
				name := strings.TrimSpace(m[1])
				var cb, tb, pct, sp float64
				fmt.Sscanf(m[2], "%f", &cb)
				fmt.Sscanf(m[4], "%f", &tb)
				fmt.Sscanf(m[6], "%f", &pct)
				fmt.Sscanf(m[7], "%f", &sp)
				fp.update(name, cb*unitToMul(m[3]), tb*unitToMul(m[5]), sp*unitToMul(m[8]), pct)
				if r.activeMgr != nil {
					pctCopy := pct
					r.activeMgr.OnFileProgress(runID, name, int64(cb*unitToMul(m[3])), int64(tb*unitToMul(m[5])), int64(sp*unitToMul(m[8])), &pctCopy)
				}
				if (tb > 0 && cb >= tb) || pct >= 100 {
					fp.markCopied(name)
					if r.activeMgr != nil {
						r.activeMgr.OnFileCopied(runID, name)
					}
					marked = true
				}
				_ = r.db.UpdateRun(runID, func(rr *store.Run) {
					if rr.Summary == nil {
						rr.Summary = map[string]any{}
					}
					rr.Summary["progressLine"] = line
				})
				if marked {
					_ = r.db.UpdateRun(runID, func(rr *store.Run) {
						if rr.Summary == nil {
							rr.Summary = map[string]any{}
						}
						prog, _ := rr.Summary["progress"].(map[string]any)
						if prog == nil {
							prog = map[string]any{}
						}
						if lst := fp.copiedList(); len(lst) > 0 {
							if nc, ok := prog["completedFiles"].(float64); !ok || float64(len(lst)) > nc {
								prog["completedFiles"] = float64(len(lst))
							}
							rr.Summary["files"] = fp.snapshot(100)
						}
						rr.Summary["progress"] = prog
					})
				}
				continue
			}
			if m := fileCopiedRe.FindStringSubmatch(line); len(m) > 0 {
				name := strings.TrimSpace(m[1])
				fp.update(name, -1, -1, -1, 100)
				fp.markCopied(name)
				if r.activeMgr != nil {
					r.activeMgr.OnFileCopied(runID, name)
				}
				marked = true
			}
			if !marked {
				if m := fileCASMatchedRe.FindStringSubmatch(line); len(m) > 0 {
					name := strings.TrimSpace(m[1])
					fp.update(name, -1, -1, -1, 100)
					fp.markCopied(name)
					if r.activeMgr != nil {
						r.activeMgr.OnFileCASMatched(runID, name)
					}
					marked = true
				}
			}
			if !marked {
				path := strings.TrimSpace(extractPathFromLogLine(line))
				msg := strings.TrimSpace(extractMsgFromLogLine(line))
				if path == "" {
					path = strings.TrimSpace(jsonObj)
				}
				if msg == "" {
					msg = strings.TrimSpace(jsonMsg)
				}
				if isCASCompatibleNotFound(path, msg, openlistCASCompatible) {
					if r.confirmCASMatch(cfg, dst, path) {
						fp.update(path, -1, -1, -1, 100)
						fp.markCopied(path)
						if r.activeMgr != nil {
							r.activeMgr.OnFileCASMatched(runID, path)
						}
						r.appendCASExclude(excludeFrom, path)
						_, _ = out.WriteString(fmt.Sprintf("NOTICE : %s: CAS compatible match after source cleanup (%s)\n", path, msg))
						marked = true
						_ = r.db.UpdateRun(runID, func(rr *store.Run) {
							if rr.Summary == nil {
								rr.Summary = map[string]any{}
							}
							prog, _ := rr.Summary["progress"].(map[string]any)
							if prog == nil {
								prog = map[string]any{}
							}
							if lst := fp.copiedList(); len(lst) > 0 {
								prog["completedFiles"] = float64(len(lst))
								rr.Summary["files"] = fp.snapshot(100)
							}
							rr.Summary["progress"] = prog
						})
						continue
					} else {
						if r.activeMgr != nil {
							r.activeMgr.OnFileFailed(runID, path, msg)
						}
						_, _ = out.WriteString(fmt.Sprintf("ERROR : %s: %s\n", path, msg))
						continue
					}
				}
				if !marked {
					if isAttemptObjectNotFoundSummary(path, msg) {
						continue
					}
					if row, _, ok := classifyRunLogRow("INFO", path, msg, map[string]int64{}, openlistCASCompatible); ok && r.activeMgr != nil {
						path := strings.TrimSpace(anyString(row["path"]))
						action := strings.ToLower(strings.TrimSpace(anyString(row["action"])))
						msg := strings.TrimSpace(anyString(row["message"]))
						switch action {
						case "deleted":
							if !isMove {
								r.activeMgr.OnFileDeleted(runID, path)
							}
						case "skipped":
							r.activeMgr.OnFileSkipped(runID, path, msg)
						case "error":
							r.activeMgr.OnFileFailed(runID, path, msg)
						}
					}
				}
			}
			if marked {
				_ = r.db.UpdateRun(runID, func(rr *store.Run) {
					if rr.Summary == nil {
						rr.Summary = map[string]any{}
					}
					prog, _ := rr.Summary["progress"].(map[string]any)
					if prog == nil {
						prog = map[string]any{}
					}
					if lst := fp.copiedList(); len(lst) > 0 {
						if nc, ok := prog["completedFiles"].(float64); !ok || float64(len(lst)) > nc {
							prog["completedFiles"] = float64(len(lst))
						}
						rr.Summary["files"] = fp.snapshot(100)
					}
					rr.Summary["progress"] = prog
				})
				continue
			}
		}
	}
	if err := s.Err(); err != nil {
		logger.Debug("progress scanner error", zap.Error(err))
	}
}

var oneLineRe = regexp.MustCompile(`(?i)^\s*(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)(?:B)?\s*/\s*(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)(?:B)?\s*,\s*(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)(?:B)?/s\s*,\s*(?:ETA\s*([0-9hms:.-]+)|ETA\s*-|([0-9]{1,3})%)`) // kept for reference

func unitToMul(u string) float64 {
	u = strings.ToUpper(u)
	switch u {
	case "K", "KI":
		return 1024
	case "M", "MI":
		return 1024 * 1024
	case "G", "GI":
		return 1024 * 1024 * 1024
	case "T", "TI":
		return 1024 * 1024 * 1024 * 1024
	case "P", "PI":
		return 1024 * 1024 * 1024 * 1024 * 1024
	default:
		return 1
	}
}

// parseUnit parses a string like "3.055 MiB" or "508 KiB" into bytes
func parseUnit(s string) float64 {
	s = strings.TrimSpace(s)
	// regex to extract number and unit
	var num float64
	var unit string
	// handle formats: "3.055MiB", "3.055 MiB", "508KiB", "508 KiB"
	re := regexp.MustCompile(`([\d.]+)\s*([KMGTP]?i?B?)`)
	m := re.FindStringSubmatch(s)
	if len(m) < 3 {
		// try just parsing as float
		fmt.Sscanf(s, "%f", &num)
		return num
	}
	fmt.Sscanf(m[1], "%f", &num)
	unit = strings.ToUpper(m[2])
	// normalize unit: "MIB" -> "MI", "MB" -> "M"
	unit = strings.TrimSuffix(unit, "B")
	if !strings.HasSuffix(unit, "I") && len(unit) > 0 && unit[len(unit)-1] == 'I' {
		// already has I suffix
	} else if strings.HasSuffix(unit, "I") {
		// keep as is
	} else {
		// no I suffix, add it for KiB, MiB etc
		if unit == "K" || unit == "M" || unit == "G" || unit == "T" || unit == "P" {
			unit = unit + "I"
		}
	}
	return num * unitToMul(unit)
}

func parseETA(s string) int {
	s = strings.TrimSpace(s)
	if s == "-" || s == "" {
		return 0
	}
	// formats: 4d9h18m | 1h2m3s | 2m3s | 45s | 01:23:45 | 12:34
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
	sec := 0
	// 显式按包含的后缀组合解析，避免 "1m26s" 被 "%dh%dm%ds" 误读成 1h
	hasD := strings.Contains(s, "d")
	hasH := strings.Contains(s, "h")
	hasM := strings.Contains(s, "m")
	hasS := strings.Contains(s, "s")
	var d, h, m, ss int
	switch {
	case hasD && hasH && hasM && hasS:
		fmt.Sscanf(s, "%dd%dh%dm%ds", &d, &h, &m, &ss)
	case hasD && hasH && hasM:
		fmt.Sscanf(s, "%dd%dh%dm", &d, &h, &m)
	case hasD && hasH && hasS:
		fmt.Sscanf(s, "%dd%dh%ds", &d, &h, &ss)
	case hasD && hasM && hasS:
		fmt.Sscanf(s, "%dd%dm%ds", &d, &m, &ss)
	case hasD && hasH:
		fmt.Sscanf(s, "%dd%dh", &d, &h)
	case hasD && hasM:
		fmt.Sscanf(s, "%dd%dm", &d, &m)
	case hasD && hasS:
		fmt.Sscanf(s, "%dd%ds", &d, &ss)
	case hasD:
		fmt.Sscanf(s, "%dd", &d)
	case hasH && hasM && hasS:
		fmt.Sscanf(s, "%dh%dm%ds", &h, &m, &ss)
	case hasH && hasM:
		fmt.Sscanf(s, "%dh%dm", &h, &m)
	case hasH && hasS:
		fmt.Sscanf(s, "%dh%ds", &h, &ss)
	case hasM && hasS:
		fmt.Sscanf(s, "%dm%ds", &m, &ss)
	case hasH:
		fmt.Sscanf(s, "%dh", &h)
	case hasM:
		fmt.Sscanf(s, "%dm", &m)
	case hasS:
		fmt.Sscanf(s, "%ds", &ss)
	}
	sec += d*24*3600 + h*3600 + m*60 + ss
	return sec
}

var bytesPairRe = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)B\s*/\s*(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)B`)
var speedTokenRe = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)(?:B)?/s`)
var pctTokenRe = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)%`)
var etaTokenRe = regexp.MustCompile(`(?i)ETA\s*([0-9dhms:.-]+|-)`)
var aggregateOneLineRe = regexp.MustCompile(`(?i)^\s*(?:\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2}\s+)?(?:INFO|NOTICE)\s*:\s*\d+(?:\.\d+)?\s*[KMGTPE]?i?B\s*/\s*\d+(?:\.\d+)?\s*[KMGTPE]?i?B\s*,\s*\d+(?:\.\d+)?%\s*,\s*\d+(?:\.\d+)?\s*[KMGTPE]?i?B/s\s*,\s*ETA\s*[0-9dhms:.-]+(?:\s*\(xfr#\d+(?:/\d+)?\))?\s*$`)

func parseOneLineProgress(line string) (map[string]any, bool) {
	l := strings.TrimSpace(line)
	// 明确排除文件级进度/文件动作日志，避免把单文件行误当成整体进度
	if fileLineRe.MatchString(l) || fileCopiedRe.MatchString(l) {
		return nil, false
	}
	// 只接受整体 one-line 统计的完整形态，避免碎片日志/文件级日志误入
	if !aggregateOneLineRe.MatchString(l) {
		return nil, false
	}
	// 提取 (xfr#a/b) 的已完成数 a 和计划总数 b
	xfrDone := float64(0)
	planned := float64(0)
	if i := strings.Index(l, "("); i >= 0 {
		// 截取括号段与主段
		paren := l[i:]
		l = strings.TrimSpace(l[:i])
		if j := strings.Index(strings.ToLower(paren), "xfr#"); j >= 0 {
			var a, b int
			n, _ := fmt.Sscanf(paren[j:], "xfr#%d/%d", &a, &b)
			if n >= 1 && a > 0 {
				xfrDone = float64(a)
			}
			if n == 2 && b > 0 {
				planned = float64(b)
			}
			if n == 0 { // 旧格式仅含 a
				var aa int
				if _, err := fmt.Sscanf(paren[j:], "xfr#%d", &aa); err == nil && aa > 0 {
					xfrDone = float64(aa)
				}
			}
		}
	}
	// 按多段拼接处理：取第一个匹配片段作为当前进度（整体进度通常在前面）
	bps := bytesPairRe.FindAllStringSubmatch(l, -1)
	if len(bps) == 0 {
		return nil, false
	}
	bp := bps[0]
	var cur, tot float64
	fmt.Sscanf(bp[1], "%f", &cur)
	fmt.Sscanf(bp[3], "%f", &tot)
	curBytes := cur * unitToMul(bp[2])
	totBytes := tot * unitToMul(bp[4])
	// 速度：取最后一个 token
	var sp float64
	spmAll := speedTokenRe.FindAllStringSubmatch(l, -1)
	var spm []string
	if len(spmAll) > 0 {
		spm = spmAll[len(spmAll)-1]
	}
	if len(spm) > 0 {
		fmt.Sscanf(spm[1], "%f", &sp)
	}
	spBytes := sp * unitToMul(spmValue(spm, 2))
	// ETA / 百分比（取最后一个）
	eta := 0
	if emAll := etaTokenRe.FindAllStringSubmatch(l, -1); len(emAll) > 0 {
		em := emAll[len(emAll)-1]
		if em[1] != "-" {
			eta = parseETA(em[1])
		}
	}
	prog := map[string]any{"bytes": curBytes, "totalBytes": totBytes, "speed": spBytes, "eta": float64(eta)}
	if pmAll := pctTokenRe.FindAllStringSubmatch(l, -1); len(pmAll) > 0 {
		pm := pmAll[len(pmAll)-1]
		var pct float64
		fmt.Sscanf(pm[1], "%f", &pct)
		prog["percentage"] = pct
	}
	if xfrDone > 0 {
		prog["completedFiles"] = xfrDone
	}
	if planned > 0 {
		prog["plannedFiles"] = planned
	}
	return prog, true
}

func spmValue(m []string, i int) string {
	if len(m) > i {
		return m[i]
	}
	return ""
}

// enrichFilesSizesAsync 异步补全文件大小，不阻塞数据库更新和 webhook
func (r *Runner) enrichFilesSizesAsync(runID int64, files []map[string]any, dst, cfg string, openlistCASCompatible bool) {
	go func() {
		if len(files) == 0 {
			return
		}
		// 限制文件数量，超过 5000 个则跳过
		if len(files) > 5000 {
			return
		}
		cr := &adapter.CmdRunner{}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		out, _, e2 := cr.Run(ctx, []string{"lsjson", dst, "--config", cfg, "--files-only", "--recursive"}...)
		if e2 != nil {
			return
		}
		var arr []map[string]any
		if json.Unmarshal([]byte(out), &arr) != nil {
			return
		}
		m := map[string]int64{}
		for _, it := range arr {
			p, _ := it["Path"].(string)
			if p == "" {
				p, _ = it["path"].(string)
			}
			var sz int64
			switch vv := it["Size"].(type) {
			case float64:
				sz = int64(vv)
			}
			if p != "" {
				p = strings.ReplaceAll(p, "\\", "/")
				m[p] = sz
				if openlistCASCompatible && isCASPath(p) {
					m[trimCASSuffix(p)] = sz
				}
			}
		}
		if len(m) == 0 {
			return
		}
		// 直接在传入的 files 上修改（切片是引用类型，修改会影响原数组）
		for i := range files {
			p := strings.ReplaceAll(fmt.Sprint(files[i]["path"]), "\\", "/")
			if sz, ok := m[p]; ok && sz > 0 {
				files[i]["sizeBytes"] = sz
			}
		}
		// 补全完成后，更新数据库中的 finalSummary
		_ = r.db.UpdateRun(runID, func(rr *store.Run) {
			if rr == nil || rr.Summary == nil {
				return
			}
			if fs, ok := rr.Summary["finalSummary"].(map[string]any); ok {
				fs["files"] = files
			}
		})
	}()
}

type openlistCASCompatPlan struct {
	ExcludeFrom       string
	SourceFiles       []string
	MatchedSource     []string
	DestinationExtras []string
}

func isOpenlistCASCompatible(run store.Run) bool {
	if run.Summary == nil {
		return false
	}
	if raw, ok := run.Summary["effectiveOptions"].(map[string]any); ok {
		if v, ok := raw["openlistCasCompatible"].(bool); ok {
			return v
		}
	}
	return false
}

func buildOpenlistCASCompatPlan(cfg, src, dst, mode string) (*openlistCASCompatPlan, error) {
	srcFiles, err := listRecursiveFilePaths(cfg, src)
	if err != nil {
		return nil, err
	}
	dstFiles, err := listRecursiveFilePaths(cfg, dst)
	if err != nil {
		return nil, err
	}
	plan := buildOpenlistCASCompatPlanFromPaths(srcFiles, dstFiles, mode)
	if len(plan.MatchedSource) > 0 {
		f, err := os.CreateTemp("", "rcloneflow-openlist-cas-*.txt")
		if err != nil {
			return nil, err
		}
		for _, p := range plan.MatchedSource {
			if _, err := f.WriteString(p + "\n"); err != nil {
				f.Close()
				os.Remove(f.Name())
				return nil, err
			}
		}
		if err := f.Close(); err != nil {
			os.Remove(f.Name())
			return nil, err
		}
		plan.ExcludeFrom = f.Name()
	}
	return plan, nil
}

func buildOpenlistCASCompatPlanFromPaths(srcFiles, dstFiles []string, mode string) *openlistCASCompatPlan {
	dstExact := make(map[string]struct{}, len(dstFiles))
	for _, p := range dstFiles {
		dstExact[p] = struct{}{}
	}
	srcExact := make(map[string]struct{}, len(srcFiles))
	for _, p := range srcFiles {
		srcExact[p] = struct{}{}
	}
	matched := make([]string, 0)
	for _, p := range srcFiles {
		if isCASPath(p) {
			if _, ok := dstExact[p]; ok {
				matched = append(matched, p)
			}
			continue
		}
		if _, ok := dstExact[p+".cas"]; ok {
			matched = append(matched, p)
		}
	}
	extras := make([]string, 0)
	if mode == "sync" {
		for _, p := range dstFiles {
			if isCASPath(p) {
				if _, ok := srcExact[p]; ok {
					continue
				}
				if _, ok := srcExact[trimCASSuffix(p)]; ok {
					continue
				}
				extras = append(extras, p)
				continue
			}
			if _, ok := srcExact[p]; ok {
				continue
			}
			extras = append(extras, p)
		}
	}
	return &openlistCASCompatPlan{SourceFiles: append([]string(nil), srcFiles...), MatchedSource: matched, DestinationExtras: extras}
}

func (p *openlistCASCompatPlan) ApplyPostActions(cfg, src, dst, originalMode string) error {
	if p == nil {
		return nil
	}
	cr := &adapter.CmdRunner{}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()
	if originalMode == "move" {
		for _, rel := range p.MatchedSource {
			if _, _, err := cr.Run(ctx, []string{"deletefile", joinRemotePath(src, rel), "--config", cfg}...); err != nil {
				return fmt.Errorf("delete matched move source %s: %w", rel, err)
			}
		}
	}
	if originalMode == "sync" {
		for _, rel := range p.DestinationExtras {
			if _, _, err := cr.Run(ctx, []string{"deletefile", joinRemotePath(dst, rel), "--config", cfg}...); err != nil {
				return fmt.Errorf("delete extra sync destination %s: %w", rel, err)
			}
		}
	}
	return nil
}

func listRecursiveFilePaths(cfg, target string) ([]string, error) {
	cr := &adapter.CmdRunner{}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()
	out, _, err := cr.Run(ctx, []string{"lsjson", target, "--config", cfg, "--files-only", "--recursive"}...)
	if err != nil {
		return nil, err
	}
	var arr []map[string]any
	if err := json.Unmarshal([]byte(out), &arr); err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(arr))
	for _, it := range arr {
		p, _ := it["Path"].(string)
		if p == "" {
			p, _ = it["path"].(string)
		}
		p = strings.TrimSpace(strings.ReplaceAll(p, "\\", "/"))
		if p != "" {
			paths = append(paths, p)
		}
	}
	return paths, nil
}

func isCASPath(p string) bool {
	return strings.HasSuffix(strings.ToLower(strings.TrimSpace(p)), ".cas")
}

func trimCASSuffix(p string) string {
	if !isCASPath(p) {
		return p
	}
	return p[:len(p)-4]
}

func normalizeVisibleTargetPaths(arr []map[string]any, openlistCASCompatible bool) map[string]struct{} {
	m := map[string]struct{}{}
	for _, it := range arr {
		p, _ := it["Path"].(string)
		if p == "" {
			p, _ = it["path"].(string)
		}
		p = strings.TrimSpace(strings.ReplaceAll(p, "\\", "/"))
		if p == "" {
			continue
		}
		m[p] = struct{}{}
		if openlistCASCompatible && isCASPath(p) {
			m[trimCASSuffix(p)] = struct{}{}
		}
	}
	return m
}

func expectedVisibleDestinationPaths(plan *openlistCASCompatPlan, originalMode string) []string {
	if plan == nil {
		return nil
	}
	if originalMode == "move" {
		return append([]string(nil), plan.MatchedSource...)
	}
	return append([]string(nil), plan.SourceFiles...)
}

func areAllExpectedPathsVisible(expected []string, visible map[string]struct{}) bool {
	for _, p := range expected {
		if _, ok := visible[p]; !ok {
			return false
		}
	}
	return true
}

func isCASCompatibleNotFound(path, msg string, openlistCASCompatible bool) bool {
	if !openlistCASCompatible {
		return false
	}
	p := strings.TrimSpace(strings.ReplaceAll(path, "\\", "/"))
	if p == "" || isCASPath(p) {
		return false
	}
	low := strings.ToLower(strings.TrimSpace(msg))
	if low == "" {
		return false
	}
	if !(strings.Contains(low, "not found") || strings.Contains(low, "no such file") || strings.Contains(low, "object not found")) {
		return false
	}
	return true
}

func defaultCASFileExists(cfg, dst, rel string) (bool, error) {
	casRel := strings.TrimPrefix(strings.ReplaceAll(rel, "\\", "/"), "/") + ".cas"
	if strings.TrimSpace(casRel) == ".cas" {
		return false, nil
	}
	cr := &adapter.CmdRunner{}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	out, _, err := cr.Run(ctx, []string{"lsjson", joinRemotePath(dst, casRel), "--config", cfg, "--files-only"}...)
	if err != nil {
		return false, nil
	}
	var arr []map[string]any
	if err := json.Unmarshal([]byte(out), &arr); err != nil {
		return false, err
	}
	for _, it := range arr {
		p, _ := it["Path"].(string)
		if p == "" {
			p, _ = it["path"].(string)
		}
		p = strings.TrimSpace(strings.ReplaceAll(p, "\\", "/"))
		if p == "" {
			continue
		}
		if p == casRel || strings.HasSuffix(p, "/"+casRel) || strings.HasSuffix(p, "/"+filepath.Base(casRel)) || p == filepath.Base(casRel) {
			return true, nil
		}
	}
	return false, nil
}

func (r *Runner) confirmCASMatch(cfg, dst, rel string) bool {
	if r == nil || r.casVerifier == nil {
		return false
	}
	delays := r.casVerifyDelays
	if len(delays) == 0 {
		delays = []time.Duration{0}
	}
	for _, d := range delays {
		if d > 0 {
			time.Sleep(d)
		}
		ok, err := r.casVerifier(cfg, dst, rel)
		if err == nil && ok {
			return true
		}
	}
	return false
}

func sanitizeRunLogLine(line string, openlistCASCompatible bool) string {
	return line
}

func splitRunLogSegments(line string) []string {
	l := strings.TrimSpace(line)
	if l == "" {
		return nil
	}
	tsRe := regexp.MustCompile(`\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2}\s+(?:INFO|NOTICE|ERROR)\s*:`)
	idx := tsRe.FindAllStringIndex(l, -1)
	if len(idx) <= 1 {
		return []string{l}
	}
	segments := make([]string, 0, len(idx))
	for i := 0; i < len(idx); i++ {
		start := idx[i][0]
		end := len(l)
		if i+1 < len(idx) {
			end = idx[i+1][0]
		}
		segments = append(segments, strings.TrimSpace(l[start:end]))
	}
	return segments
}

func parseRunLogSegment(seg string) (at, level, path, msg string, ok bool) {
	re := regexp.MustCompile(`(?:(\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2})\s+)?(INFO|NOTICE|ERROR)\s*:\s*(.+?):\s*(.+)$`)
	if m := re.FindStringSubmatch(strings.TrimSpace(seg)); len(m) > 0 {
		return strings.TrimSpace(m[1]), strings.ToUpper(strings.TrimSpace(m[2])), strings.TrimSpace(m[3]), strings.TrimSpace(m[4]), true
	}
	var rec map[string]any
	if json.Unmarshal([]byte(strings.TrimSpace(seg)), &rec) == nil {
		level = strings.ToUpper(strings.TrimSpace(fmt.Sprint(rec["level"])))
		msg = strings.TrimSpace(fmt.Sprint(rec["msg"]))
		path = strings.TrimSpace(fmt.Sprint(rec["object"]))
		at = strings.TrimSpace(fmt.Sprint(rec["time"]))
		if at == "" {
			at = strings.TrimSpace(fmt.Sprint(rec["timestamp"]))
		}
		if msg != "" {
			return at, level, path, msg, true
		}
	}
	return "", "", "", "", false
}

func mergeMoveRows(files []map[string]any) ([]map[string]any, map[string]int) {
	copiedMap := map[string]map[string]any{}
	deletedMap := map[string]map[string]any{}
	others := []map[string]any{}
	for _, f := range files {
		a := strings.ToLower(fmt.Sprint(f["action"]))
		p := fmt.Sprint(f["path"])
		switch a {
		case "copied":
			copiedMap[p] = f
		case "deleted":
			deletedMap[p] = f
		default:
			others = append(others, f)
		}
	}
	moved := []map[string]any{}
	remainingCopied := []map[string]any{}
	remainingDeleted := []map[string]any{}
	for p, c := range copiedMap {
		if _, ok := deletedMap[p]; ok {
			moved = append(moved, map[string]any{"path": p, "at": c["at"], "status": "success", "action": "Moved", "sizeBytes": c["sizeBytes"]})
			delete(deletedMap, p)
		} else {
			remainingCopied = append(remainingCopied, c)
		}
	}
	for _, d := range deletedMap {
		remainingDeleted = append(remainingDeleted, d)
	}
	merged := append([]map[string]any{}, moved...)
	merged = append(merged, remainingCopied...)
	merged = append(merged, others...)
	merged = append(merged, remainingDeleted...)
	counts := map[string]int{"copied": len(moved) + len(remainingCopied), "deleted": len(remainingDeleted), "failed": 0, "skipped": 0, "total": 0}
	for _, f := range merged {
		a := strings.ToLower(fmt.Sprint(f["action"]))
		s := strings.ToLower(fmt.Sprint(f["status"]))
		if a == "error" || s == "failed" {
			counts["failed"]++
		} else if s == "skipped" || a == "skipped" {
			counts["skipped"]++
		}
	}
	counts["total"] = counts["copied"] + counts["deleted"] + counts["failed"] + counts["skipped"]
	return merged, counts
}

func buildFinalSummaryFilesFromLog(logPath string, openlistCASCompatible bool, moveMode bool) ([]map[string]any, map[string]int) {
	files := []map[string]any{}
	counts := map[string]int{"copied": 0, "deleted": 0, "skipped": 0, "failed": 0, "total": 0}
	if logPath == "" {
		return files, counts
	}
	b, e := os.ReadFile(logPath)
	if e != nil {
		return files, counts
	}
	lines := strings.Split(string(b), "\n")
	sizes := map[string]int64{}
	for _, ln := range lines {
		if m := fileLineRe.FindStringSubmatch(ln); len(m) > 0 {
			name := strings.TrimSpace(m[1])
			var tb float64
			fmt.Sscanf(m[4], "%f", &tb)
			total := int64(tb * unitToMul(m[5]))
			if total > 0 {
				sizes[name] = total
			}
		}
	}
	for _, ln := range lines {
		for _, seg := range splitRunLogSegments(ln) {
			at, level, path, msg, ok := parseRunLogSegment(seg)
			if !ok {
				continue
			}
			row, bucket, ok := classifyRunLogRow(level, path, msg, sizes, openlistCASCompatible)
			if !ok {
				continue
			}
			row["at"] = at
			files = append(files, row)
			counts[bucket]++
			counts["total"]++
		}
	}
	if moveMode {
		return mergeMoveRows(files)
	}
	return files, counts
}

func classifyRunLogRow(level, path, msg string, sizes map[string]int64, openlistCASCompatible bool) (map[string]any, string, bool) {
	at := ""
	row := map[string]any{"path": path, "at": at, "status": "", "action": "", "sizeBytes": 0}
	if sz, ok := sizes[path]; ok {
		row["sizeBytes"] = sz
	}
	low := strings.ToLower(strings.TrimSpace(msg))
	upperLevel := strings.ToUpper(strings.TrimSpace(level))
	if isAttemptObjectNotFoundSummary(path, msg) {
		return nil, "", false
	}
	if upperLevel == "ERROR" {
		row["status"] = "failed"
		row["action"] = "Error"
		row["message"] = msg
		return row, "failed", true
	}
	switch {
	case strings.Contains(low, "cas compatible match after source cleanup"):
		row["status"] = "success"
		row["action"] = "CAS Matched"
		row["message"] = msg
		return row, "copied", true
	case strings.Contains(low, "copied"):
		row["status"] = "success"
		row["action"] = "Copied"
		return row, "copied", true
	case strings.Contains(low, "deleted") || strings.Contains(low, "removed"):
		row["status"] = "success"
		row["action"] = "Deleted"
		return row, "deleted", true
	case strings.Contains(low, "skipped"):
		row["status"] = "skipped"
		row["action"] = "Skipped"
		return row, "skipped", true
	default:
		return nil, "", false
	}
}

func isAttemptObjectNotFoundSummary(path, msg string) bool {
	if strings.TrimSpace(path) != "Attempt 1/3 failed with 5 errors and" && !strings.HasPrefix(strings.TrimSpace(path), "Attempt ") {
		return false
	}
	low := strings.ToLower(strings.TrimSpace(msg))
	return strings.Contains(low, "object not found")
}

type casAttemptAnalysis struct {
	CASMatchedPaths map[string]struct{}
	RealFailures    map[string]string
}

func analyzeCASAttemptLogSegment(path string, startOffset int64, openlistCASCompatible bool) casAttemptAnalysis {
	res := casAttemptAnalysis{CASMatchedPaths: map[string]struct{}{}, RealFailures: map[string]string{}}
	f, err := os.Open(path)
	if err != nil {
		return res
	}
	defer f.Close()
	if startOffset > 0 {
		if _, err := f.Seek(startOffset, io.SeekStart); err != nil {
			return res
		}
	}
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 0, 128*1024), 2*1024*1024)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		if m := fileCASMatchedRe.FindStringSubmatch(line); len(m) > 0 {
			res.CASMatchedPaths[strings.TrimSpace(m[1])] = struct{}{}
			delete(res.RealFailures, strings.TrimSpace(m[1]))
			continue
		}
		var rec map[string]any
		if json.Unmarshal([]byte(line), &rec) == nil {
			level := strings.ToUpper(strings.TrimSpace(anyString(rec["level"])))
			msg := strings.TrimSpace(anyString(rec["msg"]))
			obj := strings.TrimSpace(anyString(rec["object"]))
			if level == "ERROR" {
				if isCASCompatibleNotFound(obj, msg, openlistCASCompatible) {
					continue
				}
				if obj != "" {
					res.RealFailures[obj] = msg
				}
			}
			continue
		}
		at, level, p, msg, ok := parseRunLogSegment(line)
		_ = at
		if !ok {
			continue
		}
		if isAttemptObjectNotFoundSummary(p, msg) {
			continue
		}
		if strings.EqualFold(level, "ERROR") {
			if isCASCompatibleNotFound(p, msg, openlistCASCompatible) {
				continue
			}
			if strings.TrimSpace(p) != "" {
				res.RealFailures[p] = msg
			}
		}
	}
	return res
}

func configuredRetryCount(opt map[string]any) int {
	if opt != nil {
		switch v := opt["retries"].(type) {
		case int:
			if v > 0 {
				return v
			}
		case int64:
			if v > 0 {
				return int(v)
			}
		case float64:
			if v > 0 {
				return int(v)
			}
		}
	}
	return 1
}

func forceFlagValue(args []string, flag, value string) []string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag {
			args[i+1] = value
			return args
		}
	}
	return append(args, flag, value)
}

func (r *Runner) appendCASExclude(excludeFrom, rel string) {
	if r == nil || strings.TrimSpace(excludeFrom) == "" || strings.TrimSpace(rel) == "" {
		return
	}
	r.casExcludeMu.Lock()
	defer r.casExcludeMu.Unlock()
	f, err := os.OpenFile(excludeFrom, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(strings.TrimSpace(rel) + "\n")
}

func joinRemotePath(base, rel string) string {
	rel = strings.TrimPrefix(strings.ReplaceAll(rel, "\\", "/"), "/")
	if rel == "" {
		return base
	}
	if strings.HasSuffix(base, ":") || strings.HasSuffix(base, "/") {
		return base + rel
	}
	return base + "/" + rel
}

func extractPathFromLogLine(line string) string {
	m := regexp.MustCompile(`(?:(\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2})\s+)?(INFO|NOTICE|ERROR)\s*:\s*(.+?):\s*(.+)$`).FindStringSubmatch(strings.TrimSpace(line))
	if len(m) == 0 {
		return ""
	}
	return strings.TrimSpace(m[3])
}

func extractMsgFromLogLine(line string) string {
	m := regexp.MustCompile(`(?:(\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2})\s+)?(INFO|NOTICE|ERROR)\s*:\s*(.+?):\s*(.+)$`).FindStringSubmatch(strings.TrimSpace(line))
	if len(m) == 0 {
		return ""
	}
	return strings.TrimSpace(m[4])
}

func anyString(v any) string {
	s, _ := v.(string)
	return s
}

func anyFloat64(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int64:
		return float64(x)
	case int:
		return float64(x)
	default:
		return 0
	}
}
