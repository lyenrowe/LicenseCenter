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

type RSAServiceTestSuite struct {
	suite.Suite
	rsaService *services.RSAService
}

func (suite *RSAServiceTestSuite) SetupSuite() {
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
	suite.rsaService = services.NewRSAService()
}

func (suite *RSAServiceTestSuite) TearDownSuite() {
	// 清理资源
	if database.DB != nil {
		database.DB.Close()
	}
}

func (suite *RSAServiceTestSuite) TestGenerateAndSaveKeyPair() {
	// 测试生成密钥对
	privateKey, publicKey, err := suite.rsaService.GenerateAndSaveKeyPair()

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), privateKey)
	assert.NotNil(suite.T(), publicKey)

	// 验证可以获取PEM格式的公钥
	publicKeyPEM, err := suite.rsaService.GetPublicKeyPEM()
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), publicKeyPEM, "-----BEGIN PUBLIC KEY-----")
}

func (suite *RSAServiceTestSuite) TestGetPublicKeyPEM() {
	// 先生成密钥对
	_, _, err := suite.rsaService.GenerateAndSaveKeyPair()
	assert.NoError(suite.T(), err)

	// 测试获取公钥
	publicKeyPEM, err := suite.rsaService.GetPublicKeyPEM()
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), publicKeyPEM)
	assert.Contains(suite.T(), publicKeyPEM, "-----BEGIN PUBLIC KEY-----")
}

func (suite *RSAServiceTestSuite) TestSignAndVerify() {
	// 先生成密钥对
	_, _, err := suite.rsaService.GenerateAndSaveKeyPair()
	assert.NoError(suite.T(), err)

	// 测试数据
	testData := "Hello, World!"

	// 测试签名
	signature, err := suite.rsaService.SignData([]byte(testData))
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), signature)

	// 测试验证
	err = suite.rsaService.VerifySignature([]byte(testData), signature)
	assert.NoError(suite.T(), err)

	// 测试错误数据验证
	err = suite.rsaService.VerifySignature([]byte("Wrong data"), signature)
	assert.Error(suite.T(), err)
}

func (suite *RSAServiceTestSuite) TestMultipleKeyPairs() {
	// 生成第一个密钥对
	_, _, err := suite.rsaService.GenerateAndSaveKeyPair()
	assert.NoError(suite.T(), err)

	// 获取第一个公钥PEM
	publicKeyPEM1, err := suite.rsaService.GetPublicKeyPEM()
	assert.NoError(suite.T(), err)

	// 生成第二个密钥对（应该替换第一个）
	_, _, err = suite.rsaService.GenerateAndSaveKeyPair()
	assert.NoError(suite.T(), err)

	// 获取第二个公钥PEM
	publicKeyPEM2, err := suite.rsaService.GetPublicKeyPEM()
	assert.NoError(suite.T(), err)

	// 第二个密钥应该不同于第一个
	assert.NotEqual(suite.T(), publicKeyPEM1, publicKeyPEM2)

	// 当前活跃的应该是第二个密钥
	currentPublicKey, err := suite.rsaService.GetPublicKeyPEM()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), publicKeyPEM2, currentPublicKey)
}

func (suite *RSAServiceTestSuite) TestEmptyDatabase() {
	// 清空数据库中的密钥
	database.GetDB().Exec("DELETE FROM rsa_keys")

	// 在没有密钥的情况下，GetPublicKeyPEM应该返回错误
	_, err := suite.rsaService.GetPublicKeyPEM()
	assert.Error(suite.T(), err)

	// 但是SignData会自动创建密钥
	signature, err := suite.rsaService.SignData([]byte("test"))
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), signature)

	// 现在GetPublicKeyPEM应该能正常工作
	publicKeyPEM, err := suite.rsaService.GetPublicKeyPEM()
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), publicKeyPEM)
	assert.Contains(suite.T(), publicKeyPEM, "-----BEGIN PUBLIC KEY-----")
}

// 运行测试套件
func TestRSAServiceSuite(t *testing.T) {
	suite.Run(t, new(RSAServiceTestSuite))
}
