package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lyenrowe/LicenseCenter/internal/database"
	"github.com/lyenrowe/LicenseCenter/internal/models"
	"github.com/lyenrowe/LicenseCenter/pkg/crypto"
	"github.com/lyenrowe/LicenseCenter/pkg/errors"
	"github.com/lyenrowe/LicenseCenter/pkg/utils"
	"gorm.io/gorm"
)

// LicenseService 授权服务
type LicenseService struct {
	db          *gorm.DB
	rsaService  *RSAService
	authService *AuthorizationService
}

// NewLicenseService 创建授权服务实例
func NewLicenseService() *LicenseService {
	return &LicenseService{
		db:          database.GetDB(),
		rsaService:  NewRSAService(),
		authService: NewAuthorizationService(),
	}
}

// BindFile 绑定请求文件结构
type BindFile struct {
	Hostname    string    `json:"hostname"`
	MachineID   string    `json:"machine_id"`
	RequestTime time.Time `json:"request_time"`
}

// LicenseFile 授权文件结构
type LicenseFile struct {
	LicenseData LicenseData `json:"license_data"`
	Signature   string      `json:"signature"`
}

// LicenseData 授权数据结构
type LicenseData struct {
	MachineID        string    `json:"machine_id"`
	IssuedAt         time.Time `json:"issued_at"`
	ExpiresAt        time.Time `json:"expires_at"`
	LicenseType      string    `json:"license_type"`
	UnbindPrivateKey string    `json:"unbind_private_key"`
}

// UnbindFile 解绑文件结构
type UnbindFile struct {
	SignedLicense LicenseFile `json:"signed_license"`
	UnbindProof   string      `json:"unbind_proof"`
}

// EncryptedFileResponse 加密文件响应结构
type EncryptedFileResponse struct {
	EncryptedContent string `json:"encrypted_content"` // Base64编码的加密数据
	FileType         string `json:"file_type"`         // 文件类型：bind, license, unbind
}

// ActivateLicensesEncrypted 批量激活设备（返回加密文件）
func (s *LicenseService) ActivateLicensesEncrypted(authCode string, encryptedBindFiles []string) ([]EncryptedFileResponse, error) {
	// 1. 解密绑定文件
	bindFiles, err := s.DecryptBindFiles(encryptedBindFiles)
	if err != nil {
		return nil, err
	}

	// 2. 生成普通授权文件
	licenseFiles, err := s.ActivateLicenses(authCode, bindFiles)
	if err != nil {
		return nil, err
	}

	// 3. 加密授权文件
	encryptedLicenseFiles, err := s.EncryptLicenseFiles(licenseFiles)
	if err != nil {
		return nil, err
	}

	return encryptedLicenseFiles, nil
}

// TransferLicenseEncrypted 授权转移（使用加密文件）
func (s *LicenseService) TransferLicenseEncrypted(authCode string, encryptedUnbindFile, encryptedBindFile string) (*EncryptedFileResponse, error) {
	// 1. 解密文件
	unbindFile, err := s.DecryptUnbindFile(encryptedUnbindFile)
	if err != nil {
		return nil, err
	}

	bindFile, err := s.DecryptBindFile(encryptedBindFile)
	if err != nil {
		return nil, err
	}

	// 2. 执行授权转移
	newLicenseFile, err := s.TransferLicense(authCode, *unbindFile, *bindFile)
	if err != nil {
		return nil, err
	}

	// 3. 加密新的授权文件
	encryptedLicenseFile, err := s.EncryptLicenseFile(*newLicenseFile)
	if err != nil {
		return nil, err
	}

	return encryptedLicenseFile, nil
}

// DecryptBindFiles 解密绑定文件列表
func (s *LicenseService) DecryptBindFiles(encryptedBindFiles []string) ([]BindFile, error) {
	privateKey, _, err := s.rsaService.GetActiveKeyPair()
	if err != nil {
		return nil, err
	}

	var bindFiles []BindFile
	for i, encryptedData := range encryptedBindFiles {
		jsonData, err := crypto.DecryptFileFromBase64(privateKey, encryptedData)
		if err != nil {
			return nil, errors.WrapError(err, 41003, fmt.Sprintf("解密第%d个绑定文件失败", i+1))
		}

		var bindFile BindFile
		if err := json.Unmarshal(jsonData, &bindFile); err != nil {
			return nil, errors.WrapError(err, 41003, fmt.Sprintf("解析第%d个绑定文件失败", i+1))
		}

		bindFiles = append(bindFiles, bindFile)
	}

	return bindFiles, nil
}

// DecryptBindFile 解密单个绑定文件
func (s *LicenseService) DecryptBindFile(encryptedBindFile string) (*BindFile, error) {
	privateKey, _, err := s.rsaService.GetActiveKeyPair()
	if err != nil {
		return nil, err
	}

	jsonData, err := crypto.DecryptFileFromBase64(privateKey, encryptedBindFile)
	if err != nil {
		return nil, errors.WrapError(err, 41003, "解密绑定文件失败")
	}

	var bindFile BindFile
	if err := json.Unmarshal(jsonData, &bindFile); err != nil {
		return nil, errors.WrapError(err, 41003, "解析绑定文件失败")
	}

	return &bindFile, nil
}

// DecryptUnbindFile 解密解绑文件
func (s *LicenseService) DecryptUnbindFile(encryptedUnbindFile string) (*UnbindFile, error) {
	privateKey, _, err := s.rsaService.GetActiveKeyPair()
	if err != nil {
		return nil, err
	}

	jsonData, err := crypto.DecryptFileFromBase64(privateKey, encryptedUnbindFile)
	if err != nil {
		return nil, errors.WrapError(err, 41004, "解密解绑文件失败")
	}

	var unbindFile UnbindFile
	if err := json.Unmarshal(jsonData, &unbindFile); err != nil {
		return nil, errors.WrapError(err, 41004, "解析解绑文件失败")
	}

	return &unbindFile, nil
}

// EncryptLicenseFiles 加密授权文件列表
func (s *LicenseService) EncryptLicenseFiles(licenseFiles []LicenseFile) ([]EncryptedFileResponse, error) {
	_, publicKey, err := s.rsaService.GetActiveKeyPair()
	if err != nil {
		return nil, err
	}

	var encryptedFiles []EncryptedFileResponse
	for i, licenseFile := range licenseFiles {
		jsonData, err := json.Marshal(licenseFile)
		if err != nil {
			return nil, errors.WrapError(err, 50002, fmt.Sprintf("序列化第%d个授权文件失败", i+1))
		}

		encryptedContent, err := crypto.EncryptFileToBase64(publicKey, jsonData)
		if err != nil {
			return nil, errors.WrapError(err, 50002, fmt.Sprintf("加密第%d个授权文件失败", i+1))
		}

		encryptedFiles = append(encryptedFiles, EncryptedFileResponse{
			EncryptedContent: encryptedContent,
			FileType:         "license",
		})
	}

	return encryptedFiles, nil
}

// EncryptLicenseFile 加密单个授权文件
func (s *LicenseService) EncryptLicenseFile(licenseFile LicenseFile) (*EncryptedFileResponse, error) {
	_, publicKey, err := s.rsaService.GetActiveKeyPair()
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(licenseFile)
	if err != nil {
		return nil, errors.WrapError(err, 50002, "序列化授权文件失败")
	}

	encryptedContent, err := crypto.EncryptFileToBase64(publicKey, jsonData)
	if err != nil {
		return nil, errors.WrapError(err, 50002, "加密授权文件失败")
	}

	return &EncryptedFileResponse{
		EncryptedContent: encryptedContent,
		FileType:         "license",
	}, nil
}

// EncryptBindFile 加密绑定文件（客户端使用）
func (s *LicenseService) EncryptBindFile(bindFile BindFile) (*EncryptedFileResponse, error) {
	_, publicKey, err := s.rsaService.GetActiveKeyPair()
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(bindFile)
	if err != nil {
		return nil, errors.WrapError(err, 50002, "序列化绑定文件失败")
	}

	encryptedContent, err := crypto.EncryptFileToBase64(publicKey, jsonData)
	if err != nil {
		return nil, errors.WrapError(err, 50002, "加密绑定文件失败")
	}

	return &EncryptedFileResponse{
		EncryptedContent: encryptedContent,
		FileType:         "bind",
	}, nil
}

// EncryptUnbindFile 加密解绑文件（客户端使用）
func (s *LicenseService) EncryptUnbindFile(unbindFile UnbindFile) (*EncryptedFileResponse, error) {
	_, publicKey, err := s.rsaService.GetActiveKeyPair()
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(unbindFile)
	if err != nil {
		return nil, errors.WrapError(err, 50002, "序列化解绑文件失败")
	}

	encryptedContent, err := crypto.EncryptFileToBase64(publicKey, jsonData)
	if err != nil {
		return nil, errors.WrapError(err, 50002, "加密解绑文件失败")
	}

	return &EncryptedFileResponse{
		EncryptedContent: encryptedContent,
		FileType:         "unbind",
	}, nil
}

// ActivateLicenses 批量激活设备
func (s *LicenseService) ActivateLicenses(authCode string, bindFiles []BindFile) ([]LicenseFile, error) {
	// 验证授权码
	auth, err := s.authService.ValidateAuthorizationCode(authCode)
	if err != nil {
		return nil, err
	}

	// 检查席位是否足够
	if !auth.HasAvailableSeats(len(bindFiles)) {
		return nil, errors.ErrInsufficientSeats
	}

	var licenseFiles []LicenseFile
	var createdLicenses []uint

	// 逐个处理绑定文件（不使用事务避免锁定问题）
	for _, bindFile := range bindFiles {
		// 验证绑定文件
		if err := s.validateBindFile(&bindFile); err != nil {
			// 回滚已创建的授权
			s.rollbackCreatedLicenses(createdLicenses)
			return nil, err
		}

		// 检查机器是否已经激活
		var existing models.License
		err := s.db.Where("machine_id = ? AND status = ?",
			bindFile.MachineID, models.LicenseStatusActive).First(&existing).Error
		if err == nil {
			// 回滚已创建的授权
			s.rollbackCreatedLicenses(createdLicenses)
			return nil, errors.ErrDuplicateMachine
		}
		if err != gorm.ErrRecordNotFound {
			// 回滚已创建的授权
			s.rollbackCreatedLicenses(createdLicenses)
			return nil, errors.WrapError(err, 50001, "检查机器状态失败")
		}

		// 生成授权文件
		licenseFile, license, err := s.generateLicenseFile(auth, &bindFile)
		if err != nil {
			// 回滚已创建的授权
			s.rollbackCreatedLicenses(createdLicenses)
			return nil, err
		}

		// 保存到数据库
		err = s.db.Create(license).Error
		if err != nil {
			// 回滚已创建的授权
			s.rollbackCreatedLicenses(createdLicenses)
			return nil, errors.WrapError(err, 50001, "保存授权记录失败")
		}

		createdLicenses = append(createdLicenses, license.ID)
		licenseFiles = append(licenseFiles, *licenseFile)
	}

	// 最后消耗席位
	err = s.authService.ConsumeSeats(auth.ID, len(bindFiles))
	if err != nil {
		// 回滚已创建的授权
		s.rollbackCreatedLicenses(createdLicenses)
		return nil, err
	}

	return licenseFiles, nil
}

// rollbackCreatedLicenses 回滚已创建的授权记录
func (s *LicenseService) rollbackCreatedLicenses(licenseIDs []uint) {
	if len(licenseIDs) == 0 {
		return
	}

	// 删除已创建的授权记录
	s.db.Where("id IN ?", licenseIDs).Delete(&models.License{})
}

// TransferLicense 授权转移
func (s *LicenseService) TransferLicense(authCode string, unbindFile UnbindFile, bindFile BindFile) (*LicenseFile, error) {
	// 验证授权码
	auth, err := s.authService.ValidateAuthorizationCode(authCode)
	if err != nil {
		return nil, err
	}

	var newLicenseFile *LicenseFile

	// 开始事务
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 验证解绑文件
		oldLicense, err := s.validateUnbindFile(&unbindFile)
		if err != nil {
			return err
		}

		// 检查解绑的授权是否属于当前授权码
		if oldLicense.AuthorizationID != auth.ID {
			return errors.NewAppError(41004, "解绑文件不属于当前授权码")
		}

		// 验证新绑定文件
		if err := s.validateBindFile(&bindFile); err != nil {
			return err
		}

		// 检查新机器是否已经激活
		var existing models.License
		err = tx.Where("machine_id = ? AND status = ?",
			bindFile.MachineID, models.LicenseStatusActive).First(&existing).Error
		if err == nil {
			return errors.ErrDuplicateMachine
		}
		if err != gorm.ErrRecordNotFound {
			return errors.WrapError(err, 50001, "检查新机器状态失败")
		}

		// 标记旧授权为解绑状态
		oldLicense.Unbind(false)
		err = tx.Save(oldLicense).Error
		if err != nil {
			return errors.WrapError(err, 50001, "更新旧授权状态失败")
		}

		// 生成新授权文件（继承旧授权的到期时间）
		licenseFile, license, err := s.generateLicenseFileWithExpiry(auth, &bindFile, oldLicense.ExpiresAt)
		if err != nil {
			return err
		}

		// 保存新授权记录
		err = tx.Create(license).Error
		if err != nil {
			return errors.WrapError(err, 50001, "保存新授权记录失败")
		}

		newLicenseFile = licenseFile
		return nil
	})

	if err != nil {
		return nil, err
	}

	return newLicenseFile, nil
}

// ForceUnbindLicense 管理员强制解绑设备
func (s *LicenseService) ForceUnbindLicense(licenseID uint, reason string) error {
	// 先获取许可证信息（不在事务中）
	var license models.License
	err := s.db.First(&license, licenseID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrLicenseNotFound
		}
		return errors.WrapError(err, 50001, "获取授权记录失败")
	}

	if !license.CanUnbind() {
		return errors.NewAppError(41004, "授权状态不允许解绑")
	}

	// 使用事务更新许可证状态
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 重新获取许可证以确保数据一致性
		var currentLicense models.License
		err := tx.First(&currentLicense, licenseID).Error
		if err != nil {
			return errors.WrapError(err, 50001, "获取授权记录失败")
		}

		if !currentLicense.CanUnbind() {
			return errors.NewAppError(41004, "授权状态不允许解绑")
		}

		// 标记为强制解绑
		currentLicense.Unbind(true)
		return tx.Save(&currentLicense).Error
	})
	if err != nil {
		return errors.WrapError(err, 50001, "更新授权状态失败")
	}

	// 在事务外释放席位，避免长时间持有锁
	return s.authService.ReleaseSeats(license.AuthorizationID, 1)
}

// GetLicensesByAuth 获取授权码下的所有设备
func (s *LicenseService) GetLicensesByAuth(authCode string) ([]models.License, error) {
	auth, err := s.authService.GetAuthorizationByCode(authCode)
	if err != nil {
		return nil, err
	}

	var licenses []models.License
	err = s.db.Where("authorization_id = ?", auth.ID).
		Order("created_at DESC").Find(&licenses).Error
	if err != nil {
		return nil, errors.WrapError(err, 50001, "获取授权设备列表失败")
	}

	return licenses, nil
}

// validateBindFile 验证绑定文件
func (s *LicenseService) validateBindFile(bindFile *BindFile) error {
	// 验证机器ID格式
	if !utils.ValidateMachineID(bindFile.MachineID) {
		return errors.ErrInvalidBindFile
	}

	// 验证主机名
	if bindFile.Hostname == "" {
		return errors.NewAppError(41003, "主机名不能为空")
	}

	// 验证请求时间（不能太旧）
	if time.Since(bindFile.RequestTime) > 24*time.Hour {
		return errors.NewAppError(41003, "绑定请求已过期")
	}

	return nil
}

// validateUnbindFile 验证解绑文件
func (s *LicenseService) validateUnbindFile(unbindFile *UnbindFile) (*models.License, error) {
	// 验证主签名
	licenseData, err := json.Marshal(unbindFile.SignedLicense.LicenseData)
	if err != nil {
		return nil, errors.ErrInvalidUnbindFile
	}

	err = s.rsaService.VerifySignature(licenseData, unbindFile.SignedLicense.Signature)
	if err != nil {
		return nil, errors.ErrInvalidSignature
	}

	// 根据机器ID查找授权记录
	var license models.License
	err = s.db.Where("machine_id = ? AND status = ?",
		unbindFile.SignedLicense.LicenseData.MachineID,
		models.LicenseStatusActive).First(&license).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrLicenseNotFound
		}
		return nil, errors.WrapError(err, 50001, "查找授权记录失败")
	}

	// 验证解绑证明
	signedLicenseData, err := json.Marshal(unbindFile.SignedLicense)
	if err != nil {
		return nil, errors.ErrInvalidUnbindFile
	}

	// 使用一次性解绑公钥验证
	unbindPublicKey, err := crypto.LoadPublicKeyFromPEM(license.UnbindPublicKey)
	if err != nil {
		return nil, errors.WrapError(err, 50002, "解析解绑公钥失败")
	}

	err = crypto.VerifySignature(unbindPublicKey, signedLicenseData, unbindFile.UnbindProof)
	if err != nil {
		return nil, errors.ErrInvalidSignature
	}

	return &license, nil
}

// generateLicenseFile 生成授权文件
func (s *LicenseService) generateLicenseFile(auth *models.Authorization, bindFile *BindFile) (*LicenseFile, *models.License, error) {
	expiresAt := auth.CalculateExpiryDate()
	return s.generateLicenseFileWithExpiry(auth, bindFile, expiresAt)
}

// generateLicenseFileWithExpiry 生成带指定到期时间的授权文件
func (s *LicenseService) generateLicenseFileWithExpiry(auth *models.Authorization, bindFile *BindFile, expiresAt time.Time) (*LicenseFile, *models.License, error) {
	// 生成一次性解绑密钥对
	unbindKeyPair, err := crypto.GenerateRSAKeyPair(2048)
	if err != nil {
		return nil, nil, errors.WrapError(err, 50002, "生成解绑密钥对失败")
	}

	unbindPrivateKeyPEM, err := unbindKeyPair.PrivateKeyToPEM()
	if err != nil {
		return nil, nil, errors.WrapError(err, 50002, "转换解绑私钥失败")
	}

	unbindPublicKeyPEM, err := unbindKeyPair.PublicKeyToPEM()
	if err != nil {
		return nil, nil, errors.WrapError(err, 50002, "转换解绑公钥失败")
	}

	// 创建授权数据
	now := time.Now()
	licenseData := LicenseData{
		MachineID:        bindFile.MachineID,
		IssuedAt:         now,
		ExpiresAt:        expiresAt,
		LicenseType:      "FULL",
		UnbindPrivateKey: unbindPrivateKeyPEM,
	}

	// 签名授权数据
	licenseDataBytes, err := json.Marshal(licenseData)
	if err != nil {
		return nil, nil, errors.WrapError(err, 50002, "序列化授权数据失败")
	}

	signature, err := s.rsaService.SignData(licenseDataBytes)
	if err != nil {
		return nil, nil, err
	}

	// 创建授权文件
	licenseFile := &LicenseFile{
		LicenseData: licenseData,
		Signature:   signature,
	}

	// 生成授权记录的唯一标识
	licenseKey := s.generateLicenseKey(bindFile.MachineID, now)

	// 创建数据库记录
	license := &models.License{
		AuthorizationID: auth.ID,
		LicenseKey:      licenseKey,
		MachineID:       bindFile.MachineID,
		Hostname:        bindFile.Hostname,
		UnbindPublicKey: unbindPublicKeyPEM,
		IssuedAt:        now,
		ExpiresAt:       expiresAt,
		Status:          models.LicenseStatusActive,
		ActivatedAt:     now,
	}

	return licenseFile, license, nil
}

// generateLicenseKey 生成授权记录的唯一标识
func (s *LicenseService) generateLicenseKey(machineID string, issuedAt time.Time) string {
	data := fmt.Sprintf("%s:%s:%s", machineID, issuedAt.Format(time.RFC3339), uuid.New().String())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
