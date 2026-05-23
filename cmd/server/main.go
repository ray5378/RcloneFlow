package main

import (
	"fmt"
	"os"

	"rcloneflow/internal/app"
	"rcloneflow/internal/config"
)

func main() {
	if err := RunMain(); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

func RunMain() error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("配置加载失败: %w", err)
	}
	if err := app.Run(cfg); err != nil {
		return fmt.Errorf("启动失败: %w", err)
	}
	return nil
}
