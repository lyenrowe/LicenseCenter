package models

import (
	"time"
)

// License 已激活设备表模型
type License struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	AuthorizationID uint       `gorm:"not null" json:"authorization_id"`
	LicenseKey      string     `gorm:"unique;not null" json:"license_key"` // .license文件内容的哈希或唯一标识
	MachineID       string     `gorm:"not null;size:255" json:"machine_id"`
	Hostname        string     `gorm:"size:255" json:"hostname"`
	UnbindPublicKey string     `gorm:"type:text" json:"unbind_public_key"` // 用于验证解绑凭证的一次性公钥
	IssuedAt        time.Time  `gorm:"not null" json:"issued_at"`
	ExpiresAt       time.Time  `json:"expires_at"`
	Status          string     `gorm:"not null;size:50" json:"status"` // 'active', 'unbound', 'force_unbound'
	ActivatedAt     time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"activated_at"`
	UnboundAt       *time.Time `json:"unbound_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// 关联关系
	Authorization Authorization `gorm:"foreignKey:AuthorizationID" json:"authorization,omitempty"`
}

// TableName 指定表名
func (License) TableName() string {
	return "licenses"
}

// LicenseStatus 授权状态常量
const (
	LicenseStatusActive       = "active"        // 激活状态
	LicenseStatusUnbound      = "unbound"       // 客户解绑
	LicenseStatusForceUnbound = "force_unbound" // 管理员强制解绑
)

// IsActive 检查授权是否有效
func (l *License) IsActive() bool {
	return l.Status == LicenseStatusActive && time.Now().Before(l.ExpiresAt)
}

// IsExpired 检查授权是否已过期
func (l *License) IsExpired() bool {
	return time.Now().After(l.ExpiresAt)
}

// CanUnbind 检查是否可以解绑
func (l *License) CanUnbind() bool {
	return l.Status == LicenseStatusActive
}

// Unbind 执行解绑操作
func (l *License) Unbind(isForced bool) {
	now := time.Now()
	l.UnboundAt = &now

	if isForced {
		l.Status = LicenseStatusForceUnbound
	} else {
		l.Status = LicenseStatusUnbound
	}
}
