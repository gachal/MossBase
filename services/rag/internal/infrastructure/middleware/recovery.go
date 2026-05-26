package middleware

import (
	"net/http"

	"github.com/gachal/mossbase/services/rag/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery returns a middleware that recovers from panics, logs the error with
// zap, and returns a 500 JSON response using the standard response envelope.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				zap.L().Error("panic recovered",
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.Any("error", err),
				)

				msg := "internal server error"
				response.Error(c, http.StatusInternalServerError, msg)
				c.Abort()
			}
		}()
		c.Next()
	}
}
