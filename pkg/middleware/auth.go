package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/Guo-Chenxu/pay-log/config"
	"github.com/Guo-Chenxu/pay-log/pkg/auth"
	"github.com/Guo-Chenxu/pay-log/pkg/bizerr"
	"github.com/Guo-Chenxu/pay-log/pkg/logger"
	"github.com/Guo-Chenxu/pay-log/pkg/path"
	"github.com/Guo-Chenxu/pay-log/pkg/response"
)

var noAuthPath map[string]struct{} = map[string]struct{}{
	"/api/v1/auth/login": {},
	"/healthz":           {},
}

// AuthMiddleware 认证中间件.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果请求路径为空（NoRoute，用于静态文件服务），直接放行
		if c.FullPath() == "" {
			c.Next()
			return
		}

		// 不需要鉴权的接口，支持通配符（如 /swagger/*any）
		if skip := path.Contains(noAuthPath, c.FullPath()); skip {
			insertUserIDToContextWhenNoAuth(c)
			c.Next()
			return
		}

		// 不需要鉴权的接口（方法+路径匹配，如 GET /api/v1/post/:id）
		if skip := path.Contains(noAuthPath, c.Request.Method+" "+c.FullPath()); skip {
			insertUserIDToContextWhenNoAuth(c)
			c.Next()
			return
		}

		// 校验 token
		authHeader := c.GetHeader(config.GetAuthConfig().TokenName)
		if authHeader == "" {
			response.FailResponse(c, bizerr.Unauthorized.Code, bizerr.Unauthorized.Message)
			c.Abort()
			return
		}

		loginInfo, err := auth.ValidateToken(c, authHeader)
		if err != nil {
			response.FailResponse(c, bizerr.Unauthorized.Code, bizerr.Unauthorized.Message)
			c.Abort()
			return
		}

		insertUserIDToContext(c, loginInfo.LoginID)

		c.Next()
	}
}

func insertUserIDToContext(c *gin.Context, userID int64) {
	c.Request = c.Request.WithContext(logger.AddToContext(c.Request.Context(), logger.UserIDKey, userID))
}

func insertUserIDToContextWhenNoAuth(c *gin.Context) {
	authHeader := c.GetHeader(config.GetAuthConfig().TokenName)
	if authHeader == "" {
		return
	}

	loginInfo, err := auth.ValidateToken(c, authHeader)
	if err != nil {
		response.FailResponse(c, bizerr.Unauthorized.Code, bizerr.Unauthorized.Message)
		c.Abort()
		return
	}

	insertUserIDToContext(c, loginInfo.LoginID)
}
