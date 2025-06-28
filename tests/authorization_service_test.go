package tests

import (
	"testing"

	"github.com/lyenrowe/LicenseCenter/internal/config"
	"github.com/lyenrowe/LicenseCenter/internal/database"
	"github.com/lyenrowe/LicenseCenter/internal/services"
	"github.com/lyenrowe/LicenseCenter/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AuthorizationServiceTestSuite struct {
	suite.Suite
	authService *services.AuthorizationService
}

func (suite *AuthorizationServiceTestSuite) SetupSuite() {
	// 初始化测试配置
	err := config.LoadConfig("../configs/app.yaml")
	assert.NoError(suite.T(), err)

	// 初始化日志
	err = logger.InitLogger("debug", "../logs/test.log")
	assert.NoError(suite.T(), err)

	// 使用内存数据库进行测试
	config.AppConfig.Database.Driver = "sqlite"
	config.AppConfig.Database.DSN = ":memory:"

	// 初始化数据库
	err = database.InitDatabase(&config.AppConfig.Database)
	assert.NoError(suite.T(), err)

	// 执行数据库迁移
	err = database.DB.AutoMigrate()
	assert.NoError(suite.T(), err)

	// 创建索引
	err = database.DB.CreateIndexes()
	assert.NoError(suite.T(), err)

	// 初始化服务
	suite.authService = services.NewAuthorizationService()
}

func (suite *AuthorizationServiceTestSuite) TearDownSuite() {
	// 清理资源
	if database.DB != nil {
		database.DB.Close()
	}
}

func (suite *AuthorizationServiceTestSuite) SetupTest() {
	// 每个测试前清理数据
	database.GetDB().Exec("DELETE FROM authorizations")
}

func (suite *AuthorizationServiceTestSuite) TestCreateAuthorization() {
	req := &services.CreateAuthorizationRequest{
		CustomerName:  "测试客户",
		MaxSeats:      5,
		DurationYears: &[]int{1}[0],
	}

	auth, err := suite.authService.CreateAuthorization(req)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), auth)
	assert.Equal(suite.T(), "测试客户", auth.CustomerName)
	assert.Equal(suite.T(), 5, auth.MaxSeats)
	assert.Equal(suite.T(), 0, auth.UsedSeats)
	assert.Equal(suite.T(), 1, *auth.DurationYears)
	assert.NotEmpty(suite.T(), auth.AuthorizationCode)
	assert.Equal(suite.T(), 1, auth.Status)
}

func (suite *AuthorizationServiceTestSuite) TestCreateAuthorizationWithCustomCode() {
	customCode := "CUSTOM-123-ABC"
	req := &services.CreateAuthorizationRequest{
		CustomerName:      "测试客户",
		AuthorizationCode: customCode,
		MaxSeats:          3,
	}

	auth, err := suite.authService.CreateAuthorization(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), customCode, auth.AuthorizationCode)
}

func (suite *AuthorizationServiceTestSuite) TestGetAuthorizationByCode() {
	// 先创建一个授权码
	req := &services.CreateAuthorizationRequest{
		CustomerName: "测试客户",
		MaxSeats:     5,
	}
	createdAuth, err := suite.authService.CreateAuthorization(req)
	assert.NoError(suite.T(), err)

	// 通过授权码查询
	foundAuth, err := suite.authService.GetAuthorizationByCode(createdAuth.AuthorizationCode)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), createdAuth.ID, foundAuth.ID)
	assert.Equal(suite.T(), createdAuth.AuthorizationCode, foundAuth.AuthorizationCode)
}

func (suite *AuthorizationServiceTestSuite) TestGetAuthorizationByCodeNotFound() {
	_, err := suite.authService.GetAuthorizationByCode("NOT-EXIST-CODE")
	assert.Error(suite.T(), err)
}

func (suite *AuthorizationServiceTestSuite) TestUpdateAuthorization() {
	// 先创建一个授权码
	req := &services.CreateAuthorizationRequest{
		CustomerName: "原客户",
		MaxSeats:     5,
	}
	auth, err := suite.authService.CreateAuthorization(req)
	assert.NoError(suite.T(), err)

	// 更新授权码
	updateReq := &services.UpdateAuthorizationRequest{
		CustomerName: "新客户",
		MaxSeats:     &[]int{10}[0],
	}
	updatedAuth, err := suite.authService.UpdateAuthorization(auth.ID, updateReq)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "新客户", updatedAuth.CustomerName)
	assert.Equal(suite.T(), 10, updatedAuth.MaxSeats)
}

func (suite *AuthorizationServiceTestSuite) TestUpdateAuthorizationReduceSeats() {
	// 先创建一个授权码并消耗一些席位
	req := &services.CreateAuthorizationRequest{
		CustomerName: "测试客户",
		MaxSeats:     5,
	}
	auth, err := suite.authService.CreateAuthorization(req)
	assert.NoError(suite.T(), err)

	// 消耗2个席位
	err = suite.authService.ConsumeSeats(auth.ID, 2)
	assert.NoError(suite.T(), err)

	// 尝试将最大席位数减少到小于已使用席位数（应该失败）
	updateReq := &services.UpdateAuthorizationRequest{
		MaxSeats: &[]int{1}[0],
	}
	_, err = suite.authService.UpdateAuthorization(auth.ID, updateReq)
	assert.Error(suite.T(), err)
}

func (suite *AuthorizationServiceTestSuite) TestConsumeSeats() {
	// 创建授权码
	req := &services.CreateAuthorizationRequest{
		CustomerName: "测试客户",
		MaxSeats:     5,
	}
	auth, err := suite.authService.CreateAuthorization(req)
	assert.NoError(suite.T(), err)

	// 消耗席位
	err = suite.authService.ConsumeSeats(auth.ID, 2)
	assert.NoError(suite.T(), err)

	// 验证席位数
	updatedAuth, err := suite.authService.GetAuthorizationByID(auth.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, updatedAuth.UsedSeats)
}

func (suite *AuthorizationServiceTestSuite) TestConsumeSeatsBeyondLimit() {
	// 创建授权码
	req := &services.CreateAuthorizationRequest{
		CustomerName: "测试客户",
		MaxSeats:     3,
	}
	auth, err := suite.authService.CreateAuthorization(req)
	assert.NoError(suite.T(), err)

	// 尝试消耗超过限制的席位
	err = suite.authService.ConsumeSeats(auth.ID, 5)
	assert.Error(suite.T(), err)
}

func (suite *AuthorizationServiceTestSuite) TestReleaseSeats() {
	// 创建授权码并消耗席位
	req := &services.CreateAuthorizationRequest{
		CustomerName: "测试客户",
		MaxSeats:     5,
	}
	auth, err := suite.authService.CreateAuthorization(req)
	assert.NoError(suite.T(), err)

	err = suite.authService.ConsumeSeats(auth.ID, 3)
	assert.NoError(suite.T(), err)

	// 释放席位
	err = suite.authService.ReleaseSeats(auth.ID, 2)
	assert.NoError(suite.T(), err)

	// 验证席位数
	updatedAuth, err := suite.authService.GetAuthorizationByID(auth.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, updatedAuth.UsedSeats)
}

func (suite *AuthorizationServiceTestSuite) TestReleaseSeatsMoreThanUsed() {
	// 创建授权码并消耗席位
	req := &services.CreateAuthorizationRequest{
		CustomerName: "测试客户",
		MaxSeats:     5,
	}
	auth, err := suite.authService.CreateAuthorization(req)
	assert.NoError(suite.T(), err)

	err = suite.authService.ConsumeSeats(auth.ID, 2)
	assert.NoError(suite.T(), err)

	// 释放超过已使用的席位数（应该重置为0）
	err = suite.authService.ReleaseSeats(auth.ID, 5)
	assert.NoError(suite.T(), err)

	// 验证席位数应该为0
	updatedAuth, err := suite.authService.GetAuthorizationByID(auth.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, updatedAuth.UsedSeats)
}

func (suite *AuthorizationServiceTestSuite) TestValidateAuthorizationCode() {
	// 创建活跃的授权码
	req := &services.CreateAuthorizationRequest{
		CustomerName: "测试客户",
		MaxSeats:     5,
	}
	auth, err := suite.authService.CreateAuthorization(req)
	assert.NoError(suite.T(), err)

	// 验证授权码
	validAuth, err := suite.authService.ValidateAuthorizationCode(auth.AuthorizationCode)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), auth.ID, validAuth.ID)

	// 禁用授权码
	updateReq := &services.UpdateAuthorizationRequest{
		Status: &[]int{0}[0],
	}
	_, err = suite.authService.UpdateAuthorization(auth.ID, updateReq)
	assert.NoError(suite.T(), err)

	// 验证被禁用的授权码
	_, err = suite.authService.ValidateAuthorizationCode(auth.AuthorizationCode)
	assert.Error(suite.T(), err)
}

func (suite *AuthorizationServiceTestSuite) TestListAuthorizations() {
	// 创建多个授权码
	for i := 0; i < 3; i++ {
		req := &services.CreateAuthorizationRequest{
			CustomerName: "测试客户" + string(rune('A'+i)),
			MaxSeats:     5,
		}
		_, err := suite.authService.CreateAuthorization(req)
		assert.NoError(suite.T(), err)
	}

	// 查询列表
	auths, total, err := suite.authService.ListAuthorizations(1, 10, "", nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(3), total)
	assert.Len(suite.T(), auths, 3)
}

func (suite *AuthorizationServiceTestSuite) TestListAuthorizationsWithSearch() {
	// 创建授权码
	req1 := &services.CreateAuthorizationRequest{
		CustomerName: "Apple公司",
		MaxSeats:     5,
	}
	_, err := suite.authService.CreateAuthorization(req1)
	assert.NoError(suite.T(), err)

	req2 := &services.CreateAuthorizationRequest{
		CustomerName: "Google公司",
		MaxSeats:     3,
	}
	_, err = suite.authService.CreateAuthorization(req2)
	assert.NoError(suite.T(), err)

	// 搜索
	auths, total, err := suite.authService.ListAuthorizations(1, 10, "Apple", nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), total)
	assert.Len(suite.T(), auths, 1)
	assert.Equal(suite.T(), "Apple公司", auths[0].CustomerName)
}

func (suite *AuthorizationServiceTestSuite) TestGetStatistics() {
	// 创建一些测试数据
	req := &services.CreateAuthorizationRequest{
		CustomerName: "测试客户",
		MaxSeats:     10,
	}
	auth, err := suite.authService.CreateAuthorization(req)
	assert.NoError(suite.T(), err)

	// 消耗一些席位
	err = suite.authService.ConsumeSeats(auth.ID, 3)
	assert.NoError(suite.T(), err)

	// 获取统计信息
	stats, err := suite.authService.GetStatistics()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), stats["total_authorizations"])
	assert.Equal(suite.T(), int64(1), stats["active_authorizations"])
	assert.Equal(suite.T(), int64(10), stats["total_seats"])
	assert.Equal(suite.T(), int64(3), stats["used_seats"])
}

// 运行测试套件
func TestAuthorizationServiceSuite(t *testing.T) {
	suite.Run(t, new(AuthorizationServiceTestSuite))
}
