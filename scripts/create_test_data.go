package main

import (
	"fmt"
	"log"
	"time"

	"github.com/lyenrowe/LicenseCenter/internal/config"
	"github.com/lyenrowe/LicenseCenter/internal/database"
	"github.com/lyenrowe/LicenseCenter/internal/models"
	"github.com/lyenrowe/LicenseCenter/internal/services"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 加载配置
	config.LoadConfig("configs/app.yaml")

	// 初始化数据库
	err := database.InitDatabase(&config.AppConfig.Database)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 自动迁移表结构
	db := database.GetDB()
	err = db.AutoMigrate(
		&models.Authorization{},
		&models.License{},
		&models.AdminUser{},
		&models.AdminLog{},
		&models.RSAKey{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 创建测试授权码
	createTestAuthorizations()

	// 创建测试管理员
	createTestAdmin()

	fmt.Println("测试数据创建完成!")
}

func createTestAuthorizations() {
	authService := services.NewAuthorizationService()

	// 创建测试授权码
	testAuths := []struct {
		CustomerName      string
		AuthorizationCode string
		MaxSeats          int
		DurationYears     *int
		LatestExpiryDate  *time.Time
	}{
		{
			CustomerName:      "测试客户1",
			AuthorizationCode: "TEST-001-ABC",
			MaxSeats:          5,
			DurationYears:     intPtr(1),
			LatestExpiryDate:  timePtr(time.Now().AddDate(1, 0, 0)),
		},
		{
			CustomerName:      "测试客户2",
			AuthorizationCode: "TEST-002-DEF",
			MaxSeats:          10,
			DurationYears:     intPtr(2),
			LatestExpiryDate:  timePtr(time.Now().AddDate(2, 0, 0)),
		},
		{
			CustomerName:      "演示客户",
			AuthorizationCode: "DEMO-999-XYZ",
			MaxSeats:          3,
			DurationYears:     intPtr(99), // 永久授权
			LatestExpiryDate:  nil,
		},
	}

	for _, auth := range testAuths {
		req := &services.CreateAuthorizationRequest{
			CustomerName:      auth.CustomerName,
			AuthorizationCode: auth.AuthorizationCode,
			MaxSeats:          auth.MaxSeats,
			DurationYears:     auth.DurationYears,
			LatestExpiryDate:  auth.LatestExpiryDate,
		}

		_, err := authService.CreateAuthorization(req)
		if err != nil {
			log.Printf("创建授权码 %s 失败: %v", auth.AuthorizationCode, err)
		} else {
			fmt.Printf("创建授权码成功: %s (客户: %s, 席位: %d)\n",
				auth.AuthorizationCode, auth.CustomerName, auth.MaxSeats)
		}
	}
}

func createTestAdmin() {
	db := database.GetDB()

	// 检查是否已存在测试管理员
	var existingAdmin models.AdminUser
	result := db.Where("username = ?", "admin").First(&existingAdmin)
	if result.Error == nil {
		fmt.Println("测试管理员账户已存在")
		return
	}

	// 创建测试管理员密码哈希
	password := "admin123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("生成密码哈希失败: %v", err)
		return
	}

	// 创建测试管理员
	admin := models.AdminUser{
		Username:     "admin",
		PasswordHash: string(hashedPassword),
		IsActive:     true,
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Printf("创建测试管理员失败: %v", err)
		return
	}

	fmt.Printf("创建测试管理员成功: %s (密码: %s)\n", admin.Username, password)
}

func intPtr(i int) *int {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}
