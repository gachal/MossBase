package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config holds all configuration for the RAG service.
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Qdrant    QdrantConfig    `mapstructure:"qdrant"`
	Embedding EmbeddingConfig `mapstructure:"embedding"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Auth      AuthConfig      `mapstructure:"auth"`
	Log       LogConfig       `mapstructure:"log"`
	Chunker   ChunkerConfig   `mapstructure:"chunker"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port               int      `mapstructure:"port"`
	Mode               string   `mapstructure:"mode"`
	CORSAllowedOrigins []string `mapstructure:"cors_allowed_origins"`
}

// QdrantConfig holds Qdrant vector database connection settings.
type QdrantConfig struct {
	Host             string `mapstructure:"host"`
	Port             int    `mapstructure:"port"`
	APIKey           string `mapstructure:"api_key"`
	CollectionPrefix string `mapstructure:"collection_prefix"`
}

// EmbeddingConfig holds embedding provider settings.
type EmbeddingConfig struct {
	Provider   string `mapstructure:"provider"`
	Model      string `mapstructure:"model"`
	BaseURL    string `mapstructure:"base_url"`
	APIKey     string `mapstructure:"api_key"`
	Dimensions int    `mapstructure:"dimensions"`
	BatchSize  int    `mapstructure:"batch_size"`
	MaxRetries int    `mapstructure:"max_retries"`
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// AuthConfig holds API key authentication settings.
type AuthConfig struct {
	APIKeys []string `mapstructure:"api_keys"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Output string `mapstructure:"output"`
}

// ChunkerConfig holds text chunking settings.
type ChunkerConfig struct {
	ChunkSize    int `mapstructure:"chunk_size"`
	ChunkOverlap int `mapstructure:"chunk_overlap"`
}

// Load reads configuration from file and environment variables.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath("/app/configs")
	}

	// Allow environment variable overrides
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Override sensitive fields from environment variables if set
	if apiKey := os.Getenv("QDRANT_API_KEY"); apiKey != "" {
		cfg.Qdrant.APIKey = apiKey
	}
	if apiKey := os.Getenv("EMBEDDING_API_KEY"); apiKey != "" {
		cfg.Embedding.APIKey = apiKey
	}
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		cfg.Redis.Addr = addr
	}

	return &cfg, nil
}
