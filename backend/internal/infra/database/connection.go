package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"organiq/config"
)

// Connect establishes a connection to the PostgreSQL database
func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	// Configure GORM logger
	gormLogger := logger.Default
	if cfg.IsDevelopment() {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 gormLogger,
		SkipDefaultTransaction: true,
		// CORREÇÃO CRÍTICA AQUI:
		// PrepareStmt deve ser false para evitar o erro "cannot insert multiple commands into a prepared statement"
		// quando usamos migrations com múltiplos comandos (CREATE + INSERT)
		PrepareStmt:            false, 
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Get underlying sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Ping database to verify connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().
		Str("host", cfg.Database.Host).
		Str("port", cfg.Database.Port).
		Str("database", cfg.Database.Name).
		Msg("Database connection pool configured")

	return db, nil
}

// Close closes the database connection
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	log.Info().Msg("Closing database connection")
	return sqlDB.Close()
}

// RunMigrations executes database migrations from SQL files
func RunMigrations(db *gorm.DB) error {
	log.Info().Msg("Running database migrations")

	// Tenta localizar as migrations em diferentes locais
	possiblePaths := []string{
		"migrations", // Relativo ao WORKDIR /app do Docker
		"internal/infra/database/migrations", // Relativo à raiz do projeto em dev
	}

	var migrationsPath string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			migrationsPath = path
			break
		}
	}

	if migrationsPath == "" {
		log.Warn().Msg("Migrations directory not found in known paths. Skipping.")
		return nil
	}

	log.Info().Str("path", migrationsPath).Msg("Migrations directory found")

	// Ler todos os arquivos .sql do diretório
	files, err := filepath.Glob(filepath.Join(migrationsPath, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	if len(files) == 0 {
		log.Warn().Str("path", migrationsPath).Msg("No .sql files found in directory")
		return nil
	}

	// Ordenar os arquivos por nome (001_*, 002_*, etc)
	sort.Strings(files)

	// CORREÇÃO: Obter conexão raw SQL para evitar interferência do GORM nas migrations
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Executar cada migration
	for _, file := range files {
		migrationName := filepath.Base(file)
		
		log.Info().
			Str("migration", migrationName).
			Msg("Executing migration")

		// Ler conteúdo do arquivo
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", migrationName, err)
		}

		// CORREÇÃO: Usar sqlDB.Exec em vez de db.Exec (GORM)
		// Isso garante que múltiplos comandos separados por ";" sejam aceitos
		if _, err := sqlDB.Exec(string(content)); err != nil {
			// Se for erro de "already exists" ou similar, logar como warning
			if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "duplicate key") {
				log.Warn().
					Str("migration", migrationName).
					Str("error", err.Error()).
					Msg("Migration constraint/index likely already applied (skipping)")
				continue
			}
			
			return fmt.Errorf("failed to execute migration %s: %w", migrationName, err)
		}

		log.Info().
			Str("migration", migrationName).
			Msg("Migration executed successfully")
	}

	log.Info().
		Int("migrations_count", len(files)).
		Msg("All migrations completed successfully")
	
	return nil
}

// HealthCheck verifies database connectivity
func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}
