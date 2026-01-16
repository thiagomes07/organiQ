// ==========================================
// config/config.go
// ==========================================
package config

import (
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	Storage     StorageConfig
	Queue       QueueConfig
	Auth        AuthConfig
	CORS        CORSConfig
	AI          AIConfig
	Logger      LoggerConfig
	Worker      WorkerConfig
	RateLimit   RateLimitConfig
	Payment     PaymentConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port            string
	Host            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// StorageConfig holds blob storage configuration
type StorageConfig struct {
	Type string // "minio" | "s3"

	// MinIO specific
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	MinIOUseSSL    bool

	// S3 specific
	S3Bucket string
	S3Region string
}

// QueueConfig holds message queue configuration
type QueueConfig struct {
	Enabled                bool // Se false, usa NoOpQueue
	Endpoint               string
	Region                 string
	AccessKeyID            string
	SecretAccessKey        string
	ArticleGenerationQueue string
	ArticlePublishQueue    string
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret            string
	PasswordPepper       string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowCredentials bool
	MaxAge           int
}

// AIConfig holds AI service configuration
type AIConfig struct {
	Provider    string // "openai" | "anthropic"
	APIKey      string
	Model       string
	MaxTokens   int
	Temperature float64
}

// LoggerConfig holds logging configuration
type LoggerConfig struct {
	Level    string // "debug" | "info" | "warn" | "error"
	Format   string // "json" | "console"
	Output   string // "stdout" | "file"
	FilePath string
}

// WorkerConfig holds worker pool configuration
type WorkerConfig struct {
	PoolSize     int
	PollInterval time.Duration
	MaxRetries   int
	RetryBackoff string // "fixed" | "exponential"
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled  bool
	Requests int
	Window   time.Duration
}

// PaymentConfig holds payment provider configuration
type PaymentConfig struct {
	StripeSecretKey     string
	StripeWebhookSecret string
	StripeSuccessURL    string
	StripeCancelURL     string

	MercadoPagoAccessToken   string
	MercadoPagoWebhookSecret string
}
