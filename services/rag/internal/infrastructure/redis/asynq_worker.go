package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/gachal/mossbase/services/rag/internal/application/dto"
	"github.com/gachal/mossbase/services/rag/internal/application/service"
	"github.com/gachal/mossbase/services/rag/internal/infrastructure/config"
)

const (
	// TaskTypeEmbedDocument is the asynq task type for document embedding.
	TaskTypeEmbedDocument = "rag:embed_document"
	// TaskTypeDeleteDocument is the asynq task type for document deletion.
	TaskTypeDeleteDocument = "rag:delete_document"
)

// AsynqWorker manages async embedding tasks via Asynq.
type AsynqWorker struct {
	server *asynq.Server
	client *asynq.Client
	docSvc service.DocumentService
}

// NewAsynqWorker creates a new AsynqWorker with the given Redis config and document service.
func NewAsynqWorker(cfg config.RedisConfig, docSvc service.DocumentService) (*AsynqWorker, error) {
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	server := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: 4,
		Queues: map[string]int{
			"default":  6,
			"critical": 10,
		},
		RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
			return time.Duration(n*n) * time.Second
		},
	})

	client := asynq.NewClient(redisOpt)

	return &AsynqWorker{
		server: server,
		client: client,
		docSvc: docSvc,
	}, nil
}

// Start begins processing async tasks.
func (w *AsynqWorker) Start() {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskTypeEmbedDocument, w.handleEmbedDocument)
	mux.HandleFunc(TaskTypeDeleteDocument, w.handleDeleteDocument)

	zap.L().Info("starting asynq worker server")
	go func() {
		if err := w.server.Run(mux); err != nil {
			zap.L().Error("asynq server stopped with error", zap.Error(err))
		}
	}()
}

// Stop gracefully shuts down the asynq server and client.
func (w *AsynqWorker) Stop() {
	w.server.Shutdown()
	if err := w.client.Close(); err != nil {
		zap.L().Error("failed to close asynq client", zap.Error(err))
	}
}

// EnqueueIndexTask enqueues a document indexing task.
func (w *AsynqWorker) EnqueueIndexTask(ctx context.Context, req dto.IndexDocumentRequest) error {
	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal index document request: %w", err)
	}

	task := asynq.NewTask(TaskTypeEmbedDocument, payload)
	info, err := w.client.EnqueueContext(ctx, task,
		asynq.Queue("default"),
		asynq.MaxRetry(5),
		asynq.Timeout(10*time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue index task for doc %s: %w", req.DocumentID, err)
	}

	zap.L().Info("enqueued index task",
		zap.String("doc_id", req.DocumentID),
		zap.String("task_id", info.ID),
	)
	return nil
}

// EnqueueDeleteTask enqueues a document deletion task.
func (w *AsynqWorker) EnqueueDeleteTask(ctx context.Context, docID string) error {
	payload, err := json.Marshal(dto.DeleteDocumentRequest{
		DocumentID: docID,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal delete document request: %w", err)
	}

	task := asynq.NewTask(TaskTypeDeleteDocument, payload)
	info, err := w.client.EnqueueContext(ctx, task,
		asynq.Queue("default"),
		asynq.MaxRetry(3),
		asynq.Timeout(5*time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue delete task for doc %s: %w", docID, err)
	}

	zap.L().Info("enqueued delete task",
		zap.String("doc_id", docID),
		zap.String("task_id", info.ID),
	)
	return nil
}

// handleEmbedDocument processes an embedding task.
func (w *AsynqWorker) handleEmbedDocument(ctx context.Context, t *asynq.Task) error {
	var req dto.IndexDocumentRequest
	if err := json.Unmarshal(t.Payload(), &req); err != nil {
		return fmt.Errorf("failed to unmarshal embed document payload: %w", err)
	}

	zap.L().Info("processing embed document task",
		zap.String("doc_id", req.DocumentID),
		zap.String("space_id", req.SpaceID),
	)

	if err := w.docSvc.IndexDocument(ctx, req); err != nil {
		zap.L().Error("failed to index document",
			zap.String("doc_id", req.DocumentID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to index document %s: %w", req.DocumentID, err)
	}

	zap.L().Info("successfully indexed document", zap.String("doc_id", req.DocumentID))
	return nil
}

// handleDeleteDocument processes a deletion task.
func (w *AsynqWorker) handleDeleteDocument(ctx context.Context, t *asynq.Task) error {
	var req dto.DeleteDocumentRequest
	if err := json.Unmarshal(t.Payload(), &req); err != nil {
		return fmt.Errorf("failed to unmarshal delete document payload: %w", err)
	}

	zap.L().Info("processing delete document task",
		zap.String("doc_id", req.DocumentID),
	)

	if err := w.docSvc.DeleteDocument(ctx, req); err != nil {
		zap.L().Error("failed to delete document chunks",
			zap.String("doc_id", req.DocumentID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete document %s: %w", req.DocumentID, err)
	}

	zap.L().Info("successfully deleted document chunks", zap.String("doc_id", req.DocumentID))
	return nil
}
