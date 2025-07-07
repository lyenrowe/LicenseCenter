package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/lyenrowe/LicenseCenter/internal/services"
	"github.com/lyenrowe/LicenseCenter/pkg/auth"
	"github.com/lyenrowe/LicenseCenter/pkg/errors"
)

// CustomerHandler 客户端处理器
type CustomerHandler struct {
	authService    *services.AuthorizationService
	licenseService *services.LicenseService
	captchaService *services.CaptchaService
	validator      *validator.Validate
}

// NewCustomerHandler 创建客户端处理器
func NewCustomerHandler() *CustomerHandler {
	return &CustomerHandler{
		authService:    services.NewAuthorizationService(),
		licenseService: services.NewLicenseService(),
		captchaService: services.NewCaptchaService(),
		validator:      validator.New(),
	}
}

// Login 客户端登录
func (h *CustomerHandler) Login(c *gin.Context) {
	var req struct {
		AuthorizationCode string `json:"authorization_code" validate:"required"`
		CaptchaToken      string `json:"captcha_token" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
			"code":  40000,
		})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数验证失败",
			"code":  40000,
		})
		return
	}

	// 验证hCaptcha令牌
	clientIP := c.ClientIP()
	if err := h.captchaService.VerifyToken(req.CaptchaToken, clientIP); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "验证码验证失败",
				"code":  50000,
			})
		}
		return
	}

	// 验证授权码（压缩首尾空格）
	authorization, err := h.authService.ValidateAuthorizationCode(strings.TrimSpace(req.AuthorizationCode))
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "验证授权码失败",
				"code":  50000,
			})
		}
		return
	}

	// 生成JWT令牌 (使用授权码ID作为用户ID)
	jwtManager := auth.NewJWTManager()
	duration := auth.GetDefaultDuration("customer")
	token, err := jwtManager.GenerateToken(authorization.ID, authorization.AuthorizationCode, "customer", duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "生成令牌失败",
			"code":  50000,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_token": token,
		"expires_in":    int(duration.Seconds()),
		"customer_info": gin.H{
			"id":                 authorization.ID,
			"customer_name":      authorization.CustomerName,
			"authorization_code": authorization.AuthorizationCode,
			"max_seats":          authorization.MaxSeats,
			"used_seats":         authorization.UsedSeats,
		},
	})
}

// GetCaptchaConfig 获取验证码配置
func (h *CustomerHandler) GetCaptchaConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"enabled":  h.captchaService.IsEnabled(),
		"site_key": h.captchaService.GetSiteKey(),
	})
}

// Logout 客户端登出
func (h *CustomerHandler) Logout(c *gin.Context) {
	// 客户端登出不需要记录到数据库，只需要返回成功响应
	// 客户端会清除本地token
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "注销成功",
	})
}

// GetDashboard 获取客户端控制台数据
func (h *CustomerHandler) GetDashboard(c *gin.Context) {
	// 从JWT中获取授权码ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未认证",
			"code":  40100,
		})
		return
	}

	authID := userID.(uint)

	// 获取授权码信息
	authorization, err := h.authService.GetAuthorizationByID(authID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取授权信息失败",
			"code":  50000,
		})
		return
	}

	// 获取设备列表 (使用授权码而不是ID)
	licenses, err := h.licenseService.GetLicensesByAuth(authorization.AuthorizationCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取设备列表失败",
			"code":  50000,
		})
		return
	}

	// 分离活跃和历史设备
	var activeDevices []interface{}
	var historicalDevices []interface{}

	for _, license := range licenses {
		// 安全处理machine_id显示
		displayMachineID := license.MachineID
		if len(displayMachineID) > 12 {
			displayMachineID = displayMachineID[:12] + "..."
		}

		deviceInfo := gin.H{
			"id":         license.ID,
			"hostname":   license.Hostname,
			"machine_id": displayMachineID,
			"issued_at":  license.IssuedAt,
			"expires_at": license.ExpiresAt,
			"status":     license.Status,
		}

		if license.Status == "active" {
			activeDevices = append(activeDevices, deviceInfo)
		} else {
			deviceInfo["unbound_at"] = license.UnboundAt
			historicalDevices = append(historicalDevices, deviceInfo)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"authorization": gin.H{
			"customer_name":      authorization.CustomerName,
			"authorization_code": authorization.AuthorizationCode,
			"max_seats":          authorization.MaxSeats,
			"used_seats":         authorization.UsedSeats,
			"available_seats":    authorization.GetAvailableSeats(),
		},
		"devices": gin.H{
			"active":     activeDevices,
			"historical": historicalDevices,
		},
	})
}
