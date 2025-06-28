package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/licensecenter/licensecenter/internal/services"
	"github.com/licensecenter/licensecenter/pkg/errors"
)

// AuthorizationHandler 授权码处理器
type AuthorizationHandler struct {
	authService *services.AuthorizationService
	validator   *validator.Validate
}

// NewAuthorizationHandler 创建授权码处理器
func NewAuthorizationHandler() *AuthorizationHandler {
	return &AuthorizationHandler{
		authService: services.NewAuthorizationService(),
		validator:   validator.New(),
	}
}

// CreateAuthorization 创建授权码
func (h *AuthorizationHandler) CreateAuthorization(c *gin.Context) {
	var req services.CreateAuthorizationRequest
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

	auth, err := h.authService.CreateAuthorization(&req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "创建授权码失败",
				"code":  50000,
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"id":                 auth.ID,
			"customer_name":      auth.CustomerName,
			"authorization_code": auth.AuthorizationCode,
			"max_seats":          auth.MaxSeats,
			"used_seats":         auth.UsedSeats,
			"duration_years":     auth.DurationYears,
			"latest_expiry_date": auth.LatestExpiryDate,
			"status":             auth.Status,
			"created_at":         auth.CreatedAt,
		},
	})
}

// ListAuthorizations 获取授权码列表
func (h *AuthorizationHandler) ListAuthorizations(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	var status *int
	if statusStr := c.Query("status"); statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			status = &s
		}
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	auths, total, err := h.authService.ListAuthorizations(page, limit, search, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取授权码列表失败",
			"code":  50000,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": auths,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetAuthorization 获取单个授权码详情
func (h *AuthorizationHandler) GetAuthorization(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的授权码ID",
			"code":  40000,
		})
		return
	}

	auth, err := h.authService.GetAuthorizationByID(uint(id))
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "获取授权码失败",
				"code":  50000,
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": auth,
	})
}

// UpdateAuthorization 更新授权码
func (h *AuthorizationHandler) UpdateAuthorization(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的授权码ID",
			"code":  40000,
		})
		return
	}

	var req services.UpdateAuthorizationRequest
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

	auth, err := h.authService.UpdateAuthorization(uint(id), &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "更新授权码失败",
				"code":  50000,
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":                 auth.ID,
			"customer_name":      auth.CustomerName,
			"authorization_code": auth.AuthorizationCode,
			"max_seats":          auth.MaxSeats,
			"used_seats":         auth.UsedSeats,
			"duration_years":     auth.DurationYears,
			"latest_expiry_date": auth.LatestExpiryDate,
			"status":             auth.Status,
			"updated_at":         auth.UpdatedAt,
		},
	})
}

// DeleteAuthorization 删除授权码
func (h *AuthorizationHandler) DeleteAuthorization(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的授权码ID",
			"code":  40000,
		})
		return
	}

	err = h.authService.DeleteAuthorization(uint(id))
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "删除授权码失败",
				"code":  50000,
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}

// GetStatistics 获取授权码统计信息
func (h *AuthorizationHandler) GetStatistics(c *gin.Context) {
	stats, err := h.authService.GetStatistics()
	if err != nil {
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
