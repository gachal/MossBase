package qdrant

import (
	"fmt"

	pb "github.com/qdrant/go-client/qdrant"

	"github.com/gachal/mossbase/services/rag/internal/infrastructure/config"
)

// QdrantClient wraps the Qdrant high-level client for vector database operations.
type QdrantClient struct {
	client *pb.Client
	config config.QdrantConfig
}

// NewQdrantClient creates a new Qdrant client connection using the official go-client.
func NewQdrantClient(cfg config.QdrantConfig) (*QdrantClient, error) {
	client, err := pb.NewClient(&pb.Config{
		Host: cfg.Host,
		Port: cfg.Port,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to qdrant at %s:%d: %w", cfg.Host, cfg.Port, err)
	}

	return &QdrantClient{
		client: client,
		config: cfg,
	}, nil
}

// Close releases the Qdrant client connection.
func (c *QdrantClient) Close() {
	if c.client != nil {
		_ = c.client.Close()
	}
}

// GetClient returns the underlying Qdrant client.
func (c *QdrantClient) GetClient() *pb.Client {
	return c.client
}
