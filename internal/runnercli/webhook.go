package runnercli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"strconv"

	"go.uber.org/zap"
	"rcloneflow/internal/config"
	"rcloneflow/internal/logger"
)

func toInt64(v any) int64 {
	switch x := v.(type) {
	case int:
		return int64(x)
	case int64:
		return x
	case float64:
		return int64(x)
	case json.Number:
		if i, err := x.Int64(); err == nil { return i }
		if f, err := x.Float64(); err == nil { return int64(f) }
	case string:
		if s, err := strconv.ParseInt(x, 10, 64); err == nil { return s }
		if f, err := strconv.ParseFloat(x, 64); err == nil { return int64(f) }
	}
	return 0
}

func toFloat64(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case json.Number:
		if f, err := x.Float64(); err == nil { return f }
	case string:
		if f, err := strconv.ParseFloat(x, 64); err == nil { return f }
	}
	return 0
}

func formatBytes(b int64) string {
	units := []string{"B","KB","MB","GB","TB","PB"}
	f := float64(b)
	i := 0
	for f >= 1024 && i < len(units)-1 {
		f /= 1024
		i++
	}
	if f >= 100 || f == float64(int64(f)) {
		return fmt.Sprintf("%d%s", int64(f), units[i])
	}
	return fmt.Sprintf("%.1f%s", f, units[i])
}

func formatBps(bps float64) string {
	return formatBytes(int64(bps)) + "/s"
}

func baseName(p string) string {
	if p == "" { return p }
	// handle both / and \\ separators
	lastSlash := strings.LastIndex(p, "/")
	lastBack := strings.LastIndex(p, "\\")
	idx := lastSlash
	if lastBack > idx { idx = lastBack }
	if idx >= 0 && idx+1 < len(p) { return p[idx+1:] }
	return p
}

// postWebhookIfNeeded 读取任务 options，若满足条件则发送 Webhook 通知（3 秒超时，最多 3 次重试）
func (r *Runner) postWebhookIfNeeded(runID int64) {
	// 取 run 与 task
	run, err := r.db.GetRun(runID)
	if err != nil {
		logger.Warn("webhook: get run failed", zap.Error(err), zap.Int64("runID", runID))
		return
	}
	task, ok := r.db.GetTask(run.TaskID)
	if !ok {
		logger.Warn("webhook: get task failed", zap.Int64("taskID", run.TaskID))
		return
	}
	// 解析任务 options
	var opt map[string]any
	if len(task.Options) > 0 {
		_ = json.Unmarshal(task.Options, &opt)
	}
	// 支持多个通知目标：通用 Webhook + 企业微信
	postURLs := []string{}
	if v, ok := opt["webhookPostUrl"].(string); ok {
		vv := strings.TrimSpace(v)
		if vv != "" { postURLs = append(postURLs, vv) }
	}
	if v, ok := opt["wecomPostUrl"].(string); ok {
		vv := strings.TrimSpace(v)
		if vv != "" { postURLs = append(postURLs, vv) }
	}
	if len(postURLs) == 0 {
		return // 未配置则不发送
	}
	// 触发来源过滤
	notifyOn := map[string]bool{"manual": true, "schedule": true, "webhook": true}
	if m, ok := opt["webhookNotifyOn"].(map[string]any); ok {
		notifyOn["manual"] = toBool(m["manual"])
		notifyOn["schedule"] = toBool(m["schedule"])
		notifyOn["webhook"] = toBool(m["webhook"])
	}
	if !notifyOn[strings.ToLower(run.Trigger)] {
		return
	}
	// 状态过滤（A 方案）：默认 success/failed 为 true；hasTransfer 默认 false（兼容旧配置）
	statusOn := map[string]bool{"success": true, "failed": true, "hasTransfer": false}
	if m, ok := opt["webhookNotifyStatus"].(map[string]any); ok {
		if _, ok2 := m["success"]; ok2 { statusOn["success"] = toBool(m["success"]) }
		if _, ok2 := m["failed"]; ok2 { statusOn["failed"] = toBool(m["failed"]) }
		if _, ok2 := m["hasTransfer"]; ok2 { statusOn["hasTransfer"] = toBool(m["hasTransfer"]) }
	}
	st := strings.ToLower(run.Status)
	allowed := (st == "finished" && statusOn["success"]) || (st == "failed" && statusOn["failed"])
	// 勾选了"有传输"时，只要运行结束（成功或失败）且实际有传输就发送通知，
	// 不再受 success/failed 复选框限制；hasTransfer 作为主过滤条件。
	if statusOn["hasTransfer"] && (st == "finished" || st == "failed") {
		allowed = true
	}
	if !allowed { return }
	// 构造载荷（沿用约定字段）
	payload := map[string]any{}
	payload["title"] = "RcloneFlow 任务通知"
	// 中文说明字段
	trZh := map[string]string{"manual": "手动", "schedule": "定时", "webhook": "Webhook"}
	stZh := map[string]string{"finished": "完成", "failed": "失败"}
	payload["triggerZh"] = trZh[strings.ToLower(run.Trigger)]
	payload["statusZh"] = stZh[strings.ToLower(run.Status)]
	payload["summaryZh"] = fmt.Sprintf("任务 %s 已%s", run.TaskName, payload["statusZh"])
	// task/run 概要
	// 使用 run 中的名称/模式，若为空则回退到 task（避免通知里任务名为空）
	name := run.TaskName
	if strings.TrimSpace(name) == "" { name = task.Name }
	mode := run.TaskMode
	if strings.TrimSpace(mode) == "" { mode = task.Mode }
	payload["task"] = map[string]any{"id": run.TaskID, "name": name, "mode": mode}
	runMap := map[string]any{
		"id": run.ID,
		"trigger": run.Trigger,
		"status": run.Status,
		"startedAt": run.Summary["startedAt"],
		"finishedAt": run.Summary["finishedAt"],
	}
	// summary（从 finalSummary 回填）
	sum := map[string]any{}
	if fs, ok := run.Summary["finalSummary"].(map[string]any); ok {
		sum["totalCount"] = nested(fs, "counts.total")
		sum["completedCount"] = nested(fs, "counts.copied")
		sum["failedCount"] = nested(fs, "counts.failed")
		sum["skippedCount"] = nested(fs, "counts.skipped")
		sum["totalBytes"] = fs["totalBytes"]
		sum["transferredBytes"] = fs["transferredBytes"]
		sum["avgSpeedBps"] = fs["avgSpeedBps"]
		runMap["durationSeconds"] = fs["durationSec"]
		runMap["durationText"] = fs["durationText"]
		// files: 前 N 个（由 settings.json 的 WEBHOOK_MAX_FILES 控制，默认 100）
		files := []string{}
		if arr2, ok3 := fs["files"].([]any); ok3 {
			limit := config.GetWebhookMaxFiles()
			if limit <= 0 { // 0 或负数：不限制
				for i := 0; i < len(arr2); i++ {
					m, _ := arr2[i].(map[string]any)
					if m != nil { files = append(files, fmt.Sprint(m["path"])) }
				}
			} else {
				if len(arr2) < limit { limit = len(arr2) }
				for i := 0; i < limit; i++ {
					m, _ := arr2[i].(map[string]any)
					if m != nil { files = append(files, fmt.Sprint(m["path"])) }
				}
			}
		}
		payload["files"] = files
		omitted := 0
		if v := nested(fs, "counts.total"); v != nil {
			switch tot := v.(type) {
			case int:
				if tot > len(files) { omitted = tot - len(files) }
			case float64:
				it := int(tot)
				if it > len(files) { omitted = it - len(files) }
			case string:
				if it, err := strconv.Atoi(tot); err == nil {
					if it > len(files) { omitted = it - len(files) }
				}
			}
		}
		payload["omittedCount"] = omitted
	}
	payload["run"] = runMap
	payload["summary"] = sum
	if statusOn["hasTransfer"] && !hasTransferEvidence(sum) {
		return
	}
	// 发送（3 秒超时，最多 3 次）
	for _, postURL := range postURLs {
		go func(postURL string) {
			client := &http.Client{ Timeout: 6 * time.Second }
			// 默认 body 为我们自有 schema
			body, _ := json.Marshal(payload)
			// 取公共字段
			runMap, _ := payload["run"].(map[string]any)
			sumMap, _ := payload["summary"].(map[string]any)
			filesArr, _ := payload["files"].([]string)
			omitted, _ := payload["omittedCount"].(int)
			taskName := fmt.Sprint(payload["task"].(map[string]any)["name"])
			statusZh := fmt.Sprint(payload["statusZh"]) // 完成/失败
			triggerZh := fmt.Sprint(payload["triggerZh"]) // 手动/定时/Webhook
			mode := fmt.Sprint(payload["task"].(map[string]any)["mode"]) // copy/sync/move
			okLabel := "成功"
			if strings.ToLower(mode) == "move" { okLabel = "移动" }
			total := fmt.Sprint(sumMap["totalCount"]) 
			okcnt := fmt.Sprint(sumMap["completedCount"]) 
			fail := fmt.Sprint(sumMap["failedCount"]) 
			skipped := fmt.Sprint(sumMap["skippedCount"]) 
			bytesFmt := formatBytes(toInt64(sumMap["totalBytes"]))
			txbytesFmt := formatBytes(toInt64(sumMap["transferredBytes"]))
			speedFmt := formatBps(toFloat64(sumMap["avgSpeedBps"]))
			duration := fmt.Sprint(runMap["durationText"]) 
			// 若为企业微信 webhook，改装为 markdown（带字符预算）
			if u, err := url.Parse(postURL); err == nil && strings.Contains(u.Host, "qyapi.weixin.qq.com") {
				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("**任务** <font color=\"info\">%s</font> 已%s（%s）\n", taskName, statusZh, triggerZh))
				sb.WriteString(fmt.Sprintf("> 总计: %s  %s: %s  失败: %s  其他: %s\n", total, okLabel, okcnt, fail, skipped))
				sb.WriteString(fmt.Sprintf("> 体积: %s / 已传: %s\n", bytesFmt, txbytesFmt))
				sb.WriteString(fmt.Sprintf("> 均速: %s  耗时: %s\n", speedFmt, duration))
				limitChars := 3800
				cur := sb.Len()
				shown := 0
				for _, p := range filesArr {
					name := baseName(p)
					line := "> " + name + "\n"
					if cur+len(line) > limitChars { break }
					sb.WriteString(line)
					cur += len(line)
					shown++
				}
				if omitted > 0 || shown < len(filesArr) {
					sb.WriteString("> 其他 …\n")
				}
				msg := map[string]any{"msgtype": "markdown", "markdown": map[string]string{"content": sb.String()}}
				body, _ = json.Marshal(msg)
			} else {
				// 通用 Webhook：也发送与企业微信一致的 markdown 格式，但不限制文件数量
				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("**任务** <font color=\"info\">%s</font> 已%s（%s）\n", taskName, statusZh, triggerZh))
				sb.WriteString(fmt.Sprintf("> 总计: %s  %s: %s  失败: %s  其他: %s\n", total, okLabel, okcnt, fail, skipped))
				sb.WriteString(fmt.Sprintf("> 体积: %s / 已传: %s\n", bytesFmt, txbytesFmt))
				sb.WriteString(fmt.Sprintf("> 均速: %s  耗时: %s\n", speedFmt, duration))
				for _, p := range filesArr { sb.WriteString("> "+ baseName(p) + "\n") }
				msg := map[string]any{"msgtype": "markdown", "markdown": map[string]string{"content": sb.String()}}
				body, _ = json.Marshal(msg)
			}
			for i := 0; i < 3; i++ {
				req, _ := http.NewRequest("POST", postURL, bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				resp, err := client.Do(req)
				if err == nil && resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
					_ = resp.Body.Close()
					logger.Info("webhook: posted", zap.Int64("runID", runID), zap.String("url", postURL))
					return
				}
				status := 0
				if resp != nil { status = resp.StatusCode }
				if resp != nil { _ = resp.Body.Close() }
				if status >= 400 && status < 500 { break }
				time.Sleep(500 * time.Millisecond)
			}
			logger.Warn("webhook: post failed", zap.Int64("runID", runID), zap.String("url", postURL))
		}(postURL)
	}
}

func toBool(v any) bool {
	s := strings.ToLower(fmt.Sprint(v))
	return s == "1" || s == "true" || s == "yes" || s == "on"
}

func hasTransferEvidence(sum map[string]any) bool {
	return toInt64(sum["completedCount"]) > 0
}

// nested 获取 a.b.c 路径
func nested(m map[string]any, path string) any {
	cur := any(m)
	for _, key := range strings.Split(path, ".") {
		mm, ok := cur.(map[string]any)
		if !ok { return nil }
		cur, ok = mm[key]
		if !ok { return nil }
	}
	return cur
}
