package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// SessionAuthMiddleware는 세션 기반 인증을 확인하는 미들웨어 함수를 반환합니다.
func SessionAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		admin := session.Get("admin")
		if admin == nil || admin != true {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}
