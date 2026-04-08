package rclone

import (
	"bufio"
	"bytes"
	"strings"
)

// VersionCLI returns `rclone version` first line, e.g., "rclone v1.73.3"
func VersionCLI() (string, error) {
	cmd := RcloneCmd("version")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil { return "", err }
	s := buf.String()
	scanner := bufio.NewScanner(strings.NewReader(s))
	if scanner.Scan() {
		line := scanner.Text()
		return strings.TrimSpace(line), nil
	}
	return strings.TrimSpace(s), nil
}
