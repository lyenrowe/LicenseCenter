package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/lyenrowe/LicenseCenter/internal/services"
	"github.com/lyenrowe/LicenseCenter/pkg/errors"
)

// LicenseHandler 授权处理器
type LicenseHandler struct {
	licenseService *services.LicenseService
	rsaService     *services.RSAService
	validator      *validator.Validate
}

// NewLicenseHandler 创建授权处理器
func NewLicenseHandler() *LicenseHandler {
	return &LicenseHandler{
		licenseService: services.NewLicenseService(),
		rsaService:     services.NewRSAService(),
		validator:      validator.New(),
	}
}

// ActivateLicensesRequest 批量激活请求
type ActivateLicensesRequest struct {
	AuthorizationCode string              `json:"authorization_code" validate:"required"`
	BindFiles         []services.BindFile `json:"bind_files" validate:"required,min=1,dive"`
}

// TransferLicenseRequest 授权转移请求
type TransferLicenseRequest struct {
	AuthorizationCode string              `json:"authorization_code" validate:"required"`
	UnbindFile        services.UnbindFile `json:"unbind_file" validate:"required"`
	BindFile          services.BindFile   `json:"bind_file" validate:"required"`
}

// GetPublicKey 获取服务端公钥
func (h *LicenseHandler) GetPublicKey(c *gin.Context) {
	publicKeyPEM, err := h.rsaService.GetPublicKeyPEM()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取公钥失败",
			"code":  50000,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"public_key": publicKeyPEM,
	})
}

// ActivateLicenses 批量激活设备
func (h *LicenseHandler) ActivateLicenses(c *gin.Context) {
	var req ActivateLicensesRequest
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

	licenseFiles, err := h.licenseService.ActivateLicenses(req.AuthorizationCode, req.BindFiles)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "激活设备失败",
				"code":  50000,
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"license_files": licenseFiles,
		"message":       "设备激活成功",
	})
}

// TransferLicense 授权转移
func (h *LicenseHandler) TransferLicense(c *gin.Context) {
	var req TransferLicenseRequest
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

	licenseFile, err := h.licenseService.TransferLicense(req.AuthorizationCode, req.UnbindFile, req.BindFile)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "授权转移失败",
				"code":  50000,
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"license_file": licenseFile,
		"message":      "授权转移成功",
	})
}

// GetLicensesByAuth 获取授权码下的设备列表
func (h *LicenseHandler) GetLicensesByAuth(c *gin.Context) {
	authCode := c.Query("auth_code")
	if authCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "授权码不能为空",
			"code":  40000,
		})
		return
	}

	licenses, err := h.licenseService.GetLicensesByAuth(authCode)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "获取设备列表失败",
				"code":  50000,
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": licenses,
	})
}

// ForceUnbindLicense 管理员强制解绑设备
func (h *LicenseHandler) ForceUnbindLicense(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的授权ID",
			"code":  40000,
		})
		return
	}

	type UnbindRequest struct {
		Reason string `json:"reason"`
	}

	var req UnbindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Reason = "管理员强制解绑"
	}

	err = h.licenseService.ForceUnbindLicense(uint(id), req.Reason)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "强制解绑失败",
				"code":  50000,
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "设备解绑成功",
	})
}
