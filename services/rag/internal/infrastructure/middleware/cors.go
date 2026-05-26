package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS returns a middleware that validates the Origin header against the
// allowed origins list. If no origins are configured, CORS headers are not set
// and requests from browsers will be blocked by same-origin policy.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[strings.ToLower(strings.TrimSpace(o))] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if origin != "" {
			if _, ok := originSet[strings.ToLower(origin)]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				c.Header("Access-Control-Allow-Headers", "Content-Type, X-API-Key, Authorization")
				c.Header("Vary", "Origin")
			}
		}

		if c.Request.Method == http.MethodOptions {
			if origin != "" {
				if _, ok := originSet[strings.ToLower(origin)]; ok {
					c.AbortWithStatus(http.StatusNoContent)
					return
				}
			}
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}
