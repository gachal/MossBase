package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/gachal/mossbase/services/rag/pkg/response"
)

// HealthHandler handles health check requests.
type HealthHandler struct{}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck handles GET /health.
// It returns a simple status indicator for load balancers and monitoring.
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	response.Success(c, gin.H{
		"status": "ok",
	})
}
