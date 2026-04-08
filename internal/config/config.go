package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Rclone  RcloneConfig  `yaml:"rclone"`
	Server  ServerConfig  `yaml:"server"`
	Storage StorageConfig `yaml:"storage"`
	Log     LogConfig     `yaml:"log"`
	Sync    SyncConfig    `yaml:"sync"`
}

// RcloneConfig rclone连接配置
type RcloneConfig struct {
	RCURL  string        `yaml:"rc_url"`
	RCUser string        `yaml:"rc_user"`
	RCPass string        `yaml:"rc_pass"`
	Timeout time.Duration `yaml:"timeout"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Addr       string `yaml:"addr"`
	StaticDir  string `yaml:"static_dir"`
	ReadTimeout  int `yaml:"read_timeout"`  // 秒
	WriteTimeout int `yaml:"write_timeout"` // 秒
}

// SyncConfig 同步配置
type SyncConfig struct {
	PoolInterval      int `yaml:"pool_interval"`       // 任务状态同步间隔（秒）
	ScheduleInterval int `yaml:"schedule_interval"`   // 定时任务检查间隔（分钟）
	CleanupInterval  int `yaml:"cleanup_interval"`    // 历史记录清理间隔（小时），0表示不清理
	CleanupRetention int `yaml:"cleanup_retention"`   // 历史记录保留天数
}

// StorageConfig 存储配置
type StorageConfig struct {
	DataDir string `yaml:"data_dir"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `yaml:"level"`
	Output string `yaml:"output"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Rclone: RcloneConfig{
			RCURL:  "http://127.0.0.1:5572",
			Timeout: 120 * time.Second,
		},
		Server: ServerConfig{
			Addr:         ":17870",
			StaticDir:    "./web",
			ReadTimeout:  30,
			WriteTimeout: 30,
		},
		Storage: StorageConfig{
			DataDir: "./data",
		},
		Log: LogConfig{
			Level:  "info",
			Output: "stdout",
		},
		Sync: SyncConfig{
			PoolInterval:      30,      // 30秒
			ScheduleInterval:   1,      // 1分钟
			CleanupInterval:  24,      // 24小时清理一次
			CleanupRetention: 15,      // 默认保留15天
		},
	}
}

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	cfg := DefaultConfig()
	
	// 如果没有指定配置文件路径，尝试查找默认位置
	if configPath == "" {
		configPath = findConfigFile()
	}
	
	// 如果找到配置文件，则加载
	if configPath != "" {
		if err := loadFromFile(cfg, configPath); err != nil {
			return nil, fmt.Errorf("加载配置文件失败 %s: %w", configPath, err)
		}
	}
	
	// 使用环境变量覆盖配置
	loadFromEnv(cfg)
	
	return cfg, nil
}

// findConfigFile 查找配置文件
func findConfigFile() string {
	// 查找顺序
	paths := []string{
		"./config.yaml",
		"./rcloneflow.yaml",
		"./.config/rcloneflow/config.yaml",
		filepath.Join(os.Getenv("HOME"), ".config", "rcloneflow", "config.yaml"),
	}
	
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// loadFromFile 从文件加载配置
func loadFromFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, cfg)
}

// loadFromEnv 从环境变量加载配置（环境变量优先级最高）
func loadFromEnv(cfg *Config) {
	// Rclone配置
	if v := os.Getenv("RCLONE_RC_URL"); v != "" {
		cfg.Rclone.RCURL = v
	}
	if v := os.Getenv("RCLONE_RC_USER"); v != "" {
		cfg.Rclone.RCUser = v
	}
	if v := os.Getenv("RCLONE_RC_PASS"); v != "" {
		cfg.Rclone.RCPass = v
	}
	if v := os.Getenv("RCLONE_RC_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Rclone.Timeout = d
		}
	}
	
	// 服务器配置
	if v := os.Getenv("APP_ADDR"); v != "" {
		cfg.Server.Addr = v
	}
	if v := os.Getenv("APP_STATIC_DIR"); v != "" {
		cfg.Server.StaticDir = v
	}
	
	// 存储配置
	if v := os.Getenv("APP_DATA_DIR"); v != "" {
		cfg.Storage.DataDir = v
	}
	
	// 日志配置
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	}
	if v := os.Getenv("LOG_OUTPUT"); v != "" {
		cfg.Log.Output = v
	}
}

// GetRcloneAddr 返回rclone地址（用于环境变量）
func (c *Config) GetRcloneAddr() string {
	return c.Rclone.RCURL
}

// GetRcloneUser 返回rclone用户名
func (c *Config) GetRcloneUser() string {
	return c.Rclone.RCUser
}

// GetRclonePass 返回rclone密码
func (c *Config) GetRclonePass() string {
	return c.Rclone.RCPass
}

// GetRcloneTimeout 返回rclone超时时间
func (c *Config) GetRcloneTimeout() time.Duration {
	return c.Rclone.Timeout
}

// GetServerAddr 返回服务器地址
func (c *Config) GetServerAddr() string {
	return c.Server.Addr
}

// GetStaticDir 返回静态文件目录
func (c *Config) GetStaticDir() string {
	return c.Server.StaticDir
}

// GetDataDir 返回数据目录
func (c *Config) GetDataDir() string {
	return c.Storage.DataDir
}

// GetLogLevel 返回日志级别
func (c *Config) GetLogLevel() string {
	return c.Log.Level
}

// GetLogOutput 返回日志输出
func (c *Config) GetLogOutput() string {
	return c.Log.Output
}

// GetPoolInterval 返回任务状态同步间隔（秒）
func (c *Config) GetPoolInterval() int {
	return c.Sync.PoolInterval
}

// GetScheduleInterval 返回定时任务检查间隔（分钟）
func (c *Config) GetScheduleInterval() int {
	return c.Sync.ScheduleInterval
}

// GetCleanupInterval 返回历史记录清理间隔（小时），0表示不清理
func (c *Config) GetCleanupInterval() int {
	return c.Sync.CleanupInterval
}

// GetCleanupRetention 返回历史记录保留天数
func (c *Config) GetCleanupRetention() int {
	return c.Sync.CleanupRetention
}

// ToEnvMap 转换为环境变量映射（用于传递给子组件）
func (c *Config) ToEnvMap() map[string]string {
	m := make(map[string]string)
	
	if c.Rclone.RCURL != "" {
		m["RCLONE_RC_URL"] = c.Rclone.RCURL
	}
	if c.Rclone.RCUser != "" {
		m["RCLONE_RC_USER"] = c.Rclone.RCUser
	}
	if c.Rclone.RCPass != "" {
		m["RCLONE_RC_PASS"] = c.Rclone.RCPass
	}
	m["RCLONE_RC_TIMEOUT"] = c.Rclone.Timeout.String()
	
	m["APP_ADDR"] = c.Server.Addr
	m["APP_STATIC_DIR"] = c.Server.StaticDir
	m["APP_DATA_DIR"] = c.Storage.DataDir
	
	m["LOG_LEVEL"] = c.Log.Level
	m["LOG_OUTPUT"] = c.Log.Output
	
	return m
}

// String 返回配置文件的字符串表示
func (c *Config) String() string {
	var sb strings.Builder
	sb.WriteString("RcloneFlow 配置:\n")
	sb.WriteString(fmt.Sprintf("  rclone:\n"))
	sb.WriteString(fmt.Sprintf("    rc_url: %s\n", c.Rclone.RCURL))
	sb.WriteString(fmt.Sprintf("    timeout: %s\n", c.Rclone.Timeout))
	if c.Rclone.RCUser != "" {
		sb.WriteString(fmt.Sprintf("    rc_user: %s\n", c.Rclone.RCUser))
	}
	sb.WriteString(fmt.Sprintf("  server:\n"))
	sb.WriteString(fmt.Sprintf("    addr: %s\n", c.Server.Addr))
	sb.WriteString(fmt.Sprintf("    static_dir: %s\n", c.Server.StaticDir))
	sb.WriteString(fmt.Sprintf("  storage:\n"))
	sb.WriteString(fmt.Sprintf("    data_dir: %s\n", c.Storage.DataDir))
	sb.WriteString(fmt.Sprintf("  log:\n"))
	sb.WriteString(fmt.Sprintf("    level: %s\n", c.Log.Level))
	sb.WriteString(fmt.Sprintf("    output: %s\n", c.Log.Output))
	return sb.String()
}
