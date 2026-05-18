package adapter

import (
	"bytes"
	"context"
	"os"
	"os/exec"
)

// CmdRunner builds rclone CLI commands with sane defaults.
type CmdRunner struct{ Bin string }

func (r *CmdRunner) bin() string {
	if r.Bin != "" {
		return r.Bin
	}
	if b := os.Getenv("RCLONE_BIN"); b != "" {
		return b
	}
	if p, err := exec.LookPath("rclone"); err == nil {
		return p
	}
	if st, err := os.Stat("./bin/rclone"); err == nil && !st.IsDir() {
		return "./bin/rclone"
	}
	return "rclone"
}

func (r *CmdRunner) CmdContext(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, r.bin(), args...)
	cmd.Env = os.Environ()
	// 强制英文日志，避免本地化导致进度行关键字变化
	cmd.Env = append(cmd.Env, "LC_ALL=C", "LANG=C")
	if _, ok := os.LookupEnv("RCLONE_CONFIG"); !ok {
		if _, err := os.Stat("./data/rclone.conf"); err == nil {
			cmd.Env = append(cmd.Env, "RCLONE_CONFIG=./data/rclone.conf")
		}
	}
	return cmd
}

func (r *CmdRunner) Run(ctx context.Context, args ...string) (string, string, error) {
	cmd := r.CmdContext(ctx, args...)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	err := cmd.Run()
	return out.String(), errb.String(), err
}
