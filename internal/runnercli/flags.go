package runnercli

import (
	"fmt"
	"strings"
)

func buildFlagsFromOptions(opt map[string]any) []string {
	flags := []string{}
	push := func(k string, vs ...string){ flags = append(flags, k); flags = append(flags, vs...) }
	asBool := func(v any) (bool, bool){ b, ok := v.(bool); return b, ok }
	asInt := func(v any) (string, bool){
		switch x := v.(type) {
		case float64:
			return fmt.Sprintf("%d", int64(x)), true
		case int64:
			return fmt.Sprintf("%d", x), true
		case int:
			return fmt.Sprintf("%d", x), true
		case string:
			s := strings.TrimSpace(x)
			if s != "" { return s, true }
		}
		return "", false
	}
	asStr := func(v any) (string, bool){ s, ok := v.(string); if !ok { return "", false }; s = strings.TrimSpace(s); if s=="" { return "", false }; return s, true }
	asArr := func(v any) ([]string, bool){
		arr := []string{}
		switch vv := v.(type){
		case []any:
			for _, e := range vv { if s, ok := e.(string); ok { s = strings.TrimSpace(s); if s!="" { arr = append(arr, s) } } }
		case []string:
			for _, s := range vv { s = strings.TrimSpace(s); if s!="" { arr = append(arr, s) } }
		case string:
			s := strings.TrimSpace(vv)
			if s != "" {
				s = strings.ReplaceAll(s, "\r", "")
				for _, line := range strings.Split(s, "\n") {
					line = strings.TrimSpace(line)
					if line == "" { continue }
					for _, part := range strings.Split(line, ",") {
						p := strings.TrimSpace(part)
						if p != "" { arr = append(arr, p) }
					}
				}
			}
		}
		return arr, len(arr)>0
	}

	// 常用数值
	if v, ok := asInt(opt["transfers"]); ok { push("--transfers", v) }
	if v, ok := asInt(opt["checkers"]); ok { push("--checkers", v) }
	if v, ok := asInt(opt["retries"]); ok { push("--retries", v) }
	if v, ok := asInt(opt["lowLevelRetries"]); ok { push("--low-level-retries", v) }
	// bufferSize：数字→自动补 M；字符串→纯数字则补 M，带单位则原样
	if raw, ok := opt["bufferSize"]; ok {
		switch vv := raw.(type) {
		case float64:
			push("--buffer-size", fmt.Sprintf("%dM", int64(vv)))
		case int64:
			push("--buffer-size", fmt.Sprintf("%dM", vv))
		case string:
			s := strings.TrimSpace(vv)
			if s != "" {
				pureNum := true
				for _, ch := range s { if ch < '0' || ch > '9' { pureNum = false; break } }
				if pureNum { push("--buffer-size", s+"M") } else { push("--buffer-size", s) }
			}
		}
	}
	if s, ok := asStr(opt["bwLimit"]); ok { push("--bwlimit", s) }
	if s, ok := asStr(opt["bwlimit"]); ok { push("--bwlimit", s) }

	// 布尔开关
	for _, key := range []string{"ignoreExisting","checksum","sizeOnly","ignoreSize","ignoreTimes","update","noTraverse","noCheckDest","inplace","immutable","checkFirst","deleteBefore","deleteDuring","deleteAfter","trackRenames","ignoreErrors","useServerModtime","refreshTimes","deleteExcluded","dryRun","serverSideAcrossConfigs","interactive"} {
		if b, ok := asBool(opt[key]); ok && b { push("--"+toKebab(key)) }
	}
	// 特殊布尔：disableHttp2 → --disable-http2
	if b, ok := asBool(opt["disableHttp2"]); ok && b { push("--disable-http2") }
	// 路径类
	if s, ok := asStr(opt["compareDest"]); ok { push("--compare-dest", s) }
	if s, ok := asStr(opt["copyDest"]); ok { push("--copy-dest", s) }
	if s, ok := asStr(opt["backupDir"]); ok { push("--backup-dir", s) }

	// include/exclude：收集→归一→去重→拼接
	norm := func(s string) string {
		s = strings.TrimSpace(s)
		if s == "" { return s }
		if strings.HasPrefix(s, ".") && !strings.ContainsAny(s, "*?[") { return "*"+s }
		return s
	}
	// include
	{
		incSet := map[string]struct{}{}
		if arr, ok := asArr(opt["include"]); ok { for _, p := range arr { p = norm(p); if p!="" { incSet[p] = struct{}{} } } }
		if s, ok := asStr(opt["include"]); ok { s = norm(s); if s!="" { incSet[s] = struct{}{} } }
		for p := range incSet { push("--include", p) }
	}
	// exclude
	{
		excSet := map[string]struct{}{}
		if arr, ok := asArr(opt["exclude"]); ok { for _, p := range arr { p = norm(p); if p!="" { excSet[p] = struct{}{} } } }
		if s, ok := asStr(opt["exclude"]); ok { s = norm(s); if s!="" { excSet[s] = struct{}{} } }
		for p := range excSet { push("--exclude", p) }
	}

	// 超时：仅在显式配置时映射
	if s, ok := asInt(opt["timeout"]); ok { push("--timeout", s+"s") }
	if s, ok := asInt(opt["connTimeout"]); ok { push("--contimeout", s+"s") }
	if s, ok := asInt(opt["expectContinueTimeout"]); ok { push("--expect-continue-timeout", s+"s") }
	// 其他
	if s, ok := asStr(opt["bwLimit"]); ok { push("--bwlimit", s) }
	if s, ok := asInt(opt["maxTransfer"]); ok { push("--max-transfer", s) }
	if s, ok := asInt(opt["maxDuration"]); ok { push("--max-duration", s+"s") }
	if s, ok := asStr(opt["logFile"]); ok { push("--log-file", s) }
	return flags
}

func toKebab(s string) string {
	repl := map[string]string{"useServerModtime":"use-server-modtime","noCheckDest":"no-check-dest","noTraverse":"no-traverse","ignoreCase":"ignore-case","ignoreCaseSync":"ignore-case-sync","sizeOnly":"size-only","ignoreSize":"ignore-size","ignoreTimes":"ignore-times","checkFirst":"check-first","deleteBefore":"delete-before","deleteDuring":"delete-during","deleteAfter":"delete-after","trackRenames":"track-renames","ignoreErrors":"ignore-errors","bufferSize":"buffer-size","serverSideAcrossConfigs":"server-side-across-configs"}
	if v, ok := repl[s]; ok { return v }
	return strings.ToLower(strings.ReplaceAll(s, " ", "-"))
}
