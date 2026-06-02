package webdavserver

import (
	"bufio"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"rcloneflow/internal/logger"

	"go.uber.org/zap"
)

const (
	combineRemoteName = "_webdav"
	internalPort      = "127.0.0.1:17871"
	settingsKey       = "WEBDAB_ENABLED"
	encryptedPassKey  = "WEBDAB_ENCRYPTED_PASSWORD"
)

type Manager struct {
	mu         sync.Mutex
	cmd        *exec.Cmd
	cancel     context.CancelFunc
	running    bool
	dataDir    string
	configFile string
}

func NewManager(dataDir string) *Manager {
	return &Manager{
		dataDir:    dataDir,
		configFile: filepath.Join(dataDir, "rclone.conf"),
	}
}

func (m *Manager) getJWTSecret() ([]byte, error) {
	secretFile := filepath.Join(m.dataDir, ".jwt_secret")
	secretHex, err := os.ReadFile(secretFile)
	if err != nil {
		return nil, fmt.Errorf("读取JWT密钥失败: %w", err)
	}
	secretHex = []byte(strings.TrimSpace(string(secretHex)))
	secret := make([]byte, hex.DecodedLen(len(secretHex)))
	if _, err := hex.Decode(secret, secretHex); err != nil {
		return nil, fmt.Errorf("JWT密钥解码失败: %w", err)
	}
	return secret, nil
}

func deriveAESKey(jwtSecret []byte) []byte {
	h := sha256.Sum256(jwtSecret)
	return h[:]
}

func deriveWebdavPassword(jwtSecret []byte) string {
	h := sha256.Sum256(jwtSecret)
	return hex.EncodeToString(h[:16])
}

func encryptPassword(password string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decryptPassword(encrypted string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("密文太短")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func (m *Manager) readSettings() map[string]string {
	fp := filepath.Join(m.dataDir, "settings.json")
	b, err := os.ReadFile(fp)
	if err != nil || len(b) == 0 {
		return map[string]string{}
	}
	var settings map[string]string
	if json.Unmarshal(b, &settings) != nil {
		return map[string]string{}
	}
	return settings
}

func (m *Manager) writeSettings(key, value string) error {
	fp := filepath.Join(m.dataDir, "settings.json")
	settings := make(map[string]string)
	b, err := os.ReadFile(fp)
	if err == nil && len(b) > 0 {
		_ = json.Unmarshal(b, &settings)
		if settings == nil {
			settings = make(map[string]string)
		}
	}
	settings[key] = value
	data, _ := json.MarshalIndent(settings, "", "  ")
	return os.WriteFile(fp, data, 0644)
}

func (m *Manager) listRemotesFromConfig() ([]string, error) {
	f, err := os.Open(m.configFile)
	if err != nil {
		return nil, fmt.Errorf("打开rclone配置文件失败: %w", err)
	}
	defer f.Close()

	var remotes []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			name := line[1 : len(line)-1]
			if name != "" && name != combineRemoteName {
				remotes = append(remotes, name)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取rclone配置文件失败: %w", err)
	}
	return remotes, nil
}

func (m *Manager) createCombineRemoteInConfig(remotes []string) error {
	if len(remotes) == 0 {
		return fmt.Errorf("没有可用的存储节点")
	}

	upstreams := make([]string, 0, len(remotes))
	for _, name := range remotes {
		upstreams = append(upstreams, fmt.Sprintf("%s=:%s:", name, name))
	}

	content, err := os.ReadFile(m.configFile)
	if err != nil {
		return fmt.Errorf("读取rclone配置文件失败: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	skip := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.Trim(trimmed, "[]") == combineRemoteName {
			skip = true
			continue
		}
		if skip && strings.HasPrefix(trimmed, "[") {
			skip = false
		}
		if !skip {
			newLines = append(newLines, line)
		}
	}

	newLines = append(newLines, "")
	newLines = append(newLines, fmt.Sprintf("[%s]", combineRemoteName))
	newLines = append(newLines, "type = combine")
	newLines = append(newLines, fmt.Sprintf("upstreams = %s", strings.Join(upstreams, " ")))
	newLines = append(newLines, "")

	return os.WriteFile(m.configFile, []byte(strings.Join(newLines, "\n")), 0644)
}

func (m *Manager) Start(username string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("WebDAV服务已在运行中")
	}

	remotes, err := m.listRemotesFromConfig()
	if err != nil {
		return fmt.Errorf("获取远程存储列表失败: %w", err)
	}

	if err := m.createCombineRemoteInConfig(remotes); err != nil {
		return err
	}

	jwtSecret, err := m.getJWTSecret()
	if err != nil {
		return err
	}
	aesKey := deriveAESKey(jwtSecret)

	password := deriveWebdavPassword(jwtSecret)

	encryptedPass, err := encryptPassword(password, aesKey)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	if err := m.writeSettings(encryptedPassKey, encryptedPass); err != nil {
		return fmt.Errorf("保存密码失败: %w", err)
	}
	if err := m.writeSettings(settingsKey, "true"); err != nil {
		return fmt.Errorf("保存状态失败: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel

	args := []string{
		"serve", "webdav",
		combineRemoteName + ":",
		"--addr", internalPort,
		"--user", username,
		"--pass", password,
		"--config", m.configFile,
		"--no-checksum",
	}

	m.cmd = exec.CommandContext(ctx, "rclone", args...)
	m.cmd.Stdout = os.Stdout
	m.cmd.Stderr = os.Stderr

	if err := m.cmd.Start(); err != nil {
		cancel()
		m.cmd = nil
		m.cancel = nil
		return fmt.Errorf("启动WebDAV服务失败: %w", err)
	}

	m.running = true
	logger.Info("WebDAV服务已启动", zap.String("addr", internalPort))

	go func() {
		if err := m.cmd.Wait(); err != nil {
			logger.Error("WebDAV服务异常退出", zap.Error(err))
		}
		m.mu.Lock()
		m.running = false
		m.cmd = nil
		m.cancel = nil
		m.mu.Unlock()
	}()

	time.Sleep(500 * time.Millisecond)
	return nil
}

func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	if m.cancel != nil {
		m.cancel()
	}

	if m.cmd != nil && m.cmd.Process != nil {
		m.cmd.Process.Signal(os.Interrupt)
		done := make(chan struct{})
		go func() {
			m.cmd.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			m.cmd.Process.Kill()
		}
	}

	m.running = false
	m.cmd = nil
	m.cancel = nil

	if err := m.writeSettings(settingsKey, "false"); err != nil {
		logger.Error("保存WebDAV关闭状态失败", zap.Error(err))
	}

	logger.Info("WebDAV服务已停止")
	return nil
}

func (m *Manager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

func (m *Manager) GetEncryptedPassword() string {
	settings := m.readSettings()
	return settings[encryptedPassKey]
}

func (m *Manager) GetDecryptedPassword() (string, error) {
	encrypted := m.GetEncryptedPassword()
	if encrypted == "" {
		return "", fmt.Errorf("未找到加密密码")
	}
	jwtSecret, err := m.getJWTSecret()
	if err != nil {
		return "", err
	}
	aesKey := deriveAESKey(jwtSecret)
	return decryptPassword(encrypted, aesKey)
}

func (m *Manager) IsEnabled() bool {
	settings := m.readSettings()
	return settings[settingsKey] == "true"
}

func (m *Manager) GetPassword() (string, error) {
	pass := m.GetEncryptedPassword()
	if pass != "" {
		return m.GetDecryptedPassword()
	}
	jwtSecret, err := m.getJWTSecret()
	if err != nil {
		return "", err
	}
	return deriveWebdavPassword(jwtSecret), nil
}

func (m *Manager) AutoRestore(username string) {
	if !m.IsEnabled() {
		return
	}

	_, err := m.GetPassword()
	if err != nil {
		logger.Error("WebDAV自动恢复失败: 密码获取错误", zap.Error(err))
		m.writeSettings(settingsKey, "false")
		return
	}

	if err := m.Start(username); err != nil {
		logger.Error("WebDAV自动恢复失败", zap.Error(err))
		m.writeSettings(settingsKey, "false")
		return
	}

	logger.Info("WebDAV服务已自动恢复")
}