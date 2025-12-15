// internal/infra/storage/factory.go
package storage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog/log"

	"organiq/config"
)

// NewStorageService cria nova instância de StorageService baseado em configuração
func NewStorageService(cfg *config.Config) (StorageService, error) {
	log.Info().Str("type", cfg.Storage.Type).Msg("Inicializando Storage")

	switch cfg.Storage.Type {
	case "minio":
		return newMinIOStorage(cfg)
	case "s3":
		return newS3Storage(cfg)
	default:
		return nil, fmt.Errorf("storage type inválido: %s", cfg.Storage.Type)
	}
}

// newMinIOStorage cria cliente para MinIO (local development)
func newMinIOStorage(cfg *config.Config) (StorageService, error) {
	log.Debug().Msg("Configurando MinIO Storage")

	// Construir endpoint
	endpoint := cfg.Storage.MinIOEndpoint
	if len(endpoint) == 0 {
		return nil, fmt.Errorf("MINIO_ENDPOINT não configurado")
	}

	// Credenciais estáticas para MinIO local
	credProvider := credentials.NewStaticCredentialsProvider(
		cfg.Storage.MinIOAccessKey,
		cfg.Storage.MinIOSecretKey,
		"", // sessionToken (vazio para MinIO)
	)

	// Criar cliente S3 apontando para MinIO
	client := s3.NewFromConfig(aws.Config{
		Region:       "us-east-1", // MinIO requer region mesmo que não use
		Credentials:  credProvider,
		BaseEndpoint: aws.String("http://" + endpoint),
	}, func(o *s3.Options) {
		o.UsePathStyle = true // MinIO usa path-style URLs
	})

	storage := NewS3Storage(
		client,
		cfg.Storage.MinIOBucket,
		"us-east-1",
		endpoint,
		cfg.Storage.MinIOUseSSL,
		"", // publicURL (MinIO local não precisa)
	)

	// Health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*60) // 5 minutos timeout
	defer cancel()

	if err := storage.HealthCheck(ctx); err != nil {
		log.Error().Err(err).Msg("MinIO HealthCheck falhou")
		return nil, err
	}

	log.Info().Str("bucket", cfg.Storage.MinIOBucket).Msg("MinIO Storage inicializado com sucesso")
	return storage, nil
}

// newS3Storage cria cliente para AWS S3 (production)
func newS3Storage(cfg *config.Config) (StorageService, error) {
	log.Debug().Msg("Configurando AWS S3 Storage")

	// Usar credenciais padrão (IAM role em ECS, ou env vars)
	// AWS SDK v2 busca credenciais automaticamente
	awsCfg, err := newAWSConfig(cfg.Storage.S3Region)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao criar AWS config")
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg)

	// Para S3, publicURL seria o CloudFront ou acesso direto
	publicURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com", cfg.Storage.S3Bucket, cfg.Storage.S3Region)

	storage := NewS3Storage(
		client,
		cfg.Storage.S3Bucket,
		cfg.Storage.S3Region,
		"",   // endpoint vazio para AWS S3
		true, // useSSL = true para AWS
		publicURL,
	)

	// Health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*60)
	defer cancel()

	if err := storage.HealthCheck(ctx); err != nil {
		log.Error().Err(err).Msg("S3 HealthCheck falhou")
		return nil, err
	}

	log.Info().Str("bucket", cfg.Storage.S3Bucket).Str("region", cfg.Storage.S3Region).Msg("S3 Storage inicializado com sucesso")
	return storage, nil
}

// newAWSConfig cria configuração AWS com credenciais automáticas
func newAWSConfig(region string) (aws.Config, error) {
	// Usar config loader padrão do SDK v2
	// Busca credenciais em: IAM role, env vars, ~/.aws/credentials
	cfg, err := newDefaultAWSConfig(region)
	if err != nil {
		return aws.Config{}, err
	}
	return cfg, nil
}

// Placeholder - você pode expandir com lógica customizada se necessário
func newDefaultAWSConfig(region string) (aws.Config, error) {
	// Para este projeto, retornar config vazia (SDK buscará credenciais automaticamente)
	// Se precisar de customização, adicione aqui
	return aws.Config{
		Region: region,
	}, nil
}
