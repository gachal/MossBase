package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func Pagination() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		if page < 1 {
			page = 1
		}
		if pageSize < 1 {
			pageSize = 20
		}
		if pageSize > 100 {
			pageSize = 100
		}
		c.Set("page", page)
		c.Set("page_size", pageSize)
		c.Set("offset", (page - 1) * pageSize)
		c.Next()
	}
}
