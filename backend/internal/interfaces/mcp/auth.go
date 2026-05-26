package mcp

import (
	"context"
	"crypto/subtle"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type contextKey string

const userIDKey contextKey = "mcp_user_id"

// MCPAuth handles API Key authentication for MCP requests.
type MCPAuth struct {
	enabled       bool
	validKeys     [][]byte
	defaultUserID uint64
}

// NewMCPAuth creates a new auth handler. Empty keys disables auth (local dev mode).
func NewMCPAuth(keys []string, defaultUserID uint64) *MCPAuth {
	validKeys := make([][]byte, 0, len(keys))
	for _, k := range keys {
		trimmed := strings.TrimSpace(k)
		if trimmed != "" {
			validKeys = append(validKeys, []byte(trimmed))
		}
	}
	if len(validKeys) == 0 {
		zap.L().Warn("MCP auth disabled: no API keys configured. All requests will use default user.")
	}

	return &MCPAuth{
		enabled:       len(validKeys) > 0,
		validKeys:     validKeys,
		defaultUserID: defaultUserID,
	}
}

// Authenticate validates the API key and returns the user ID.
// Returns default user ID when auth is disabled.
func (a *MCPAuth) Authenticate(apiKey string) (uint64, error) {
	if !a.enabled {
		return a.defaultUserID, nil
	}
	if apiKey == "" {
		return 0, fmt.Errorf("API key is required")
	}
	apiKeyBytes := []byte(apiKey)
	for _, valid := range a.validKeys {
		if subtle.ConstantTimeCompare(apiKeyBytes, valid) == 1 {
			return a.defaultUserID, nil
		}
	}
	return 0, fmt.Errorf("invalid API key")
}

// WithUserID injects the userID into the context.
func WithUserID(ctx context.Context, userID uint64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID extracts the userID from context. Returns defaultUserID if not found.
func GetUserID(ctx context.Context, defaultUserID uint64) uint64 {
	if v, ok := ctx.Value(userIDKey).(uint64); ok {
		return v
	}
	return defaultUserID
}

// UserID returns the default user ID for authenticated requests.
func (a *MCPAuth) UserID() uint64 {
	return a.defaultUserID
}

// HTTPMiddleware returns an HTTP middleware that validates API keys.
// If auth is disabled, the request passes through without checks.
func (a *MCPAuth) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.enabled {
			next.ServeHTTP(w, r)
			return
		}

		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			http.Error(w, "API key required", http.StatusUnauthorized)
			return
		}

		if _, err := a.Authenticate(apiKey); err != nil {
			http.Error(w, "Invalid API key", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
