package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gachal/mossbase/backend/pkg/response"
)

type visitor struct {
	count    int
	expiryAt time.Time
}

func InstallRateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	visitors := make(map[string]*visitor)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		v, exists := visitors[ip]
		if !exists || time.Now().After(v.expiryAt) {
			visitors[ip] = &visitor{count: 1, expiryAt: time.Now().Add(window)}
			mu.Unlock()
			c.Next()
			return
		}
		v.count++
		if v.count > maxRequests {
			mu.Unlock()
			response.Error(c, http.StatusTooManyRequests, "too many requests, please wait")
			c.Abort()
			return
		}
		mu.Unlock()
		c.Next()
	}
}
