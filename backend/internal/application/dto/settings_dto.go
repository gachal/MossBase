package dto

const (
	SentinelUnchanged = "__UNCHANGED__"
	SentinelClear     = "__CLEAR__"
)

type MCPSettings struct {
	Enabled       bool     `json:"enabled"`
	Transport     string   `json:"transport"`
	HTTPPort      int      `json:"http_port"`
	APIKeys       []string `json:"api_keys"`
	APIKeysMasked bool     `json:"api_keys_masked"`
	DefaultUserID uint64   `json:"default_user_id"`
}

type RAGSettings struct {
	Enabled       bool   `json:"enabled"`
	BaseURL       string `json:"base_url"`
	APIKey        string `json:"api_key"`
	APIKeyMasked  bool   `json:"api_key_masked"`
	Timeout       int    `json:"timeout"`
}

type SettingsResponse struct {
	MCP MCPSettings `json:"mcp"`
	RAG RAGSettings `json:"rag"`
}

type MCPSettingsRequest struct {
	Enabled       bool     `json:"enabled"`
	Transport     string   `json:"transport" binding:"omitempty,oneof=stdio http both"`
	HTTPPort      int      `json:"http_port" binding:"omitempty,min=1,max=65535"`
	APIKeys       []string `json:"api_keys"`
	APIKeysAction string  `json:"api_keys_action" binding:"omitempty,oneof=keep replace clear"`
	DefaultUserID uint64   `json:"default_user_id"`
}

type RAGSettingsRequest struct {
	Enabled bool   `json:"enabled"`
	BaseURL string `json:"base_url" binding:"omitempty,url"`
	APIKey  string `json:"api_key"`
	Timeout int    `json:"timeout" binding:"omitempty,min=1,max=300"`
}

type SettingsRequest struct {
	MCP *MCPSettingsRequest `json:"mcp,omitempty"`
	RAG *RAGSettingsRequest `json:"rag,omitempty"`
}

type TestRAGRequest struct {
	BaseURL    string `json:"base_url" binding:"required,url"`
	APIKey     string `json:"api_key"`
	UseSavedKey bool  `json:"use_saved_key"`
}

type TestRAGResponse struct {
	Connected bool   `json:"connected"`
	Message   string `json:"message"`
}
