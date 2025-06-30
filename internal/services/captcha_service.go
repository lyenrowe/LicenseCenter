package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lyenrowe/LicenseCenter/internal/config"
	"github.com/lyenrowe/LicenseCenter/pkg/errors"
	"github.com/lyenrowe/LicenseCenter/pkg/logger"
	"go.uber.org/zap"
)

// CaptchaService 验证码服务
type CaptchaService struct {
	config *config.CaptchaConfig
	client *http.Client
}

// NewCaptchaService 创建验证码服务实例
func NewCaptchaService() *CaptchaService {
	return &CaptchaService{
		config: &config.AppConfig.Captcha,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// HCaptchaResponse hCaptcha验证响应结构
type HCaptchaResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts,omitempty"`
	Hostname    string   `json:"hostname,omitempty"`
	Credit      bool     `json:"credit,omitempty"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
}

// VerifyToken 验证hCaptcha令牌
func (s *CaptchaService) VerifyToken(token string, remoteIP string) error {
	// 如果未启用验证码，直接返回成功
	if !s.config.Enabled {
		logger.GetLogger().Info("验证码功能已禁用，跳过验证")
		return nil
	}

	// 检查是否为开发模式的降级token
	if strings.HasPrefix(token, "fallback_captcha_token_") {
		logger.GetLogger().Warn("检测到降级验证码令牌", zap.String("token_prefix", "fallback_captcha_token_"))
		// 在开发环境允许降级验证码，生产环境应该拒绝
		if config.AppConfig.Server.Mode == "debug" {
			return nil
		} else {
			return errors.NewAppError(40020, "生产环境不允许使用降级验证码")
		}
	}

	// 验证参数
	if token == "" {
		return errors.NewAppError(40021, "验证码令牌不能为空")
	}

	if s.config.SecretKey == "" {
		logger.GetLogger().Error("hCaptcha密钥未配置")
		return errors.NewAppError(50010, "验证码服务配置错误")
	}

	// 准备请求数据
	data := url.Values{}
	data.Set("secret", s.config.SecretKey)
	data.Set("response", token)
	if remoteIP != "" {
		data.Set("remoteip", remoteIP)
	}

	// 发送验证请求
	resp, err := s.client.PostForm("https://hcaptcha.com/siteverify", data)
	if err != nil {
		logger.GetLogger().Error("hCaptcha验证请求失败", zap.Error(err))
		return errors.NewAppError(50011, "验证码验证请求失败")
	}
	defer resp.Body.Close()

	// 解析响应
	var hcaptchaResp HCaptchaResponse
	if err := json.NewDecoder(resp.Body).Decode(&hcaptchaResp); err != nil {
		logger.GetLogger().Error("解析hCaptcha响应失败", zap.Error(err))
		return errors.NewAppError(50012, "验证码响应解析失败")
	}

	// 检查验证结果
	if !hcaptchaResp.Success {
		logger.GetLogger().Warn("hCaptcha验证失败",
			zap.Strings("error_codes", hcaptchaResp.ErrorCodes),
			zap.String("token", token[:min(len(token), 20)]+"..."))

		// 根据错误代码返回更具体的错误信息
		if len(hcaptchaResp.ErrorCodes) > 0 {
			switch hcaptchaResp.ErrorCodes[0] {
			case "missing-input-secret":
				return errors.NewAppError(50013, "验证码服务密钥缺失")
			case "invalid-input-secret":
				return errors.NewAppError(50014, "验证码服务密钥无效")
			case "missing-input-response":
				return errors.NewAppError(40022, "验证码响应缺失")
			case "invalid-input-response":
				return errors.NewAppError(40023, "验证码响应无效")
			case "timeout-or-duplicate":
				return errors.NewAppError(40024, "验证码已过期或重复使用")
			default:
				return errors.NewAppError(40025, fmt.Sprintf("验证码验证失败: %s", hcaptchaResp.ErrorCodes[0]))
			}
		}

		return errors.NewAppError(40026, "人机验证失败")
	}

	logger.GetLogger().Info("hCaptcha验证成功",
		zap.String("hostname", hcaptchaResp.Hostname),
		zap.String("challenge_ts", hcaptchaResp.ChallengeTS))

	return nil
}

// IsEnabled 检查验证码是否启用
func (s *CaptchaService) IsEnabled() bool {
	return s.config.Enabled
}

// GetSiteKey 获取站点密钥（用于前端显示）
func (s *CaptchaService) GetSiteKey() string {
	return s.config.SiteKey
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
