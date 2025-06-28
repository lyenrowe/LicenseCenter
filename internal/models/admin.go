package models

import (
	"time"
)

// AdminUser 管理员用户表模型
type AdminUser struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Username     string     `gorm:"unique;not null;size:50" json:"username" validate:"required,max=50"`
	PasswordHash string     `gorm:"not null;size:255" json:"-"` // 不在JSON中返回密码
	TOTPSecret   string     `gorm:"size:32" json:"-"`           // TOTP密钥，不在JSON中返回
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	LastLogin    *time.Time `json:"last_login"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// 关联关系
	Logs []AdminLog `gorm:"foreignKey:AdminID" json:"logs,omitempty"`
}

// TableName 指定表名
func (AdminUser) TableName() string {
	return "admin_users"
}

// AdminLog 管理员操作日志表模型
type AdminLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	AdminID    *uint     `json:"admin_id"`                        // 可为空，系统操作时为空
	Action     string    `gorm:"not null;size:100" json:"action"` // 操作类型
	TargetType string    `gorm:"size:50" json:"target_type"`      // 操作对象类型
	TargetID   string    `gorm:"size:100" json:"target_id"`       // 操作对象ID
	Details    string    `gorm:"type:text" json:"details"`        // 操作详情JSON
	IPAddress  string    `gorm:"size:45" json:"ip_address"`       // IP地址
	CreatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`

	// 关联关系
	Admin *AdminUser `gorm:"foreignKey:AdminID" json:"admin,omitempty"`
}

// TableName 指定表名
func (AdminLog) TableName() string {
	return "admin_logs"
}

// LogAction 日志操作类型常量
const (
	// 认证相关
	LogActionLogin       = "login"
	LogActionLogout      = "logout"
	LogActionLoginFailed = "login_failed"

	// 授权码管理
	LogActionCreateAuth  = "create_authorization"
	LogActionUpdateAuth  = "update_authorization"
	LogActionDisableAuth = "disable_authorization"
	LogActionEnableAuth  = "enable_authorization"

	// 设备管理
	LogActionForceUnbind  = "force_unbind_device"
	LogActionViewCustomer = "view_customer_details"

	// 系统管理
	LogActionBackup       = "system_backup"
	LogActionKeyRotation  = "key_rotation"
	LogActionConfigUpdate = "config_update"
)

// LogTargetType 日志目标类型常量
const (
	LogTargetAuth    = "authorization"
	LogTargetLicense = "license"
	LogTargetAdmin   = "admin"
	LogTargetSystem  = "system"
)
