package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gachal/mossbase/services/rag/pkg/response"
	"github.com/gin-gonic/gin"
)

// APIKeyAuth returns a middleware that validates API key authentication.
// It reads the key from the X-API-Key header or Authorization: Bearer xxx header.
// If the apiKeys list is empty, authentication is skipped (for development).
func APIKeyAuth(apiKeys []string) gin.HandlerFunc {
	validKeys := make([][]byte, 0, len(apiKeys))
	for _, k := range apiKeys {
		trimmed := strings.TrimSpace(k)
		if trimmed != "" {
			validKeys = append(validKeys, []byte(trimmed))
		}
	}

	return func(c *gin.Context) {
		// Skip auth when no keys are configured (development mode).
		if len(validKeys) == 0 {
			c.Next()
			return
		}

		apiKey := extractAPIKey(c)
		if apiKey == "" {
			response.Error(c, http.StatusUnauthorized, "API key is required")
			c.Abort()
			return
		}

		apiKeyBytes := []byte(apiKey)
		matched := false
		for _, valid := range validKeys {
			if subtle.ConstantTimeCompare(apiKeyBytes, valid) == 1 {
				matched = true
				break
			}
		}

		if !matched {
			response.Error(c, http.StatusUnauthorized, "invalid API key")
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractAPIKey reads the API key from the X-API-Key header first, falling
// back to the Authorization: Bearer xxx header.
func extractAPIKey(c *gin.Context) string {
	if key := c.GetHeader("X-API-Key"); key != "" {
		return strings.TrimSpace(key)
	}

	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	}

	return ""
}
