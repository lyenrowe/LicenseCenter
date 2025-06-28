package models

import (
	"time"
)

// Authorization 授权码表模型
type Authorization struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	CustomerName      string     `gorm:"not null;size:255" json:"customer_name" validate:"required,max=255"`
	AuthorizationCode string     `gorm:"unique;not null;size:255" json:"authorization_code" validate:"required,max=255"`
	MaxSeats          int        `gorm:"not null" json:"max_seats" validate:"required,min=1"`
	UsedSeats         int        `gorm:"default:0" json:"used_seats"`
	DurationYears     *int       `json:"duration_years" validate:"omitempty,min=1"`
	LatestExpiryDate  *time.Time `json:"latest_expiry_date"`
	Status            int        `gorm:"default:1" json:"status"` // 1:有效 0:禁用
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`

	// 关联关系
	Licenses []License `gorm:"foreignKey:AuthorizationID" json:"licenses,omitempty"`
}

// TableName 指定表名
func (Authorization) TableName() string {
	return "authorizations"
}

// IsActive 检查授权码是否有效
func (a *Authorization) IsActive() bool {
	return a.Status == 1
}

// HasAvailableSeats 检查是否有可用席位
func (a *Authorization) HasAvailableSeats(required int) bool {
	return a.MaxSeats-a.UsedSeats >= required
}

// GetAvailableSeats 获取可用席位数
func (a *Authorization) GetAvailableSeats() int {
	return a.MaxSeats - a.UsedSeats
}

// CalculateExpiryDate 计算授权到期时间
func (a *Authorization) CalculateExpiryDate() time.Time {
	now := time.Now()

	// 如果设置了最晚到期时间
	if a.LatestExpiryDate != nil {
		// 如果设置了授权年限，取两者较早的时间
		if a.DurationYears != nil {
			durationExpiry := now.AddDate(*a.DurationYears, 0, 0)
			if a.LatestExpiryDate.Before(durationExpiry) {
				return *a.LatestExpiryDate
			}
			return durationExpiry
		}
		return *a.LatestExpiryDate
	}

	// 如果只设置了授权年限
	if a.DurationYears != nil {
		return now.AddDate(*a.DurationYears, 0, 0)
	}

	// 默认1年
	return now.AddDate(1, 0, 0)
}
