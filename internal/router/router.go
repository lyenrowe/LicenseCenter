package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lyenrowe/LicenseCenter/internal/handlers"
	"github.com/lyenrowe/LicenseCenter/internal/middleware"
	"github.com/lyenrowe/LicenseCenter/pkg/logger"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	// 全局中间件
	r.Use(gin.Recovery())
	r.Use(middleware.LoggingMiddleware(logger.GetLogger()))
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// 创建处理器实例
	adminHandler := handlers.NewAdminHandler()
	authHandler := handlers.NewAuthorizationHandler()
	licenseHandler := handlers.NewLicenseHandler()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "License Center is running",
		})
	})

	// API v1 路由组
	v1 := r.Group("/api")

	// 公开接口（无需认证）
	{
		// 管理员登录
		v1.POST("/admin/login", adminHandler.AdminLogin)

		// 获取服务端公钥
		v1.GET("/public-key", licenseHandler.GetPublicKey)
	}

	// 客户端接口（无需认证，但需要有效授权码）
	licenses := v1.Group("/licenses")
	{
		// 批量激活设备
		licenses.POST("/activate", licenseHandler.ActivateLicenses)

		// 授权转移
		licenses.POST("/transfer", licenseHandler.TransferLicense)

		// 根据授权码获取设备列表
		licenses.GET("", licenseHandler.GetLicensesByAuth)
	}

	// 管理员接口（需要JWT认证 + 管理员权限）
	admin := v1.Group("/admin")
	admin.Use(middleware.JWTAuthMiddleware())
	admin.Use(middleware.AdminAuthMiddleware())
	admin.Use(middleware.AdminActionLoggingMiddleware())
	{
		// 控制台仪表板
		admin.GET("/dashboard", adminHandler.GetDashboard)

		// 管理员管理
		admins := admin.Group("/admins")
		{
			admins.POST("", adminHandler.CreateAdmin)
			admins.GET("", adminHandler.ListAdmins)
			admins.PUT("/:id", adminHandler.UpdateAdmin)
			admins.DELETE("/:id", adminHandler.DeleteAdmin)

			// 双因素认证
			admins.POST("/:id/totp/enable", adminHandler.EnableTOTP)
			admins.DELETE("/:id/totp", adminHandler.DisableTOTP)
		}

		// 授权码管理
		authorizations := admin.Group("/authorizations")
		{
			authorizations.POST("", authHandler.CreateAuthorization)
			authorizations.GET("", authHandler.ListAuthorizations)
			authorizations.GET("/:id", authHandler.GetAuthorization)
			authorizations.PUT("/:id", authHandler.UpdateAuthorization)
			authorizations.DELETE("/:id", authHandler.DeleteAuthorization)
			authorizations.GET("/statistics", authHandler.GetStatistics)
		}

		// 设备管理
		adminLicenses := admin.Group("/licenses")
		{
			// 强制解绑设备
			adminLicenses.DELETE("/:id/unbind", licenseHandler.ForceUnbindLicense)
		}

		// 操作日志
		admin.GET("/logs", adminHandler.GetLogs)

		// RSA密钥管理
		rsa := admin.Group("/rsa")
		{
			// 获取当前公钥
			rsa.GET("/public-key", licenseHandler.GetPublicKey)

			// 轮换密钥（暂时先提供接口，后续可实现）
			rsa.POST("/rotate", func(c *gin.Context) {
				c.JSON(http.StatusNotImplemented, gin.H{
					"error": "密钥轮换功能暂未实现",
					"code":  50100,
				})
			})
		}
	}

	return r
}
