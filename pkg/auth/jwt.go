package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/licensecenter/licensecenter/internal/config"
)

// Claims JWT声明结构
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	UserType string `json:"user_type"` // "customer" 或 "admin"
	jwt.RegisteredClaims
}

// JWTManager JWT管理器
type JWTManager struct {
	secretKey []byte
}

// NewJWTManager 创建JWT管理器
func NewJWTManager() *JWTManager {
	return &JWTManager{
		secretKey: []byte(config.AppConfig.Security.JWTSecret),
	}
}

// GenerateToken 生成JWT令牌
func (j *JWTManager) GenerateToken(userID uint, username, userType string, duration time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		Username: username,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "LicenseCenter",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateToken 验证JWT令牌
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("解析令牌失败: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("无效的令牌")
}

// RefreshToken 刷新JWT令牌
func (j *JWTManager) RefreshToken(tokenString string, duration time.Duration) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// 检查令牌是否即将过期（剩余时间少于1小时）
	if time.Until(claims.ExpiresAt.Time) > time.Hour {
		return "", fmt.Errorf("令牌尚未到期，无需刷新")
	}

	return j.GenerateToken(claims.UserID, claims.Username, claims.UserType, duration)
}

// GetDefaultDuration 获取默认令牌有效期
func GetDefaultDuration(userType string) time.Duration {
	switch userType {
	case "admin":
		return time.Duration(config.AppConfig.Security.AdminSessionTimeout) * time.Second
	case "customer":
		return time.Duration(config.AppConfig.Security.SessionTimeout) * time.Second
	default:
		return time.Hour // 默认1小时
	}
}
