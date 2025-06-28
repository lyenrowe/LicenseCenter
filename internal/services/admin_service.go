package services

import (
	"encoding/json"
	"time"

	"github.com/licensecenter/licensecenter/internal/database"
	"github.com/licensecenter/licensecenter/internal/models"
	"github.com/licensecenter/licensecenter/pkg/errors"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AdminService 管理员服务
type AdminService struct {
	db *gorm.DB
}

// NewAdminService 创建管理员服务实例
func NewAdminService() *AdminService {
	return &AdminService{
		db: database.GetDB(),
	}
}

// AdminLoginRequest 管理员登录请求
type AdminLoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	TOTPCode string `json:"totp_code,omitempty"`
}

// CreateAdminRequest 创建管理员请求
type CreateAdminRequest struct {
	Username string `json:"username" validate:"required,max=50"`
	Password string `json:"password" validate:"required,min=6"`
}

// UpdateAdminRequest 更新管理员请求
type UpdateAdminRequest struct {
	Password string `json:"password,omitempty" validate:"omitempty,min=6"`
	IsActive *bool  `json:"is_active,omitempty"`
}

// AdminLogin 管理员登录
func (s *AdminService) AdminLogin(req *AdminLoginRequest) (*models.AdminUser, error) {
	var admin models.AdminUser
	err := s.db.Where("username = ?", req.Username).First(&admin).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrInvalidCredentials
		}
		return nil, errors.WrapError(err, 50001, "查询管理员失败")
	}

	// 检查账户状态
	if !admin.IsActive {
		return nil, errors.NewAppError(40006, "账户已被禁用")
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	// 验证TOTP（如果启用）
	if admin.TOTPSecret != "" {
		if req.TOTPCode == "" {
			return nil, errors.ErrInvalidTOTP
		}

		valid := totp.Validate(req.TOTPCode, admin.TOTPSecret)
		if !valid {
			return nil, errors.ErrInvalidTOTP
		}
	}

	// 更新最后登录时间
	now := time.Now()
	admin.LastLogin = &now
	s.db.Save(&admin)

	return &admin, nil
}

// CreateAdmin 创建管理员
func (s *AdminService) CreateAdmin(req *CreateAdminRequest) (*models.AdminUser, error) {
	// 检查用户名是否已存在
	var existing models.AdminUser
	err := s.db.Where("username = ?", req.Username).First(&existing).Error
	if err == nil {
		return nil, errors.NewAppError(40002, "用户名已存在")
	}
	if err != gorm.ErrRecordNotFound {
		return nil, errors.WrapError(err, 50001, "检查用户名唯一性失败")
	}

	// 生成密码哈希
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.WrapError(err, 50002, "生成密码哈希失败")
	}

	// 创建管理员
	admin := &models.AdminUser{
		Username:     req.Username,
		PasswordHash: string(passwordHash),
		IsActive:     true,
	}

	err = s.db.Create(admin).Error
	if err != nil {
		return nil, errors.WrapError(err, 50001, "创建管理员失败")
	}

	return admin, nil
}

// GetAdminByID 根据ID获取管理员
func (s *AdminService) GetAdminByID(id uint) (*models.AdminUser, error) {
	var admin models.AdminUser
	err := s.db.First(&admin, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewAppError(40006, "管理员不存在")
		}
		return nil, errors.WrapError(err, 50001, "获取管理员失败")
	}

	return &admin, nil
}

// UpdateAdmin 更新管理员
func (s *AdminService) UpdateAdmin(id uint, req *UpdateAdminRequest) (*models.AdminUser, error) {
	var admin models.AdminUser
	err := s.db.First(&admin, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewAppError(40006, "管理员不存在")
		}
		return nil, errors.WrapError(err, 50001, "获取管理员失败")
	}

	// 更新密码
	if req.Password != "" {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.WrapError(err, 50002, "生成密码哈希失败")
		}
		admin.PasswordHash = string(passwordHash)
	}

	// 更新状态
	if req.IsActive != nil {
		admin.IsActive = *req.IsActive
	}

	err = s.db.Save(&admin).Error
	if err != nil {
		return nil, errors.WrapError(err, 50001, "更新管理员失败")
	}

	return &admin, nil
}

// ListAdmins 获取管理员列表
func (s *AdminService) ListAdmins(page, limit int) ([]models.AdminUser, int64, error) {
	var admins []models.AdminUser
	var total int64

	// 获取总数
	err := s.db.Model(&models.AdminUser{}).Count(&total).Error
	if err != nil {
		return nil, 0, errors.WrapError(err, 50001, "获取管理员总数失败")
	}

	// 分页查询
	offset := (page - 1) * limit
	err = s.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&admins).Error
	if err != nil {
		return nil, 0, errors.WrapError(err, 50001, "获取管理员列表失败")
	}

	return admins, total, nil
}

// EnableTOTP 启用TOTP双因素认证
func (s *AdminService) EnableTOTP(adminID uint) (string, error) {
	var admin models.AdminUser
	err := s.db.First(&admin, adminID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", errors.NewAppError(40006, "管理员不存在")
		}
		return "", errors.WrapError(err, 50001, "获取管理员失败")
	}

	// 生成TOTP密钥
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "LicenseCenter",
		AccountName: admin.Username,
	})
	if err != nil {
		return "", errors.WrapError(err, 50002, "生成TOTP密钥失败")
	}

	// 保存密钥
	admin.TOTPSecret = key.Secret()
	err = s.db.Save(&admin).Error
	if err != nil {
		return "", errors.WrapError(err, 50001, "保存TOTP密钥失败")
	}

	return key.URL(), nil
}

// DisableTOTP 禁用TOTP双因素认证
func (s *AdminService) DisableTOTP(adminID uint) error {
	return s.db.Model(&models.AdminUser{}).Where("id = ?", adminID).Update("totp_secret", "").Error
}

// LogAction 记录管理员操作日志
func (s *AdminService) LogAction(adminID *uint, action, targetType, targetID, ipAddress string, details interface{}) error {
	var detailsJSON string
	if details != nil {
		detailsBytes, err := json.Marshal(details)
		if err != nil {
			return errors.WrapError(err, 50002, "序列化操作详情失败")
		}
		detailsJSON = string(detailsBytes)
	}

	log := &models.AdminLog{
		AdminID:    adminID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Details:    detailsJSON,
		IPAddress:  ipAddress,
	}

	err := s.db.Create(log).Error
	if err != nil {
		return errors.WrapError(err, 50001, "记录操作日志失败")
	}

	return nil
}

// GetLogs 获取操作日志
func (s *AdminService) GetLogs(page, limit int, action, targetType string, adminID *uint) ([]models.AdminLog, int64, error) {
	var logs []models.AdminLog
	var total int64

	query := s.db.Model(&models.AdminLog{}).Preload("Admin")

	// 筛选条件
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}
	if adminID != nil {
		query = query.Where("admin_id = ?", *adminID)
	}

	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, errors.WrapError(err, 50001, "获取日志总数失败")
	}

	// 分页查询
	offset := (page - 1) * limit
	err = query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error
	if err != nil {
		return nil, 0, errors.WrapError(err, 50001, "获取操作日志失败")
	}

	return logs, total, nil
}

// DeleteAdmin 删除管理员（软删除）
func (s *AdminService) DeleteAdmin(id uint) error {
	// 检查是否是最后一个活跃管理员
	var activeCount int64
	err := s.db.Model(&models.AdminUser{}).Where("is_active = ? AND id != ?", true, id).Count(&activeCount).Error
	if err != nil {
		return errors.WrapError(err, 50001, "检查活跃管理员数量失败")
	}

	if activeCount == 0 {
		return errors.NewAppError(40006, "不能删除最后一个活跃管理员")
	}

	// 禁用管理员
	err = s.db.Model(&models.AdminUser{}).Where("id = ?", id).Update("is_active", false).Error
	if err != nil {
		return errors.WrapError(err, 50001, "删除管理员失败")
	}

	return nil
}

// GetDashboardStats 获取管理员控制台统计信息
func (s *AdminService) GetDashboardStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取授权统计
	authService := NewAuthorizationService()
	authStats, err := authService.GetStatistics()
	if err != nil {
		return nil, err
	}

	// 获取今日新增授权码
	var todayAuths int64
	today := time.Now().Truncate(24 * time.Hour)
	err = s.db.Model(&models.Authorization{}).Where("created_at >= ?", today).Count(&todayAuths).Error
	if err != nil {
		return nil, errors.WrapError(err, 50001, "获取今日新增授权码失败")
	}

	// 获取今日新增设备
	var todayDevices int64
	err = s.db.Model(&models.License{}).Where("activated_at >= ?", today).Count(&todayDevices).Error
	if err != nil {
		return nil, errors.WrapError(err, 50001, "获取今日新增设备失败")
	}

	// 获取即将过期的授权（30天内）
	expiringSoon := time.Now().AddDate(0, 0, 30)
	var expiringLicenses int64
	err = s.db.Model(&models.License{}).Where("status = ? AND expires_at <= ?",
		models.LicenseStatusActive, expiringSoon).Count(&expiringLicenses).Error
	if err != nil {
		return nil, errors.WrapError(err, 50001, "获取即将过期授权失败")
	}

	// 合并统计信息
	for k, v := range authStats {
		stats[k] = v
	}
	stats["today_new_authorizations"] = todayAuths
	stats["today_new_devices"] = todayDevices
	stats["expiring_licenses"] = expiringLicenses

	return stats, nil
}
