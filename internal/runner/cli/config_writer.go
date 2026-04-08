package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteConfigINI 将远端配置以 INI 形式写入到 workDir/rclone.conf。
// remotes 的结构：name -> (key->value)。
func WriteConfigINI(workDir string, remotes map[string]map[string]string) (string, error) {
	if workDir == "" { return "", fmt.Errorf("workDir 不能为空") }
	if err := os.MkdirAll(workDir, 0o755); err != nil { return "", err }
	cfg := filepath.Join(workDir, "rclone.conf")
	var b strings.Builder
	for name, kv := range remotes {
		b.WriteString("[" + name + "]\n")
		for k, v := range kv {
			// 简单转义（避免多行），后续可扩展
			v = strings.ReplaceAll(v, "\n", " ")
			b.WriteString(fmt.Sprintf("%s = %s\n", k, v))
		}
		b.WriteString("\n")
	}
	if err := os.WriteFile(cfg, []byte(b.String()), 0o600); err != nil { return "", err }
	return cfg, nil
}
