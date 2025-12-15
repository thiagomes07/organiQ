// internal/infra/queue/interface.go
package queue

import (
	"context"
	"time"
)

// QueueService define contrato para operações com filas de mensagens
type QueueService interface {
	// SendMessage envia mensagem para fila
	SendMessage(ctx context.Context, queueName string, message []byte) error

	// ReceiveMessages recebe mensagens da fila
	// maxMessages: número máximo de mensagens a receber (1-10)
	// Retorna lista de mensagens com receipt handles
	ReceiveMessages(ctx context.Context, queueName string, maxMessages int) ([]*Message, error)

	// DeleteMessage deleta mensagem da fila usando receipt handle
	DeleteMessage(ctx context.Context, queueName string, receiptHandle string) error

	// ChangeMessageVisibility altera timeout de visibilidade da mensagem
	// Útil para dar mais tempo de processamento
	ChangeMessageVisibility(ctx context.Context, queueName string, receiptHandle string, visibilityTimeout int) error

	// SendMessageBatch envia múltiplas mensagens em batch
	SendMessageBatch(ctx context.Context, queueName string, messages [][]byte) error

	// GetQueueURL obtém URL da fila (para operações customizadas)
	GetQueueURL(ctx context.Context, queueName string) (string, error)

	// HealthCheck verifica se queue está acessível
	HealthCheck(ctx context.Context, queueName string) error
}

// Message representa uma mensagem recebida da fila
type Message struct {
	ID            string    // Message ID
	Body          []byte    // Conteúdo da mensagem
	ReceiptHandle string    // Handle para deletar ou modificar
	Attributes    map[string]string
	MD5OfBody     string
	ReceivedCount int       // Quantas vezes foi recebida
	FirstReceived time.Time // Quando foi primeira vez recebida
}

// BatchMessageInput para envios em batch
type BatchMessageInput struct {
	ID   string
	Body []byte
}
