package handlers

import (
	"encoding/json"
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
	// 从JWT中获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未认证",
			"code":  40100,
		})
		return
	}

	// 解析上传的文件
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "文件上传失败",
			"code":  40000,
		})
		return
	}

	bindFiles := form.File["bind_files"]
	if len(bindFiles) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请上传至少一个.bind文件",
			"code":  40000,
		})
		return
	}

	// 解析bind文件内容
	var bindFileContents []services.BindFile
	for _, fileHeader := range bindFiles {
		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无法读取文件: " + fileHeader.Filename,
				"code":  40000,
			})
			return
		}
		defer file.Close()

		content := make([]byte, fileHeader.Size)
		_, err = file.Read(content)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "读取文件内容失败: " + fileHeader.Filename,
				"code":  40000,
			})
			return
		}

		var bindFile services.BindFile
		if err := json.Unmarshal(content, &bindFile); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "文件格式错误: " + fileHeader.Filename,
				"code":  40000,
			})
			return
		}
		bindFileContents = append(bindFileContents, bindFile)
	}

	// 激活设备 (临时使用现有方法，后续需要在service中实现)
	// TODO: 需要在LicenseService中实现ActivateLicensesByUserID方法
	_ = userID           // 避免未使用变量警告
	_ = bindFileContents // 避免未使用变量警告

	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "功能开发中",
		"code":  50100,
	})
}

// TransferLicense 授权转移
func (h *LicenseHandler) TransferLicense(c *gin.Context) {
	// 从JWT中获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未认证",
			"code":  40100,
		})
		return
	}

	// 解析上传的文件
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "文件上传失败",
			"code":  40000,
		})
		return
	}

	// 获取unbind文件
	unbindFiles := form.File["unbind_file"]
	if len(unbindFiles) != 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请上传一个.unbind文件",
			"code":  40000,
		})
		return
	}

	// 获取bind文件
	bindFiles := form.File["bind_file"]
	if len(bindFiles) != 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请上传一个.bind文件",
			"code":  40000,
		})
		return
	}

	// TODO: 解析文件内容并调用service方法
	_ = userID      // 避免未使用变量警告
	_ = unbindFiles // 避免未使用变量警告
	_ = bindFiles   // 避免未使用变量警告

	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "功能开发中",
		"code":  50100,
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

// DownloadLicense 下载license文件
func (h *LicenseHandler) DownloadLicense(c *gin.Context) {
	licenseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的授权ID",
			"code":  40000,
		})
		return
	}

	// 从JWT中获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未认证",
			"code":  40100,
		})
		return
	}

	// TODO: 验证用户是否有权限下载该license文件
	// TODO: 从数据库获取license文件内容并返回
	_ = userID    // 避免未使用变量警告
	_ = licenseID // 避免未使用变量警告

	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "功能开发中",
		"code":  50100,
	})
}
