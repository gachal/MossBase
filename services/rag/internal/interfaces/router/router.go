package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gachal/mossbase/services/rag/internal/application/service"
	"github.com/gachal/mossbase/services/rag/internal/infrastructure/config"
	"github.com/gachal/mossbase/services/rag/internal/infrastructure/middleware"
	"github.com/gachal/mossbase/services/rag/internal/interfaces/handler"
)

// Setup registers all HTTP routes and middleware on the given gin.Engine.
func Setup(engine *gin.Engine, cfg *config.Config, docSvc service.DocumentService) {
	// Global middleware
	engine.Use(middleware.Recovery())
	engine.Use(middleware.CORS(cfg.Server.CORSAllowedOrigins))

	// Health check endpoint (no auth required)
	healthHandler := handler.NewHealthHandler()
	engine.GET("/health", healthHandler.HealthCheck)

	// API v1 group with API key authentication
	docHandler := handler.NewDocumentHandler(docSvc)
	searchHandler := handler.NewSearchHandler(docSvc)

	v1 := engine.Group("/api/v1")
	v1.Use(middleware.APIKeyAuth(cfg.Auth.APIKeys))
	{
		v1.POST("/documents", docHandler.IndexDocument)
		v1.DELETE("/documents/:id", docHandler.DeleteDocument)
		v1.POST("/search", searchHandler.Search)
	}
}
