package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"sync"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/gachal/mossbase/backend/internal/application/dto"
	"github.com/gachal/mossbase/backend/internal/infrastructure/config"
	"github.com/gachal/mossbase/backend/pkg/hash"
	"github.com/gachal/mossbase/backend/pkg/migration"
)

var installMu sync.Mutex

var dbNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,63}$`)

func validateDBName(name string) error {
	if !dbNameRegex.MatchString(name) {
		return fmt.Errorf("invalid database name: must match [a-zA-Z_][a-zA-Z0-9_]{0,63}")
	}
	return nil
}

type InstallService struct{}

func NewInstallService() *InstallService {
	return &InstallService{}
}

func (s *InstallService) GetStatus() bool {
	return config.IsInstalled()
}

func (s *InstallService) TestDatabase(ctx context.Context, input dto.DatabaseInput) (*dto.TestDBResponse, error) {
	if err := validateDBName(input.DBName); err != nil {
		return &dto.TestDBResponse{Connected: false, Error: err.Error()}, nil
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local",
		input.Username, input.Password, input.Host, input.Port)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return &dto.TestDBResponse{
			Connected: false,
			Error:     fmt.Sprintf("connect failed: %v", err),
		}, nil
	}
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	sqlDB, err := db.DB()
	if err != nil {
		return &dto.TestDBResponse{
			Connected: false,
			Error:     fmt.Sprintf("get sql.DB: %v", err),
		}, nil
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return &dto.TestDBResponse{
			Connected: false,
			Error:     fmt.Sprintf("ping failed: %v", err),
		}, nil
	}

	var version string
	db.Raw("SELECT VERSION()").Scan(&version)

	return &dto.TestDBResponse{
		Connected: true,
		Version:   version,
	}, nil
}

func (s *InstallService) Execute(ctx context.Context, req dto.InstallRequest) error {
	installMu.Lock()
	defer installMu.Unlock()

	if config.IsInstalled() {
		return fmt.Errorf("system is already installed")
	}

	if err := validateDBName(req.Database.DBName); err != nil {
		return err
	}

	db, err := s.connectDB(req.Database)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	if err := s.ensureDatabase(db, req.Database.DBName); err != nil {
		return fmt.Errorf("ensure database: %w", err)
	}

	targetDB, err := s.connectTargetDB(req.Database)
	if err != nil {
		return fmt.Errorf("connect target database: %w", err)
	}
	defer func() {
		sqlDB, _ := targetDB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	if s.checkExistingData(targetDB) {
		zap.L().Info("database already contains data, skipping migration and admin creation")
	} else {
		if err := migration.Run(targetDB, "migrations"); err != nil {
			return fmt.Errorf("run migrations: %w", err)
		}

		if err := s.createAdminUser(targetDB, req.Admin); err != nil {
			return fmt.Errorf("create admin user: %w", err)
		}
	}

	secret, err := generateJWTSecret()
	if err != nil {
		return fmt.Errorf("generate jwt secret: %w", err)
	}

	cfg := s.buildConfig(req, secret)

	if err := config.Save("configs/config.yaml", cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	if err := config.MarkInstalled(); err != nil {
		return fmt.Errorf("mark installed: %w", err)
	}

	zap.L().Info("installation completed successfully")
	return nil
}

func (s *InstallService) connectDB(input dto.DatabaseInput) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local",
		input.Username, input.Password, input.Host, input.Port)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
}

func (s *InstallService) ensureDatabase(db *gorm.DB, dbname string) error {
	return db.Exec(fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		dbname,
	)).Error
}

func (s *InstallService) connectTargetDB(input dto.DatabaseInput) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		input.Username, input.Password, input.Host, input.Port, input.DBName)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
}

func (s *InstallService) createAdminUser(db *gorm.DB, admin dto.AdminInput) error {
	passwordHash, err := hash.HashPassword(admin.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return db.Exec(`
		INSERT INTO users (email, username, password_hash, role, status)
		VALUES (?, ?, ?, 'admin', 'active')
	`, admin.Email, admin.Username, passwordHash).Error
}

func (s *InstallService) checkExistingData(db *gorm.DB) bool {
	var count int64
	db.Raw("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
	return count > 0
}

func (s *InstallService) buildConfig(req dto.InstallRequest, jwtSecret string) *config.Config {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8033,
			Mode: "release",
		},
		Database: config.DatabaseConfig{
			Host:         req.Database.Host,
			Port:         req.Database.Port,
			Username:     req.Database.Username,
			Password:     req.Database.Password,
			DBName:       req.Database.DBName,
			Charset:      "utf8mb4",
			MaxIdleConns: 10,
			MaxOpenConns: 100,
			LogLevel:     "warn",
		},
		JWT: config.JWTConfig{
			Secret:      jwtSecret,
			ExpiryHours: 24,
		},
		Log: config.LogConfig{
			Level:  "info",
			Output: "stdout",
		},
		RAG: config.RAGConfig{
			Enabled: false,
			Timeout: 30,
		},
		MCP: config.MCPConfig{
			Enabled:   false,
			Transport: "stdio",
			HTTPPort:  8095,
		},
		Upload: config.UploadConfig{
			Dir:       "./uploads",
			MaxSizeMB: 5,
			BaseURL:   "/uploads",
		},
	}

	if req.RAG != nil && req.RAG.Enabled {
		cfg.RAG = config.RAGConfig{
			Enabled: true,
			BaseURL: req.RAG.BaseURL,
			APIKey:  req.RAG.APIKey,
			Timeout: 30,
		}
	}

	if req.MCP != nil && req.MCP.Enabled {
		apiKeys := req.MCP.APIKeys
		if len(apiKeys) == 0 {
			apiKeys = []string{}
		}
		cfg.MCP = config.MCPConfig{
			Enabled:   true,
			Transport: req.MCP.Transport,
			HTTPPort:  req.MCP.HTTPPort,
			APIKeys:   apiKeys,
		}
	}

	return cfg
}

func generateJWTSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
