package services

import (
	"crypto/rsa"

	"github.com/lyenrowe/LicenseCenter/internal/database"
	"github.com/lyenrowe/LicenseCenter/internal/models"
	"github.com/lyenrowe/LicenseCenter/pkg/crypto"
	"github.com/lyenrowe/LicenseCenter/pkg/errors"
	"gorm.io/gorm"
)

// RSAService RSA密钥管理服务
type RSAService struct {
	db *gorm.DB
}

// NewRSAService 创建RSA服务实例
func NewRSAService() *RSAService {
	return &RSAService{
		db: database.GetDB(),
	}
}

// GetActiveKeyPair 获取当前活跃的RSA密钥对
func (s *RSAService) GetActiveKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	var rsaKey models.RSAKey
	err := s.db.Where("is_active = ?", true).First(&rsaKey).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果没有活跃密钥，创建一个新的
			return s.GenerateAndSaveKeyPair()
		}
		return nil, nil, errors.WrapError(err, 50001, "获取RSA密钥失败")
	}

	// 解析私钥
	privateKey, err := crypto.LoadPrivateKeyFromPEM(rsaKey.PrivateKey)
	if err != nil {
		return nil, nil, errors.WrapError(err, 50002, "解析RSA私钥失败")
	}

	// 解析公钥
	publicKey, err := crypto.LoadPublicKeyFromPEM(rsaKey.PublicKey)
	if err != nil {
		return nil, nil, errors.WrapError(err, 50002, "解析RSA公钥失败")
	}

	return privateKey, publicKey, nil
}

// GenerateAndSaveKeyPair 生成并保存新的RSA密钥对
func (s *RSAService) GenerateAndSaveKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	// 生成新的RSA密钥对
	keyPair, err := crypto.GenerateRSAKeyPair(2048)
	if err != nil {
		return nil, nil, errors.WrapError(err, 50002, "生成RSA密钥对失败")
	}

	// 转换为PEM格式
	privateKeyPEM, err := keyPair.PrivateKeyToPEM()
	if err != nil {
		return nil, nil, errors.WrapError(err, 50002, "转换私钥为PEM格式失败")
	}

	publicKeyPEM, err := keyPair.PublicKeyToPEM()
	if err != nil {
		return nil, nil, errors.WrapError(err, 50002, "转换公钥为PEM格式失败")
	}

	// 开始数据库事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 将现有的活跃密钥设为非活跃
	err = tx.Model(&models.RSAKey{}).Where("is_active = ?", true).Update("is_active", false).Error
	if err != nil {
		tx.Rollback()
		return nil, nil, errors.WrapError(err, 50001, "更新旧密钥状态失败")
	}

	// 保存新密钥
	newKey := models.RSAKey{
		PrivateKey: privateKeyPEM,
		PublicKey:  publicKeyPEM,
		IsActive:   true,
	}

	err = tx.Create(&newKey).Error
	if err != nil {
		tx.Rollback()
		return nil, nil, errors.WrapError(err, 50001, "保存新RSA密钥失败")
	}

	// 提交事务
	err = tx.Commit().Error
	if err != nil {
		return nil, nil, errors.WrapError(err, 50001, "提交事务失败")
	}

	return keyPair.PrivateKey, keyPair.PublicKey, nil
}

// GetPublicKeyPEM 获取当前活跃的公钥PEM格式
func (s *RSAService) GetPublicKeyPEM() (string, error) {
	var rsaKey models.RSAKey
	err := s.db.Where("is_active = ?", true).First(&rsaKey).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", errors.ErrCryptoError
		}
		return "", errors.WrapError(err, 50001, "获取RSA公钥失败")
	}

	return rsaKey.PublicKey, nil
}

// SignData 使用当前活跃的私钥签名数据
func (s *RSAService) SignData(data []byte) (string, error) {
	privateKey, _, err := s.GetActiveKeyPair()
	if err != nil {
		return "", err
	}

	signature, err := crypto.SignData(privateKey, data)
	if err != nil {
		return "", errors.WrapError(err, 50002, "RSA签名失败")
	}

	return signature, nil
}

// VerifySignature 使用公钥验证签名
func (s *RSAService) VerifySignature(data []byte, signature string) error {
	_, publicKey, err := s.GetActiveKeyPair()
	if err != nil {
		return err
	}

	err = crypto.VerifySignature(publicKey, data, signature)
	if err != nil {
		return errors.WrapError(err, 50002, "RSA签名验证失败")
	}

	return nil
}

// ListKeys 列出所有RSA密钥
func (s *RSAService) ListKeys() ([]models.RSAKey, error) {
	var keys []models.RSAKey
	err := s.db.Order("created_at DESC").Find(&keys).Error
	if err != nil {
		return nil, errors.WrapError(err, 50001, "获取RSA密钥列表失败")
	}

	return keys, nil
}

// RotateKeys 轮换密钥（生成新密钥并设为活跃）
func (s *RSAService) RotateKeys() error {
	_, _, err := s.GenerateAndSaveKeyPair()
	return err
}
