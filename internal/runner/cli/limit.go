package cli

import (
	"os"
	"strconv"
)

// 运行并发上限（简单信号量），默认 2，可通过环境变量 RCLONE_MAX_PROCS 配置。
var (
	sem = make(chan struct{}, maxProcsFromEnv())
)

func maxProcsFromEnv() int {
	v := os.Getenv("RCLONE_MAX_PROCS")
	if v == "" { return 2 }
	if n, err := strconv.Atoi(v); err == nil && n > 0 { return n }
	return 2
}

func acquire() { sem <- struct{}{} }
func release() { <-sem }
