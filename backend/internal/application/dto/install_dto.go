package dto

type InstallRequest struct {
	Database DatabaseInput `json:"database" binding:"required"`
	Admin    AdminInput    `json:"admin" binding:"required"`
	MCP      *MCPInput     `json:"mcp"`
	RAG      *RAGInput     `json:"rag"`
}

type DatabaseInput struct {
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port" binding:"required,min=1,max=65535"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password"`
	DBName   string `json:"dbname" binding:"required"`
}

type AdminInput struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=2,max=50"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

type MCPInput struct {
	Enabled   bool     `json:"enabled"`
	Transport string   `json:"transport"`
	HTTPPort  int      `json:"http_port"`
	APIKeys   []string `json:"api_keys"`
}

type RAGInput struct {
	Enabled bool   `json:"enabled"`
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
}

type InstallStatusResponse struct {
	Installed bool `json:"installed"`
}

type TestDBResponse struct {
	Connected bool   `json:"connected"`
	Version   string `json:"version,omitempty"`
	Error     string `json:"error,omitempty"`
}
