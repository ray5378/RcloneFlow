package runnercli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"rcloneflow/internal/logger"
	"rcloneflow/internal/store"
)

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
						if tb > 0 {
							computed := (cb / tb) * 100
							pctPtr = &computed
						} else if v, ok := item["percentage"].(float64); ok {
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
							_ = r.updater.UpdateRun(runID, func(rr *store.Run) {
								if rr.Summary == nil {
									rr.Summary = map[string]any{}
								}
								rr.Summary["currentFile"] = currentFile
							})
						}
					}
					if len(currentFiles) > 0 {
						_ = r.updater.UpdateRun(runID, func(rr *store.Run) {
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
						if _, ok3 := prog["percentage"]; !ok3 {
							prog["percentage"] = v
						}
					}
				}
				recomputeProgressPct(prog)
				_ = r.updater.UpdateRun(runID, func(rr *store.Run) {
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
				r.broadcaster.Broadcast("run_progress", map[string]any{
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
			recomputeProgressPct(prog)
			_ = r.updater.UpdateRun(runID, func(rr *store.Run) {
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
			r.broadcaster.Broadcast("run_progress", map[string]any{
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
				recomputeProgressPct(prog)
				_ = r.updater.UpdateRun(runID, func(rr *store.Run) {
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
				r.broadcaster.Broadcast("run_progress", map[string]any{
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
				_ = r.updater.UpdateRun(runID, func(rr *store.Run) {
					if rr.Summary == nil {
						rr.Summary = map[string]any{}
					}
					rr.Summary["progressLine"] = line
				})
				if marked {
					_ = r.updater.UpdateRun(runID, func(rr *store.Run) {
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
				if isRunObjectNotFoundSummary(path, msg) || isAttemptObjectNotFoundSummary(path, msg) {
					continue
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
						_ = r.updater.UpdateRun(runID, func(rr *store.Run) {
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
				_ = r.updater.UpdateRun(runID, func(rr *store.Run) {
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