package entity

import "time"

// Chunk represents a text segment extracted from a document for vector embedding.
type Chunk struct {
	ID         string
	DocumentID string
	ChunkIndex int
	Title      string
	Content    string
	Embedding  []float32
	Metadata   map[string]string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// SearchResult represents a scored chunk returned from vector search.
type SearchResult struct {
	Chunk Chunk
	Score float32
}
