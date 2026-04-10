package app

import (
	"fmt"
	"strings"
)

// buildFlagsFromOptions converts task options map into rclone CLI flags.
func buildFlagsFromOptions(opt map[string]any) []string {
	flags := []string{}
	push := func(k string, vs ...string){ flags = append(flags, k); flags = append(flags, vs...) }
	asBool := func(v any) (bool, bool){ b, ok := v.(bool); return b, ok }
	asInt := func(v any) (string, bool){ switch x:=v.(type){ case float64: return fmt.Sprintf("%d", int64(x)), true; case int64: return fmt.Sprintf("%d", x), true; case string: if x!="" { return x, true }; default: } ; return "", false }
	asStr := func(v any) (string, bool){ s, ok := v.(string); if !ok || s=="" { return "", false }; return s, true }
	asArr := func(v any) ([]string, bool){
		arr := []string{}
		switch vv := v.(type){
		case []any:
			for _, e := range vv { if s, ok := e.(string); ok && s!="" { arr = append(arr, s) } }
		case []string:
			arr = vv
		}
		return arr, len(arr)>0
	}
	// common scalars
	if v, ok := asInt(opt["transfers"]); ok { push("--transfers", v) }
	if v, ok := asInt(opt["checkers"]); ok { push("--checkers", v) }
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
	// flags
	for _, key := range []string{"ignoreExisting","checksum","sizeOnly","ignoreSize","ignoreTimes","update","noTraverse","noCheckDest","inplace","immutable","checkFirst","deleteBefore","deleteDuring","deleteAfter","trackRenames","ignoreErrors","useServerModtime","refreshTimes","deleteExcluded","dryRun","serverSideAcrossConfigs"} {
		if b, ok := asBool(opt[key]); ok && b { push("--"+toKebab(key)) }
	}
	// path options
	if s, ok := asStr(opt["compareDest"]); ok { push("--compare-dest", s) }
	if s, ok := asStr(opt["copyDest"]); ok { push("--copy-dest", s) }
	// include/exclude arrays
	if arr, ok := asArr(opt["include"]); ok { for _, p := range arr { push("--include", p) } }
	if arr, ok := asArr(opt["exclude"]); ok { for _, p := range arr { push("--exclude", p) } }
	// timeouts
	if s, ok := asInt(opt["timeout"]); ok { push("--timeout", s+"s") }
	if s, ok := asInt(opt["connTimeout"]); ok { push("--contimeout", s+"s") }
	if s, ok := asInt(opt["expectContinueTimeout"]); ok { push("--expect-continue-timeout", s+"s") }
	return flags
}

func toKebab(s string) string {
	repl := map[string]string{"useServerModtime":"use-server-modtime","noCheckDest":"no-check-dest","noTraverse":"no-traverse","ignoreCase":"ignore-case","ignoreCaseSync":"ignore-case-sync","sizeOnly":"size-only","ignoreSize":"ignore-size","ignoreTimes":"ignore-times","checkFirst":"check-first","deleteBefore":"delete-before","deleteDuring":"delete-during","deleteAfter":"delete-after","trackRenames":"track-renames","ignoreErrors":"ignore-errors","bufferSize":"buffer-size","serverSideAcrossConfigs":"server-side-across-configs"}
	if v, ok := repl[s]; ok { return v }
	return strings.ToLower(strings.ReplaceAll(s, " ", "-"))
}
