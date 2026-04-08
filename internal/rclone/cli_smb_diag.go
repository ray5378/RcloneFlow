package rclone

import (
	"bytes"
	"os/exec"
)

// TestSMBRoot 尝试列举 smb 根（匿名/无密码），用于快速诊断 SMB exit status 3。
// 返回标准输出/标准错误，便于前端展示。
func TestSMBRoot(host string, share string, user string) (string, string, error) {
	fs := "smb:"
	args := []string{"-vv", "lsd", fs}
	if host != "" { args = append(args, "--smb-host", host) }
	if share != "" { args = append(args, "--smb-share", share) }
	if user != "" { args = append(args, "--smb-user", user) }
	cmd := exec.Command("rclone", args...)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	err := cmd.Run()
	return out.String(), errb.String(), err
}
