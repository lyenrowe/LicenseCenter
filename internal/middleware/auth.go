package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/licensecenter/licensecenter/pkg/auth"
)

// JWTAuthMiddleware JWT认证中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	jwtManager := auth.NewJWTManager()

	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "缺少认证令牌",
				"code":  40001,
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的认证令牌格式",
				"code":  40001,
			})
			c.Abort()
			return
		}

		// 提取令牌
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// 验证令牌
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的认证令牌",
				"code":  40001,
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_type", claims.UserType)

		c.Next()
	}
}

// AdminAuthMiddleware 管理员认证中间件
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists || userType != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "需要管理员权限",
				"code":  40003,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CustomerAuthMiddleware 客户认证中间件
func CustomerAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists || userType != "customer" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "需要客户权限",
				"code":  40003,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
