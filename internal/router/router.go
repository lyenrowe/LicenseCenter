package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lyenrowe/LicenseCenter/internal/config"
	"github.com/lyenrowe/LicenseCenter/internal/handlers"
	"github.com/lyenrowe/LicenseCenter/internal/middleware"
	"github.com/lyenrowe/LicenseCenter/pkg/logger"
	"go.uber.org/zap"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	// 从配置文件设置Gin模式
	ginMode := config.AppConfig.Logging.GinMode
	switch ginMode {
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.ReleaseMode) // 默认使用release模式
	}

	r := gin.New()

	// 记录路由设置开始
	logger.GetLogger().Info("开始设置路由",
		zap.String("gin_mode", ginMode))

	// 全局中间件
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// 记录panic详情
		logger.GetLogger().Error("HTTP请求发生panic",
			zap.Any("panic", recovered),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()))

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "服务器内部错误",
			"code":  50000,
		})
	}))
	r.Use(middleware.LoggingMiddleware(logger.GetLogger()))
	r.Use(middleware.ErrorHandlerMiddleware())
	r.Use(middleware.ErrorResponseHandler())
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// 静态文件服务
	r.Static("/static", "./web/dist")
	r.StaticFile("/", "./web/dist/index.html")

	// 创建处理器实例
	adminHandler := handlers.NewAdminHandler()
	customerHandler := handlers.NewCustomerHandler()
	authHandler := handlers.NewAuthorizationHandler()
	licenseHandler := handlers.NewLicenseHandler()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "License Center is running",
		})
	})

	// API路由组
	api := r.Group("/api")
	{
		// 客户端相关路由 (无需认证)
		api.POST("/login", customerHandler.Login) // 客户端登录

		// 客户端认证后的路由
		customerAuth := api.Group("/", middleware.JWTAuthMiddleware(), middleware.CustomerAuthMiddleware())
		{
			customerAuth.POST("/logout", customerHandler.Logout) // 客户端登出
		}

		// 客户端控制台路由
		client := api.Group("/client", middleware.JWTAuthMiddleware(), middleware.CustomerAuthMiddleware())
		{
			client.GET("/dashboard", customerHandler.GetDashboard) // 客户端控制台
		}

		// 管理员路由组
		admin := api.Group("/admin")
		{
			// 管理员登录 (无需认证)
			admin.POST("/login", adminHandler.Login)

			// 需要管理员认证的路由
			adminAuth := admin.Group("/", middleware.JWTAuthMiddleware(), middleware.AdminAuthMiddleware(), middleware.AdminActionLoggingMiddleware())
			{
				adminAuth.POST("/logout", adminHandler.Logout) // 管理员登出
				adminAuth.GET("/dashboard", adminHandler.GetDashboard)

				// 管理员管理
				adminAuth.POST("/admins", adminHandler.CreateAdmin)
				adminAuth.GET("/admins", adminHandler.ListAdmins)
				adminAuth.PUT("/admins/:id", adminHandler.UpdateAdmin)
				adminAuth.DELETE("/admins/:id", adminHandler.DeleteAdmin)

				// 双因素认证
				adminAuth.POST("/totp/enable/:id", adminHandler.EnableTOTP)
				adminAuth.POST("/totp/disable/:id", adminHandler.DisableTOTP)
				adminAuth.GET("/totp/info/:id", adminHandler.GetTOTPSetupInfo)
				adminAuth.POST("/totp/verify/:id", adminHandler.VerifyTOTPSetup)

				// 操作日志
				adminAuth.GET("/logs", adminHandler.GetLogs)

				// 授权码管理
				adminAuth.POST("/authorizations", authHandler.CreateAuthorization)
				adminAuth.GET("/authorizations", authHandler.ListAuthorizations)
				adminAuth.GET("/authorizations/:id/details", authHandler.GetAuthorizationDetails)
				adminAuth.PUT("/authorizations/:id", authHandler.UpdateAuthorization)
				adminAuth.DELETE("/authorizations/:id", authHandler.DeleteAuthorization)

				// 设备管理
				adminAuth.POST("/licenses/:id/force-unbind", licenseHandler.ForceUnbindLicense)
			}
		}

		// 客户端操作路由 (需要客户端认证)
		actions := api.Group("/actions", middleware.JWTAuthMiddleware(), middleware.CustomerAuthMiddleware())
		{
			actions.POST("/activate-licenses", licenseHandler.ActivateLicenses)
			actions.POST("/transfer-license", licenseHandler.TransferLicense)
		}

		// 许可证相关路由 (需要JWT认证，但不区分管理员或客户端)
		licenses := api.Group("/licenses", middleware.JWTAuthMiddleware())
		{
			licenses.GET("/:id/download", licenseHandler.DownloadLicense)
		}

		// 公开接口
		api.GET("/public-key", licenseHandler.GetPublicKey)
	}

	return r
}
