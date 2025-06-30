package tests

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/lyenrowe/LicenseCenter/internal/config"
	"github.com/lyenrowe/LicenseCenter/internal/database"
	"github.com/lyenrowe/LicenseCenter/internal/services"
	"github.com/lyenrowe/LicenseCenter/pkg/crypto"
	"github.com/lyenrowe/LicenseCenter/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HybridEncryptionTestSuite struct {
	suite.Suite
	licenseService *services.LicenseService
	authService    *services.AuthorizationService
	rsaService     *services.RSAService
	privateKey     *rsa.PrivateKey
	publicKey      *rsa.PublicKey
}

func (suite *HybridEncryptionTestSuite) SetupSuite() {
	// 初始化测试配置
	err := config.LoadConfig("../configs/app.yaml")
	assert.NoError(suite.T(), err)

	// 初始化日志
	err = logger.InitLogger("debug", "../logs/hybrid_test.log")
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
	suite.licenseService = services.NewLicenseService()
	suite.authService = services.NewAuthorizationService()
	suite.rsaService = services.NewRSAService()

	// 生成RSA密钥对
	suite.privateKey, suite.publicKey, err = suite.rsaService.GenerateAndSaveKeyPair()
	assert.NoError(suite.T(), err)
}

func (suite *HybridEncryptionTestSuite) TearDownSuite() {
	// 清理资源
	if database.DB != nil {
		database.DB.Close()
	}
}

func (suite *HybridEncryptionTestSuite) TestHybridEncryptDecrypt() {
	// 测试数据
	testData := []byte(`{"test": "data", "number": 123, "array": [1, 2, 3]}`)

	// 测试加密
	encryptedData, err := crypto.HybridEncrypt(suite.publicKey, testData)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), encryptedData)

	// 测试解密
	decryptedData, err := crypto.HybridDecrypt(suite.privateKey, encryptedData)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), testData, decryptedData)
}

func (suite *HybridEncryptionTestSuite) TestEncryptFileToBase64() {
	// 测试数据
	testJSON := map[string]interface{}{
		"hostname":     "test-host",
		"machine_id":   "1234567890abcdef1234567890abcdef", // 32位十六进制字符串（MD5格式）
		"request_time": time.Now(),
	}

	jsonData, err := json.Marshal(testJSON)
	assert.NoError(suite.T(), err)

	// 测试Base64加密
	base64Encrypted, err := crypto.EncryptFileToBase64(suite.publicKey, jsonData)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), base64Encrypted)

	// 测试Base64解密
	decryptedData, err := crypto.DecryptFileFromBase64(suite.privateKey, base64Encrypted)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), jsonData, decryptedData)

	// 验证解密后的JSON
	var decryptedJSON map[string]interface{}
	err = json.Unmarshal(decryptedData, &decryptedJSON)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), testJSON["hostname"], decryptedJSON["hostname"])
	assert.Equal(suite.T(), testJSON["machine_id"], decryptedJSON["machine_id"])
}

func (suite *HybridEncryptionTestSuite) TestEncryptedBindFileFlow() {
	// 创建测试授权码
	auth, err := suite.authService.CreateAuthorization(&services.CreateAuthorizationRequest{
		CustomerName:      "测试客户",
		AuthorizationCode: "TEST-ENCRYPT-001",
		MaxSeats:          5,
	})
	assert.NoError(suite.T(), err)

	// 创建绑定文件数据
	bindFile := services.BindFile{
		Hostname:    "test-host-encrypted",
		MachineID:   "abc123456789abcd1234567890abcdef", // 32位十六进制字符串（MD5格式）
		RequestTime: time.Now(),
	}

	// 加密绑定文件
	encryptedBindFile, err := suite.licenseService.EncryptBindFile(bindFile)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "bind", encryptedBindFile.FileType)
	assert.NotEmpty(suite.T(), encryptedBindFile.EncryptedContent)

	// 解密绑定文件
	decryptedBindFile, err := suite.licenseService.DecryptBindFile(encryptedBindFile.EncryptedContent)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), bindFile.Hostname, decryptedBindFile.Hostname)
	assert.Equal(suite.T(), bindFile.MachineID, decryptedBindFile.MachineID)

	// 测试加密批量激活流程
	encryptedBindFiles := []string{encryptedBindFile.EncryptedContent}
	encryptedLicenseFiles, err := suite.licenseService.ActivateLicensesEncrypted(auth.AuthorizationCode, encryptedBindFiles)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), encryptedLicenseFiles, 1)
	assert.Equal(suite.T(), "license", encryptedLicenseFiles[0].FileType)
	assert.NotEmpty(suite.T(), encryptedLicenseFiles[0].EncryptedContent)

	// 验证席位消耗
	updatedAuth, err := suite.authService.GetAuthorizationByCode(auth.AuthorizationCode)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, updatedAuth.UsedSeats)
}

func (suite *HybridEncryptionTestSuite) TestEncryptedTransferFlow() {
	// 创建测试授权码
	auth, err := suite.authService.CreateAuthorization(&services.CreateAuthorizationRequest{
		CustomerName:      "传输测试客户",
		AuthorizationCode: "TEST-TRANSFER-001",
		MaxSeats:          3,
	})
	assert.NoError(suite.T(), err)

	// 第一步：激活旧设备
	oldBindFile := services.BindFile{
		Hostname:    "old-device",
		MachineID:   "def456789abcdef01234567890abcdef", // 32位十六进制字符串（MD5格式）
		RequestTime: time.Now(),
	}

	licenseFiles, err := suite.licenseService.ActivateLicenses(auth.AuthorizationCode, []services.BindFile{oldBindFile})
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), licenseFiles, 1)

	// 第二步：生成解绑文件
	unbindFile := services.UnbindFile{
		SignedLicense: licenseFiles[0],
		UnbindProof:   "test-unbind-proof", // 实际应该是正确的签名
	}

	// 加密解绑文件
	encryptedUnbindFile, err := suite.licenseService.EncryptUnbindFile(unbindFile)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "unbind", encryptedUnbindFile.FileType)

	// 第三步：创建新设备绑定文件
	newBindFile := services.BindFile{
		Hostname:    "new-device",
		MachineID:   "fed987654321dcba0987654321dcba09ab", // 32位十六进制字符串（MD5格式）
		RequestTime: time.Now(),
	}

	encryptedNewBindFile, err := suite.licenseService.EncryptBindFile(newBindFile)
	assert.NoError(suite.T(), err)

	// 解密验证
	decryptedUnbindFile, err := suite.licenseService.DecryptUnbindFile(encryptedUnbindFile.EncryptedContent)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), oldBindFile.MachineID, decryptedUnbindFile.SignedLicense.LicenseData.MachineID)

	decryptedNewBindFile, err := suite.licenseService.DecryptBindFile(encryptedNewBindFile.EncryptedContent)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newBindFile.Hostname, decryptedNewBindFile.Hostname)
	assert.Equal(suite.T(), newBindFile.MachineID, decryptedNewBindFile.MachineID)
}

func (suite *HybridEncryptionTestSuite) TestEncryptionPerformance() {
	// 测试不同大小数据的加密性能
	testSizes := []int{100, 1000, 10000, 100000} // 字节

	for _, size := range testSizes {
		suite.T().Run(fmt.Sprintf("Size_%d_bytes", size), func(t *testing.T) {
			// 生成测试数据
			testData := make([]byte, size)
			for i := range testData {
				testData[i] = byte(i % 256)
			}

			// 测量加密时间
			start := time.Now()
			encryptedData, err := crypto.HybridEncrypt(suite.publicKey, testData)
			encryptDuration := time.Since(start)

			assert.NoError(t, err)
			assert.NotEmpty(t, encryptedData)

			// 测量解密时间
			start = time.Now()
			decryptedData, err := crypto.HybridDecrypt(suite.privateKey, encryptedData)
			decryptDuration := time.Since(start)

			assert.NoError(t, err)
			assert.Equal(t, testData, decryptedData)

			t.Logf("Size: %d bytes, Encrypt: %v, Decrypt: %v", size, encryptDuration, decryptDuration)
		})
	}
}

func (suite *HybridEncryptionTestSuite) TestEncryptionErrorHandling() {
	testData := []byte("test data")

	// 测试无效的公钥
	invalidPublicKey := &rsa.PublicKey{}
	_, err := crypto.HybridEncrypt(invalidPublicKey, testData)
	assert.Error(suite.T(), err)

	// 测试无效的私钥
	invalidPrivateKey := &rsa.PrivateKey{}
	_, err = crypto.HybridDecrypt(invalidPrivateKey, []byte("invalid data"))
	assert.Error(suite.T(), err)

	// 测试无效的加密数据格式
	_, err = crypto.HybridDecrypt(suite.privateKey, []byte("too short"))
	assert.Error(suite.T(), err)

	// 测试无效的Base64数据
	_, err = crypto.DecryptFileFromBase64(suite.privateKey, "invalid base64!")
	assert.Error(suite.T(), err)
}

// 运行测试套件
func TestHybridEncryptionSuite(t *testing.T) {
	suite.Run(t, new(HybridEncryptionTestSuite))
}
