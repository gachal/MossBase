package service

import (
	"context"

	"github.com/gachal/mossbase/services/rag/internal/application/dto"
)

// DocumentService defines the interface for document indexing and search operations.
type DocumentService interface {
	// IndexDocument chunks a document and stores its embeddings in the vector store.
	IndexDocument(ctx context.Context, req dto.IndexDocumentRequest) error

	// DeleteDocument removes all chunks for a document from the vector store.
	DeleteDocument(ctx context.Context, req dto.DeleteDocumentRequest) error

	// Search performs a similarity search across chunks in the given space.
	Search(ctx context.Context, req dto.SearchRequest) (*dto.SearchResponse, error)
}
