package queue

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/rs/zerolog/log"
)

// SQSQueue implementa QueueService usando AWS SQS ou LocalStack
type SQSQueue struct {
	client            *sqs.Client
	queueURLs         map[string]string // Cache de URLs das filas (queueName -> URL)
	visibilityTimeout int32             // Default visibility timeout (segundos)
	maxReceiveCount   int32
	waitTimeSeconds   int32
}

// NewSQSQueue cria nova instância para AWS SQS/LocalStack
func NewSQSQueue(
	client *sqs.Client,
	visibilityTimeout int32,
) QueueService {
	if visibilityTimeout <= 0 {
		visibilityTimeout = 30 // default 30 segundos
	}

	return &SQSQueue{
		client:            client,
		queueURLs:         make(map[string]string),
		visibilityTimeout: visibilityTimeout,
		maxReceiveCount:   3,
		waitTimeSeconds:   20, // Long polling
	}
}

// SendMessage envia mensagem para fila
func (sq *SQSQueue) SendMessage(ctx context.Context, queueName string, message []byte) error {
	log.Debug().
		Str("queue_name", queueName).
		Int("message_size", len(message)).
		Msg("SQSQueue SendMessage iniciado")

	// Validar entrada
	if len(queueName) == 0 {
		log.Error().Msg("SQSQueue SendMessage: queueName não pode estar vazio")
		return fmt.Errorf("queueName não pode estar vazio")
	}

	if len(message) == 0 {
		log.Error().Msg("SQSQueue SendMessage: message não pode estar vazio")
		return fmt.Errorf("message não pode estar vazio")
	}

	// Obter URL da fila
	queueURL, err := sq.getQueueURL(ctx, queueName)
	if err != nil {
		log.Error().
			Err(err).
			Str("queue_name", queueName).
			Msg("SQSQueue SendMessage erro ao obter URL da fila")
		return err
	}

	// Enviar mensagem
	msgBody := string(message)
	result, err := sq.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &queueURL,
		MessageBody: &msgBody,
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("queue_name", queueName).
			Msg("SQSQueue SendMessage erro ao enviar")
		return fmt.Errorf("erro ao enviar mensagem: %w", err)
	}

	log.Info().
		Str("queue_name", queueName).
		Str("message_id", *result.MessageId).
		Msg("SQSQueue SendMessage bem-sucedido")

	return nil
}

// ReceiveMessages recebe mensagens da fila
func (sq *SQSQueue) ReceiveMessages(
	ctx context.Context,
	queueName string,
	maxMessages int,
) ([]*Message, error) {
	log.Debug().
		Str("queue_name", queueName).
		Int("max_messages", maxMessages).
		Msg("SQSQueue ReceiveMessages iniciado")

	// Validar entrada
	if len(queueName) == 0 {
		log.Error().Msg("SQSQueue ReceiveMessages: queueName não pode estar vazio")
		return nil, fmt.Errorf("queueName não pode estar vazio")
	}

	if maxMessages < 1 || maxMessages > 10 {
		log.Warn().Int("max_messages", maxMessages).Msg("SQSQueue ReceiveMessages: maxMessages inválido, usando 10")
		maxMessages = 10
	}

	// Obter URL da fila
	queueURL, err := sq.getQueueURL(ctx, queueName)
	if err != nil {
		log.Error().
			Err(err).
			Str("queue_name", queueName).
			Msg("SQSQueue ReceiveMessages erro ao obter URL")
		return nil, err
	}

	// Receber mensagens
	maxMsgs := int32(maxMessages)
	result, err := sq.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:              &queueURL,
		MaxNumberOfMessages:   maxMsgs,
		VisibilityTimeout:     sq.visibilityTimeout,
		WaitTimeSeconds:       sq.waitTimeSeconds,
		MessageAttributeNames: []string{"All"},
		AttributeNames:        []types.QueueAttributeName{types.QueueAttributeNameAll},
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("queue_name", queueName).
			Msg("SQSQueue ReceiveMessages erro ao receber")
		return nil, fmt.Errorf("erro ao receber mensagens: %w", err)
	}

	if len(result.Messages) == 0 {
		log.Debug().Str("queue_name", queueName).Msg("SQSQueue ReceiveMessages: nenhuma mensagem disponível")
		return []*Message{}, nil
	}

	// Converter para formato interno
	messages := make([]*Message, 0, len(result.Messages))

	for _, sqsMsg := range result.Messages {
		// Extrair atributos
		attributes := make(map[string]string)
		if sqsMsg.Attributes != nil {
			for k, v := range sqsMsg.Attributes {
				attributes[string(k)] = v
			}
		}

		// Contar quantas vezes foi recebido
		receivedCount := 1
		if approxReceiveCount, ok := attributes["ApproximateReceiveCount"]; ok {
			if count, err := strconv.Atoi(approxReceiveCount); err == nil {
				receivedCount = count
			}
		}

		// Tentar extrair timestamp de primeira recepção
		firstReceivedTime := time.Now()
		if sentTime, ok := attributes["SentTimestamp"]; ok {
			if timestamp, err := strconv.ParseInt(sentTime, 10, 64); err == nil {
				firstReceivedTime = time.UnixMilli(timestamp)
			}
		}

		msg := &Message{
			ID:            *sqsMsg.MessageId,
			Body:          []byte(*sqsMsg.Body),
			ReceiptHandle: *sqsMsg.ReceiptHandle,
			Attributes:    attributes,
			MD5OfBody:     *sqsMsg.MD5OfBody,
			ReceivedCount: receivedCount,
			FirstReceived: firstReceivedTime,
		}

		messages = append(messages, msg)
	}

	log.Info().
		Str("queue_name", queueName).
		Int("received_count", len(messages)).
		Msg("SQSQueue ReceiveMessages bem-sucedido")

	return messages, nil
}

// DeleteMessage deleta mensagem da fila
func (sq *SQSQueue) DeleteMessage(
	ctx context.Context,
	queueName string,
	receiptHandle string,
) error {
	log.Debug().
		Str("queue_name", queueName).
		Str("receipt_handle", receiptHandle[:20]+"..."). // Truncar para log
		Msg("SQSQueue DeleteMessage iniciado")

	// Validar entrada
	if len(queueName) == 0 {
		log.Error().Msg("SQSQueue DeleteMessage: queueName não pode estar vazio")
		return fmt.Errorf("queueName não pode estar vazio")
	}

	if len(receiptHandle) == 0 {
		log.Error().Msg("SQSQueue DeleteMessage: receiptHandle não pode estar vazio")
		return fmt.Errorf("receiptHandle não pode estar vazio")
	}

	// Obter URL da fila
	queueURL, err := sq.getQueueURL(ctx, queueName)
	if err != nil {
		log.Error().
			Err(err).
			Str("queue_name", queueName).
			Msg("SQSQueue DeleteMessage erro ao obter URL")
		return err
	}

	// Deletar mensagem
	_, err = sq.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &queueURL,
		ReceiptHandle: &receiptHandle,
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("queue_name", queueName).
			Msg("SQSQueue DeleteMessage erro ao deletar")
		return fmt.Errorf("erro ao deletar mensagem: %w", err)
	}

	log.Debug().
		Str("queue_name", queueName).
		Msg("SQSQueue DeleteMessage bem-sucedido")

	return nil
}

// ChangeMessageVisibility altera timeout de visibilidade da mensagem
func (sq *SQSQueue) ChangeMessageVisibility(
	ctx context.Context,
	queueName string,
	receiptHandle string,
	visibilityTimeout int,
) error {
	log.Debug().
		Str("queue_name", queueName).
		Int("visibility_timeout", visibilityTimeout).
		Msg("SQSQueue ChangeMessageVisibility iniciado")

	// Validar entrada
	if len(queueName) == 0 {
		log.Error().Msg("SQSQueue ChangeMessageVisibility: queueName não pode estar vazio")
		return fmt.Errorf("queueName não pode estar vazio")
	}

	if len(receiptHandle) == 0 {
		log.Error().Msg("SQSQueue ChangeMessageVisibility: receiptHandle não pode estar vazio")
		return fmt.Errorf("receiptHandle não pode estar vazio")
	}

	if visibilityTimeout < 0 || visibilityTimeout > 43200 {
		log.Error().
			Int("visibility_timeout", visibilityTimeout).
			Msg("SQSQueue ChangeMessageVisibility: visibilityTimeout fora do intervalo válido (0-43200)")
		return fmt.Errorf("visibilityTimeout deve estar entre 0 e 43200 segundos")
	}

	// Obter URL da fila
	queueURL, err := sq.getQueueURL(ctx, queueName)
	if err != nil {
		log.Error().
			Err(err).
			Str("queue_name", queueName).
			Msg("SQSQueue ChangeMessageVisibility erro ao obter URL")
		return err
	}

	// Mudar visibilidade
	_, err = sq.client.ChangeMessageVisibility(ctx, &sqs.ChangeMessageVisibilityInput{
		QueueUrl:          &queueURL,
		ReceiptHandle:     &receiptHandle,
		VisibilityTimeout: int32(visibilityTimeout),
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("queue_name", queueName).
			Msg("SQSQueue ChangeMessageVisibility erro")
		return fmt.Errorf("erro ao mudar visibilidade: %w", err)
	}

	log.Debug().
		Str("queue_name", queueName).
		Int("visibility_timeout", visibilityTimeout).
		Msg("SQSQueue ChangeMessageVisibility bem-sucedido")

	return nil
}

// SendMessageBatch envia múltiplas mensagens em batch
func (sq *SQSQueue) SendMessageBatch(
	ctx context.Context,
	queueName string,
	messages [][]byte,
) error {
	log.Debug().
		Str("queue_name", queueName).
		Int("batch_size", len(messages)).
		Msg("SQSQueue SendMessageBatch iniciado")

	// Validar entrada
	if len(queueName) == 0 {
		log.Error().Msg("SQSQueue SendMessageBatch: queueName não pode estar vazio")
		return fmt.Errorf("queueName não pode estar vazio")
	}

	if len(messages) == 0 {
		log.Error().Msg("SQSQueue SendMessageBatch: messages não pode estar vazio")
		return fmt.Errorf("messages não pode estar vazio")
	}

	if len(messages) > 10 {
		log.Error().
			Int("batch_size", len(messages)).
			Msg("SQSQueue SendMessageBatch: máximo 10 mensagens por batch")
		return fmt.Errorf("máximo 10 mensagens por batch")
	}

	// Obter URL da fila
	queueURL, err := sq.getQueueURL(ctx, queueName)
	if err != nil {
		log.Error().
			Err(err).
			Str("queue_name", queueName).
			Msg("SQSQueue SendMessageBatch erro ao obter URL")
		return err
	}

	// Converter para formato SQS
	entries := make([]types.SendMessageBatchRequestEntry, len(messages))
	for i, msg := range messages {
		id := strconv.Itoa(i)
		body := string(msg)
		entries[i] = types.SendMessageBatchRequestEntry{
			Id:          &id,
			MessageBody: &body,
		}
	}

	// Enviar batch
	result, err := sq.client.SendMessageBatch(ctx, &sqs.SendMessageBatchInput{
		QueueUrl: &queueURL,
		Entries:  entries,
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("queue_name", queueName).
			Msg("SQSQueue SendMessageBatch erro ao enviar")
		return fmt.Errorf("erro ao enviar batch: %w", err)
	}

	if len(result.Failed) > 0 {
		log.Warn().
			Str("queue_name", queueName).
			Int("failed_count", len(result.Failed)).
			Int("successful_count", len(result.Successful)).
			Msg("SQSQueue SendMessageBatch: algumas mensagens falharam")

		// Logar cada falha
		for _, failed := range result.Failed {
			log.Error().
				Str("entry_id", *failed.Id).
				Str("error_code", *failed.Code).
				Str("error_message", *failed.Message).
				Msg("SQSQueue SendMessageBatch falha de mensagem")
		}
	}

	log.Info().
		Str("queue_name", queueName).
		Int("successful_count", len(result.Successful)).
		Int("failed_count", len(result.Failed)).
		Msg("SQSQueue SendMessageBatch bem-sucedido")

	return nil
}

// GetQueueURL obtém URL da fila
func (sq *SQSQueue) GetQueueURL(ctx context.Context, queueName string) (string, error) {
	log.Debug().
		Str("queue_name", queueName).
		Msg("SQSQueue GetQueueURL iniciado")

	return sq.getQueueURL(ctx, queueName)
}

// HealthCheck verifica se SQS está acessível
func (sq *SQSQueue) HealthCheck(ctx context.Context, queueName string) error {
	log.Debug().
		Str("queue_name", queueName).
		Msg("SQSQueue HealthCheck iniciado")

	// Validar que conseguimos obter a URL da fila
	_, err := sq.getQueueURL(ctx, queueName)
	if err != nil {
		log.Error().
			Err(err).
			Str("queue_name", queueName).
			Msg("SQSQueue HealthCheck falhou")
		return fmt.Errorf("erro ao verificar saúde de SQS: %w", err)
	}

	log.Info().
		Str("queue_name", queueName).
		Msg("SQSQueue HealthCheck bem-sucedido")

	return nil
}

// ============================================
// PRIVATE METHODS
// ============================================

// getQueueURL obtém URL da fila com cache
func (sq *SQSQueue) getQueueURL(ctx context.Context, queueName string) (string, error) {
	// Verificar cache
	if cachedURL, exists := sq.queueURLs[queueName]; exists {
		return cachedURL, nil
	}

	// Buscar URL do SQS
	result, err := sq.client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: &queueName,
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("queue_name", queueName).
			Msg("SQSQueue getQueueURL erro ao obter URL")
		return "", fmt.Errorf("erro ao obter URL da fila: %w", err)
	}

	// Cachear
	sq.queueURLs[queueName] = *result.QueueUrl

	log.Debug().
		Str("queue_name", queueName).
		Str("queue_url", *result.QueueUrl).
		Msg("SQSQueue getQueueURL cacheado")

	return *result.QueueUrl, nil
}

// ListQueues lista filas (para debugging)
func (sq *SQSQueue) ListQueues(ctx context.Context, prefix string) ([]string, error) {
	log.Debug().
		Str("prefix", prefix).
		Msg("SQSQueue ListQueues iniciado")

	result, err := sq.client.ListQueues(ctx, &sqs.ListQueuesInput{
		QueueNamePrefix: &prefix,
	})

	if err != nil {
		log.Error().
			Err(err).
			Msg("SQSQueue ListQueues erro")
		return nil, fmt.Errorf("erro ao listar filas: %w", err)
	}

	queueNames := make([]string, 0, len(result.QueueUrls))

	for _, queueURL := range result.QueueUrls {
		// Extrair nome da fila da URL
		parts := strings.Split(queueURL, "/")
		if len(parts) > 0 {
			queueNames = append(queueNames, parts[len(parts)-1])
		}
	}

	log.Debug().
		Int("queue_count", len(queueNames)).
		Msg("SQSQueue ListQueues bem-sucedido")

	return queueNames, nil
}

// PurgeQueue limpa todas as mensagens da fila
func (sq *SQSQueue) PurgeQueue(ctx context.Context, queueName string) error {
	log.Debug().
		Str("queue_name", queueName).
		Msg("SQSQueue PurgeQueue iniciado")

	// Obter URL da fila
	queueURL, err := sq.getQueueURL(ctx, queueName)
	if err != nil {
		return err
	}

	// Purgar fila
	_, err = sq.client.PurgeQueue(ctx, &sqs.PurgeQueueInput{
		QueueUrl: &queueURL,
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("queue_name", queueName).
			Msg("SQSQueue PurgeQueue erro")
		return fmt.Errorf("erro ao purgar fila: %w", err)
	}

	log.Warn().
		Str("queue_name", queueName).
		Msg("SQSQueue PurgeQueue bem-sucedido - fila foi purgada")

	return nil
}

// GetQueueAttributes obtém atributos da fila
func (sq *SQSQueue) GetQueueAttributes(ctx context.Context, queueName string) (map[string]string, error) {
	log.Debug().
		Str("queue_name", queueName).
		Msg("SQSQueue GetQueueAttributes iniciado")

	// Obter URL da fila
	queueURL, err := sq.getQueueURL(ctx, queueName)
	if err != nil {
		return nil, err
	}

	// Obter atributos
	result, err := sq.client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       &queueURL,
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameAll},
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("queue_name", queueName).
			Msg("SQSQueue GetQueueAttributes erro")
		return nil, fmt.Errorf("erro ao obter atributos: %w", err)
	}

	attributes := make(map[string]string)
	for k, v := range result.Attributes {
		attributes[string(k)] = v
	}

	log.Debug().
		Str("queue_name", queueName).
		Int("attribute_count", len(attributes)).
		Msg("SQSQueue GetQueueAttributes bem-sucedido")

	return attributes, nil
}
