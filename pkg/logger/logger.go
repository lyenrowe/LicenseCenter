package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger      *zap.Logger
	ErrorLogger *zap.Logger
)

// LogConfig 日志配置
type LogConfig struct {
	Level        string
	AppLogFile   string
	ErrorLogFile string
}

// InitLogger 初始化日志器
func InitLogger(level, logFile string) error {
	config := &LogConfig{
		Level:        level,
		AppLogFile:   logFile,
		ErrorLogFile: getErrorLogFile(logFile),
	}
	return InitLoggerWithConfig(config)
}

// InitLoggerWithConfig 使用配置初始化日志器
func InitLoggerWithConfig(config *LogConfig) error {
	// 配置日志级别
	var zapLevel zapcore.Level
	switch config.Level {
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
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// 确保日志目录存在
	if err := ensureLogDir(config.AppLogFile); err != nil {
		return err
	}
	if err := ensureLogDir(config.ErrorLogFile); err != nil {
		return err
	}

	// 创建应用日志文件输出
	appFile, err := os.OpenFile(config.AppLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	// 创建错误日志文件输出
	errorFile, err := os.OpenFile(config.ErrorLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	// 应用日志核心（app.log + 控制台）
	appFileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	appCore := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapLevel),
		zapcore.NewCore(appFileEncoder, zapcore.AddSync(appFile), zapLevel),
	)

	// 错误日志核心（error.log，只记录error级别）
	errorFileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	errorCore := zapcore.NewCore(
		errorFileEncoder,
		zapcore.AddSync(errorFile),
		zapcore.ErrorLevel,
	)

	// 创建logger
	Logger = zap.New(appCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	ErrorLogger = zap.New(errorCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

// GetLogger 获取应用日志器实例
func GetLogger() *zap.Logger {
	if Logger == nil {
		// 如果未初始化，使用默认配置
		Logger, _ = zap.NewDevelopment()
	}
	return Logger
}

// GetErrorLogger 获取错误日志器实例
func GetErrorLogger() *zap.Logger {
	if ErrorLogger == nil {
		// 如果未初始化，使用默认配置
		ErrorLogger, _ = zap.NewDevelopment()
	}
	return ErrorLogger
}

// getErrorLogFile 根据应用日志文件路径生成错误日志文件路径
func getErrorLogFile(appLogFile string) string {
	dir := filepath.Dir(appLogFile)
	ext := filepath.Ext(appLogFile)
	name := filepath.Base(appLogFile)
	nameWithoutExt := name[:len(name)-len(ext)]

	return filepath.Join(dir, nameWithoutExt+"_error"+ext)
}

// ensureLogDir 确保日志目录存在
func ensureLogDir(logFile string) error {
	dir := filepath.Dir(logFile)
	return os.MkdirAll(dir, 0755)
}

// Sync 同步日志缓冲区
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
	if ErrorLogger != nil {
		ErrorLogger.Sync()
	}
}
