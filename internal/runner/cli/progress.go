package cli

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 解析 rclone 的进度输出（两类）：
// 1) --stats-one-line / --stats-one-line-date 的单行文本
// 2) --use-json-log 的 JSONL（当 msg=="stats" 时）

var (
	// 示例：
	// 2026/04/06 12:34:56 INFO  : 
	// Transferred:    1.234 GiB / 10.000 GiB, 12%,  1.23 MiB/s, ETA 02:34:56
	reOneLine = regexp.MustCompile(`Transferred:\s+([0-9A-Za-z\./]+)\s*/\s*([0-9A-Za-z\./]+),\s*([0-9\.]+)%\s*,\s*([0-9A-Za-z\./]+)\s*/s,\s*ETA\s*([0-9:]+)`) 
)

// DerivedProgress 为上层消费的统一结构。
type DerivedProgress struct {
	Bytes            float64    `json:"bytes"`
	TotalBytes       float64    `json:"totalBytes"`
	Percent          float64    `json:"percent"`
	SpeedBytesPerSec float64    `json:"speedBps"`
	EtaSeconds       int64      `json:"etaSec"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

// ParseProgressLine 解析一行输出，成功返回 (progress, true)。
func ParseProgressLine(line string) (DerivedProgress, bool) {
	line = strings.TrimSpace(line)
	if line == "" { return DerivedProgress{}, false }

	// JSON 路径
	if strings.HasPrefix(line, "{") {
		if p, ok := parseJSONProgress(line); ok {
			return p, true
		}
	}
	// 单行文本路径
	if p, ok := parseOneLineProgress(line); ok {
		return p, true
	}
	return DerivedProgress{}, false
}

// 解析 JSON stats 日志。
func parseJSONProgress(line string) (DerivedProgress, bool) {
	var m map[string]any
	if err := json.Unmarshal([]byte(line), &m); err != nil { return DerivedProgress{}, false }
	if msg, _ := m["msg"].(string); msg != "stats" { return DerivedProgress{}, false }
	// 常见结构：{"msg":"stats","stats":{"bytes":123,"totalBytes":999,"speed":12345,"eta":123,"percentage":12.3}}
	stat, _ := m["stats"].(map[string]any)
	if stat == nil { return DerivedProgress{}, false }
	p := DerivedProgress{UpdatedAt: time.Now()}
	p.Bytes = toInt64(stat["bytes"]) 
	if p.TotalBytes = toInt64(stat["totalBytes"]); p.TotalBytes == 0 {
		// 某些版本字段名不同，兼容 total / total_bytes
		if v := toInt64(stat["total"]); v > 0 { p.TotalBytes = v }
		if v := toInt64(stat["total_bytes"]); v > 0 { p.TotalBytes = v }
	}
	if v := toFloat64(stat["percentage"]); v > 0 { p.Percent = v }
	// 速度字段可能为 speed / speedAvg / bytesPerSecond
	if v := toFloat64(stat["speed"]); v > 0 { p.SpeedBytesPerSec = v }
	if v := toFloat64(stat["speedAvg"]); v > 0 { p.SpeedBytesPerSec = v }
	if v := toFloat64(stat["bytesPerSecond"]); v > 0 { p.SpeedBytesPerSec = v }
	if v := toInt64(stat["eta"]); v > 0 { p.EtaSeconds = v }
	return p, true
}

// 解析单行文本 stats。
func parseOneLineProgress(line string) (DerivedProgress, bool) {
	// 允许前缀时间/级别，截取 "Transferred:" 起始
	idx := strings.Index(line, "Transferred:")
	if idx >= 0 { line = line[idx:] }
	m := reOneLine.FindStringSubmatch(line)
	if len(m) != 6 { return DerivedProgress{}, false }
	p := DerivedProgress{UpdatedAt: time.Now()}
	p.Bytes = parseHumanBytes(m[1])
	p.TotalBytes = parseHumanBytes(m[2])
	p.Percent, _ = strconv.ParseFloat(m[3], 64)
	p.SpeedBytesPerSec = parseHumanBytes(m[4])
	p.EtaSeconds = parseEta(m[5])
	return p, true
}

// parseEta 将 "HH:MM:SS" 或 "MM:SS" 转为秒。
func parseEta(s string) int64 {
	parts := strings.Split(s, ":")
	if len(parts) == 2 { // MM:SS
		m, _ := strconv.Atoi(parts[0])
		sec, _ := strconv.Atoi(parts[1])
		return int64(m*60 + sec)
	}
	if len(parts) == 3 { // HH:MM:SS
		h, _ := strconv.Atoi(parts[0])
		m, _ := strconv.Atoi(parts[1])
		sec, _ := strconv.Atoi(parts[2])
		return int64(h*3600 + m*60 + sec)
	}
	return 0
}

// 解析带单位的容量（B/KiB/MiB/GiB/TiB 或 KB/MB/GB/TB）。
func parseHumanBytes(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" { return 0 }
	// 去掉 "/s" 等速率标注
	s = strings.TrimSuffix(s, "/s")
	// 兼容逗号/空格
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, " ", "")
	// 判单位
	units := []struct{ suf string; mul float64 }{
		{"TiB", 1024 * 1024 * 1024 * 1024},
		{"GiB", 1024 * 1024 * 1024},
		{"MiB", 1024 * 1024},
		{"KiB", 1024},
		{"TB", 1000 * 1000 * 1000 * 1000},
		{"GB", 1000 * 1000 * 1000},
		{"MB", 1000 * 1000},
		{"KB", 1000},
		{"B", 1},
	}
	for _, u := range units {
		if strings.HasSuffix(s, u.suf) {
			v := strings.TrimSuffix(s, u.suf)
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f * u.mul
			}
			return 0
		}
	}
	// 无单位，按字节解析
	if f, err := strconv.ParseFloat(s, 64); err == nil { return f }
	return 0
}

func toInt64(v any) int64 {
	switch x := v.(type) {
	case float64:
		return int64(x)
	case int64:
		return x
	case int:
		return int64(x)
	case string:
		if n, err := strconv.ParseInt(x, 10, 64); err == nil { return n }
	}
	return 0
}

func toFloat64(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int64:
		return float64(x)
	case int:
		return float64(x)
	case string:
		if n, err := strconv.ParseFloat(x, 64); err == nil { return n }
	}
	return 0
}
