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

	"go.uber.org/zap"

	"github.com/gachal/mossbase/backend/internal/application/service"
	"github.com/gachal/mossbase/backend/internal/infrastructure/config"
	"github.com/gachal/mossbase/backend/internal/infrastructure/database"
	applogger "github.com/gachal/mossbase/backend/internal/infrastructure/logger"
	"github.com/gachal/mossbase/backend/internal/infrastructure/persistence/repository"
	"github.com/gachal/mossbase/backend/internal/infrastructure/rag"
	mcpiface "github.com/gachal/mossbase/backend/internal/interfaces/mcp"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	applogger.Init(cfg.Log.Level, cfg.Log.Output)

	if !cfg.MCP.Enabled {
		zap.L().Fatal("MCP server is not enabled. Set MOSS_MCP_ENABLED=true or update config.yaml")
	}

	zap.L().Info("MCP config loaded",
		zap.Bool("enabled", cfg.MCP.Enabled),
		zap.String("transport", cfg.MCP.Transport),
		zap.Int("http_port", cfg.MCP.HTTPPort),
		zap.Int("api_keys_count", len(cfg.MCP.APIKeys)),
		zap.Uint64("default_user_id", cfg.MCP.DefaultUserID),
	)

	db, err := database.NewMySQL(cfg.Database)
	if err != nil {
		zap.L().Fatal("connect database", zap.Error(err))
	}

	pageRepo := repository.NewPageRepository(db)
	pageVersionRepo := repository.NewPageVersionRepository(db)
	spaceRepo := repository.NewSpaceRepository(db)
	spaceMemberRepo := repository.NewSpaceMemberRepository(db)
	userRepo := repository.NewUserRepository(db)

	var ragClient *rag.RAGClient
	if cfg.RAG.Enabled {
		ragClient = rag.NewRAGClient(cfg.RAG)
		if ragClient != nil {
			zap.L().Info("RAG client enabled", zap.String("base_url", cfg.RAG.BaseURL))
		}
	}

	pageSvc := service.NewPageService(pageRepo, pageVersionRepo, ragClient)
	spaceSvc := service.NewSpaceService(spaceRepo, spaceMemberRepo, userRepo)

	mcpAuth := mcpiface.NewMCPAuth(cfg.MCP.APIKeys, cfg.MCP.DefaultUserID)

	mcpSrv := mcpiface.NewMCPServer(pageSvc, spaceSvc, mcpAuth, spaceMemberRepo)
	server := mcpSrv.Setup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	transport := cfg.MCP.Transport
	if transport == "" {
		transport = "stdio"
	}

	zap.L().Info("MCP server starting", zap.String("transport", transport))

	switch transport {
	case "http":
		addr := fmt.Sprintf(":%d", cfg.MCP.HTTPPort)
		handler := mcpsdk.NewStreamableHTTPHandler(func(_ *http.Request) *mcpsdk.Server {
			return server
		}, nil)

		srv := &http.Server{
			Addr:              addr,
			Handler:           mcpAuth.HTTPMiddleware(handler),
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			MaxHeaderBytes:    1 << 20,
		}

		go func() {
			<-sigCh
			zap.L().Info("shutting down MCP HTTP server")
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer shutdownCancel()
			if err := srv.Shutdown(shutdownCtx); err != nil {
				zap.L().Error("HTTP server shutdown error", zap.Error(err))
			}
			cancel()
		}()

		zap.L().Info("MCP HTTP server listening", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			zap.L().Fatal("MCP HTTP server failed", zap.Error(err))
		}

	case "both":
		httpAddr := fmt.Sprintf(":%d", cfg.MCP.HTTPPort)
		handler := mcpsdk.NewStreamableHTTPHandler(func(_ *http.Request) *mcpsdk.Server {
			return server
		}, nil)

		srv := &http.Server{
			Addr:              httpAddr,
			Handler:           mcpAuth.HTTPMiddleware(handler),
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			MaxHeaderBytes:    1 << 20,
		}

		go func() {
			zap.L().Info("MCP HTTP server listening", zap.String("addr", httpAddr))
			if err := srv.ListenAndServe(); err != http.ErrServerClosed {
				zap.L().Error("MCP HTTP server failed", zap.Error(err))
			}
		}()

		go func() {
			<-sigCh
			zap.L().Info("shutting down MCP server")
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer shutdownCancel()
			srv.Shutdown(shutdownCtx)
			cancel()
		}()

		if err := server.Run(ctx, &mcpsdk.StdioTransport{}); err != nil {
			zap.L().Fatal("MCP stdio server failed", zap.Error(err))
		}

	default:
		go func() {
			<-sigCh
			zap.L().Info("shutting down MCP server")
			cancel()
		}()
		if err := server.Run(ctx, &mcpsdk.StdioTransport{}); err != nil {
			zap.L().Fatal("MCP server failed", zap.Error(err))
		}
	}
}
