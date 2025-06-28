package tests

import (
	"testing"
	"time"

	"github.com/lyenrowe/LicenseCenter/internal/config"
	"github.com/lyenrowe/LicenseCenter/internal/database"
	"github.com/lyenrowe/LicenseCenter/internal/services"
	"github.com/lyenrowe/LicenseCenter/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type LicenseServiceTestSuite struct {
	suite.Suite
	licenseService *services.LicenseService
	authService    *services.AuthorizationService
	rsaService     *services.RSAService
}

func (suite *LicenseServiceTestSuite) SetupSuite() {
	// 初始化测试配置
	err := config.LoadConfig("../configs/app.yaml")
	assert.NoError(suite.T(), err)

	// 初始化日志
	err = logger.InitLogger("debug", "../logs/test.log")
	assert.NoError(suite.T(), err)

	// 使用内存数据库进行测试
	config.AppConfig.Database.Driver = "sqlite"
	config.AppConfig.Database.DSN = ":memory:"

	// 初始化数据库（包含迁移）
	err = database.InitDatabase(&config.AppConfig.Database)
	assert.NoError(suite.T(), err)

	// 初始化服务
	suite.licenseService = services.NewLicenseService()
	suite.authService = services.NewAuthorizationService()
	suite.rsaService = services.NewRSAService()
}

func (suite *LicenseServiceTestSuite) SetupTest() {
	// 每个测试用例重新初始化数据库（确保表存在）
	config.AppConfig.Database.Driver = "sqlite"
	config.AppConfig.Database.DSN = ":memory:"

	// 重新初始化数据库
	err := database.InitDatabase(&config.AppConfig.Database)
	if err != nil {
		suite.T().Fatalf("Failed to reinitialize database: %v", err)
	}

	// 执行数据库迁移
	err = database.DB.AutoMigrate()
	if err != nil {
		suite.T().Fatalf("Failed to migrate database: %v", err)
	}

	// 创建索引
	err = database.DB.CreateIndexes()
	if err != nil {
		suite.T().Fatalf("Failed to create indexes: %v", err)
	}

	// 重新初始化服务
	suite.licenseService = services.NewLicenseService()
	suite.authService = services.NewAuthorizationService()
	suite.rsaService = services.NewRSAService()
}

func (suite *LicenseServiceTestSuite) TestActivateLicenses() {
	// 创建授权码
	auth, err := suite.authService.CreateAuthorization(&services.CreateAuthorizationRequest{
		CustomerName:      "测试客户",
		AuthorizationCode: "TEST-123-ABC",
		MaxSeats:          5,
	})
	assert.NoError(suite.T(), err)

	// 生成RSA密钥对
	_, _, err = suite.rsaService.GenerateAndSaveKeyPair()
	assert.NoError(suite.T(), err)

	// 测试设备激活
	machineID := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4" // 32位MD5格式
	bindFiles := []services.BindFile{
		{
			Hostname:    "test-host",
			MachineID:   machineID,
			RequestTime: time.Now(),
		},
	}

	licenseFiles, err := suite.licenseService.ActivateLicenses(auth.AuthorizationCode, bindFiles)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), licenseFiles, 1)
	assert.NotEmpty(suite.T(), licenseFiles[0].LicenseData.MachineID)
	assert.Equal(suite.T(), machineID, licenseFiles[0].LicenseData.MachineID)

	// 验证席位消耗
	updatedAuth, err := suite.authService.GetAuthorizationByCode(auth.AuthorizationCode)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, updatedAuth.UsedSeats)
}

func (suite *LicenseServiceTestSuite) TestActivateLicensesExceedSeats() {
	// 创建只有1个席位的授权码
	auth, err := suite.authService.CreateAuthorization(&services.CreateAuthorizationRequest{
		CustomerName:      "测试客户",
		AuthorizationCode: "TEST-123-ABC",
		MaxSeats:          1,
	})
	assert.NoError(suite.T(), err)

	// 生成RSA密钥对
	_, _, err = suite.rsaService.GenerateAndSaveKeyPair()
	assert.NoError(suite.T(), err)

	// 尝试激活超过席位数的设备
	bindFiles := []services.BindFile{
		{
			Hostname:    "test-host-1",
			MachineID:   "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4", // 32位MD5格式
			RequestTime: time.Now(),
		},
		{
			Hostname:    "test-host-2",
			MachineID:   "b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5", // 32位MD5格式
			RequestTime: time.Now(),
		},
	}

	_, err = suite.licenseService.ActivateLicenses(auth.AuthorizationCode, bindFiles)
	assert.Error(suite.T(), err)
}

func (suite *LicenseServiceTestSuite) TestGetLicensesByAuth() {
	// 创建授权码
	auth, err := suite.authService.CreateAuthorization(&services.CreateAuthorizationRequest{
		CustomerName:      "测试客户",
		AuthorizationCode: "TEST-123-ABC",
		MaxSeats:          5,
	})
	assert.NoError(suite.T(), err)

	// 生成RSA密钥对
	_, _, err = suite.rsaService.GenerateAndSaveKeyPair()
	assert.NoError(suite.T(), err)

	// 激活多台设备
	machineID1 := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4" // 32位MD5格式
	machineID2 := "b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5" // 32位MD5格式
	bindFiles := []services.BindFile{
		{
			Hostname:    "test-host-1",
			MachineID:   machineID1,
			RequestTime: time.Now(),
		},
		{
			Hostname:    "test-host-2",
			MachineID:   machineID2,
			RequestTime: time.Now(),
		},
	}

	_, err = suite.licenseService.ActivateLicenses(auth.AuthorizationCode, bindFiles)
	assert.NoError(suite.T(), err)

	// 获取设备列表
	licenses, err := suite.licenseService.GetLicensesByAuth(auth.AuthorizationCode)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), licenses, 2)

	// 验证设备信息
	machineIDs := []string{licenses[0].MachineID, licenses[1].MachineID}
	assert.Contains(suite.T(), machineIDs, machineID1)
	assert.Contains(suite.T(), machineIDs, machineID2)
}

func TestLicenseServiceSuite(t *testing.T) {
	suite.Run(t, new(LicenseServiceTestSuite))
}
