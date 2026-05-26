package dto

// IndexDocumentRequest is the request to index a document into the vector store.
type IndexDocumentRequest struct {
	DocumentID string            `json:"document_id"`
	SpaceID    string            `json:"space_id"`
	Title      string            `json:"title"`
	Content    string            `json:"content"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// DeleteDocumentRequest is the request to delete a document from the vector store.
type DeleteDocumentRequest struct {
	DocumentID string `json:"document_id"`
	SpaceID    string `json:"space_id"`
}

// SearchRequest is the request to search for similar chunks.
type SearchRequest struct {
	SpaceID string            `json:"space_id"`
	Query   string            `json:"query"`
	TopK    int               `json:"top_k,omitempty"`
	Filter  map[string]string `json:"filter,omitempty"`
}

// SearchResponse is the response for a search query.
type SearchResponse struct {
	Results []ScoredChunk `json:"results"`
}

// ScoredChunk represents a chunk with its relevance score.
type ScoredChunk struct {
	ChunkID    string            `json:"chunk_id"`
	DocumentID string            `json:"document_id"`
	Title      string            `json:"title"`
	Content    string            `json:"content"`
	Score      float32           `json:"score"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}
