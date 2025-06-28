package main

import (
	"fmt"
	"log"

	"github.com/lyenrowe/LicenseCenter/internal/config"
	"github.com/lyenrowe/LicenseCenter/internal/database"
	"github.com/lyenrowe/LicenseCenter/internal/services"
	"github.com/lyenrowe/LicenseCenter/pkg/logger"
)

func main() {
	// 初始化配置
	if err := config.LoadConfig("configs/app.yaml"); err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	// 初始化日志
	if err := logger.InitLogger("info", "logs/app.log"); err != nil {
		log.Fatalf("日志初始化失败: %v", err)
	}

	// 初始化数据库
	if err := database.InitDatabase(&config.AppConfig.Database); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 执行数据库迁移
	if err := database.DB.AutoMigrate(); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 创建索引
	if err := database.DB.CreateIndexes(); err != nil {
		log.Fatalf("创建索引失败: %v", err)
	}

	// 创建服务实例
	adminService := services.NewAdminService()
	rsaService := services.NewRSAService()

	fmt.Println("开始初始化系统...")

	// 1. 生成RSA密钥对
	fmt.Println("生成RSA密钥对...")
	_, _, err := rsaService.GenerateAndSaveKeyPair()
	if err != nil {
		log.Fatalf("生成RSA密钥对失败: %v", err)
	}
	fmt.Println("✓ RSA密钥对生成成功")

	// 2. 创建默认管理员
	fmt.Println("创建默认管理员...")
	adminReq := &services.CreateAdminRequest{
		Username: "admin",
		Password: "admin123",
	}

	admin, err := adminService.CreateAdmin(adminReq)
	if err != nil {
		// 如果管理员已存在，跳过
		fmt.Printf("⚠ 默认管理员可能已存在: %v\n", err)
	} else {
		fmt.Printf("✓ 默认管理员创建成功 - 用户名: %s\n", admin.Username)
	}

	fmt.Println("\n系统初始化完成!")
	fmt.Println("默认管理员账号:")
	fmt.Println("  用户名: admin")
	fmt.Println("  密码: admin123")
	fmt.Println("\n请及时修改默认密码!")
}
