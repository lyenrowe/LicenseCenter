package models

import (
	"time"
)

// RSAKey RSA密钥表模型
type RSAKey struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	PrivateKey string    `gorm:"not null;type:text" json:"-"`          // 私钥，不在JSON中返回
	PublicKey  string    `gorm:"not null;type:text" json:"public_key"` // 公钥
	IsActive   bool      `gorm:"default:true" json:"is_active"`        // 是否为当前活跃密钥
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName 指定表名
func (RSAKey) TableName() string {
	return "rsa_keys"
}

// SystemConfig 系统配置表模型
type SystemConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ConfigKey   string    `gorm:"unique;not null;size:100" json:"config_key"`
	ConfigValue string    `gorm:"type:text" json:"config_value"`
	Description string    `gorm:"type:text" json:"description"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName 指定表名
func (SystemConfig) TableName() string {
	return "system_config"
}

// 系统配置键常量
const (
	ConfigSessionTimeout      = "session_timeout"
	ConfigMaxBindFiles        = "max_bind_files_per_request"
	ConfigCaptchaEnabled      = "captcha_enabled"
	ConfigBackupRetentionDays = "backup_retention_days"
	ConfigMaintenanceMode     = "maintenance_mode"
	ConfigSystemVersion       = "system_version"
)
