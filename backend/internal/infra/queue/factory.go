// internal/infra/queue/factory.go
package queue

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/rs/zerolog/log"

	"organiq/config"
)

// NewQueueService cria nova instância de QueueService baseado em configuração
func NewQueueService(cfg *config.Config) (QueueService, error) {
	log.Info().Msg("Inicializando Queue Service")

	// AWS SDK v2 funciona tanto para SQS quanto para LocalStack
	// A diferença é apenas o endpoint
	client, err := newSQSClient(cfg)
	if err != nil {
		return nil, err
	}

	queue := NewSQSQueue(client, 30) // 30 segundos visibility timeout padrão

	if cfg.Queue.Endpoint != "" {
		bootstrapCtx, cancelBootstrap := context.WithTimeout(context.Background(), 10*time.Second)
		if err := NewLocalstackBootstrapper(client).EnsureQueues(bootstrapCtx, []string{
			cfg.Queue.ArticleGenerationQueue,
			cfg.Queue.ArticlePublishQueue,
		}); err != nil {
			cancelBootstrap()
			return nil, err
		}
		cancelBootstrap()
	}

	// Health check nas principais filas
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := queue.HealthCheck(ctx, cfg.Queue.ArticleGenerationQueue); err != nil {
		log.Error().
			Err(err).
			Str("queue", cfg.Queue.ArticleGenerationQueue).
			Msg("HealthCheck falhou na fila de geração")
		return nil, err
	}

	if err := queue.HealthCheck(ctx, cfg.Queue.ArticlePublishQueue); err != nil {
		log.Error().
			Err(err).
			Str("queue", cfg.Queue.ArticlePublishQueue).
			Msg("HealthCheck falhou na fila de publicação")
		return nil, err
	}

	log.Info().Msg("Queue Service inicializado com sucesso")
	return queue, nil
}

// newSQSClient cria cliente SQS com endpoint customizável
func newSQSClient(cfg *config.Config) (*sqs.Client, error) {
	log.Debug().Msg("Criando cliente SQS")

	// Credenciais (para LocalStack, podem ser dummy)
	credProvider := credentials.NewStaticCredentialsProvider(
		cfg.Queue.AccessKeyID,
		cfg.Queue.SecretAccessKey,
		"",
	)

	// Criar config AWS v2
	awsCfg := aws.Config{
		Region:      cfg.Queue.Region,
		Credentials: credProvider,
	}

	// Se temos endpoint customizado (LocalStack), configurar
	if len(cfg.Queue.Endpoint) > 0 {
		log.Debug().Str("endpoint", cfg.Queue.Endpoint).Msg("Usando endpoint customizado para SQS")
		awsCfg.BaseEndpoint = aws.String(cfg.Queue.Endpoint)
	}

	client := sqs.NewFromConfig(awsCfg)

	log.Info().
		Str("region", cfg.Queue.Region).
		Str("endpoint", cfg.Queue.Endpoint).
		Msg("Cliente SQS criado com sucesso")

	return client, nil
}
