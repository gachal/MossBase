package migration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Run(db *gorm.DB, migrationsDir string) error {
	if err := createSchemaMigrationsTable(db); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}
	sort.Strings(sqlFiles)

	applied := 0
	for _, name := range sqlFiles {
		n, err := applyMigration(db, migrationsDir, name)
		if err != nil {
			return err
		}
		applied += n
	}

	zap.L().Info("migrations completed", zap.Int("applied", applied), zap.Int("total", len(sqlFiles)))
	return nil
}

func createSchemaMigrationsTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`).Error
}

func applyMigration(db *gorm.DB, dir, filename string) (int, error) {
	var count int64
	db.Table("schema_migrations").Where("version = ?", filename).Count(&count)
	if count > 0 {
		return 0, nil
	}

	content, err := os.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		return 0, fmt.Errorf("read migration %s: %w", filename, err)
	}

	tx := db.Begin()
	if tx.Error != nil {
		return 0, fmt.Errorf("begin transaction for %s: %w", filename, tx.Error)
	}

	stmts := splitStatements(string(content))
	for _, stmt := range stmts {
		if err := tx.Exec(stmt).Error; err != nil {
			if isAlreadyExistsError(err) {
				zap.L().Info("migration object already exists, skipping",
					zap.String("version", filename), zap.String("stmt", truncate(stmt, 80)))
				continue
			}
			tx.Rollback()
			return 0, fmt.Errorf("execute migration %s: %w", filename, err)
		}
	}

	if err := tx.Table("schema_migrations").Create(map[string]interface{}{"version": filename}).Error; err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("record migration %s: %w", filename, err)
	}

	if err := tx.Commit().Error; err != nil {
		return 0, fmt.Errorf("commit migration %s: %w", filename, err)
	}

	zap.L().Info("migration applied", zap.String("version", filename))
	return 1, nil
}

func splitStatements(content string) []string {
	var lines []string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}
		lines = append(lines, line)
	}
	joined := strings.Join(lines, "\n")
	var result []string
	for _, part := range strings.Split(joined, ";") {
		if stmt := strings.TrimSpace(part); stmt != "" {
			result = append(result, stmt)
		}
	}
	return result
}

// MySQL error codes for "already exists" scenarios
const (
	errTableExists  = 1050 // CREATE TABLE already exists
	errDuplicateKey = 1061 // Duplicate key name (index already exists)
	errDupFieldName = 1060 // Duplicate column name
)

func isAlreadyExistsError(err error) bool {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		return me.Number == errTableExists || me.Number == errDuplicateKey || me.Number == errDupFieldName
	}
	return false
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
