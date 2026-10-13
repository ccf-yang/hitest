package middleware

import (
	psessionUtils "hitest/internal/core/init/session/psession"

	"github.com/gin-gonic/gin"
)

// LoadOperationUser 返回当前请求的用户名（来自伪会话或 API 认证上下文）。
func LoadOperationUser(c *gin.Context) string {
	if sessionUser, ok := c.Get(psessionUtils.GinContextSessionUserKey); ok {
		if psession, ok := sessionUser.(psessionUtils.SessionUser); ok {
			return psession.Name
		}
	}
	if apiUsername, ok := c.Get("API_AUTH_USERNAME"); ok {
		if username, ok := apiUsername.(string); ok {
			return username
		}
	}
	return ""
}
