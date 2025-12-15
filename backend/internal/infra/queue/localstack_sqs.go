package queue

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/rs/zerolog/log"
)

// LocalstackBootstrapper garante que as filas existam quando usamos LocalStack.
type LocalstackBootstrapper struct {
	client *sqs.Client
}

// NewLocalstackBootstrapper cria nova instância.
func NewLocalstackBootstrapper(client *sqs.Client) *LocalstackBootstrapper {
	return &LocalstackBootstrapper{client: client}
}

// EnsureQueues cria filas que não existirem.
func (b *LocalstackBootstrapper) EnsureQueues(ctx context.Context, queueNames []string) error {
	for _, queueName := range queueNames {
		if err := b.ensureQueue(ctx, queueName); err != nil {
			return err
		}
	}
	return nil
}

func (b *LocalstackBootstrapper) ensureQueue(ctx context.Context, queueName string) error {
	if queueName == "" {
		return fmt.Errorf("queue name cannot be empty")
	}

	const maxAttempts = 3
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		_, err := b.client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{QueueName: &queueName})
		if err == nil {
			log.Debug().Str("queue", queueName).Msg("LocalStack queue already exists")
			return nil
		}

		var notFound *types.QueueDoesNotExist
		if errors.As(err, &notFound) {
			createErr := b.createQueue(ctx, queueName)
			if createErr == nil {
				return nil
			}
			lastErr = createErr
		} else {
			lastErr = err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt) * 200 * time.Millisecond):
		}
	}

	return fmt.Errorf("failed to ensure queue %s: %w", queueName, lastErr)
}

func (b *LocalstackBootstrapper) createQueue(ctx context.Context, queueName string) error {
	log.Info().Str("queue", queueName).Msg("Creating LocalStack SQS queue")
	attrs := map[string]string{
		"VisibilityTimeout": "30",
	}
	_, err := b.client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName:  &queueName,
		Attributes: attrs,
	})
	if err != nil {
		log.Error().Err(err).Str("queue", queueName).Msg("failed to create LocalStack queue")
		return err
	}
	return nil
}
