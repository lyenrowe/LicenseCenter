package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/lyenrowe/LicenseCenter/internal/services"
	"github.com/lyenrowe/LicenseCenter/pkg/auth"
	"github.com/lyenrowe/LicenseCenter/pkg/errors"
	"github.com/lyenrowe/LicenseCenter/pkg/logger"
	"go.uber.org/zap"
)

// AdminHandler 管理员处理器
type AdminHandler struct {
	adminService *services.AdminService
	authService  *services.AuthorizationService
	rsaService   *services.RSAService
	validator    *validator.Validate
}

// NewAdminHandler 创建管理员处理器
func NewAdminHandler() *AdminHandler {
	return &AdminHandler{
		adminService: services.NewAdminService(),
		authService:  services.NewAuthorizationService(),
		rsaService:   services.NewRSAService(),
		validator:    validator.New(),
	}
}

// Login 管理员登录
func (h *AdminHandler) Login(c *gin.Context) {
	var req services.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
			"code":  40000,
		})
		return
	}

	// 验证请求参数
	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数验证失败",
			"code":  40000,
		})
		return
	}

	// 执行登录
	admin, err := h.adminService.AdminLogin(&req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "登录失败",
				"code":  50000,
			})
		}
		return
	}

	// 生成JWT令牌
	jwtManager := auth.NewJWTManager()
	duration := auth.GetDefaultDuration("admin")
	token, err := jwtManager.GenerateToken(admin.ID, admin.Username, "admin", duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "生成令牌失败",
			"code":  50000,
		})
		return
	}

	// 记录登录日志
	go h.adminService.LogAction(&admin.ID, "login", "admin", strconv.FormatUint(uint64(admin.ID), 10), c.ClientIP(), gin.H{
		"username": admin.Username,
	})

	c.JSON(http.StatusOK, gin.H{
		"token":      token,
		"expires_in": int(duration.Seconds()),
		"admin": gin.H{
			"id":       admin.ID,
			"username": admin.Username,
		},
	})
}

// Logout 用户登出（管理员和客户端通用）
func (h *AdminHandler) Logout(c *gin.Context) {
	// 从JWT中间件获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未认证",
			"code":  40100,
		})
		return
	}

	username, _ := c.Get("username")
	userType, _ := c.Get("user_type")

	// 记录登出日志（只有管理员才记录到admin_logs表）
	if userType == "admin" {
		adminID := userID.(uint)
		go h.adminService.LogAction(&adminID, "logout", "admin", strconv.FormatUint(uint64(adminID), 10), c.ClientIP(), gin.H{
			"username": username,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "注销成功",
	})
}

// GetDashboard 获取控制台仪表板数据
func (h *AdminHandler) GetDashboard(c *gin.Context) {
	stats, err := h.adminService.GetDashboardStats()
	if err != nil {
		// 记录详细的错误信息
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")
		logger.GetLogger().Error("获取控制台统计信息失败",
			zap.Error(err),
			zap.String("method", "GetDashboard"),
			zap.String("user_id", fmt.Sprintf("%v", userID)),
			zap.String("username", fmt.Sprintf("%v", username)))

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取统计信息失败",
			"code":  50000,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}

// CreateAdmin 创建管理员
func (h *AdminHandler) CreateAdmin(c *gin.Context) {
	var req services.CreateAdminRequest
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

	admin, err := h.adminService.CreateAdmin(&req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "创建管理员失败",
				"code":  50000,
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"id":         admin.ID,
			"username":   admin.Username,
			"is_active":  admin.IsActive,
			"created_at": admin.CreatedAt,
		},
	})
}

// ListAdmins 获取管理员列表
func (h *AdminHandler) ListAdmins(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	admins, total, err := h.adminService.ListAdmins(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取管理员列表失败",
			"code":  50000,
		})
		return
	}

	// 转换数据格式（隐藏敏感信息）
	var adminList []gin.H
	for _, admin := range admins {
		adminList = append(adminList, gin.H{
			"id":         admin.ID,
			"username":   admin.Username,
			"is_active":  admin.IsActive,
			"last_login": admin.LastLogin,
			"created_at": admin.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": adminList,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// UpdateAdmin 更新管理员
func (h *AdminHandler) UpdateAdmin(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的管理员ID",
			"code":  40000,
		})
		return
	}

	var req services.UpdateAdminRequest
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

	admin, err := h.adminService.UpdateAdmin(uint(id), &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "更新管理员失败",
				"code":  50000,
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":         admin.ID,
			"username":   admin.Username,
			"is_active":  admin.IsActive,
			"updated_at": admin.UpdatedAt,
		},
	})
}

// DeleteAdmin 删除管理员
func (h *AdminHandler) DeleteAdmin(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的管理员ID",
			"code":  40000,
		})
		return
	}

	err = h.adminService.DeleteAdmin(uint(id))
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "删除管理员失败",
				"code":  50000,
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}

// EnableTOTP 启用双因素认证
func (h *AdminHandler) EnableTOTP(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的管理员ID",
			"code":  40000,
		})
		return
	}

	qrCodeURL, err := h.adminService.EnableTOTP(uint(id))
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "启用双因素认证失败",
				"code":  50000,
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"qr_code_url": qrCodeURL,
		"message":     "双因素认证已启用，请扫描二维码",
	})
}

// DisableTOTP 禁用双因素认证
func (h *AdminHandler) DisableTOTP(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的管理员ID",
			"code":  40000,
		})
		return
	}

	err = h.adminService.DisableTOTP(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "禁用双因素认证失败",
			"code":  50000,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "双因素认证已禁用",
	})
}

// GetLogs 获取操作日志
func (h *AdminHandler) GetLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	action := c.Query("action")
	targetType := c.Query("target_type")

	var adminID *uint
	if adminIDStr := c.Query("admin_id"); adminIDStr != "" {
		if id, err := strconv.ParseUint(adminIDStr, 10, 32); err == nil {
			adminIDUint := uint(id)
			adminID = &adminIDUint
		}
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	logs, total, err := h.adminService.GetLogs(page, limit, action, targetType, adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取操作日志失败",
			"code":  50000,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": logs,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}
