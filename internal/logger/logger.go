package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 统一日志接口
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Sync() error
}

// zapLogger zap日志实现
type zapLogger struct {
	sugar *zap.SugaredLogger
}

// New 创建日志实例
func New(level, output string) (*zapLogger, error) {
	// 解析日志级别
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// 配置编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 根据输出目标配置
	var encoder zapcore.Encoder
	if output == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 配置写入器
	var writeSyncer zapcore.WriteSyncer
	if output == "stdout" {
		writeSyncer = zapcore.AddSync(os.Stdout)
	} else if output == "stderr" {
		writeSyncer = zapcore.AddSync(os.Stderr)
	} else {
		// 确保目录存在
		dir := filepath.Dir(output)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("创建日志目录失败: %w", err)
		}
		file, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("打开日志文件失败: %w", err)
		}
		writeSyncer = zapcore.AddSync(file)
	}

	// 创建核心
	core := zapcore.NewCore(encoder, writeSyncer, zapLevel)

	// 创建logger
	l := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar := l.Sugar()

	return &zapLogger{sugar: sugar}, nil
}

// Debug 调试级别
func (l *zapLogger) Debug(msg string, fields ...zap.Field) {
	l.sugar.Debugw(msg)
}

// Info 信息级别
func (l *zapLogger) Info(msg string, fields ...zap.Field) {
	l.sugar.Infow(msg)
}

// Warn 警告级别
func (l *zapLogger) Warn(msg string, fields ...zap.Field) {
	l.sugar.Warnw(msg)
}

// Error 错误级别
func (l *zapLogger) Error(msg string, fields ...zap.Field) {
	l.sugar.Errorw(msg)
}

// Fatal 致命错误级别
func (l *zapLogger) Fatal(msg string, fields ...zap.Field) {
	l.sugar.Fatalw(msg)
}

// Sync 刷新日志
func (l *zapLogger) Sync() error {
	return l.sugar.Sync()
}

// ============ 全局日志实例 ============

var globalLogger Logger

// Init 初始化全局日志
func Init(level, output string) error {
	logger, err := New(level, output)
	if err != nil { return err }
	globalLogger = logger
	return nil
}

// HotSet 运行时切换日志级别/输出
func HotSet(level, output string) error {
	return Init(level, output)
}

// Get 获取全局日志实例
func Get() Logger {
	if globalLogger == nil {
		// 默认配置
		globalLogger, _ = New("info", "stdout")
	}
	return globalLogger
}

// Debug 调试级别
func Debug(msg string, fields ...zap.Field) {
	Get().Debug(msg, fields...)
}

// Info 信息级别
func Info(msg string, fields ...zap.Field) {
	Get().Info(msg, fields...)
}

// Warn 警告级别
func Warn(msg string, fields ...zap.Field) {
	Get().Warn(msg, fields...)
}

// Error 错误级别
func Error(msg string, fields ...zap.Field) {
	Get().Error(msg, fields...)
}

// Fatal 致命错误级别
func Fatal(msg string, fields ...zap.Field) {
	Get().Fatal(msg, fields...)
}

// Sync 刷新日志
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// ============ 便捷函数 ============

// With 创建一个带字段的日志实例
func With(fields ...zap.Field) Logger {
	return &withLogger{fields: fields}
}

type withLogger struct {
	fields []zap.Field
}

func (l *withLogger) Debug(msg string, fields ...zap.Field) {
	Get().Debug(msg, append(l.fields, fields...)...)
}
func (l *withLogger) Info(msg string, fields ...zap.Field) {
	Get().Info(msg, append(l.fields, fields...)...)
}
func (l *withLogger) Warn(msg string, fields ...zap.Field) {
	Get().Warn(msg, append(l.fields, fields...)...)
}
func (l *withLogger) Error(msg string, fields ...zap.Field) {
	Get().Error(msg, append(l.fields, fields...)...)
}
func (l *withLogger) Fatal(msg string, fields ...zap.Field) {
	Get().Fatal(msg, append(l.fields, fields...)...)
}
func (l *withLogger) Sync() error {
	return Get().Sync()
}
