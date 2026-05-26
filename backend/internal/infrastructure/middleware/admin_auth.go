package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gachal/mossbase/backend/pkg/response"
)

func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists || role != "admin" {
			response.Error(c, http.StatusForbidden, "admin access required")
			c.Abort()
			return
		}
		c.Next()
	}
}
