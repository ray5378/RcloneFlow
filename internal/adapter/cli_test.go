package adapter

import (
	"context"
	"os"
	"testing"
)

func TestCmdRunner_Bin_Default(t *testing.T) {
	r := &CmdRunner{}
	b := r.bin()
	if b == "" {
		t.Fatal("bin() returned empty string")
	}
}

func TestCmdRunner_Bin_Custom(t *testing.T) {
	r := &CmdRunner{Bin: "/custom/rclone"}
	if got := r.bin(); got != "/custom/rclone" {
		t.Fatalf("bin() = %q, want %q", got, "/custom/rclone")
	}
}

func TestCmdRunner_Bin_EnvOverride(t *testing.T) {
	orig := os.Getenv("RCLONE_BIN")
	defer os.Setenv("RCLONE_BIN", orig)

	os.Setenv("RCLONE_BIN", "/env/rclone")
	r := &CmdRunner{}
	if got := r.bin(); got != "/env/rclone" {
		t.Fatalf("bin() = %q, want %q", got, "/env/rclone")
	}
}

func TestCmdRunner_CmdContext(t *testing.T) {
	r := &CmdRunner{Bin: "echo"}
	ctx := context.Background()
	cmd := r.CmdContext(ctx, "hello")
	if cmd == nil {
		t.Fatal("CmdContext returned nil")
	}
	// exec.CommandContext resolves to full path, just check it contains "echo"
	if cmd.Path == "" {
		t.Fatal("cmd.Path is empty")
	}
}

func TestCmdRunner_Run(t *testing.T) {
	r := &CmdRunner{Bin: "echo"}
	ctx := context.Background()
	stdout, stderr, err := r.Run(ctx, "hello")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if stdout == "" {
		t.Fatal("Run() returned empty stdout")
	}
	_ = stderr
}

func TestCmdRunner_Run_Error(t *testing.T) {
	r := &CmdRunner{Bin: "false"}
	ctx := context.Background()
	_, _, err := r.Run(ctx)
	if err == nil {
		t.Fatal("Run() expected error, got nil")
	}
}
