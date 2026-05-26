package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/gachal/mossbase/backend/internal/infrastructure/config"
)

// RAGClientInterface defines the contract for RAG service communication.
type RAGClientInterface interface {
	IndexDocument(ctx context.Context, docID string, spaceID uint64, title string, content string) error
	DeleteDocument(ctx context.Context, docID string, spaceID uint64) error
	SemanticSearch(ctx context.Context, spaceID uint64, query string, limit int) (*RAGSearchResponse, error)
}

// RAGClient is an HTTP client for the RAG microservice.
type RAGClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewRAGClient creates a new RAGClient from the given configuration.
// Returns nil if the baseURL is invalid.
func NewRAGClient(cfg config.RAGConfig) *RAGClient {
	if cfg.BaseURL == "" {
		return nil
	}
	parsed, err := url.Parse(cfg.BaseURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		zap.L().Warn("RAG client: invalid base_url, RAG disabled", zap.String("base_url", cfg.BaseURL))
		return nil
	}

	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &RAGClient{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// indexRequest is the JSON body for the index document API.
type indexRequest struct {
	DocumentID string `json:"document_id"`
	SpaceID    uint64 `json:"space_id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
}

// searchRequest is the JSON body for the semantic search API.
type searchRequest struct {
	SpaceID uint64 `json:"space_id"`
	Query   string `json:"query"`
	TopK    int    `json:"top_k"`
}

// IndexDocument sends a document to the RAG service for indexing.
// Errors are logged but not returned to avoid blocking page operations.
func (c *RAGClient) IndexDocument(ctx context.Context, docID string, spaceID uint64, title string, content string) error {
	body := indexRequest{
		DocumentID: docID,
		SpaceID:    spaceID,
		Title:      title,
		Content:    content,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		zap.L().Warn("RAG IndexDocument: marshal error", zap.Error(err))
		return nil
	}

	url := fmt.Sprintf("%s/api/v1/documents", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		zap.L().Warn("RAG IndexDocument: create request error", zap.Error(err))
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		zap.L().Warn("RAG IndexDocument: request error",
			zap.String("doc_id", docID),
			zap.Error(err),
		)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		zap.L().Warn("RAG IndexDocument: non-success status",
			zap.String("doc_id", docID),
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(respBody)),
		)
		return nil
	}

	zap.L().Debug("RAG IndexDocument: success", zap.String("doc_id", docID))
	return nil
}

// DeleteDocument removes a document from the RAG service index.
// Errors are logged but not returned to avoid blocking page operations.
func (c *RAGClient) DeleteDocument(ctx context.Context, docID string, spaceID uint64) error {
	encodedDocID := url.PathEscape(docID)
	reqURL := fmt.Sprintf("%s/api/v1/documents/%s?space_id=%d", c.baseURL, encodedDocID, spaceID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		zap.L().Warn("RAG DeleteDocument: create request error", zap.Error(err))
		return nil
	}
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		zap.L().Warn("RAG DeleteDocument: request error",
			zap.String("doc_id", docID),
			zap.Error(err),
		)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		zap.L().Warn("RAG DeleteDocument: non-success status",
			zap.String("doc_id", docID),
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(respBody)),
		)
		return nil
	}

	zap.L().Debug("RAG DeleteDocument: success", zap.String("doc_id", docID))
	return nil
}

// SemanticSearch performs a semantic search against the RAG service.
// This method returns errors since the caller needs to know if search failed.
func (c *RAGClient) SemanticSearch(ctx context.Context, spaceID uint64, query string, limit int) (*RAGSearchResponse, error) {
	body := searchRequest{
		SpaceID: spaceID,
		Query:   query,
		TopK:    limit,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("RAG SemanticSearch: marshal error: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/search", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("RAG SemanticSearch: create request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("RAG SemanticSearch: request error: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("RAG SemanticSearch: read response error: %w", err)
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("RAG SemanticSearch: non-success status %d: %s", resp.StatusCode, string(respBody))
	}

	var result RAGSearchResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("RAG SemanticSearch: unmarshal error: %w", err)
	}

	return &result, nil
}
