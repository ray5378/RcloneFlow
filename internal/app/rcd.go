package app

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"rcloneflow/internal/logger"

	"os/exec"
)

// startEmbeddedRC starts `rclone rcd` for config/metadata only.
// addr like "127.0.0.1:5572". user/pass for basic auth. configPath shared with CLI.
func startEmbeddedRC(addr, user, pass, configPath string, wait time.Duration) error {
	if configPath == "" {
		configPath = ensureConfigPath()
	}
	// ensure env has RCLONE_CONFIG for child
	_ = os.Setenv("RCLONE_CONFIG", configPath)

	args := []string{"rcd", "--rc-addr", addr}
	if user != "" {
		args = append(args, "--rc-user", user)
	}
	if pass != "" {
		args = append(args, "--rc-pass", pass)
	}
	cmd := exec.Command("rclone", args...)
	// inherit env
	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		return err
	}

	// health wait /rc/noop
	base := "http://" + addr
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(wait)
	for time.Now().Before(deadline) {
		req, _ := http.NewRequest(http.MethodPost, strings.TrimRight(base, "/")+"/rc/noop", nil)
		if user != "" || pass != "" {
			req.SetBasicAuth(user, pass)
		}
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode < 300 {
			_ = resp.Body.Close()
			return nil
		}
		time.Sleep(300 * time.Millisecond)
	}
	return fmt.Errorf("embedded rclone rcd not responding at %s", addr)
}

func ensureConfigPath() string {
	dataDir := os.Getenv("APP_DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	_ = os.MkdirAll(dataDir, 0o755)
	p := filepath.Join(dataDir, "rclone.conf")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		// rclone.conf 使用 INI 语法，不能写成 JSON “{}”；空文件即可
		_ = os.WriteFile(p, []byte("\n"), 0o644)
	} else if b, err := os.ReadFile(p); err == nil {
		trim := strings.TrimSpace(string(b))
		if trim == "{}" || strings.HasPrefix(trim, "{") {
			// 如果误写为 JSON，重置为空避免 rclone 解析报错
			_ = os.WriteFile(p, []byte("\n"), 0o644)
		}
	}
	return p
}

func maybeStartEmbeddedRC() {
	embed := os.Getenv("EMBED_RC")
	if embed != "" && !(strings.EqualFold(embed, "true") || embed == "1") {
		logger.Info("内置 RC 关闭（EMBED_RC=false）")
		return
	}
	addr := os.Getenv("RCLONE_RC_URL")
	if addr == "" {
		addr = "http://127.0.0.1:5572"
	}
	addr = strings.TrimPrefix(addr, "http://")
	user := os.Getenv("RCLONE_RC_USER")
	if user == "" {
		user = "rc"
	}
	pass := os.Getenv("RCLONE_RC_PASS")
	if pass == "" {
		pass = "rcpass"
	}
	cfg := ensureConfigPath()
	if err := startEmbeddedRC(addr, user, pass, cfg, 10*time.Second); err != nil {
		logger.Warn("内置 RC 启动失败", zap.Error(err), zap.String("addr", addr))
		return
	}
	// export for adapter
	_ = os.Setenv("RCLONE_RC_URL", "http://"+addr)
	_ = os.Setenv("RCLONE_RC_USER", user)
	_ = os.Setenv("RCLONE_RC_PASS", pass)
	_ = os.Setenv("RCLONE_CONFIG", cfg)
	logger.Info("内置 RC 已就绪", zap.String("addr", "http://"+addr))
}
