package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/gachal/mossbase/backend/internal/infrastructure/config"
	"github.com/gachal/mossbase/backend/internal/infrastructure/database"
	applogger "github.com/gachal/mossbase/backend/internal/infrastructure/logger"
	"github.com/gachal/mossbase/backend/internal/infrastructure/middleware"
	"github.com/gachal/mossbase/backend/internal/infrastructure/persistence/repository"
	"github.com/gachal/mossbase/backend/internal/infrastructure/rag"
	"github.com/gachal/mossbase/backend/internal/application/service"
	"github.com/gachal/mossbase/backend/internal/interfaces/handler"
	"github.com/gachal/mossbase/backend/internal/interfaces/router"
)

func main() {
	if config.IsInstalled() {
		runNormalMode()
	} else {
		runInstallMode()
	}
}

func runNormalMode() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	applogger.Init(cfg.Log.Level, cfg.Log.Output)

	db, err := database.NewMySQL(cfg.Database)
	if err != nil {
		zap.L().Fatal("connect database", zap.Error(err))
	}

	userRepo := repository.NewUserRepository(db)
	spaceRepo := repository.NewSpaceRepository(db)
	spaceMemberRepo := repository.NewSpaceMemberRepository(db)
	pageRepo := repository.NewPageRepository(db)
	pageVersionRepo := repository.NewPageVersionRepository(db)

	var ragClient *rag.RAGClient
	if cfg.RAG.Enabled {
		ragClient = rag.NewRAGClient(cfg.RAG)
		zap.L().Info("RAG client enabled", zap.String("base_url", cfg.RAG.BaseURL))
	}

	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(
		middleware.Recovery(),
		middleware.RequestLogger(),
		middleware.CORS(),
	)

	router.Setup(engine, cfg, userRepo, spaceRepo, spaceMemberRepo, pageRepo, pageVersionRepo, ragClient)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	zap.L().Info("server starting", zap.String("addr", addr), zap.String("mode", "normal"))
	if err := engine.Run(addr); err != nil {
		zap.L().Fatal("server failed", zap.Error(err))
	}
}

func runInstallMode() {
	cfg, _ := config.LoadMinimal()
	applogger.Init(cfg.Log.Level, cfg.Log.Output)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(
		middleware.Recovery(),
		middleware.RequestLogger(),
		middleware.CORS(),
	)

	shutdownCh := make(chan struct{}, 1)
	installSvc := service.NewInstallService()
	installH := handler.NewInstallHandler(installSvc, shutdownCh)
	router.SetupInstall(engine, installH)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	go func() {
		zap.L().Info("server starting", zap.String("addr", addr), zap.String("mode", "install"))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-shutdownCh:
		zap.L().Info("installation completed, shutting down for restart")
	case <-quit:
		zap.L().Info("received shutdown signal")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	os.Exit(0)
}
