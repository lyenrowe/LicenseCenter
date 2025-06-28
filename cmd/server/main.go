package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/licensecenter/licensecenter/internal/config"
	"github.com/licensecenter/licensecenter/internal/database"
	"github.com/licensecenter/licensecenter/internal/router"
	"github.com/licensecenter/licensecenter/pkg/logger"
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

	// 创建必要的目录
	createDirectories()

	// 初始化数据库
	if err := database.InitDatabase(&config.AppConfig.Database); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 自动迁移数据库表
	if err := database.DB.AutoMigrate(); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 创建索引
	if err := database.DB.CreateIndexes(); err != nil {
		log.Fatalf("创建数据库索引失败: %v", err)
	}

	// 初始化系统数据
	if err := initSystemData(); err != nil {
		log.Fatalf("初始化系统数据失败: %v", err)
	}

	// 设置路由
	r := router.SetupRouter()

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	log.Printf("服务器启动在 %s", addr)

	// 优雅关闭
	go func() {
		if err := r.Run(addr); err != nil {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")

	// 关闭数据库连接
	if err := database.DB.Close(); err != nil {
		log.Printf("关闭数据库连接失败: %v", err)
	}

	log.Println("服务器已关闭")
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
	log.Println("系统数据初始化完成")
	return nil
}
