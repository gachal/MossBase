package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/gachal/mossbase/services/rag/internal/infrastructure/config"
)

// OpenAIEmbeddingProvider calls an OpenAI-compatible API to generate embeddings.
type OpenAIEmbeddingProvider struct {
	baseURL    string
	apiKey     string
	model      string
	dimensions int
	maxRetries int
	httpClient *http.Client
}

// embeddingRequest is the request body sent to the embedding API.
type embeddingRequest struct {
	Model      string   `json:"model"`
	Input      []string `json:"input"`
	Dimensions int      `json:"dimensions,omitempty"`
}

// embeddingResponse is the response body from the embedding API.
type embeddingResponse struct {
	Data []embeddingData `json:"data"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// embeddingData holds a single embedding vector from the API response.
type embeddingData struct {
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// apiErrorResponse represents an error response from the API.
type apiErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// NewOpenAIEmbeddingProvider creates a new OpenAI-compatible embedding client.
func NewOpenAIEmbeddingProvider(cfg config.EmbeddingConfig) *OpenAIEmbeddingProvider {
	return &OpenAIEmbeddingProvider{
		baseURL:    cfg.BaseURL,
		apiKey:     cfg.APIKey,
		model:      cfg.Model,
		dimensions: cfg.Dimensions,
		maxRetries: cfg.MaxRetries,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetEmbedding generates an embedding vector for a single text input.
func (p *OpenAIEmbeddingProvider) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	results, err := p.GetEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no embedding returned for text")
	}

	return results[0], nil
}

// GetEmbeddings generates embedding vectors for multiple text inputs, splitting into batches of 100.
func (p *OpenAIEmbeddingProvider) GetEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	const batchSize = 100
	allEmbeddings := make([][]float32, 0, len(texts))

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		embeddings, err := p.callEmbeddingAPI(ctx, batch)
		if err != nil {
			return nil, fmt.Errorf("embedding batch starting at index %d: %w", i, err)
		}

		allEmbeddings = append(allEmbeddings, embeddings...)
	}

	return allEmbeddings, nil
}

// callEmbeddingAPI makes the HTTP request with retry logic.
func (p *OpenAIEmbeddingProvider) callEmbeddingAPI(ctx context.Context, texts []string) ([][]float32, error) {
	var lastErr error

	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
			zap.L().Warn("retrying embedding API call",
				zap.Int("attempt", attempt),
				zap.Duration("backoff", backoff),
			)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		embeddings, err := p.doRequest(ctx, texts)
		if err == nil {
			return embeddings, nil
		}

		lastErr = err
		zap.L().Warn("embedding API call failed",
			zap.Int("attempt", attempt),
			zap.Error(err),
		)
	}

	return nil, fmt.Errorf("embedding API failed after %d retries: %w", p.maxRetries, lastErr)
}

// doRequest performs a single HTTP request to the embedding API.
func (p *OpenAIEmbeddingProvider) doRequest(ctx context.Context, texts []string) ([][]float32, error) {
	reqBody := embeddingRequest{
		Model:      p.model,
		Input:      texts,
		Dimensions: p.dimensions,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	url := fmt.Sprintf("%s/embeddings", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call embedding API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp apiErrorResponse
		if jsonErr := json.Unmarshal(respBody, &errResp); jsonErr == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("embedding API error (status %d): %s", resp.StatusCode, errResp.Error.Message)
		}
		return nil, fmt.Errorf("embedding API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var embResp embeddingResponse
	if err := json.Unmarshal(respBody, &embResp); err != nil {
		return nil, fmt.Errorf("failed to decode embedding response: %w", err)
	}

	// Convert float64 to float32 and preserve order
	embeddings := make([][]float32, len(texts))
	for _, data := range embResp.Data {
		if data.Index < 0 || data.Index >= len(embeddings) {
			return nil, fmt.Errorf("unexpected embedding index %d", data.Index)
		}
		vec := make([]float32, len(data.Embedding))
		for j, v := range data.Embedding {
			vec[j] = float32(v)
		}
		embeddings[data.Index] = vec
	}

	for i, emb := range embeddings {
		if emb == nil {
			return nil, fmt.Errorf("missing embedding for input index %d", i)
		}
	}

	return embeddings, nil
}
