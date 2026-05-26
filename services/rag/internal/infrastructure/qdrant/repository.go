package qdrant

import (
	"context"
	"fmt"

	pb "github.com/qdrant/go-client/qdrant"
	"go.uber.org/zap"

	"github.com/gachal/mossbase/services/rag/internal/domain/entity"
	"github.com/gachal/mossbase/services/rag/internal/domain/repository"
)

// QdrantRepository implements repository.VectorRepository using the Qdrant Go client.
type QdrantRepository struct {
	client *QdrantClient
}

// NewQdrantRepository creates a new QdrantRepository.
func NewQdrantRepository(client *QdrantClient) repository.VectorRepository {
	return &QdrantRepository{
		client: client,
	}
}

// CreateCollection creates a new vector collection with cosine distance.
func (r *QdrantRepository) CreateCollection(ctx context.Context, name string, vectorSize uint64) error {
	err := r.client.GetClient().CreateCollection(ctx, &pb.CreateCollection{
		CollectionName: name,
		VectorsConfig: &pb.VectorsConfig{
			Config: &pb.VectorsConfig_Params{
				Params: &pb.VectorParams{
					Size:     vectorSize,
					Distance: pb.Distance_Cosine,
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create collection %s: %w", name, err)
	}

	zap.L().Info("created qdrant collection", zap.String("name", name), zap.Uint64("vector_size", vectorSize))
	return nil
}

// ListCollections returns all collection names from Qdrant.
func (r *QdrantRepository) ListCollections(ctx context.Context) ([]string, error) {
	names, err := r.client.GetClient().ListCollections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	return names, nil
}

// CollectionExists checks whether a collection with the given name exists.
func (r *QdrantRepository) CollectionExists(ctx context.Context, name string) (bool, error) {
	exists, err := r.client.GetClient().CollectionExists(ctx, name)
	if err != nil {
		return false, fmt.Errorf("failed to check collection existence: %w", err)
	}

	return exists, nil
}

// UpsertPoints converts entity.Chunk slices to Qdrant point structs and upserts them.
func (r *QdrantRepository) UpsertPoints(ctx context.Context, collectionName string, chunks []entity.Chunk) error {
	points := make([]*pb.PointStruct, 0, len(chunks))

	for i := range chunks {
		chunk := chunks[i]

		payload := map[string]*pb.Value{
			"doc_id":      pb.NewValueString(chunk.DocumentID),
			"chunk_index": pb.NewValueInt(int64(chunk.ChunkIndex)),
			"title":       pb.NewValueString(chunk.Title),
			"content":     pb.NewValueString(chunk.Content),
		}

		for k, v := range chunk.Metadata {
			payload[k] = pb.NewValueString(v)
		}

		point := &pb.PointStruct{
			Id:      pb.NewIDUUID(chunk.ID),
			Vectors: pb.NewVectors(chunk.Embedding...),
			Payload: payload,
		}

		points = append(points, point)
	}

	_, err := r.client.GetClient().Upsert(ctx, &pb.UpsertPoints{
		CollectionName: collectionName,
		Points:         points,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert points into %s: %w", collectionName, err)
	}

	zap.L().Info("upserted points", zap.String("collection", collectionName), zap.Int("count", len(points)))
	return nil
}

// DeletePointsByDocID deletes all points matching the given document ID.
func (r *QdrantRepository) DeletePointsByDocID(ctx context.Context, collectionName string, docID string) error {
	_, err := r.client.GetClient().Delete(ctx, &pb.DeletePoints{
		CollectionName: collectionName,
		Points: &pb.PointsSelector{
			PointsSelectorOneOf: &pb.PointsSelector_Filter{
				Filter: &pb.Filter{
					Must: []*pb.Condition{
						{
							ConditionOneOf: &pb.Condition_Field{
								Field: &pb.FieldCondition{
									Key: "doc_id",
									Match: &pb.Match{
										MatchValue: &pb.Match_Keyword{
											Keyword: docID,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete points by doc_id %s from %s: %w", docID, collectionName, err)
	}

	zap.L().Info("deleted points by doc_id", zap.String("collection", collectionName), zap.String("doc_id", docID))
	return nil
}

// Search performs a similarity search and returns scored results.
func (r *QdrantRepository) Search(ctx context.Context, collectionName string, queryVector []float32, limit int, filter map[string]string) ([]entity.SearchResult, error) {
	searchFilter := buildFilter(filter)

	results, err := r.client.GetClient().Query(ctx, &pb.QueryPoints{
		CollectionName: collectionName,
		Query:          pb.NewQuery(queryVector...),
		Limit:          pb.PtrOf(uint64(limit)),
		WithPayload:    pb.NewWithPayload(true),
		Filter:         searchFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search collection %s: %w", collectionName, err)
	}

	searchResults := make([]entity.SearchResult, 0, len(results))
	for _, scored := range results {
		chunk := scoredPointToChunk(scored)
		searchResults = append(searchResults, entity.SearchResult{
			Chunk: chunk,
			Score: scored.GetScore(),
		})
	}

	return searchResults, nil
}

// buildFilter constructs a Qdrant filter from a string map.
func buildFilter(filter map[string]string) *pb.Filter {
	if len(filter) == 0 {
		return nil
	}

	conditions := make([]*pb.Condition, 0, len(filter))
	for k, v := range filter {
		conditions = append(conditions, &pb.Condition{
			ConditionOneOf: &pb.Condition_Field{
				Field: &pb.FieldCondition{
					Key: k,
					Match: &pb.Match{
						MatchValue: &pb.Match_Keyword{
							Keyword: v,
						},
					},
				},
			},
		})
	}

	return &pb.Filter{Must: conditions}
}

// scoredPointToChunk converts a Qdrant scored point to an entity.Chunk.
func scoredPointToChunk(sp *pb.ScoredPoint) entity.Chunk {
	payload := sp.GetPayload()

	chunk := entity.Chunk{}

	if id, ok := payload["doc_id"]; ok {
		chunk.DocumentID = id.GetStringValue()
	}
	if idx, ok := payload["chunk_index"]; ok {
		chunk.ChunkIndex = int(idx.GetIntegerValue())
	}
	if title, ok := payload["title"]; ok {
		chunk.Title = title.GetStringValue()
	}
	if content, ok := payload["content"]; ok {
		chunk.Content = content.GetStringValue()
	}

	// Extract remaining payload fields as metadata (excluding core fields)
	coreFields := map[string]bool{"doc_id": true, "chunk_index": true, "title": true, "content": true}
	chunk.Metadata = make(map[string]string)
	for k, v := range payload {
		if !coreFields[k] {
			chunk.Metadata[k] = v.GetStringValue()
		}
	}

	return chunk
}
