package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gachal/mossbase/services/rag/internal/application/dto"
	"github.com/gachal/mossbase/services/rag/internal/domain/repository"
	"github.com/gachal/mossbase/services/rag/internal/infrastructure/embedding"
	"github.com/gachal/mossbase/services/rag/pkg/chunker"
)

// DocumentServiceImpl implements DocumentService with vector storage, embedding, and chunking.
type DocumentServiceImpl struct {
	vectorRepo       repository.VectorRepository
	embedder         *embedding.OpenAIEmbeddingProvider
	chunker          *chunker.TextChunker
	collectionPrefix string
	dimensions       int
}

// NewDocumentService creates a new DocumentServiceImpl with the given dependencies.
func NewDocumentService(
	vectorRepo repository.VectorRepository,
	embedder *embedding.OpenAIEmbeddingProvider,
	ch *chunker.TextChunker,
	prefix string,
	dims int,
) DocumentService {
	return &DocumentServiceImpl{
		vectorRepo:       vectorRepo,
		embedder:         embedder,
		chunker:          ch,
		collectionPrefix: prefix,
		dimensions:       dims,
	}
}

// IndexDocument splits the document into chunks, generates embeddings, and upserts them.
func (s *DocumentServiceImpl) IndexDocument(ctx context.Context, req dto.IndexDocumentRequest) error {
	collectionName := fmt.Sprintf("%s-%s", s.collectionPrefix, req.SpaceID)

	// Ensure the collection exists
	exists, err := s.vectorRepo.CollectionExists(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to check collection existence: %w", err)
	}
	if !exists {
		if createErr := s.vectorRepo.CreateCollection(ctx, collectionName, uint64(s.dimensions)); createErr != nil {
			return fmt.Errorf("failed to create collection %s: %w", collectionName, createErr)
		}
	}

	// Chunk the document content
	chunks := s.chunker.Chunk(req.Title, req.Content)
	if len(chunks) == 0 {
		zap.L().Warn("no chunks produced for document",
			zap.String("document_id", req.DocumentID),
			zap.String("space_id", req.SpaceID),
		)
		return nil
	}

	// Use composite doc_id: "page-{spaceID}-{documentID}"
	compositeDocID := fmt.Sprintf("page-%s-%s", req.SpaceID, req.DocumentID)
	now := time.Now()

	// Collect chunk content strings for batch embedding
	texts := make([]string, len(chunks))
	for i, c := range chunks {
		texts[i] = c.Content
	}

	// Generate embeddings in batch
	embeddings, err := s.embedder.GetEmbeddings(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings for document %s: %w", req.DocumentID, err)
	}

	// Assign IDs, doc_id, embeddings, and metadata to each chunk
	for i := range chunks {
		chunks[i].ID = uuid.New().String()
		chunks[i].DocumentID = compositeDocID
		chunks[i].CreatedAt = now
		chunks[i].UpdatedAt = now

		if i < len(embeddings) {
			chunks[i].Embedding = embeddings[i]
		}

		// Merge request metadata into chunk metadata
		if chunks[i].Metadata == nil {
			chunks[i].Metadata = make(map[string]string)
		}
		for k, v := range req.Metadata {
			chunks[i].Metadata[k] = v
		}
		// Always store space_id in metadata for filtering
		chunks[i].Metadata["space_id"] = req.SpaceID
	}

	// Upsert all chunks into the vector store
	if err := s.vectorRepo.UpsertPoints(ctx, collectionName, chunks); err != nil {
		return fmt.Errorf("failed to upsert chunks for document %s: %w", req.DocumentID, err)
	}

	zap.L().Info("indexed document",
		zap.String("document_id", req.DocumentID),
		zap.String("space_id", req.SpaceID),
		zap.Int("chunk_count", len(chunks)),
	)
	return nil
}

// DeleteDocument removes all chunks for a document from the vector store.
func (s *DocumentServiceImpl) DeleteDocument(ctx context.Context, req dto.DeleteDocumentRequest) error {
	collectionName := fmt.Sprintf("%s-%s", s.collectionPrefix, req.SpaceID)
	compositeDocID := fmt.Sprintf("page-%s-%s", req.SpaceID, req.DocumentID)

	if err := s.vectorRepo.DeletePointsByDocID(ctx, collectionName, compositeDocID); err != nil {
		return fmt.Errorf("failed to delete document %s: %w", req.DocumentID, err)
	}

	zap.L().Info("deleted document",
		zap.String("document_id", req.DocumentID),
		zap.String("space_id", req.SpaceID),
	)
	return nil
}

// Search generates a query embedding and performs similarity search against the vector store.
func (s *DocumentServiceImpl) Search(ctx context.Context, req dto.SearchRequest) (*dto.SearchResponse, error) {
	collectionName := fmt.Sprintf("%s-%s", s.collectionPrefix, req.SpaceID)

	// Generate query embedding
	queryEmbedding, err := s.embedder.GetEmbedding(ctx, req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Build filter with space_id
	filter := make(map[string]string)
	filter["space_id"] = req.SpaceID
	for k, v := range req.Filter {
		filter[k] = v
	}

	topK := req.TopK
	if topK <= 0 {
		topK = 10
	}

	// Perform similarity search
	searchResults, err := s.vectorRepo.Search(ctx, collectionName, queryEmbedding, topK, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search collection %s: %w", collectionName, err)
	}

	// Convert entity.SearchResult to dto.ScoredChunk
	scoredChunks := make([]dto.ScoredChunk, 0, len(searchResults))
	for _, result := range searchResults {
		scoredChunks = append(scoredChunks, dto.ScoredChunk{
			ChunkID:    result.Chunk.ID,
			DocumentID: result.Chunk.DocumentID,
			Title:      result.Chunk.Title,
			Content:    result.Chunk.Content,
			Score:      result.Score,
			Metadata:   result.Chunk.Metadata,
		})
	}

	return &dto.SearchResponse{
		Results: scoredChunks,
	}, nil
}
