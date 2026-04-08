package rclone

import (
	"bytes"
	"os/exec"
)

// TestWebDAV 以 -vv 测试 webdav 根目录连通性（不传密码，防止泄漏）。
func TestWebDAV(url, user string) (string, string, error) {
	args := []string{"-vv", "lsd", "webdav:", "--webdav-url", url, "--webdav-vendor", "other"}
	if user != "" { args = append(args, "--webdav-user", user) }
	cmd := exec.Command("rclone", args...)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	err := cmd.Run()
	return out.String(), errb.String(), err
}
