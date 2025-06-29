package services

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/lyenrowe/LicenseCenter/internal/database"
	"github.com/lyenrowe/LicenseCenter/internal/models"
	"github.com/lyenrowe/LicenseCenter/pkg/errors"
	"gorm.io/gorm"
)

// 全局计数器，用于确保授权码唯一性
var authCodeCounter int64

// AuthorizationService 授权码管理服务
type AuthorizationService struct {
	db *gorm.DB
}

// NewAuthorizationService 创建授权码服务实例
func NewAuthorizationService() *AuthorizationService {
	return &AuthorizationService{
		db: database.GetDB(),
	}
}

// CreateAuthorizationRequest 创建授权码请求结构
type CreateAuthorizationRequest struct {
	CustomerName      string     `json:"customer_name" validate:"required,max=255"`
	AuthorizationCode string     `json:"authorization_code,omitempty" validate:"omitempty,max=255"`
	MaxSeats          int        `json:"max_seats" validate:"required,min=1"`
	DurationYears     *int       `json:"duration_years,omitempty" validate:"omitempty,min=1"`
	LatestExpiryDate  *time.Time `json:"latest_expiry_date,omitempty"`
}

// UpdateAuthorizationRequest 更新授权码请求结构
type UpdateAuthorizationRequest struct {
	CustomerName     string     `json:"customer_name,omitempty" validate:"omitempty,max=255"`
	MaxSeats         *int       `json:"max_seats,omitempty" validate:"omitempty,min=1"`
	DurationYears    *int       `json:"duration_years,omitempty" validate:"omitempty,min=1"`
	LatestExpiryDate *time.Time `json:"latest_expiry_date,omitempty"`
	Status           *int       `json:"status,omitempty" validate:"omitempty,oneof=0 1"`
}

// CreateAuthorization 创建新的授权码
func (s *AuthorizationService) CreateAuthorization(req *CreateAuthorizationRequest) (*models.Authorization, error) {
	// 如果没有提供授权码，自动生成
	if req.AuthorizationCode == "" {
		req.AuthorizationCode = s.generateAuthorizationCode()
	}

	// 检查授权码是否已存在
	var existing models.Authorization
	err := s.db.Where("authorization_code = ?", req.AuthorizationCode).First(&existing).Error
	if err == nil {
		return nil, errors.NewAppError(41008, "授权码已存在")
	}
	if err != gorm.ErrRecordNotFound {
		return nil, errors.WrapError(err, 50001, "检查授权码唯一性失败")
	}

	// 创建授权码
	auth := &models.Authorization{
		CustomerName:      req.CustomerName,
		AuthorizationCode: req.AuthorizationCode,
		MaxSeats:          req.MaxSeats,
		UsedSeats:         0,
		DurationYears:     req.DurationYears,
		LatestExpiryDate:  req.LatestExpiryDate,
		Status:            1, // 默认启用
	}

	err = s.db.Create(auth).Error
	if err != nil {
		return nil, errors.WrapError(err, 50001, "创建授权码失败")
	}

	return auth, nil
}

// GetAuthorizationByCode 根据授权码获取授权信息
func (s *AuthorizationService) GetAuthorizationByCode(code string) (*models.Authorization, error) {
	var auth models.Authorization
	err := s.db.Where("authorization_code = ?", code).First(&auth).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrAuthCodeNotFound
		}
		return nil, errors.WrapError(err, 50001, "获取授权码失败")
	}

	return &auth, nil
}

// GetAuthorizationByID 根据ID获取授权信息
func (s *AuthorizationService) GetAuthorizationByID(id uint) (*models.Authorization, error) {
	var auth models.Authorization
	err := s.db.Preload("Licenses").First(&auth, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrAuthCodeNotFound
		}
		return nil, errors.WrapError(err, 50001, "获取授权码失败")
	}

	return &auth, nil
}

// UpdateAuthorization 更新授权码
func (s *AuthorizationService) UpdateAuthorization(id uint, req *UpdateAuthorizationRequest) (*models.Authorization, error) {
	var auth models.Authorization
	err := s.db.First(&auth, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrAuthCodeNotFound
		}
		return nil, errors.WrapError(err, 50001, "获取授权码失败")
	}

	// 更新字段
	if req.CustomerName != "" {
		auth.CustomerName = req.CustomerName
	}
	if req.MaxSeats != nil {
		// 只能增加席位，不能减少
		if *req.MaxSeats < auth.UsedSeats {
			return nil, errors.NewAppError(41001, "最大席位数不能小于已使用席位数")
		}
		auth.MaxSeats = *req.MaxSeats
	}
	if req.DurationYears != nil {
		auth.DurationYears = req.DurationYears
	}
	if req.LatestExpiryDate != nil {
		auth.LatestExpiryDate = req.LatestExpiryDate
	}
	if req.Status != nil {
		auth.Status = *req.Status
	}

	err = s.db.Save(&auth).Error
	if err != nil {
		return nil, errors.WrapError(err, 50001, "更新授权码失败")
	}

	return &auth, nil
}

// ListAuthorizations 获取授权码列表
func (s *AuthorizationService) ListAuthorizations(page, limit int, search string, status *int) ([]models.Authorization, int64, error) {
	var auths []models.Authorization
	var total int64

	query := s.db.Model(&models.Authorization{})

	// 搜索条件
	if search != "" {
		query = query.Where("customer_name LIKE ? OR authorization_code LIKE ?",
			"%"+search+"%", "%"+search+"%")
	}

	// 状态筛选
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, errors.WrapError(err, 50001, "获取授权码总数失败")
	}

	// 分页查询
	offset := (page - 1) * limit
	err = query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&auths).Error
	if err != nil {
		return nil, 0, errors.WrapError(err, 50001, "获取授权码列表失败")
	}

	return auths, total, nil
}

// ValidateAuthorizationCode 验证授权码是否有效
func (s *AuthorizationService) ValidateAuthorizationCode(code string) (*models.Authorization, error) {
	auth, err := s.GetAuthorizationByCode(code)
	if err != nil {
		return nil, err
	}

	// 检查状态
	if !auth.IsActive() {
		return nil, errors.ErrAuthCodeDisabled
	}

	return auth, nil
}

// ConsumeSeats 消耗席位
func (s *AuthorizationService) ConsumeSeats(authID uint, count int) error {
	// 先检查可用席位
	var auth models.Authorization
	err := s.db.First(&auth, authID).Error
	if err != nil {
		return err
	}

	if !auth.HasAvailableSeats(count) {
		return errors.ErrInsufficientSeats
	}

	// 直接更新使用的席位数
	err = s.db.Model(&auth).Update("used_seats", gorm.Expr("used_seats + ?", count)).Error
	if err != nil {
		return err
	}

	return nil
}

// ReleaseSeats 释放席位
func (s *AuthorizationService) ReleaseSeats(authID uint, count int) error {
	// 使用原子更新，确保不会小于0
	err := s.db.Model(&models.Authorization{}).
		Where("id = ?", authID).
		Update("used_seats", gorm.Expr("CASE WHEN used_seats - ? < 0 THEN 0 ELSE used_seats - ? END", count, count)).Error

	return err
}

// DeleteAuthorization 删除授权码（软删除）
func (s *AuthorizationService) DeleteAuthorization(id uint) error {
	// 检查是否有活跃的授权设备
	var activeCount int64
	err := s.db.Model(&models.License{}).Where("authorization_id = ? AND status = ?",
		id, models.LicenseStatusActive).Count(&activeCount).Error
	if err != nil {
		return errors.WrapError(err, 50001, "检查活跃授权失败")
	}

	if activeCount > 0 {
		return errors.NewAppError(41001, "存在活跃授权设备，无法删除")
	}

	// 软删除
	err = s.db.Model(&models.Authorization{}).Where("id = ?", id).Update("status", 0).Error
	if err != nil {
		return errors.WrapError(err, 50001, "删除授权码失败")
	}

	return nil
}

// generateAuthorizationCode 生成授权码
func (s *AuthorizationService) generateAuthorizationCode() string {
	// 生成20位授权码，格式：ABCD-EFGH-IJKL-MNOP-QRST（4个字符一组，分成5组）
	// 使用UUID + 原子计数器 + 时间戳确保唯一性

	// 原子递增计数器
	counter := atomic.AddInt64(&authCodeCounter, 1)

	// 获取当前时间戳（纳秒）
	timestamp := time.Now().UnixNano()

	// 生成UUID
	id := uuid.New().String()
	// 移除UUID中的连字符，只保留字母和数字
	cleanID := strings.ReplaceAll(id, "-", "")

	// 组合: UUID前12位 + 4位计数器 + 4位时间戳
	counterHex := fmt.Sprintf("%04X", counter&0xFFFF)
	timestampHex := fmt.Sprintf("%04X", timestamp&0xFFFF)
	code := strings.ToUpper(cleanID[:12] + counterHex + timestampHex)

	// 格式化为 XXXX-XXXX-XXXX-XXXX-XXXX 格式（20位字符，每段4位，共5段）
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		code[0:4],
		code[4:8],
		code[8:12],
		code[12:16],
		code[16:20])
}

// GetStatistics 获取授权码统计信息
func (s *AuthorizationService) GetStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总授权码数
	var totalAuths int64
	err := s.db.Model(&models.Authorization{}).Count(&totalAuths).Error
	if err != nil {
		return nil, errors.WrapError(err, 50001, "获取授权码总数失败")
	}

	// 活跃授权码数
	var activeAuths int64
	err = s.db.Model(&models.Authorization{}).Where("status = ?", 1).Count(&activeAuths).Error
	if err != nil {
		return nil, errors.WrapError(err, 50001, "获取活跃授权码数失败")
	}

	// 总席位数和已用席位数
	var totalSeats, usedSeats int64
	err = s.db.Model(&models.Authorization{}).Where("status = ?", 1).
		Select("COALESCE(SUM(max_seats), 0), COALESCE(SUM(used_seats), 0)").
		Row().Scan(&totalSeats, &usedSeats)
	if err != nil {
		return nil, errors.WrapError(err, 50001, "获取席位统计失败")
	}

	// 活跃设备数
	var activeDevices int64
	err = s.db.Model(&models.License{}).Where("status = ?", models.LicenseStatusActive).Count(&activeDevices).Error
	if err != nil {
		return nil, errors.WrapError(err, 50001, "获取活跃设备数失败")
	}

	stats["total_authorizations"] = totalAuths
	stats["active_authorizations"] = activeAuths
	stats["total_seats"] = totalSeats
	stats["used_seats"] = usedSeats
	stats["active_devices"] = activeDevices

	return stats, nil
}
