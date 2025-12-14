
// ==========================================
// config/env.go
// ==========================================
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	
	"github.com/joho/godotenv"
)

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if exists (development)
	_ = godotenv.Load()
	
	cfg := &Config{
		Environment: getEnv("ENV", "development"),
		
		Server: ServerConfig{
			Port:            getEnv("SERVER_PORT", "8080"),
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:     parseDuration("SERVER_READ_TIMEOUT", "30s"),
			WriteTimeout:    parseDuration("SERVER_WRITE_TIMEOUT", "30s"),
			IdleTimeout:     parseDuration("SERVER_IDLE_TIMEOUT", "120s"),
			ShutdownTimeout: parseDuration("SERVER_SHUTDOWN_TIMEOUT", "30s"),
		},
		
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "organiq"),
			Password:        getEnv("DB_PASSWORD", "dev_password"),
			Name:            getEnv("DB_NAME", "organiq_dev"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    parseInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    parseInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: parseDuration("DB_CONN_MAX_LIFETIME", "5m"),
		},
		
		Storage: StorageConfig{
			Type:            getEnv("STORAGE_TYPE", "minio"),
			MinIOEndpoint:   getEnv("MINIO_ENDPOINT", "localhost:9000"),
			MinIOAccessKey:  getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			MinIOSecretKey:  getEnv("MINIO_SECRET_KEY", "minioadmin"),
			MinIOBucket:     getEnv("MINIO_BUCKET", "organiq-dev-brand-files"),
			MinIOUseSSL:     parseBool("MINIO_USE_SSL", false),
			S3Bucket:        getEnv("S3_BUCKET", "organiq-prod-brand-files"),
			S3Region:        getEnv("S3_REGION", "us-east-1"),
		},
		
		Queue: QueueConfig{
			Endpoint:               getEnv("QUEUE_ENDPOINT", "http://localhost:4566"),
			Region:                 getEnv("AWS_REGION", "us-east-1"),
			AccessKeyID:            getEnv("AWS_ACCESS_KEY_ID", "test"),
			SecretAccessKey:        getEnv("AWS_SECRET_ACCESS_KEY", "test"),
			ArticleGenerationQueue: getEnv("ARTICLE_GENERATION_QUEUE", "article-generation-queue"),
			ArticlePublishQueue:    getEnv("ARTICLE_PUBLISH_QUEUE", "article-publish-queue"),
		},
		
		Auth: AuthConfig{
			JWTSecret:            requireEnv("JWT_SECRET"),
			PasswordPepper:       requireEnv("PASSWORD_PEPPER"),
			AccessTokenDuration:  parseDuration("ACCESS_TOKEN_DURATION", "15m"),
			RefreshTokenDuration: parseDuration("REFRESH_TOKEN_DURATION", "168h"),
		},
		
		CORS: CORSConfig{
			AllowedOrigins:   parseStringSlice("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
			AllowCredentials: parseBool("CORS_ALLOW_CREDENTIALS", true),
			MaxAge:           parseInt("CORS_MAX_AGE", 3600),
		},
		
		AI: AIConfig{
			Provider:    getEnv("AI_PROVIDER", "openai"),
			APIKey:      requireEnv("AI_API_KEY"),
			Model:       getEnv("AI_MODEL", "gpt-4-turbo-preview"),
			MaxTokens:   parseInt("AI_MAX_TOKENS", 4000),
			Temperature: parseFloat("AI_TEMPERATURE", 0.7),
		},
		
		Logger: LoggerConfig{
			Level:    getEnv("LOG_LEVEL", "debug"),
			Format:   getEnv("LOG_FORMAT", "json"),
			Output:   getEnv("LOG_OUTPUT", "stdout"),
			FilePath: getEnv("LOG_FILE_PATH", "/var/log/organiq/app.log"),
		},
		
		Worker: WorkerConfig{
			PoolSize:     parseInt("WORKER_POOL_SIZE", 5),
			PollInterval: parseDuration("WORKER_QUEUE_POLL_INTERVAL", "10s"),
			MaxRetries:   parseInt("WORKER_MAX_RETRIES", 3),
			RetryBackoff: getEnv("WORKER_RETRY_BACKOFF", "exponential"),
		},
		
		RateLimit: RateLimitConfig{
			Enabled:  parseBool("RATE_LIMIT_ENABLED", true),
			Requests: parseInt("RATE_LIMIT_REQUESTS", 1000),
			Window:   parseDuration("RATE_LIMIT_WINDOW", "1m"),
		},
		
		Payment: PaymentConfig{
			StripeSecretKey:          getEnv("STRIPE_SECRET_KEY", ""),
			StripeWebhookSecret:      getEnv("STRIPE_WEBHOOK_SECRET", ""),
			StripeSuccessURL:         getEnv("STRIPE_SUCCESS_URL", ""),
			StripeCancelURL:          getEnv("STRIPE_CANCEL_URL", ""),
			MercadoPagoAccessToken:   getEnv("MERCADOPAGO_ACCESS_TOKEN", ""),
			MercadoPagoWebhookSecret: getEnv("MERCADOPAGO_WEBHOOK_SECRET", ""),
		},
	}
	
	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return cfg, nil
}

// Validate checks if configuration is valid
func (c *Config) Validate() error {
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	
	if len(c.Auth.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	
	if c.Auth.PasswordPepper == "" {
		return fmt.Errorf("PASSWORD_PEPPER is required")
	}
	
	if len(c.Auth.PasswordPepper) < 32 {
		return fmt.Errorf("PASSWORD_PEPPER must be at least 32 characters")
	}
	
	if c.AI.APIKey == "" {
		return fmt.Errorf("AI_API_KEY is required")
	}
	
	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func requireEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("Environment variable %s is required", key))
	}
	return value
}

func parseInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func parseFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func parseBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func parseDuration(key, defaultValue string) time.Duration {
	value := getEnv(key, defaultValue)
	duration, err := time.ParseDuration(value)
	if err != nil {
		panic(fmt.Sprintf("Invalid duration for %s: %s", key, value))
	}
	return duration
}

func parseStringSlice(key, defaultValue string) []string {
	value := getEnv(key, defaultValue)
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}