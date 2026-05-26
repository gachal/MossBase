package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server" yaml:"server"`
	Database DatabaseConfig `mapstructure:"database" yaml:"database"`
	JWT      JWTConfig      `mapstructure:"jwt" yaml:"jwt"`
	Log      LogConfig      `mapstructure:"log" yaml:"log"`
	RAG      RAGConfig      `mapstructure:"rag" yaml:"rag"`
	MCP      MCPConfig      `mapstructure:"mcp" yaml:"mcp"`
}

type RAGConfig struct {
	Enabled bool   `mapstructure:"enabled" yaml:"enabled"`
	BaseURL string `mapstructure:"base_url" yaml:"base_url"`
	APIKey  string `mapstructure:"api_key" yaml:"api_key"`
	Timeout int    `mapstructure:"timeout" yaml:"timeout"`
}

type MCPConfig struct {
	Enabled       bool     `mapstructure:"enabled" yaml:"enabled"`
	Transport     string   `mapstructure:"transport" yaml:"transport"`
	HTTPPort      int      `mapstructure:"http_port" yaml:"http_port"`
	APIKeys       []string `mapstructure:"api_keys" yaml:"api_keys"`
	DefaultUserID uint64   `mapstructure:"default_user_id" yaml:"default_user_id"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port" yaml:"port"`
	Mode string `mapstructure:"mode" yaml:"mode"`
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host" yaml:"host"`
	Port         int    `mapstructure:"port" yaml:"port"`
	Username     string `mapstructure:"username" yaml:"username"`
	Password     string `mapstructure:"password" yaml:"password"`
	DBName       string `mapstructure:"dbname" yaml:"dbname"`
	Charset      string `mapstructure:"charset" yaml:"charset"`
	MaxIdleConns int    `mapstructure:"max_idle_conns" yaml:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns" yaml:"max_open_conns"`
	LogLevel     string `mapstructure:"log_level" yaml:"log_level"`
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		d.Username, d.Password, d.Host, d.Port, d.DBName, d.Charset)
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret" yaml:"secret"`
	ExpiryHours int    `mapstructure:"expiry_hours" yaml:"expiry_hours"`
}

type LogConfig struct {
	Level  string `mapstructure:"level" yaml:"level"`
	Output string `mapstructure:"output" yaml:"output"`
}

func Load(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetEnvPrefix("MOSS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	zap.L().Info("config loaded", zap.String("path", configPath))
	return &cfg, nil
}

func LoadMinimal() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Port: 8033,
			Mode: "release",
		},
		Log: LogConfig{
			Level:  "info",
			Output: "stdout",
		},
		Database: DatabaseConfig{
			Charset:      "utf8mb4",
			MaxIdleConns: 10,
			MaxOpenConns: 100,
			LogLevel:     "warn",
		},
		JWT: JWTConfig{
			ExpiryHours: 24,
		},
	}, nil
}
