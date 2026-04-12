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
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"go.uber.org/zap"
	"rcloneflow/internal/adapter"
	"rcloneflow/internal/config"
	"rcloneflow/internal/logger"
	"rcloneflow/internal/store"
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
	// Resolve config path
	dataDir := os.Getenv("APP_DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	cfg := os.Getenv("RCLONE_CONFIG")
	if cfg == "" {
		cfg = filepath.Join(dataDir, "rclone.conf")
	}
	// Base args：非交互环境使用 --stats-one-line（不与 --progress 同用）
	// 降低默认日志级别：从 -vv 改为 -v，显著减少日志行数和解析/写库开销
	args := []string{cmdName, src, dst, "-v", "--stats", "5s", "--stats-one-line", "--config", cfg}
	// 可选：启用 JSON 日志（某些后端可能不兼容，默认关闭）
	if strings.EqualFold(os.Getenv("RCLONE_USE_JSON_LOG"), "true") || os.Getenv("RCLONE_USE_JSON_LOG") == "1" {
		args = append(args, "--use-json-log", "--log-level", "INFO", "--stats-log-level", "INFO")
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
	// 二次兜底：如 --buffer-size/--bwlimit 后是纯数字，自动补单位（M）
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--buffer-size" || args[i] == "--bwlimit" {
			n := strings.TrimSpace(args[i+1])
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

	// Preflight (approx) total count/size if enabled
	if strings.EqualFold(strings.TrimSpace(config.GetPrecheckMode()), "size") {
		if b, c, e := sizeOf(&adapter.CmdRunner{}, cfg, src, effOpt); e == nil {
			_ = r.db.UpdateRun(run.ID, func(rr *store.Run) {
				if rr.Summary == nil {
					rr.Summary = map[string]any{}
				}
				rr.Summary["preflight"] = map[string]any{"totalCount": c, "totalBytes": b}
			})
		}
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
	go r.consume(run.ID, outR, stderrFile, false, fileStats)
	go r.consume(run.ID, errR, stderrFile, true, fileStats)
	go func() {
		err := cmd.Wait()
		outW.Close()
		errW.Close()
		stderrFile.Close()
		if err != nil || (cmd.ProcessState != nil && !cmd.ProcessState.Success()) {
			_ = r.db.UpdateRun(run.ID, func(rr *store.Run) {
				rr.Status = "failed"
				if rr.Summary == nil {
					rr.Summary = map[string]any{}
				}
				rr.Summary["finished"] = true
				rr.Summary["success"] = false
				fin := time.Now().Local()
				// 若无 progress 但有 stableProgress，则回填，便于历史详情展示
				if _, ok := rr.Summary["progress"]; !ok {
					if sp, ok2 := rr.Summary["stableProgress"].(map[string]any); ok2 {
						rr.Summary["progress"] = sp
					}
				}
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
				// 体量/均速
				var prog map[string]any
				if p, ok := rr.Summary["progress"].(map[string]any); ok {
					prog = p
				} else if sp, ok := rr.Summary["stableProgress"].(map[string]any); ok {
					prog = sp
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
					if b, e := os.ReadFile(p); e == nil {
						lines := strings.Split(string(b), "\n")
						// pre-scan file total sizes to populate sizeBytes in finalSummary.files
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

						re := regexp.MustCompile(`(?:(\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2})\s+)?(INFO|NOTICE|ERROR)\s*:\s*(.+?):\s*(.+)$`)
						tsRe := regexp.MustCompile(`\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2}\s+(?:INFO|NOTICE|ERROR)\s*:`)
						for _, ln := range lines {
							l := strings.TrimSpace(ln)
							if l == "" {
								continue
							}
							segments := []string{}
							idx := tsRe.FindAllStringIndex(l, -1)
							if len(idx) > 1 {
								for i := 0; i < len(idx); i++ {
									startI := idx[i][0]
									endI := len(l)
									if i+1 < len(idx) {
										endI = idx[i+1][0]
									}
									segments = append(segments, strings.TrimSpace(l[startI:endI]))
								}
							} else {
								segments = []string{l}
							}
							for _, seg := range segments {
								m := re.FindStringSubmatch(seg)
								if len(m) == 0 {
									continue
								}
								at := strings.TrimSpace(m[1])
								level := strings.ToUpper(strings.TrimSpace(m[2]))
								path := strings.TrimSpace(m[3])
								msg := strings.TrimSpace(m[4])
								row := map[string]any{"path": path, "at": at, "status": "", "action": "", "sizeBytes": 0}
								if sz, ok := sizes[path]; ok {
									row["sizeBytes"] = sz
								}
								low := strings.ToLower(msg)
								if level == "ERROR" {
									row["status"] = "failed"
									row["action"] = "Error"
									counts["failed"]++
								} else {
									switch {
									case strings.Contains(low, "copied"):
										row["status"] = "success"
										row["action"] = "Copied"
										counts["copied"]++
									case strings.Contains(low, "deleted") || strings.Contains(low, "removed"):
										row["status"] = "success"
										row["action"] = "Deleted"
										counts["deleted"]++
									case strings.Contains(low, "skipped"):
										row["status"] = "skipped"
										row["action"] = "Skipped"
										counts["skipped"]++
									default:
										continue
									}
								}
								files = append(files, row)
								counts["total"]++
							}
						}
					}
				}
				// 如果是 move 模式：将成对的 Copied+Deleted 合并为 Moved，并调整计数（失败态也应用）
				if strings.ToLower(cmdName) == "move" {
					copiedMap := map[string]map[string]any{}
					deletedMap := map[string]map[string]any{}
					for _, f := range files {
						a := strings.ToLower(fmt.Sprint(f["action"]))
						p := fmt.Sprint(f["path"])
						if a == "copied" { copiedMap[p] = f }
						if a == "deleted" { deletedMap[p] = f }
					}
					moved := []map[string]any{}
					others := []map[string]any{}
					for p, c := range copiedMap {
						if _, ok := deletedMap[p]; ok {
							moved = append(moved, map[string]any{"path": p, "at": c["at"], "status": "success", "action": "Moved", "sizeBytes": c["sizeBytes"]})
							delete(deletedMap, p)
						} else {
							others = append(others, c)
						}
					}
					for _, d := range deletedMap { others = append(others, d) }
					files = append(moved, others...)
					counts["copied"] = len(moved)
					counts["deleted"] = 0
					counts["total"] = counts["copied"] + counts["failed"] + counts["skipped"]
				}
				// Fallback enrichment: lsjson target to fill missing sizeBytes
				if len(files) > 0 {
					cr := &adapter.CmdRunner{}
					if out, _, e2 := cr.Run(context.Background(), []string{"lsjson", dst, "--config", cfg, "--files-only", "--recursive"}...); e2 == nil {
						var arr []map[string]any
						if json.Unmarshal([]byte(out), &arr) == nil {
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
								}
							}
							if len(m) > 0 {
								for i := range files {
									if szAny, ok := files[i]["sizeBytes"]; !ok || fmt.Sprint(szAny) == "0" {
										p := strings.ReplaceAll(fmt.Sprint(files[i]["path"]), "\\", "/")
										if sz, ok2 := m[p]; ok2 && sz > 0 {
											files[i]["sizeBytes"] = sz
										}
									}
								}
							}
						}
					}
				}
				rr.Summary["finalSummary"] = map[string]any{"counts": counts, "files": files, "startAt": finalSummary["startAt"], "finishedAt": finalSummary["finishedAt"], "durationSec": durSec, "durationText": humanDuration(durSec), "result": "failed", "transferredBytes": bytes, "totalBytes": total, "avgSpeedBps": avg}
			})
			r.mu.Lock()
			delete(r.procs, run.ID)
			r.mu.Unlock()
			return
		}
		// WebDAV 完成确认（copy/sync/move 通用）：仅做可见性检查，不改变 rclone 语义
		if isWebDAVUnderlying(cfg, dstRemote) {
			interval := config.GetFinishWaitInterval()
			timeout := config.GetFinishWaitTimeout()
			if timeout > 0 {
				vr := &adapter.CmdRunner{}
				deadline := time.Now().Add(timeout)
				for time.Now().Before(deadline) {
					allOk := true
					// 简化：检查目标目录可读可见；不做逐文件 size 校验
					args := []string{"lsjson", dst, "--config", cfg, "--files-only"}
					out, _, e := vr.Run(context.Background(), args...)
					if e != nil {
						allOk = false
					}
					var arr []map[string]any
					if json.Unmarshal([]byte(out), &arr) != nil {
						allOk = false
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
			// 若无 progress 但有 stableProgress，则将最后稳态快照固化为 progress（供历史页展示）
			if _, ok := rr.Summary["progress"]; !ok {
				if sp, ok2 := rr.Summary["stableProgress"].(map[string]any); ok2 {
					rr.Summary["progress"] = sp
				}
			}
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
			// 体量/均速（从 progress 或 stableProgress 回填）
			var prog map[string]any
			if p, ok := rr.Summary["progress"].(map[string]any); ok {
				prog = p
			} else if sp, ok := rr.Summary["stableProgress"].(map[string]any); ok {
				prog = sp
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
				if b, e := os.ReadFile(p); e == nil {
					lines := strings.Split(string(b), "\n")
					re := regexp.MustCompile(`(?:(\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2})\s+)?(INFO|NOTICE|ERROR)\s*:\s*(.+?):\s*(.+)$`)
					tsRe := regexp.MustCompile(`\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2}\s+(?:INFO|NOTICE|ERROR)\s*:`)
					for _, ln := range lines {
						l := strings.TrimSpace(ln)
						if l == "" {
							continue
						}
						segments := []string{}
						idx := tsRe.FindAllStringIndex(l, -1)
						if len(idx) > 1 {
							for i := 0; i < len(idx); i++ {
								startI := idx[i][0]
								endI := len(l)
								if i+1 < len(idx) {
									endI = idx[i+1][0]
								}
								segments = append(segments, strings.TrimSpace(l[startI:endI]))
							}
						} else {
							segments = []string{l}
						}
						for _, seg := range segments {
							m := re.FindStringSubmatch(seg)
							if len(m) == 0 {
								continue
							}
							at := strings.TrimSpace(m[1])
							level := strings.ToUpper(strings.TrimSpace(m[2]))
							path := strings.TrimSpace(m[3])
							msg := strings.TrimSpace(m[4])
							row := map[string]any{"path": path, "at": at, "status": "", "action": "", "sizeBytes": 0}
							low := strings.ToLower(msg)
							if level == "ERROR" {
								row["status"] = "failed"
								row["action"] = "Error"
								counts["failed"]++
							} else {
								switch {
								case strings.Contains(low, "copied"):
									row["status"] = "success"
									row["action"] = "Copied"
									counts["copied"]++
								case strings.Contains(low, "deleted") || strings.Contains(low, "removed"):
									row["status"] = "success"
									row["action"] = "Deleted"
									counts["deleted"]++
								case strings.Contains(low, "skipped"):
									row["status"] = "skipped"
									row["action"] = "Skipped"
									counts["skipped"]++
								default:
									continue
								}
							}
							files = append(files, row)
							counts["total"]++
						}
					}
					// 如果是 move 模式：将成对的 Copied+Deleted 合并为 Moved，并调整计数
					if strings.ToLower(cmdName) == "move" {
						copiedMap := map[string]map[string]any{}
						deletedMap := map[string]map[string]any{}
						for _, f := range files {
							a := strings.ToLower(fmt.Sprint(f["action"]))
							p := fmt.Sprint(f["path"])
							if a == "copied" {
								copiedMap[p] = f
							}
							if a == "deleted" {
								deletedMap[p] = f
							}
						}
						moved := []map[string]any{}
						others := []map[string]any{}
						for p, c := range copiedMap {
							if _, ok := deletedMap[p]; ok {
								moved = append(moved, map[string]any{"path": p, "at": c["at"], "status": "success", "action": "Moved", "sizeBytes": 0})
								delete(deletedMap, p)
							} else {
								others = append(others, c)
							}
						}
						for p, d := range deletedMap {
							_ = p
							others = append(others, d)
						}
						files = append(moved, others...)
						// 计数：将 copied 置为 moved 数，deleted 置 0（成功= copied）
						counts["copied"] = len(moved)
						counts["deleted"] = 0
						counts["total"] = counts["copied"] + counts["failed"] + counts["skipped"]
					}
				}
			}
			// Fallback enrichment: lsjson target to fill missing sizeBytes
			if len(files) > 0 {
				cr := &adapter.CmdRunner{}
				if out, _, e2 := cr.Run(context.Background(), []string{"lsjson", dst, "--config", cfg, "--files-only", "--recursive"}...); e2 == nil {
					var arr []map[string]any
					if json.Unmarshal([]byte(out), &arr) == nil {
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
							}
						}
						if len(m) > 0 {
							for i := range files {
								if szAny, ok := files[i]["sizeBytes"]; !ok || fmt.Sprint(szAny) == "0" {
									p := strings.ReplaceAll(fmt.Sprint(files[i]["path"]), "\\", "/")
									if sz, ok2 := m[p]; ok2 && sz > 0 {
										files[i]["sizeBytes"] = sz
									}
								}
							}
						}
					}
				}
			}
			finalSummary["counts"] = counts
			finalSummary["files"] = files
			rr.Summary["finalSummary"] = finalSummary
		})
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

func sizeOf(r *adapter.CmdRunner, cfg, target string, opts map[string]any) (bytes int64, count int64, err error) {
	// Prefer JSON when available. Fallback to parsing text if needed.
	args := []string{"size", target, "--config", cfg, "--json"}
	// 透传过滤/列表相关参数以贴近任务过滤集合
	if opts != nil {
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
					if s := strings.ToLower(fmt.Sprint(v)); s == "true" || s == "1" {
						args = append(args, "--fast-list")
					}
					continue
				}
				args = append(args, flag, fmt.Sprint(v))
			}
		}
	}
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

var fileLineRe = regexp.MustCompile(`(?i)INFO\s*:\s*([^:]+):\s*(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)B\s*/\s*(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)B,\s*(\d+(?:\.\d+)?)%?,\s*(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)B/s`)
var fileCopiedRe = regexp.MustCompile(`(?i)INFO\s*:\s*([^:]+):\s*Copied\s*\(new\)`)

func (r *Runner) consume(runID int64, rd io.Reader, out *os.File, parseStats bool, fp *fileProgress) {
	wantParse := parseStats
	s := bufio.NewScanner(rd)
	s.Buffer(make([]byte, 0, 128*1024), 2*1024*1024)
	for s.Scan() {
		line := s.Text()
		if len(line) > 0 {
			_, _ = out.WriteString(line + "\n")
		}
		// 1) JSON 行（极少数情况下）
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
			if len(prog) > 0 {
				_ = r.db.UpdateRun(runID, func(rr *store.Run) {
					if rr.Summary == nil {
						rr.Summary = map[string]any{}
					}
					rr.Summary["progress"] = prog
					if b, ok := prog["bytes"].(float64); ok {
						rr.BytesTransferred = int64(b)
					}
					if sp, ok := prog["speed"].(float64); ok {
						rr.Speed = fmt.Sprintf("%d B/s", int64(sp))
					}
				})
				continue
			}
		}
		// 2) 文本 one-line 解析（仅在需要时）
		if wantParse {
			if prog, ok := parseOneLineProgress(line); ok {
				_ = r.db.UpdateRun(runID, func(rr *store.Run) {
					if rr.Summary == nil {
						rr.Summary = map[string]any{}
					}
					// preserve non-decreasing completedFiles; fallback to copied list if needed
					if prev, ok := rr.Summary["progress"].(map[string]any); ok {
						if pc, ok2 := prev["completedFiles"].(float64); ok2 {
							if nc, ok3 := prog["completedFiles"].(float64); ok3 {
								if nc < pc { prog["completedFiles"] = pc }
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
					// 同步到 stableProgress，前端统一读取 DB-only 稳态（包含 completedFiles）
					rr.Summary["stableProgress"] = prog
					rr.BytesTransferred = int64(prog["bytes"].(float64))
					rr.Speed = fmt.Sprintf("%d B/s", int64(prog["speed"].(float64)))
					// 同步部分文件列表快照（最近 100 条）
					if fp != nil {
						rr.Summary["files"] = fp.snapshot(100)
					}
				})
			}
			// 解析文件级进度（INFO: name: cur/total, pct, speed）
			if fp != nil {
				if m := fileLineRe.FindStringSubmatch(line); len(m) > 0 {
					name := strings.TrimSpace(m[1])
					var cb, tb, pct, sp float64
					fmt.Sscanf(m[2], "%f", &cb)
					fmt.Sscanf(m[4], "%f", &tb)
					fmt.Sscanf(m[6], "%f", &pct)
					fmt.Sscanf(m[7], "%f", &sp)
					fp.update(name, cb*unitToMul(m[3]), tb*unitToMul(m[5]), sp*unitToMul(m[8]), pct)
					// 若该文件进度已达到或超过 100%，也视为已完成（补齐非英文环境缺少 "Copied (new)" 的情况）
					if (tb > 0 && cb >= tb) || pct >= 100 {
						fp.markCopied(name)
					}
					continue
				}
				if m := fileCopiedRe.FindStringSubmatch(line); len(m) > 0 {
					name := strings.TrimSpace(m[1])
					fp.update(name, -1, -1, -1, 100)
					fp.markCopied(name)
				}
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

func parseETA(s string) int {
	s = strings.TrimSpace(s)
	if s == "-" || s == "" {
		return 0
	}
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

var bytesPairRe = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)(?:B)?\s*/\s*(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)(?:B)?`)
var speedTokenRe = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*([KMGTPE]?i?)(?:B)?/s`)
var pctTokenRe = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)%`)
var etaTokenRe = regexp.MustCompile(`(?i)ETA\s*([0-9hms:.-]+|-)`)

func parseOneLineProgress(line string) (map[string]any, bool) {
	l := strings.TrimSpace(line)
	// 提取 (xfr#a/b) 的已完成数 a
	xfrDone := float64(0)
	if i := strings.Index(l, "("); i >= 0 {
		// 截取括号段与主段
		paren := l[i:]
		l = strings.TrimSpace(l[:i])
		if j := strings.Index(strings.ToLower(paren), "xfr#"); j >= 0 {
			var a int
			fmt.Sscanf(paren[j:], "xfr#%d", &a)
			if a > 0 {
				xfrDone = float64(a)
			}
		}
	}
	// 按多段拼接处理：取最后一个匹配片段作为当前进度
	bps := bytesPairRe.FindAllStringSubmatch(l, -1)
	if len(bps) == 0 {
		return nil, false
	}
	bp := bps[len(bps)-1]
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
	return prog, true
}

func spmValue(m []string, i int) string {
	if len(m) > i {
		return m[i]
	}
	return ""
}
