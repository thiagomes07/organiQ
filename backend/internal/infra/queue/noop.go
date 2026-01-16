// internal/infra/queue/noop.go
package queue

import (
	"context"

	"github.com/rs/zerolog/log"
)

// NoOpQueue implementa QueueService sem fazer nada
// Usado quando filas não estão configuradas (desenvolvimento simplificado)
type NoOpQueue struct{}

// NewNoOpQueue cria nova instância de NoOpQueue
func NewNoOpQueue() *NoOpQueue {
	return &NoOpQueue{}
}

func (q *NoOpQueue) SendMessage(ctx context.Context, queueName string, message []byte) error {
	log.Warn().
		Str("queue", queueName).
		Msg("NoOpQueue: SendMessage chamado - mensagem ignorada (fila não configurada)")
	return nil
}

func (q *NoOpQueue) ReceiveMessages(ctx context.Context, queueName string, maxMessages int) ([]*Message, error) {
	// Retorna lista vazia - não há mensagens
	return []*Message{}, nil
}

func (q *NoOpQueue) DeleteMessage(ctx context.Context, queueName string, receiptHandle string) error {
	return nil
}

func (q *NoOpQueue) ChangeMessageVisibility(ctx context.Context, queueName string, receiptHandle string, visibilityTimeout int) error {
	return nil
}

func (q *NoOpQueue) SendMessageBatch(ctx context.Context, queueName string, messages [][]byte) error {
	log.Warn().
		Str("queue", queueName).
		Int("count", len(messages)).
		Msg("NoOpQueue: SendMessageBatch chamado - mensagens ignoradas (fila não configurada)")
	return nil
}

func (q *NoOpQueue) GetQueueURL(ctx context.Context, queueName string) (string, error) {
	return "", nil
}

func (q *NoOpQueue) HealthCheck(ctx context.Context, queueName string) error {
	// Sempre retorna sucesso - fila não está configurada mas isso não é erro
	return nil
}
