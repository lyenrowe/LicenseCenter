package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lyenrowe/LicenseCenter/internal/config"
	"github.com/lyenrowe/LicenseCenter/internal/services"
	"go.uber.org/zap"
)

// LoggingMiddleware 请求日志中间件
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 根据配置决定是否记录HTTP请求日志
		if config.AppConfig.Logging.EnableHTTPLog {
			// 使用zap logger记录请求信息
			logger.Info("HTTP请求",
				zap.String("method", param.Method),
				zap.String("path", param.Path),
				zap.Int("status", param.StatusCode),
				zap.Duration("latency", param.Latency),
				zap.String("client_ip", param.ClientIP),
				zap.String("user_agent", param.Request.UserAgent()),
				zap.Int("body_size", param.BodySize),
			)
		}

		// 返回空字符串避免重复输出到控制台
		return ""
	})
}

// AdminActionLoggingMiddleware 管理员操作日志中间件
func AdminActionLoggingMiddleware() gin.HandlerFunc {
	adminService := services.NewAdminService()

	return func(c *gin.Context) {
		// 只记录管理员的操作
		userType, exists := c.Get("user_type")
		if !exists || userType != "admin" {
			c.Next()
			return
		}

		// 获取管理员ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		adminID := userID.(uint)

		// 记录请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 执行请求
		c.Next()

		// 只记录修改操作（POST, PUT, DELETE）
		method := c.Request.Method
		if method != "POST" && method != "PUT" && method != "DELETE" {
			return
		}

		// 解析操作类型和目标
		action, targetType, targetID := parseActionFromRequest(c, requestBody)
		if action == "" {
			return
		}

		// 准备详情
		details := map[string]interface{}{
			"method": method,
			"path":   c.Request.URL.Path,
			"status": c.Writer.Status(),
		}

		// 添加请求体（如果不是密码相关）
		if len(requestBody) > 0 && !strings.Contains(string(requestBody), "password") {
			var bodyJSON interface{}
			if json.Unmarshal(requestBody, &bodyJSON) == nil {
				details["request_body"] = bodyJSON
			}
		}

		// 记录操作日志
		adminService.LogAction(&adminID, action, targetType, targetID, c.ClientIP(), details)
	}
}

// parseActionFromRequest 从请求中解析操作类型和目标
func parseActionFromRequest(c *gin.Context, requestBody []byte) (action, targetType, targetID string) {
	path := c.Request.URL.Path
	method := c.Request.Method

	// 登出操作
	if path == "/api/admin/logout" && method == "POST" {
		action = "logout"
		targetType = "admin"
		return action, targetType, targetID
	}

	// 管理员相关操作
	if strings.HasPrefix(path, "/api/admin/admins") {
		targetType = "admin"
		switch method {
		case "POST":
			action = "create_admin"
		case "PUT":
			action = "update_admin"
			targetID = extractIDFromPath(path)
		case "DELETE":
			action = "delete_admin"
			targetID = extractIDFromPath(path)
		}
	}

	// 授权码相关操作
	if strings.HasPrefix(path, "/api/admin/authorizations") {
		targetType = "authorization"
		switch method {
		case "POST":
			action = "create_authorization"
		case "PUT":
			action = "update_authorization"
			targetID = extractIDFromPath(path)
		case "DELETE":
			action = "delete_authorization"
			targetID = extractIDFromPath(path)
		}
	}

	// 设备相关操作
	if strings.HasPrefix(path, "/api/admin/licenses") {
		targetType = "license"
		if method == "DELETE" || strings.Contains(path, "unbind") {
			action = "force_unbind_license"
			targetID = extractIDFromPath(path)
		}
	}

	// RSA密钥相关操作
	if strings.HasPrefix(path, "/api/admin/rsa") {
		targetType = "rsa_key"
		if method == "POST" && strings.Contains(path, "rotate") {
			action = "rotate_rsa_keys"
		}
	}

	// 客户端授权操作
	if strings.HasPrefix(path, "/api/licenses") {
		targetType = "license"
		switch {
		case method == "POST" && strings.Contains(path, "activate"):
			action = "activate_licenses"
		case method == "POST" && strings.Contains(path, "transfer"):
			action = "transfer_license"
		}
	}

	return action, targetType, targetID
}

// extractIDFromPath 从路径中提取ID
func extractIDFromPath(path string) string {
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if _, err := strconv.Atoi(part); err == nil {
			return part
		}
	}
	return ""
}
