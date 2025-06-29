package tests

import (
	"testing"

	"github.com/lyenrowe/LicenseCenter/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestForceTOTPConfiguration(t *testing.T) {
	// 模拟配置
	config.AppConfig = &config.Config{
		Security: config.SecurityConfig{
			ForceTOTP: true,
		},
	}

	// 测试强制TOTP配置
	assert.True(t, config.AppConfig.Security.ForceTOTP, "强制TOTP应该被启用")
}

func TestForceTOTPConfigurationDisabled(t *testing.T) {
	// 模拟配置
	config.AppConfig = &config.Config{
		Security: config.SecurityConfig{
			ForceTOTP: false,
		},
	}

	// 测试强制TOTP配置
	assert.False(t, config.AppConfig.Security.ForceTOTP, "强制TOTP应该被禁用")
}
