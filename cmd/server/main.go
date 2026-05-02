package main

import (
	"os"

	"rcloneflow/internal/app"
	"rcloneflow/internal/config"
)

func main() {
	// 加载配置
	cfg, err := config.Load("")
	if err != nil {
		os.Stderr.WriteString("配置加载失败: " + err.Error() + "\n")
		os.Exit(1)
	}

	// 启动服务
	if err := app.Run(cfg); err != nil {
		os.Stderr.WriteString("启动失败: " + err.Error() + "\n")
		os.Exit(1)
	}
}
