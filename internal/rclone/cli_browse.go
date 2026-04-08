package rclone

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

// ListEntry 与 rclone lsjson 输出字段对齐的最小子集。
type ListEntry struct {
	Path  string `json:"Path"`
	Name  string `json:"Name"`
	Size  int64  `json:"Size"`
	IsDir bool   `json:"IsDir"`
}

// LsJSON 调用 rclone lsjson 列目录，返回条目列表。
func LsJSON(fs string, path string, extraArgs ...string) ([]ListEntry, error) {
	full := fs
	if path != "" { full = fmt.Sprintf("%s:%s", fs, path) }
	args := []string{"lsjson", full}
	args = append(args, extraArgs...)
	cmd := exec.Command("rclone", args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil { return nil, err }
	var out []ListEntry
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil { return nil, err }
	return out, nil
}
