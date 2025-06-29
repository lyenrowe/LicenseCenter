package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lyenrowe/LicenseCenter/internal/config"
	"github.com/lyenrowe/LicenseCenter/internal/database"
	"github.com/lyenrowe/LicenseCenter/internal/router"
	"github.com/lyenrowe/LicenseCenter/internal/services"
	"github.com/lyenrowe/LicenseCenter/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AdminHandlerTestSuite struct {
	suite.Suite
	app        *gin.Engine
	adminToken string
	authID     uint
}

func (suite *AdminHandlerTestSuite) SetupSuite() {
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

	// 初始化路由
	suite.app = router.SetupRouter()

	// 创建管理员并获取token
	suite.adminToken = suite.createAdminAndGetToken()
}

func (suite *AdminHandlerTestSuite) TearDownSuite() {
	// 清理资源
	if database.DB != nil {
		database.DB.Close()
	}
}

func (suite *AdminHandlerTestSuite) SetupTest() {
	// 每个测试前清理数据
	database.GetDB().Exec("DELETE FROM authorizations")
	database.GetDB().Exec("DELETE FROM licenses")
}

func (suite *AdminHandlerTestSuite) createAdminAndGetToken() string {
	// 创建管理员
	adminService := services.NewAdminService()
	_, err := adminService.CreateAdmin(&services.CreateAdminRequest{
		Username: "testadmin",
		Password: "password123",
	})
	assert.NoError(suite.T(), err)

	// 登录获取token
	loginData := map[string]string{
		"username": "testadmin",
		"password": "password123",
	}
	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/api/admin/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.app.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	token := response["token"].(string)
	assert.NotEmpty(suite.T(), token)
	return token
}

func (suite *AdminHandlerTestSuite) createTestAuthorization() uint {
	authService := services.NewAuthorizationService()
	auth, err := authService.CreateAuthorization(&services.CreateAuthorizationRequest{
		CustomerName: "测试客户",
		MaxSeats:     5,
	})
	assert.NoError(suite.T(), err)
	return auth.ID
}

func (suite *AdminHandlerTestSuite) TestGetAuthorizationDetails() {
	// 先创建一个测试授权码
	authID := suite.createTestAuthorization()

	url := fmt.Sprintf("/api/admin/authorizations/%d/details", authID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)

	w := httptest.NewRecorder()
	suite.app.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "测试客户", data["customer_name"])
	assert.NotEmpty(suite.T(), data["authorization_code"])
	assert.Equal(suite.T(), float64(5), data["max_seats"])
	assert.Equal(suite.T(), float64(0), data["used_seats"])
	assert.NotNil(suite.T(), data["devices"])

	// 设备列表应该为空（没有激活的设备）
	devices := data["devices"].([]interface{})
	assert.Len(suite.T(), devices, 0)
}

func (suite *AdminHandlerTestSuite) TestGetAuthorizationDetailsNotFound() {
	req, _ := http.NewRequest("GET", "/api/admin/authorizations/99999/details", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)

	w := httptest.NewRecorder()
	suite.app.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *AdminHandlerTestSuite) TestForceUnbindLicense() {
	// 先创建一个测试授权码
	authID := suite.createTestAuthorization()

	// 首先需要激活一个设备
	licenseService := services.NewLicenseService()

	// 获取授权码
	authService := services.NewAuthorizationService()
	auth, err := authService.GetAuthorizationByID(authID)
	assert.NoError(suite.T(), err)

	// 模拟激活设备
	bindFiles := []services.BindFile{
		{
			MachineID: "test-machine-001",
			Hostname:  "test-host",
		},
	}

	licenseFiles, err := licenseService.ActivateLicenses(auth.AuthorizationCode, bindFiles)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), licenseFiles, 1)

	// 获取激活的设备ID
	licenses, err := licenseService.GetLicensesByAuth(auth.AuthorizationCode)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), licenses, 1)
	licenseID := licenses[0].ID

	// 测试强制解绑
	unbindData := map[string]string{
		"reason": "测试强制解绑",
	}
	jsonData, _ := json.Marshal(unbindData)

	url := fmt.Sprintf("/api/admin/licenses/%d/force-unbind", licenseID)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)

	w := httptest.NewRecorder()
	suite.app.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "设备解绑成功", response["message"])

	// 验证设备已被解绑
	licenses, err = licenseService.GetLicensesByAuth(auth.AuthorizationCode)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), licenses, 1)
	assert.Equal(suite.T(), "unbound", licenses[0].Status)
}

func (suite *AdminHandlerTestSuite) TestForceUnbindLicenseNotFound() {
	unbindData := map[string]string{
		"reason": "测试解绑不存在的设备",
	}
	jsonData, _ := json.Marshal(unbindData)

	req, _ := http.NewRequest("POST", "/api/admin/licenses/99999/force-unbind", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)

	w := httptest.NewRecorder()
	suite.app.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func TestAdminHandlerSuite(t *testing.T) {
	suite.Run(t, new(AdminHandlerTestSuite))
}
