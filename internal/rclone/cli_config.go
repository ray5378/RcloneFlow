package rclone

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// CLIConfig 封装 rclone config 的最小等价操作。

type CLIConfig struct{}

func NewCLIConfig() *CLIConfig { return &CLIConfig{} }

// addConfigFlag 根据是否提供 configPath 决定是否追加 --config 参数。
func addConfigFlag(args []string, configPath string) []string {
	if configPath != "" {
		args = append(args, "--config", configPath)
	}
	return args
}

// Dump：rclone config dump [--config path]
func (c *CLIConfig) Dump(configPath string) (string, error) {
	args := []string{"config", "dump"}
	args = addConfigFlag(args, configPath)
	if configPath != "" { _ = os.Setenv("RCLONE_CONFIG", configPath) }
	cmd := RcloneCmd(args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil { return "", err }
	return buf.String(), nil
}

// Create：rclone config create <name> <type> k v ... [--config path]
func (c *CLIConfig) Create(configPath, name, typ string, params map[string]any) error {
	args := []string{"config", "create", name, typ}
	for k, v := range params {
		args = append(args, k)
		args = append(args, fmt.Sprintf("%v", v))
	}
	args = addConfigFlag(args, configPath)
	if configPath != "" { _ = os.Setenv("RCLONE_CONFIG", configPath) }
	cmd := RcloneCmd(args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	return cmd.Run()
}

// Delete：rclone config delete <name> [--config path]
func (c *CLIConfig) Delete(configPath, name string) error {
	args := []string{"config", "delete", name}
	args = addConfigFlag(args, configPath)
	if configPath != "" { _ = os.Setenv("RCLONE_CONFIG", configPath) }
	cmd := RcloneCmd(args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	return cmd.Run()
}

// TestRemote：通过列举根路径验证连通性。
func (c *CLIConfig) TestRemote(name, configPath string) error {
	args := []string{"lsd", name + ":"}
	args = addConfigFlag(args, configPath)
	if configPath != "" { _ = os.Setenv("RCLONE_CONFIG", configPath) }
	cmd := RcloneCmd(args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	return cmd.Run()
}

// ListNames：从 dump 结果中枚举 remote 名称
func (c *CLIConfig) ListNames(configPath string) ([]string, error) {
	out, err := c.Dump(configPath)
	if err != nil { return nil, err }
	m := map[string]map[string]any{}
	if err := json.Unmarshal([]byte(out), &m); err != nil { return nil, err }
	names := make([]string, 0, len(m))
	for k := range m { names = append(names, k) }
	return names, nil
}
