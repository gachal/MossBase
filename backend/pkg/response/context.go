package response

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetUserID(c *gin.Context) uint64 {
	val, exists := c.Get("userID")
	if !exists {
		return 0
	}
	switch v := val.(type) {
	case uint64:
		return v
	case float64:
		return uint64(v)
	case int:
		return uint64(v)
	case string:
		id, _ := strconv.ParseUint(v, 10, 64)
		return id
	}
	return 0
}
