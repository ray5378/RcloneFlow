package rclone

import (
	"os"
	"os/exec"
	"path/filepath"
)

// RclonePath resolves the rclone binary path for local/dev runs.
// Precedence: env RCLONE_BIN > PATH lookup > ./bin/rclone > ./rclone > system paths.
func RclonePath() string {
	if p := os.Getenv("RCLONE_BIN"); p != "" { return p }
	if p, err := exec.LookPath("rclone"); err == nil { return p }
	cands := []string{"./bin/rclone", "./rclone", "/usr/local/bin/rclone", "/usr/bin/rclone"}
	for _, c := range cands {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	return "rclone"
}

// RcloneCmd constructs exec.Cmd with resolved rclone binary.
func RcloneCmd(args ...string) *exec.Cmd {
	return exec.Command(RclonePath(), args...)
}
