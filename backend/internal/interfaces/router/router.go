package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gachal/mossbase/backend/internal/application/service"
	"github.com/gachal/mossbase/backend/internal/domain/repository"
	"github.com/gachal/mossbase/backend/internal/infrastructure/config"
	"github.com/gachal/mossbase/backend/internal/infrastructure/middleware"
	"github.com/gachal/mossbase/backend/internal/infrastructure/rag"
	"github.com/gachal/mossbase/backend/internal/interfaces/handler"
	"github.com/gachal/mossbase/backend/pkg/response"
)

func Setup(
	engine *gin.Engine,
	cfg *config.Config,
	userRepo repository.UserRepository,
	spaceRepo repository.SpaceRepository,
	spaceMemberRepo repository.SpaceMemberRepository,
	pageRepo repository.PageRepository,
	pageVersionRepo repository.PageVersionRepository,
	ragClient *rag.RAGClient,
) {
	// 静态文件服务（上传的图片）
	if cfg.Upload.Dir != "" {
		baseURL := cfg.Upload.BaseURL
		if baseURL == "" {
			baseURL = "/uploads"
		}
		engine.Static(baseURL, cfg.Upload.Dir)
	}

	api := engine.Group("/api/v1")

	// Block install endpoints in normal mode
	api.Use(blockInstallEndpoints())

	// Services
	userSvc := service.NewUserService(userRepo, cfg.JWT.Secret, cfg.JWT.ExpiryHours)
	spaceSvc := service.NewSpaceService(spaceRepo, spaceMemberRepo, userRepo)
	pageSvc := service.NewPageService(pageRepo, pageVersionRepo, ragClient)
	versionSvc := service.NewPageVersionService(pageRepo, pageVersionRepo)
	adminSvc := service.NewAdminService(userRepo, spaceRepo, spaceMemberRepo, pageRepo, "configs/config.yaml")

	// Handlers
	userH := handler.NewUserHandler(userSvc)
	spaceH := handler.NewSpaceHandler(spaceSvc)
	pageH := handler.NewPageHandler(pageSvc)
	versionH := handler.NewPageVersionHandler(versionSvc)
	adminH := handler.NewAdminHandler(adminSvc)
	uploadH := handler.NewUploadHandler(&cfg.Upload)

	// Public: auth
	auth := api.Group("/auth")
	auth.POST("/register", userH.Register)
	auth.POST("/login", userH.Login)

	// Authenticated routes
	authorized := api.Group("")
	authorized.Use(middleware.Auth(cfg.JWT.Secret))
	authorized.Use(middleware.Pagination())

	// User profile
	user := authorized.Group("/user")
	user.GET("/profile", userH.GetProfile)
	user.PUT("/profile", userH.UpdateProfile)

	// Upload
	uploadGroup := authorized.Group("/upload")
	uploadGroup.POST("/avatar", uploadH.UploadAvatar)
	uploadGroup.POST("/space-cover", uploadH.UploadSpaceCover)

	// Space routes
	spaces := authorized.Group("/spaces")
	spaces.POST("", spaceH.Create)
	spaces.GET("", spaceH.List)

	// Space-scoped routes (member check)
	spaceDetail := spaces.Group("/:spaceId")
	spaceDetail.Use(middleware.SpaceAuth(spaceMemberRepo, false))
	spaceDetail.GET("", spaceH.GetByID)

	// Space admin routes
	spaceAdmin := spaces.Group("/:spaceId")
	spaceAdmin.Use(middleware.SpaceAuth(spaceMemberRepo, true))
	spaceAdmin.PUT("", spaceH.Update)
	spaceAdmin.DELETE("", spaceH.Delete)
	spaceAdmin.POST("/members", spaceH.AddMember)
	spaceAdmin.DELETE("/members/:userId", spaceH.RemoveMember)

	// Space member routes (member check only)
	spaceMembers := spaces.Group("/:spaceId")
	spaceMembers.Use(middleware.SpaceAuth(spaceMemberRepo, false))
	spaceMembers.GET("/members", spaceH.ListMembers)

	// Page routes under space (member check)
	spacePages := spaces.Group("/:spaceId/pages")
	spacePages.Use(middleware.SpaceAuth(spaceMemberRepo, false))
	spacePages.POST("", pageH.Create)
	spacePages.GET("/tree", pageH.GetTree)
	spacePages.GET("/search", pageH.Search)
	spacePages.GET("/semantic-search", pageH.SemanticSearch)

	pageDetail := spacePages.Group("/:pageId")
	pageDetail.GET("", pageH.GetByID)
	pageDetail.PUT("", pageH.Update)
	pageDetail.DELETE("", pageH.Delete)
	pageDetail.PUT("/move", pageH.MovePage)

	// Version routes
	pageDetail.GET("/versions", versionH.ListVersions)
	pageDetail.GET("/versions/diff", versionH.GetDiff)
	pageDetail.GET("/versions/:versionNumber", versionH.GetVersion)
	pageDetail.POST("/versions/:versionNumber/restore", versionH.RestoreVersion)

	// Admin routes
	admin := authorized.Group("/admin")
	admin.Use(middleware.AdminAuth())
	admin.GET("/dashboard", adminH.Dashboard)
	admin.GET("/users", adminH.ListUsers)
	admin.PUT("/users/:userId/role", adminH.UpdateUserRole)
	admin.PUT("/users/:userId/status", adminH.UpdateUserStatus)
	admin.GET("/spaces", adminH.ListSpaces)
	admin.GET("/spaces/:spaceId", adminH.GetSpaceDetail)
	admin.DELETE("/spaces/:spaceId", adminH.DeleteSpace)
	admin.GET("/pages", adminH.ListPages)
	admin.DELETE("/pages/:pageId", adminH.DeletePage)
	admin.GET("/settings", adminH.GetSettings)
	admin.PUT("/settings", adminH.UpdateSettings)
	admin.POST("/settings/test-rag", adminH.TestRAGConnection)
}

func SetupInstall(engine *gin.Engine, installH *handler.InstallHandler) {
	api := engine.Group("/api/v1")

	install := api.Group("/install")
	install.Use(middleware.InstallRateLimit(20, time.Minute))
	install.GET("/status", installH.GetStatus)
	install.POST("/test-db", installH.TestDatabase)
	install.POST("/execute", installH.Execute)

	engine.NoRoute(func(c *gin.Context) {
		response.Error(c, http.StatusServiceUnavailable, "installation required")
	})
}

func blockInstallEndpoints() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/api/v1/install/execute" || path == "/api/v1/install/test-db" {
			response.Error(c, http.StatusNotFound, "not found")
			c.Abort()
			return
		}
		c.Next()
	}
}
