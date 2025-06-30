package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/lyenrowe/LicenseCenter/pkg/errors"
	"github.com/lyenrowe/LicenseCenter/pkg/logger"
	"go.uber.org/zap"
)

// ErrorHandlerMiddleware 统一错误处理中间件
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(error); ok {
			handleError(c, err)
		} else {
			// 处理panic
			logger.GetLogger().Error("服务器发生panic",
				zap.Any("panic", recovered),
				zap.String("stack", string(debug.Stack())),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("client_ip", c.ClientIP()),
			)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "服务器内部错误",
				"code":  50000,
			})
		}
		c.Abort()
	})
}

// ErrorResponseHandler 处理业务层返回的错误
func ErrorResponseHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误需要处理
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			handleError(c, err)
		}
	}
}

// handleError 统一错误处理函数
func handleError(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		// 业务错误 - 记录到app.log
		logLevel := determineLogLevel(appErr)
		logBusinessError(c, appErr, logLevel)

		// 如果还没有返回响应，则返回错误响应
		if !c.Writer.Written() {
			c.JSON(appErr.HTTPStatus(), gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		}
	} else {
		// 系统错误 - 记录到error.log
		logSystemError(c, err)

		if !c.Writer.Written() {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "服务器内部错误",
				"code":  50000,
			})
		}
	}
}

// determineLogLevel 根据错误类型确定日志级别
func determineLogLevel(appErr *errors.AppError) string {
	code := appErr.Code
	switch {
	case code >= 50000: // 系统错误
		return "error"
	case code >= 40000 && code < 41000: // 客户端错误
		return "warn"
	case code >= 41000 && code < 42000: // 业务逻辑错误
		return "warn"
	case code >= 43000: // 资源不存在
		return "info"
	default:
		return "warn"
	}
}

// logBusinessError 记录业务错误到app.log
func logBusinessError(c *gin.Context, appErr *errors.AppError, level string) {
	fields := []zap.Field{
		zap.String("error_type", "business"),
		zap.Int("error_code", appErr.Code),
		zap.String("error_message", appErr.Message),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	}

	// 添加用户信息（如果存在）
	if username, exists := c.Get("username"); exists {
		fields = append(fields, zap.String("username", username.(string)))
	}
	if userType, exists := c.Get("user_type"); exists {
		fields = append(fields, zap.String("user_type", userType.(string)))
	}

	// 业务错误记录到应用日志
	switch level {
	case "error":
		logger.GetLogger().Error("业务错误", fields...)
	case "warn":
		logger.GetLogger().Warn("业务警告", fields...)
	case "info":
		logger.GetLogger().Info("业务信息", fields...)
	default:
		logger.GetLogger().Warn("业务错误", fields...)
	}
}

// logSystemError 记录系统错误到error.log
func logSystemError(c *gin.Context, err error) {
	// 系统错误既记录到应用日志也记录到错误日志
	fields := []zap.Field{
		zap.String("error_type", "system"),
		zap.Error(err),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
		zap.String("stack", string(debug.Stack())),
	}

	// 记录到应用日志
	logger.GetLogger().Error("系统错误", fields...)

	// 同时记录到专门的错误日志文件
	logger.GetErrorLogger().Error("系统错误", fields...)
}

// AbortWithError 工具函数，用于在处理器中快速返回错误
func AbortWithError(c *gin.Context, err error) {
	c.Error(err)
	c.Abort()
}
