package tests

import (
	"testing"

	"github.com/lyenrowe/LicenseCenter/internal/config"
	"github.com/lyenrowe/LicenseCenter/internal/services"
)

func TestCaptchaService_VerifyToken(t *testing.T) {
	// 初始化配置
	config.AppConfig = &config.Config{
		Captcha: config.CaptchaConfig{
			Enabled:   true,
			SiteKey:   "test-site-key",
			SecretKey: "test-secret-key",
		},
		Server: config.ServerConfig{
			Mode: "debug",
		},
	}

	// 保存原始配置
	originalConfig := config.AppConfig.Captcha
	originalServerMode := config.AppConfig.Server.Mode
	defer func() {
		config.AppConfig.Captcha = originalConfig
		config.AppConfig.Server.Mode = originalServerMode
	}()

	service := services.NewCaptchaService()

	t.Run("验证码功能禁用时应跳过验证", func(t *testing.T) {
		// 禁用验证码
		config.AppConfig.Captcha.Enabled = false

		err := service.VerifyToken("any_token", "127.0.0.1")
		if err != nil {
			t.Errorf("验证码禁用时应该跳过验证，但收到错误: %v", err)
		}
	})

	t.Run("开发环境应允许降级验证码", func(t *testing.T) {
		// 启用验证码
		config.AppConfig.Captcha.Enabled = true
		// 设置为开发模式
		originalMode := config.AppConfig.Server.Mode
		config.AppConfig.Server.Mode = "debug"
		defer func() {
			config.AppConfig.Server.Mode = originalMode
		}()

		err := service.VerifyToken("fallback_captcha_token_12345", "127.0.0.1")
		if err != nil {
			t.Errorf("开发环境应允许降级验证码，但收到错误: %v", err)
		}
	})

	t.Run("生产环境应拒绝降级验证码", func(t *testing.T) {
		// 启用验证码
		config.AppConfig.Captcha.Enabled = true
		// 设置为生产模式
		originalMode := config.AppConfig.Server.Mode
		config.AppConfig.Server.Mode = "release"
		defer func() {
			config.AppConfig.Server.Mode = originalMode
		}()

		err := service.VerifyToken("fallback_captcha_token_12345", "127.0.0.1")
		if err == nil {
			t.Error("生产环境应拒绝降级验证码，但验证通过了")
		}
	})

	t.Run("空token应返回错误", func(t *testing.T) {
		// 启用验证码
		config.AppConfig.Captcha.Enabled = true

		err := service.VerifyToken("", "127.0.0.1")
		if err == nil {
			t.Error("空token应返回错误，但验证通过了")
		}
	})

	t.Run("未配置密钥应返回错误", func(t *testing.T) {
		// 启用验证码但不设置密钥
		config.AppConfig.Captcha.Enabled = true
		config.AppConfig.Captcha.SecretKey = ""

		err := service.VerifyToken("valid_token", "127.0.0.1")
		if err == nil {
			t.Error("未配置密钥应返回错误，但验证通过了")
		}
	})
}

func TestCaptchaService_Configuration(t *testing.T) {
	// 初始化配置
	config.AppConfig = &config.Config{
		Captcha: config.CaptchaConfig{
			Enabled:   true,
			SiteKey:   "test-site-key",
			SecretKey: "test-secret-key",
		},
	}

	// 保存原始配置
	originalConfig := config.AppConfig.Captcha
	defer func() {
		config.AppConfig.Captcha = originalConfig
	}()

	service := services.NewCaptchaService()

	t.Run("IsEnabled应返回正确的状态", func(t *testing.T) {
		config.AppConfig.Captcha.Enabled = true
		if !service.IsEnabled() {
			t.Error("验证码启用时IsEnabled应返回true")
		}

		config.AppConfig.Captcha.Enabled = false
		if service.IsEnabled() {
			t.Error("验证码禁用时IsEnabled应返回false")
		}
	})

	t.Run("GetSiteKey应返回配置的站点密钥", func(t *testing.T) {
		testSiteKey := "test-site-key-12345"
		config.AppConfig.Captcha.SiteKey = testSiteKey

		siteKey := service.GetSiteKey()
		if siteKey != testSiteKey {
			t.Errorf("GetSiteKey应返回 %s，但返回了 %s", testSiteKey, siteKey)
		}
	})
}
