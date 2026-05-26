package repository

import (
	"context"

	"github.com/gachal/mossbase/services/rag/internal/domain/entity"
)

// VectorRepository defines the interface for vector storage operations.
type VectorRepository interface {
	// CreateCollection creates a new vector collection with the given name and vector size.
	CreateCollection(ctx context.Context, name string, vectorSize uint64) error

	// ListCollections returns all collection names.
	ListCollections(ctx context.Context) ([]string, error)

	// CollectionExists checks whether a collection with the given name exists.
	CollectionExists(ctx context.Context, name string) (bool, error)

	// UpsertPoints inserts or updates vector points in the specified collection.
	UpsertPoints(ctx context.Context, collectionName string, chunks []entity.Chunk) error

	// DeletePointsByDocID deletes all points belonging to the given document ID.
	DeletePointsByDocID(ctx context.Context, collectionName string, docID string) error

	// Search performs a similarity search and returns scored results.
	Search(ctx context.Context, collectionName string, queryVector []float32, limit int, filter map[string]string) ([]entity.SearchResult, error)
}
