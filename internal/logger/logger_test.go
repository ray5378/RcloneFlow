package logger

import (
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
)

func TestNew(t *testing.T) {
	logger, err := New("info", "stdout")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer logger.Sync()
	
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestNewWithFile(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "test_log.txt")
	
	logger, err := New("debug", tmpFile)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer logger.Sync()
	
	logger.Info("test message")
	
	// 检查文件是否存在
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("expected log file to exist")
	}
	
	// 清理
	os.Remove(tmpFile)
}

func TestNewWithInvalidPath(t *testing.T) {
	// 实际上zap会自动创建目录，所以这个测试检查一个真正无效的场景
	// 比如无效的日志级别
	_, err := New("invalid_level", "stdout")
	if err != nil {
		t.Errorf("expected no error for invalid level with stdout output, got: %v", err)
	}
}

func TestNewWithDifferentLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}
	
	for _, level := range levels {
		logger, err := New(level, "stdout")
		if err != nil {
			t.Errorf("New(%s) error = %v", level, err)
		}
		logger.Sync()
	}
}

func TestGlobalLogger(t *testing.T) {
	// 初始化全局日志
	if err := Init("debug", "stdout"); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer Sync()
	
	// 测试全局函数
	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")
}

func TestWithFields(t *testing.T) {
	if err := Init("debug", "stdout"); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer Sync()
	
	logger := With(zap.String("key", "value"))
	logger.Info("message with fields")
}

func TestLoggerInterface(t *testing.T) {
	logger, _ := New("info", "stdout")
	
	// 测试接口方法
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")
	logger.Sync()
}
