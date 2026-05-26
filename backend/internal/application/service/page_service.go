package service

import (
	"context"

	"github.com/gachal/mossbase/backend/internal/application/dto"
	"github.com/gachal/mossbase/backend/internal/infrastructure/rag"
)

// RAGIndexer defines the RAG operations needed by PageService.
type RAGIndexer interface {
	IndexDocument(ctx context.Context, docID string, spaceID uint64, title string, content string) error
	DeleteDocument(ctx context.Context, docID string, spaceID uint64) error
	SemanticSearch(ctx context.Context, spaceID uint64, query string, limit int) (*rag.RAGSearchResponse, error)
}

type PageService interface {
	Create(ctx context.Context, spaceID, userID uint64, req dto.CreatePageRequest) (*dto.PageResponse, error)
	Update(ctx context.Context, pageID, userID uint64, req dto.UpdatePageRequest) (*dto.PageResponse, error)
	Delete(ctx context.Context, pageID uint64) error
	GetByID(ctx context.Context, pageID uint64) (*dto.PageResponse, error)
	GetTreeBySpace(ctx context.Context, spaceID uint64) ([]*dto.PageTreeResponse, error)
	MovePage(ctx context.Context, pageID uint64, req dto.MovePageRequest) (*dto.PageResponse, error)
	Search(ctx context.Context, spaceID uint64, query string, page, pageSize int) (*dto.SearchResultResponse, error)
	SemanticSearch(ctx context.Context, spaceID uint64, query string, limit int) (*dto.SemanticSearchResponse, error)
}
