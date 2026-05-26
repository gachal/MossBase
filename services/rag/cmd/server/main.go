package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/gachal/mossbase/services/rag/internal/application/service"
	"github.com/gachal/mossbase/services/rag/internal/infrastructure/config"
	"github.com/gachal/mossbase/services/rag/internal/infrastructure/embedding"
	"github.com/gachal/mossbase/services/rag/internal/infrastructure/logger"
	qdrantinf "github.com/gachal/mossbase/services/rag/internal/infrastructure/qdrant"
	"github.com/gachal/mossbase/services/rag/internal/interfaces/router"
	"github.com/gachal/mossbase/services/rag/pkg/chunker"
)

func main() {
	// Load configuration
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(cfg.Log.Level, cfg.Log.Output)

	// Set Gin mode
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin engine
	engine := gin.New()

	// Initialize Qdrant client
	qdrantClient, err := qdrantinf.NewQdrantClient(cfg.Qdrant)
	if err != nil {
		zap.L().Fatal("failed to connect to Qdrant", zap.Error(err))
	}
	defer qdrantClient.Close()

	// Initialize Qdrant repository
	vectorRepo := qdrantinf.NewQdrantRepository(qdrantClient)

	// Initialize OpenAI embedding provider
	embedder := embedding.NewOpenAIEmbeddingProvider(cfg.Embedding)

	// Initialize text chunker
	textChunker := chunker.NewTextChunker(cfg.Chunker)

	// Initialize document service
	collectionPrefix := cfg.Qdrant.CollectionPrefix
	if collectionPrefix == "" {
		collectionPrefix = "mossbase"
	}
	dimensions := cfg.Embedding.Dimensions
	if dimensions <= 0 {
		dimensions = 1536
	}

	docSvc := service.NewDocumentService(vectorRepo, embedder, textChunker, collectionPrefix, dimensions)

	// Setup routes
	router.Setup(engine, cfg, docSvc)

	// Configure HTTP server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		zap.L().Info("starting RAG HTTP server", zap.String("addr", addr))
		if listenErr := srv.ListenAndServe(); listenErr != nil && listenErr != http.ErrServerClosed {
			zap.L().Fatal("failed to start server", zap.Error(listenErr))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	zap.L().Info("received shutdown signal", zap.String("signal", sig.String()))

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if shutdownErr := srv.Shutdown(ctx); shutdownErr != nil {
		zap.L().Fatal("server forced to shutdown", zap.Error(shutdownErr))
	}

	zap.L().Info("server exited gracefully")
}
