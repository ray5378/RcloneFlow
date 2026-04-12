package main

import (
	"rcloneflow/internal/app"
	"rcloneflow/internal/config"
)

func main() {
	cfg, _ := config.Load("")
	_ = app.Run(cfg)
}
