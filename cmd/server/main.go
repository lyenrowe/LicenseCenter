package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/lyenrowe/LicenseCenter/internal/config"
	"github.com/lyenrowe/LicenseCenter/internal/database"
	"github.com/lyenrowe/LicenseCenter/internal/router"
	"github.com/lyenrowe/LicenseCenter/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	configPath := getConfigPath()
	if err := config.LoadConfig(configPath); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志
	if err := logger.InitLogger(config.AppConfig.Logging.Level, config.AppConfig.Logging.File); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	// 使用zap logger记录启动信息
	zapLogger := logger.GetLogger()
	zapLogger.Info("License Center 服务器启动中...",
		zap.String("config_path", configPath),
		zap.String("log_level", config.AppConfig.Logging.Level),
		zap.String("log_file", config.AppConfig.Logging.File))

	// 创建必要的目录
	createDirectories()

	// 初始化数据库
	if err := database.InitDatabase(&config.AppConfig.Database); err != nil {
		zapLogger.Fatal("初始化数据库失败", zap.Error(err))
	}
	zapLogger.Info("数据库初始化成功")

	// 自动迁移数据库表
	if err := database.DB.AutoMigrate(); err != nil {
		zapLogger.Fatal("数据库迁移失败", zap.Error(err))
	}
	zapLogger.Info("数据库迁移完成")

	// 创建索引
	if err := database.DB.CreateIndexes(); err != nil {
		zapLogger.Fatal("创建数据库索引失败", zap.Error(err))
	}
	zapLogger.Info("数据库索引创建完成")

	// 初始化系统数据
	if err := initSystemData(); err != nil {
		zapLogger.Fatal("初始化系统数据失败", zap.Error(err))
	}

	// 设置路由
	r := router.SetupRouter()
	zapLogger.Info("路由设置完成")

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	zapLogger.Info("服务器即将启动", zap.String("address", addr))

	// 优雅关闭
	go func() {
		if err := r.Run(addr); err != nil {
			zapLogger.Fatal("服务器启动失败", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("正在关闭服务器...")

	// 关闭数据库连接
	if err := database.DB.Close(); err != nil {
		zapLogger.Error("关闭数据库连接失败", zap.Error(err))
	}

	zapLogger.Info("服务器已关闭")
}

// getConfigPath 获取配置文件路径
func getConfigPath() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}
	return "configs/app.yaml"
}

// createDirectories 创建必要的目录
func createDirectories() {
	dirs := []string{
		config.AppConfig.System.DataDir,
		config.AppConfig.System.UploadDir,
		"logs",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("创建目录 %s 失败: %v", dir, err)
		}
	}
}

// initSystemData 初始化系统数据
func initSystemData() error {
	// 这里可以添加初始化系统数据的逻辑
	// 比如创建默认管理员账户、初始RSA密钥等

	zapLogger := logger.GetLogger()

	// 演示不同级别的日志
	zapLogger.Debug("这是DEBUG级别日志 - 用于详细的调试信息")
	zapLogger.Info("这是INFO级别日志 - 用于一般信息记录")
	zapLogger.Warn("这是WARN级别日志 - 用于警告信息")
	zapLogger.Error("这是ERROR级别日志示例 - 用于错误信息", zap.String("demo", "这不是真实错误"))

	zapLogger.Info("系统数据初始化完成")
	return nil
}
