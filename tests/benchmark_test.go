package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/licensecenter/licensecenter/internal/config"
	"github.com/licensecenter/licensecenter/internal/database"
	"github.com/licensecenter/licensecenter/internal/services"
	"github.com/licensecenter/licensecenter/pkg/logger"
)

// BenchmarkRSAKeyGeneration 测试RSA密钥生成性能
func BenchmarkRSAKeyGeneration(b *testing.B) {
	// 初始化测试环境
	setupBenchmark(b)

	rsaService := services.NewRSAService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := rsaService.GenerateAndSaveKeyPair()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSignData 测试数据签名性能
func BenchmarkSignData(b *testing.B) {
	// 初始化测试环境
	setupBenchmark(b)

	rsaService := services.NewRSAService()

	// 生成密钥对
	_, _, err := rsaService.GenerateAndSaveKeyPair()
	if err != nil {
		b.Fatal(err)
	}

	testData := []byte("This is test data for signing performance benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := rsaService.SignData(testData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkAuthorizationCreation 测试授权码创建性能
func BenchmarkAuthorizationCreation(b *testing.B) {
	// 初始化测试环境
	setupBenchmark(b)

	authService := services.NewAuthorizationService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := &services.CreateAuthorizationRequest{
			CustomerName: "测试客户",
			MaxSeats:     5,
		}
		_, err := authService.CreateAuthorization(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLicenseActivation 测试设备激活性能
func BenchmarkLicenseActivation(b *testing.B) {
	// 初始化测试环境
	setupBenchmark(b)

	authService := services.NewAuthorizationService()
	licenseService := services.NewLicenseService()
	rsaService := services.NewRSAService()

	// 创建授权码
	req := &services.CreateAuthorizationRequest{
		CustomerName: "测试客户",
		MaxSeats:     1000, // 足够大的席位数
	}
	auth, err := authService.CreateAuthorization(req)
	if err != nil {
		b.Fatal(err)
	}

	// 生成RSA密钥对
	_, _, err = rsaService.GenerateAndSaveKeyPair()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 为每次迭代生成唯一的机器ID
		machineID := fmt.Sprintf("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c%03x", i%4096) // 32位MD5格式，使用i作为变化
		bindFiles := []services.BindFile{
			{
				Hostname:    "test-host",
				MachineID:   machineID,
				RequestTime: time.Now(),
			},
		}

		_, err := licenseService.ActivateLicenses(auth.AuthorizationCode, bindFiles)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// setupBenchmark 设置基准测试环境
func setupBenchmark(b *testing.B) {
	// 初始化测试配置
	err := config.LoadConfig("../configs/app.yaml")
	if err != nil {
		b.Fatal(err)
	}

	// 初始化日志
	err = logger.InitLogger("error", "../logs/benchmark.log") // 使用error级别减少日志输出
	if err != nil {
		b.Fatal(err)
	}

	// 使用内存数据库进行测试
	config.AppConfig.Database.Driver = "sqlite"
	config.AppConfig.Database.DSN = ":memory:"

	// 初始化数据库
	err = database.InitDatabase(&config.AppConfig.Database)
	if err != nil {
		b.Fatal(err)
	}

	// 执行数据库迁移
	err = database.DB.AutoMigrate()
	if err != nil {
		b.Fatal(err)
	}

	// 创建索引
	err = database.DB.CreateIndexes()
	if err != nil {
		b.Fatal(err)
	}
}
