package rclone

import (
	"bytes"
	"fmt"
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

// Create 最小等价：rclone config create <name> <type> k v k v ... --config=path
func (c *CLIConfig) Create(configPath, name, typ string, params map[string]any) error {
	args := []string{"config", "create", name, typ, "--config", configPath}
	for k, v := range params {
		args = append(args, k)
		args = append(args, fmt.Sprintf("%v", v))
	}
	cmd := exec.Command("rclone", args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	return cmd.Run()
}

// TestRemote 通过列举根路径验证远端是否可访问。
func (c *CLIConfig) TestRemote(name, configPath string) error {
	// 使用 lsd 检查根（也可用 lsjson），带 --config
	cmd := exec.Command("rclone", "lsd", name+":", "--config", configPath)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	return cmd.Run()
}
