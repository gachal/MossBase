package rag

// RAGSearchResponse represents the response from the RAG semantic search API.
type RAGSearchResponse struct {
	Results []RAGSearchResultItem `json:"results"`
}

// RAGSearchResultItem represents a single result item from semantic search.
type RAGSearchResultItem struct {
	DocumentID string  `json:"document_id"`
	Title      string  `json:"title"`
	Content    string  `json:"content"`
	Score      float32 `json:"score"`
}
