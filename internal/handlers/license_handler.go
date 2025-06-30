package handlers

import (
	"archive/zip"
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/lyenrowe/LicenseCenter/internal/services"
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
	// 获取用户名（授权码）
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户信息不完整",
			"code":  40100,
		})
		return
	}

	authCode := username.(string)

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

	// 读取并解密bind文件内容
	var encryptedBindFiles []string
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

		// 将文件内容作为加密数据处理
		encryptedBindFiles = append(encryptedBindFiles, string(content))
	}

	// 使用LicenseService的加密激活方法
	encryptedLicenseFiles, err := h.licenseService.ActivateLicensesEncrypted(authCode, encryptedBindFiles)
	if err != nil {
		// 直接使用gin的Error方法，让错误处理中间件统一处理
		c.Error(err)
		return
	}

	// 创建ZIP文件包含所有license文件
	zipBuffer, err := h.createLicenseZip(encryptedLicenseFiles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建授权文件包失败",
			"code":  50000,
		})
		return
	}

	// 返回ZIP文件
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename=licenses.zip")
	c.Header("Content-Length", fmt.Sprintf("%d", len(zipBuffer)))
	c.Data(http.StatusOK, "application/zip", zipBuffer)
}

// createLicenseZip 创建包含所有license文件的ZIP包
func (h *LicenseHandler) createLicenseZip(encryptedLicenseFiles []services.EncryptedFileResponse) ([]byte, error) {
	var zipBuffer bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuffer)

	for i, licenseFile := range encryptedLicenseFiles {
		fileName := fmt.Sprintf("license_%d.license", i+1)
		fileWriter, err := zipWriter.Create(fileName)
		if err != nil {
			zipWriter.Close()
			return nil, err
		}

		_, err = fileWriter.Write([]byte(licenseFile.EncryptedContent))
		if err != nil {
			zipWriter.Close()
			return nil, err
		}
	}

	err := zipWriter.Close()
	if err != nil {
		return nil, err
	}

	return zipBuffer.Bytes(), nil
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
		c.Error(err)
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
		c.Error(err)
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
