package tests

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lyenrowe/LicenseCenter/internal/config"
	"github.com/lyenrowe/LicenseCenter/internal/database"
	"github.com/lyenrowe/LicenseCenter/internal/models"
	"github.com/lyenrowe/LicenseCenter/internal/router"
	"github.com/lyenrowe/LicenseCenter/internal/services"
	"github.com/lyenrowe/LicenseCenter/pkg/crypto"
	"github.com/lyenrowe/LicenseCenter/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testApp    *gin.Engine
	testServer *httptest.Server
	baseURL    string
)

// setupTestEnvironment 设置测试环境
func setupTestEnvironment(t *testing.T) {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 初始化测试配置
	err := config.LoadConfig("../configs/app.yaml")
	require.NoError(t, err)

	// 初始化日志
	err = logger.InitLogger("error", "../logs/transfer_test.log")
	require.NoError(t, err)

	// 使用内存数据库进行测试
	config.AppConfig.Database.Driver = "sqlite"
	config.AppConfig.Database.DSN = ":memory:"

	// 初始化数据库
	err = database.InitDatabase(&config.AppConfig.Database)
	require.NoError(t, err)

	// 执行数据库迁移
	err = database.DB.AutoMigrate()
	require.NoError(t, err)

	// 创建索引
	err = database.DB.CreateIndexes()
	require.NoError(t, err)

	// 初始化路由
	testApp = router.SetupRouter()

	// 创建默认管理员和RSA密钥
	setupDefaultData(t)
}

// teardownTestEnvironment 清理测试环境
func teardownTestEnvironment(t *testing.T) {
	if testServer != nil {
		testServer.Close()
	}
}

// startTestServer 启动测试服务器
func startTestServer() {
	testServer = httptest.NewServer(testApp)
	baseURL = testServer.URL
}

// setupDefaultData 设置默认数据
func setupDefaultData(t *testing.T) {
	// 创建默认管理员
	adminService := services.NewAdminService()
	req := &services.CreateAdminRequest{
		Username: "admin",
		Password: "admin123",
	}
	_, err := adminService.CreateAdmin(req)
	require.NoError(t, err)

	// 生成RSA密钥对
	rsaService := services.NewRSAService()
	_, _, err = rsaService.GenerateAndSaveKeyPair()
	require.NoError(t, err)
}

// generateTestMachineID 生成测试用的机器ID
func generateTestMachineID(suffix string) string {
	data := fmt.Sprintf("test-machine-%s-%d", suffix, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:32] // 返回32位MD5格式
}

// createTestAuthorization 创建测试授权码
func createTestAuthorization(t *testing.T, authCode, customerName string, maxSeats int) {
	authService := services.NewAuthorizationService()
	durationYears := 1
	req := &services.CreateAuthorizationRequest{
		CustomerName:      customerName,
		AuthorizationCode: authCode,
		MaxSeats:          maxSeats,
		DurationYears:     &durationYears,
	}
	_, err := authService.CreateAuthorization(req)
	require.NoError(t, err)
}

// loginAndGetToken 登录并获取token
func loginAndGetToken(t *testing.T, authCode string) string {
	loginData := map[string]string{
		"authorization_code": authCode,
		"captcha_token":      "fallback_captcha_token_test", // 使用开发模式的降级token
	}

	jsonData, _ := json.Marshal(loginData)
	resp, err := http.Post(baseURL+"/api/login", "application/json", bytes.NewReader(jsonData))
	require.NoError(t, err)
	defer resp.Body.Close()

	// 如果失败，打印响应内容以便调试
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Logf("Login failed with status %d: %s", resp.StatusCode, string(body))
	}

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	token, ok := result["session_token"].(string)
	require.True(t, ok, "session_token not found in response")

	return token
}

// activateDevice 激活设备并返回license_key
func activateDevice(t *testing.T, token, machineID, hostname string) string {
	// 创建bind文件
	bindData := services.BindFile{
		Hostname:    hostname,
		MachineID:   machineID,
		RequestTime: time.Now().UTC(),
	}

	// 加密绑定文件
	licenseService := services.NewLicenseService()
	encryptedFile, err := licenseService.EncryptBindFile(bindData)
	require.NoError(t, err)

	// 保存到临时文件
	tmpFile := filepath.Join(os.TempDir(), "test_activate.bind")
	err = os.WriteFile(tmpFile, []byte(encryptedFile.EncryptedContent), 0644)
	require.NoError(t, err)
	defer os.Remove(tmpFile)

	// 创建multipart请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("bind_files", filepath.Base(tmpFile))
	require.NoError(t, err)

	content, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	_, err = part.Write(content)
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	// 发送请求
	req, err := http.NewRequest("POST", baseURL+"/api/actions/activate-licenses", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	// 从数据库获取license_key
	var license models.License
	err = database.GetDB().Where("machine_id = ? AND status = ?", machineID, models.LicenseStatusActive).First(&license).Error
	require.NoError(t, err)

	return license.LicenseKey
}

// TestTransferLicense 测试完整的授权转移流程
func TestTransferLicense(t *testing.T) {
	// 设置测试环境
	setupTestEnvironment(t)
	defer teardownTestEnvironment(t)

	// 启动测试服务器
	go startTestServer()
	time.Sleep(2 * time.Second) // 等待服务器启动

	// 测试数据
	authCode := "TEST-TRANSFER-001"
	oldMachineID := generateTestMachineID("old")
	newMachineID := generateTestMachineID("new")

	// 步骤1: 创建测试授权码
	createTestAuthorization(t, authCode, "转移测试客户", 5)

	// 步骤2: 登录获取token
	token := loginAndGetToken(t, authCode)

	// 步骤3: 激活旧设备
	oldLicenseKey := activateDevice(t, token, oldMachineID, "OLD-DEVICE")

	// 步骤4: 生成解绑文件
	unbindFile := generateUnbindFile(t, oldLicenseKey, oldMachineID)

	// 步骤5: 生成新设备绑定文件
	bindFile := generateBindFile(t, newMachineID, "NEW-DEVICE")

	// 步骤6: 执行授权转移
	transferLicense(t, token, unbindFile, bindFile)

	// 步骤7: 验证结果
	verifyTransferResult(t, authCode, oldMachineID, newMachineID)
}

// generateUnbindFile 生成真实的解绑文件（包含正确的签名）
func generateUnbindFile(t *testing.T, licenseKey, machineID string) string {
	// 从数据库获取license记录以获取解绑私钥
	var license models.License
	err := database.GetDB().Where("license_key = ? AND machine_id = ?", licenseKey, machineID).First(&license).Error
	require.NoError(t, err)

	// 构造需要签名的数据
	unbindTime := time.Now().UTC()
	signData := fmt.Sprintf("%s:%s:%s:%s",
		licenseKey,
		machineID,
		unbindTime.Format(time.RFC3339),
		"OLD-DEVICE")

	// 使用解绑私钥签名
	unbindPrivateKey, err := crypto.LoadPrivateKeyFromPEM(license.UnbindPrivateKey)
	require.NoError(t, err)

	signature, err := crypto.SignData(unbindPrivateKey, []byte(signData))
	require.NoError(t, err)

	unbindData := services.UnbindFile{
		LicenseKey: licenseKey,
		MachineID:  machineID,
		UnbindMetadata: services.UnbindMetadata{
			UnbindTime:    unbindTime,
			Hostname:      "OLD-DEVICE",
			ClientVersion: "1.0.0",
			UnbindReason:  "device_replacement",
		},
		UnbindProof: signature,
	}

	// 加密解绑文件
	licenseService := services.NewLicenseService()
	encryptedFile, err := licenseService.EncryptUnbindFile(unbindData)
	require.NoError(t, err)

	// 保存到临时文件
	tmpFile := filepath.Join(os.TempDir(), "test.unbind")
	err = os.WriteFile(tmpFile, []byte(encryptedFile.EncryptedContent), 0644)
	require.NoError(t, err)

	return tmpFile
}

// generateBindFile 生成绑定文件
func generateBindFile(t *testing.T, machineID, hostname string) string {
	bindData := services.BindFile{
		Hostname:    hostname,
		MachineID:   machineID,
		RequestTime: time.Now().UTC(),
	}

	// 加密绑定文件
	licenseService := services.NewLicenseService()
	encryptedFile, err := licenseService.EncryptBindFile(bindData)
	require.NoError(t, err)

	// 保存到临时文件
	tmpFile := filepath.Join(os.TempDir(), "test.bind")
	err = os.WriteFile(tmpFile, []byte(encryptedFile.EncryptedContent), 0644)
	require.NoError(t, err)

	return tmpFile
}

// transferLicense 执行授权转移
func transferLicense(t *testing.T, token, unbindFile, bindFile string) {
	// 创建multipart请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加unbind文件
	unbindPart, err := writer.CreateFormFile("unbind_file", filepath.Base(unbindFile))
	require.NoError(t, err)
	unbindContent, err := os.ReadFile(unbindFile)
	require.NoError(t, err)
	_, err = unbindPart.Write(unbindContent)
	require.NoError(t, err)

	// 添加bind文件
	bindPart, err := writer.CreateFormFile("bind_file", filepath.Base(bindFile))
	require.NoError(t, err)
	bindContent, err := os.ReadFile(bindFile)
	require.NoError(t, err)
	_, err = bindPart.Write(bindContent)
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	// 发送请求
	req, err := http.NewRequest("POST", baseURL+"/api/actions/transfer-license", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// 验证响应
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 保存新的license文件
	licenseData, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.NotEmpty(t, licenseData)

	// 清理临时文件
	os.Remove(unbindFile)
	os.Remove(bindFile)
}

// verifyTransferResult 验证转移结果
func verifyTransferResult(t *testing.T, authCode, oldMachineID, newMachineID string) {
	// 获取授权信息
	authService := services.NewAuthorizationService()
	auth, err := authService.GetAuthorizationByCode(authCode)
	require.NoError(t, err)

	// 验证席位数没有变化（转移不应该改变已用席位数）
	assert.Equal(t, 1, auth.UsedSeats, "转移后已用席位数应该保持不变")

	// 验证旧设备状态
	licenseService := services.NewLicenseService()
	licenses, err := licenseService.GetLicensesByAuth(authCode)
	require.NoError(t, err)

	var oldLicense, newLicense *models.License
	for i := range licenses {
		if licenses[i].MachineID == oldMachineID {
			oldLicense = &licenses[i]
		} else if licenses[i].MachineID == newMachineID {
			newLicense = &licenses[i]
		}
	}

	// 验证旧设备已解绑
	require.NotNil(t, oldLicense, "应该找到旧设备记录")
	assert.Equal(t, models.LicenseStatusUnbound, oldLicense.Status, "旧设备应该是解绑状态")
	assert.NotNil(t, oldLicense.UnboundAt, "应该有解绑时间")

	// 验证新设备已激活
	require.NotNil(t, newLicense, "应该找到新设备记录")
	assert.Equal(t, models.LicenseStatusActive, newLicense.Status, "新设备应该是激活状态")
	assert.Equal(t, oldLicense.ExpiresAt.Unix(), newLicense.ExpiresAt.Unix(), "新设备应该继承旧设备的到期时间")
}

// TestTransferLicenseErrors 测试授权转移的错误情况
func TestTransferLicenseErrors(t *testing.T) {
	// 设置测试环境
	setupTestEnvironment(t)
	defer teardownTestEnvironment(t)

	// 启动测试服务器
	go startTestServer()
	time.Sleep(2 * time.Second)

	authCode := "TEST-TRANSFER-ERR-001"
	createTestAuthorization(t, authCode, "错误测试客户", 1)
	token := loginAndGetToken(t, authCode)

	t.Run("无效的解绑文件", func(t *testing.T) {
		// 创建无效的解绑文件
		invalidUnbind := filepath.Join(os.TempDir(), "invalid.unbind")
		os.WriteFile(invalidUnbind, []byte("invalid content"), 0644)
		defer os.Remove(invalidUnbind)

		// 创建有效的绑定文件
		bindFile := generateBindFile(t, generateTestMachineID("test"), "TEST-DEVICE")
		defer os.Remove(bindFile)

		// 尝试转移
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// 添加文件
		unbindPart, _ := writer.CreateFormFile("unbind_file", "invalid.unbind")
		unbindContent, _ := os.ReadFile(invalidUnbind)
		unbindPart.Write(unbindContent)

		bindPart, _ := writer.CreateFormFile("bind_file", "test.bind")
		bindContent, _ := os.ReadFile(bindFile)
		bindPart.Write(bindContent)

		writer.Close()

		req, _ := http.NewRequest("POST", baseURL+"/api/actions/transfer-license", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// 应该返回错误
		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})
}
