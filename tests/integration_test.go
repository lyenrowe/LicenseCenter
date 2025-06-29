package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lyenrowe/LicenseCenter/internal/config"
	"github.com/lyenrowe/LicenseCenter/internal/database"
	"github.com/lyenrowe/LicenseCenter/internal/router"
	"github.com/lyenrowe/LicenseCenter/internal/services"
	"github.com/lyenrowe/LicenseCenter/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	app        *gin.Engine
	adminToken string
}

func (suite *IntegrationTestSuite) SetupSuite() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 初始化测试配置
	err := config.LoadConfig("../configs/app.yaml")
	assert.NoError(suite.T(), err)

	// 初始化日志
	err = logger.InitLogger("error", "../logs/integration_test.log")
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

	// 初始化路由
	suite.app = router.SetupRouter()

	// 创建默认管理员和RSA密钥
	suite.setupDefaultData()
}

func (suite *IntegrationTestSuite) setupDefaultData() {
	// 创建默认管理员
	adminService := services.NewAdminService()
	req := &services.CreateAdminRequest{
		Username: "admin",
		Password: "admin123",
	}
	_, err := adminService.CreateAdmin(req)
	assert.NoError(suite.T(), err)

	// 生成RSA密钥对
	rsaService := services.NewRSAService()
	_, _, err = rsaService.GenerateAndSaveKeyPair()
	assert.NoError(suite.T(), err)

	// 登录获取令牌
	suite.loginAdmin()
}

func (suite *IntegrationTestSuite) loginAdmin() {
	loginData := map[string]interface{}{
		"username": "admin",
		"password": "admin123",
	}

	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/api/admin/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.app.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// 管理员登录直接返回token，不在data字段中
	suite.adminToken = response["token"].(string)
	assert.NotEmpty(suite.T(), suite.adminToken)
}

func (suite *IntegrationTestSuite) TestCompleteWorkflow() {
	// 1. 测试健康检查
	suite.testHealthCheck()

	// 2. 测试获取公钥
	publicKey := suite.testGetPublicKey()
	assert.NotEmpty(suite.T(), publicKey)

	// 3. 测试管理员仪表板
	dashboard := suite.testAdminDashboard()
	assert.NotNil(suite.T(), dashboard)

	// 4. 测试创建授权码
	authCode := suite.testCreateAuthorization()
	assert.NotEmpty(suite.T(), authCode)

	// 5. 测试设备激活
	licenseKey := suite.testDeviceActivation(authCode)
	assert.NotEmpty(suite.T(), licenseKey)

	// 6. 测试查询设备列表
	devices := suite.testGetDeviceList(authCode)
	assert.Len(suite.T(), devices, 1)

	// 7. 验证仪表板数据更新
	suite.testDashboardAfterActivation()
}

func (suite *IntegrationTestSuite) testHealthCheck() {
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	suite.app.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "ok", response["status"])
}

func (suite *IntegrationTestSuite) testGetPublicKey() string {
	req, _ := http.NewRequest("GET", "/api/public-key", nil)
	w := httptest.NewRecorder()
	suite.app.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	publicKey := response["public_key"].(string)
	assert.Contains(suite.T(), publicKey, "-----BEGIN PUBLIC KEY-----")
	return publicKey
}

func (suite *IntegrationTestSuite) testAdminDashboard() map[string]interface{} {
	req, _ := http.NewRequest("GET", "/api/admin/dashboard", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.app.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	data := response["data"].(map[string]interface{})
	assert.Contains(suite.T(), data, "total_authorizations")
	assert.Contains(suite.T(), data, "active_devices")
	return data
}

func (suite *IntegrationTestSuite) testCreateAuthorization() string {
	authData := map[string]interface{}{
		"customer_name": "集成测试客户",
		"max_seats":     5,
	}

	jsonData, _ := json.Marshal(authData)
	req, _ := http.NewRequest("POST", "/api/admin/authorizations", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)

	w := httptest.NewRecorder()
	suite.app.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	data := response["data"].(map[string]interface{})
	authCode := data["authorization_code"].(string)
	assert.NotEmpty(suite.T(), authCode)
	return authCode
}

func (suite *IntegrationTestSuite) testDeviceActivation(authCode string) string {
	activationData := map[string]interface{}{
		"authorization_code": authCode,
		"bind_files": []map[string]interface{}{
			{
				"hostname":     "integration-test-host",
				"machine_id":   "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4", // 32位MD5格式
				"request_time": time.Now().Format(time.RFC3339),
			},
		},
	}

	jsonData, _ := json.Marshal(activationData)
	req, _ := http.NewRequest("POST", "/api/license/activate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)

	w := httptest.NewRecorder()
	suite.app.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	licenseFiles := response["license_files"].([]interface{})
	assert.Len(suite.T(), licenseFiles, 1)

	licenseFile := licenseFiles[0].(map[string]interface{})
	licenseData := licenseFile["license_data"].(map[string]interface{})

	// 验证授权数据的关键字段
	machineID := licenseData["machine_id"].(string)
	assert.Equal(suite.T(), "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4", machineID)

	// 返回机器ID作为标识符（替代 license_key）
	return machineID
}

func (suite *IntegrationTestSuite) testGetDeviceList(authCode string) []interface{} {
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/licenses?auth_code=%s", authCode), nil)
	w := httptest.NewRecorder()
	suite.app.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	devices := response["data"].([]interface{})
	return devices
}

func (suite *IntegrationTestSuite) testDashboardAfterActivation() {
	dashboard := suite.testAdminDashboard()

	// 验证激活后的数据
	activeDevices := dashboard["active_devices"].(float64)
	totalAuthorizations := dashboard["total_authorizations"].(float64)
	todayNewDevices := dashboard["today_new_devices"].(float64)

	assert.Equal(suite.T(), float64(1), activeDevices)
	assert.Equal(suite.T(), float64(1), totalAuthorizations)
	assert.Equal(suite.T(), float64(1), todayNewDevices)
}

func (suite *IntegrationTestSuite) TestErrorHandling() {
	// 测试无效的授权码激活
	suite.testInvalidAuthorizationActivation()

	// 测试未授权的管理员访问
	suite.testUnauthorizedAdminAccess()
}

func (suite *IntegrationTestSuite) testInvalidAuthorizationActivation() {
	activationData := map[string]interface{}{
		"authorization_code": "INVALID-CODE",
		"bind_files": []map[string]interface{}{
			{
				"hostname":     "test-host",
				"machine_id":   "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
				"request_time": time.Now().Format(time.RFC3339),
			},
		},
	}

	jsonData, _ := json.Marshal(activationData)
	req, _ := http.NewRequest("POST", "/api/license/activate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)

	w := httptest.NewRecorder()
	suite.app.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *IntegrationTestSuite) testUnauthorizedAdminAccess() {
	req, _ := http.NewRequest("GET", "/api/admin/dashboard", nil)
	// 不设置Authorization头

	w := httptest.NewRecorder()
	suite.app.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
