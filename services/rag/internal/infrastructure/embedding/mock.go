package embedding

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
)

// MockEmbeddingProvider returns deterministic pseudo-random vectors for testing.
type MockEmbeddingProvider struct {
	Dimensions int
}

// NewMockEmbeddingProvider creates a mock embedding provider with fixed dimensions.
func NewMockEmbeddingProvider(dimensions int) *MockEmbeddingProvider {
	return &MockEmbeddingProvider{
		Dimensions: dimensions,
	}
}

// GetEmbedding returns a deterministic pseudo-random vector for a single text.
func (p *MockEmbeddingProvider) GetEmbedding(_ context.Context, text string) ([]float32, error) {
	return p.generateVector(text), nil
}

// GetEmbeddings returns deterministic pseudo-random vectors for multiple texts.
func (p *MockEmbeddingProvider) GetEmbeddings(_ context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	results := make([][]float32, len(texts))
	for i, text := range texts {
		results[i] = p.generateVector(text)
	}

	return results, nil
}

// generateVector creates a deterministic pseudo-random vector based on text hash.
func (p *MockEmbeddingProvider) generateVector(text string) []float32 {
	vec := make([]float32, p.Dimensions)

	// Generate enough hash bytes to fill the vector
	// 4 bytes per float32, 32 bytes per sha256 hash
	hashCount := (p.Dimensions*4 + 31) / 32
	hashBytes := make([]byte, 0, hashCount*32)

	for i := 0; i < hashCount; i++ {
		seed := fmt.Sprintf("%s:%d", text, i)
		h := sha256.Sum256([]byte(seed))
		hashBytes = append(hashBytes, h[:]...)
	}

	// Convert hash bytes to float32 values in range [-1, 1]
	for i := 0; i < p.Dimensions; i++ {
		offset := i * 4
		if offset+4 > len(hashBytes) {
			offset = 0
		}
		bits := binary.LittleEndian.Uint32(hashBytes[offset : offset+4])
		// Normalize to [-1, 1] range
		vec[i] = float32(bits) / float32(0xFFFFFFFF)
	}

	// Normalize vector to unit length (cosine similarity compatibility)
	return normalizeVector(vec)
}

// normalizeVector scales a vector to unit length.
func normalizeVector(vec []float32) []float32 {
	var sumSq float64
	for _, v := range vec {
		sumSq += float64(v * v)
	}

	if sumSq == 0 {
		return vec
	}

	length := math.Sqrt(sumSq)
	result := make([]float32, len(vec))
	for i, v := range vec {
		result[i] = v / float32(length)
	}

	return result
}
