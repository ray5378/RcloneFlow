package rclone

import (
	"bytes"
	"os/exec"
)

// CLIConfig 封装 rclone config 的最小等价操作（占位）。

type CLIConfig struct{}

func NewCLIConfig() *CLIConfig { return &CLIConfig{} }

// Dump 最小等价：rclone config dump --config=path
func (c *CLIConfig) Dump(configPath string) (string, error) {
	cmd := exec.Command("rclone", "config", "dump", "--config", configPath)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil { return "", err }
	return buf.String(), nil
}
