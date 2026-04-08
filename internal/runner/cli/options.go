// 映射前端「高级选项」到 rclone 命令行旗标的占位定义。
// 后续提交会逐步把字段补齐，并提供安全的 args 生成函数（不拼接字符串）。
package cli

// TaskCLIOptions 代表一次任务可配置的核心选项（精简占位）。
type TaskCLIOptions struct {
	// 基本路径
	Src string // 源，例如："FNOS:/HDD/xxx"
	Dst string // 目标，例如："crypt:/dst"

	// 传输/流式
	Transfers            int    // --transfers
	MultiThreadStreams   int    // --multi-thread-streams
	MultiThreadCutoff    string // --multi-thread-cutoff（如 "64M"）
	BufferSize           string // --buffer-size（如 "16M"）

	// 对比/校验
	UseServerModtime bool // --use-server-modtime
	SizeOnly        bool // --size-only

	// 可靠性/时间
	Retries          int    // --retries
	LowLevelRetries  int    // --low-level-retries
	Timeout          string // --timeout（如 "12h"）
	ConnTimeout      string // --contimeout（如 "60s"）
	ExpectContTimeout string // --expect-continue-timeout（如 "10s"）

	// 统计/日志
	StatsInterval string // --stats（如 "5s"）
	JSONLog       bool   // --use-json-log
	LogLevel      string // --log-level（NOTICE/INFO/DEBUG）
}

// 生成 rclone 命令行参数的占位函数。
// 注意：仅返回参数切片，由 os/exec 以 args 方式调用，避免命令注入。
func BuildCopyArgs(o TaskCLIOptions) []string {
	args := []string{"copy", o.Src, o.Dst}
	if o.Transfers > 0 {
		args = append(args, "--transfers", itoa(o.Transfers))
	}
	if o.MultiThreadStreams > 0 {
		args = append(args, "--multi-thread-streams", itoa(o.MultiThreadStreams))
	}
	if o.MultiThreadCutoff != "" { args = append(args, "--multi-thread-cutoff", o.MultiThreadCutoff) }
	if o.BufferSize != "" { args = append(args, "--buffer-size", o.BufferSize) }
	if o.UseServerModtime { args = append(args, "--use-server-modtime") }
	if o.SizeOnly { args = append(args, "--size-only") }
	if o.Retries > 0 { args = append(args, "--retries", itoa(o.Retries)) }
	if o.LowLevelRetries > 0 { args = append(args, "--low-level-retries", itoa(o.LowLevelRetries)) }
	if o.Timeout != "" { args = append(args, "--timeout", o.Timeout) }
	if o.ConnTimeout != "" { args = append(args, "--contimeout", o.ConnTimeout) }
	if o.ExpectContTimeout != "" { args = append(args, "--expect-continue-timeout", o.ExpectContTimeout) }
	if o.StatsInterval != "" { args = append(args, "--stats", o.StatsInterval) }
	if o.JSONLog { args = append(args, "--use-json-log") }
	if o.LogLevel != "" { args = append(args, "--log-level", o.LogLevel) }
	return args
}

// 精简版 itoa，后续会替换为 strconv.Itoa。
func itoa(n int) string { return fmtInt(n) }

// 为了不引入多余依赖，这里放一个极简整数转字符串（后续用 strconv 替换）。
func fmtInt(n int) string {
	// 简化实现：仅用于占位，避免编译器未使用报错；后续提交改为 strconv.Itoa
	b := []byte{}
	if n == 0 { return "0" }
	neg := false
	if n < 0 { neg = true; n = -n }
	for n > 0 {
		d := n % 10
		b = append([]byte{byte('0'+d)}, b...)
		n /= 10
	}
	if neg { b = append([]byte{'-'}, b...) }
	return string(b)
}
