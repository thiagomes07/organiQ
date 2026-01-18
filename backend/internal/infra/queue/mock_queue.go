// internal/infra/queue/mock_queue.go
package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"organiq/internal/domain/entity"
	"organiq/internal/domain/repository"
)

// MockQueue implementa QueueService simulando processamento assíncrono
// Use para desenvolvimento/testes sem depender de AI ou WordPress
//
// IMPORTANTE: Esta implementação é apenas para desenvolvimento.
// Em produção, configure MOCK_AI_GENERATION=false para usar o fluxo real.
type MockQueue struct {
	articleJobRepo  repository.ArticleJobRepository
	articleIdeaRepo repository.ArticleIdeaRepository
	processingDelay time.Duration
}

// MockQueueConfig configuração do MockQueue
type MockQueueConfig struct {
	ArticleJobRepo  repository.ArticleJobRepository
	ArticleIdeaRepo repository.ArticleIdeaRepository
	ProcessingDelay time.Duration // Delay antes de "completar" o processamento (padrão: 30s)
}

// NewMockQueue cria nova instância de MockQueue
func NewMockQueue(cfg MockQueueConfig) *MockQueue {
	delay := cfg.ProcessingDelay
	if delay == 0 {
		delay = 30 * time.Second // Default: 30 segundos
	}

	log.Info().
		Dur("processing_delay", delay).
		Msg("MockQueue inicializado - modo de desenvolvimento ativo")

	return &MockQueue{
		articleJobRepo:  cfg.ArticleJobRepo,
		articleIdeaRepo: cfg.ArticleIdeaRepo,
		processingDelay: delay,
	}
}

// SendMessage intercepta mensagens e simula processamento assíncrono
func (q *MockQueue) SendMessage(ctx context.Context, queueName string, message []byte) error {
	log.Debug().
		Str("queue", queueName).
		Msg("MockQueue: recebendo mensagem")

	// Parse da mensagem para extrair informações
	var payload map[string]interface{}
	if err := json.Unmarshal(message, &payload); err != nil {
		log.Error().Err(err).Msg("MockQueue: erro ao fazer parse da mensagem")
		return nil // Não retornar erro para não quebrar o fluxo
	}

	// Extrair tipo de job
	jobType, _ := payload["type"].(string)

	switch jobType {
	case "generate_ideas":
		go q.processGenerateIdeas(payload)
	case "publish_articles":
		go q.processPublishArticles(payload)
	default:
		log.Warn().
			Str("type", jobType).
			Msg("MockQueue: tipo de job desconhecido, ignorando")
	}

	return nil
}

// processGenerateIdeas simula a geração de ideias de artigos
func (q *MockQueue) processGenerateIdeas(payload map[string]interface{}) {
	jobIDStr, _ := payload["jobID"].(string)
	userIDStr, _ := payload["userID"].(string)

	log.Info().
		Str("job_id", jobIDStr).
		Dur("delay", q.processingDelay).
		Msg("MockQueue: iniciando simulação de geração de ideias")

	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		log.Error().Err(err).Msg("MockQueue: job_id inválido")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Error().Err(err).Msg("MockQueue: user_id inválido")
		return
	}

	// Criar contexto com timeout generoso
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 1. Atualizar status para "processing"
	if err := q.articleJobRepo.UpdateStatus(ctx, jobID, entity.JobStatusProcessing, 10); err != nil {
		log.Error().Err(err).Msg("MockQueue: erro ao atualizar status para processing")
		return
	}

	log.Debug().
		Str("job_id", jobIDStr).
		Msg("MockQueue: status atualizado para processing")

	// 2. Simular delay de processamento (dividido para updates de progress)
	progressSteps := []struct {
		progress int
		delay    time.Duration
	}{
		{30, q.processingDelay / 3},
		{60, q.processingDelay / 3},
		{90, q.processingDelay / 3},
	}

	for _, step := range progressSteps {
		time.Sleep(step.delay)
		if err := q.articleJobRepo.UpdateStatus(ctx, jobID, entity.JobStatusProcessing, step.progress); err != nil {
			log.Warn().Err(err).Int("progress", step.progress).Msg("MockQueue: erro ao atualizar progress")
		}
		log.Debug().
			Str("job_id", jobIDStr).
			Int("progress", step.progress).
			Msg("MockQueue: progress atualizado")
	}

	// 3. Gerar ideias mockadas
	mockIdeas := q.generateMockIdeas(userID, jobID)

	// 4. Salvar ideias no banco
	if err := q.articleIdeaRepo.CreateBatch(ctx, mockIdeas); err != nil {
		log.Error().Err(err).Msg("MockQueue: erro ao salvar ideias mockadas")
		_ = q.articleJobRepo.UpdateError(ctx, jobID, "erro ao salvar ideias: "+err.Error())
		return
	}

	log.Info().
		Str("job_id", jobIDStr).
		Int("ideas_count", len(mockIdeas)).
		Msg("MockQueue: ideias mockadas salvas")

	// 5. Atualizar status para "completed"
	if err := q.articleJobRepo.UpdateStatus(ctx, jobID, entity.JobStatusCompleted, 100); err != nil {
		log.Error().Err(err).Msg("MockQueue: erro ao atualizar status para completed")
		return
	}

	log.Info().
		Str("job_id", jobIDStr).
		Msg("MockQueue: processamento de geração de ideias concluído com sucesso")
}

// generateMockIdeas cria ideias de artigos mockadas para testes
func (q *MockQueue) generateMockIdeas(userID, jobID uuid.UUID) []*entity.ArticleIdea {
	// Ideias mockadas realistas para testes
	mockTitles := []struct {
		title   string
		summary string
	}{
		{
			title:   "10 Estratégias Comprovadas para Aumentar o Tráfego Orgânico do Seu Blog",
			summary: "Descubra técnicas de SEO e marketing de conteúdo que vão impulsionar seu tráfego orgânico de forma sustentável.",
		},
		{
			title:   "Como Criar Conteúdo que Converte: Guia Completo para 2026",
			summary: "Aprenda a estruturar seus artigos para maximizar conversões e engajamento do público-alvo.",
		},
		{
			title:   "O Poder das Palavras-Chave de Cauda Longa: Por Que Você Deveria Usá-las",
			summary: "Entenda como palavras-chave específicas podem trazer tráfego mais qualificado para seu negócio.",
		},
		{
			title:   "Marketing de Conteúdo B2B: Estratégias que Funcionam em 2026",
			summary: "Explore táticas avançadas de content marketing focadas no mercado business-to-business.",
		},
		{
			title:   "5 Erros Comuns de SEO que Estão Prejudicando Seu Site (e Como Corrigi-los)",
			summary: "Identifique e corrija os problemas mais frequentes que impedem seu site de rankear bem no Google.",
		},
	}

	ideas := make([]*entity.ArticleIdea, len(mockTitles))
	now := time.Now()

	for i, mock := range mockTitles {
		ideas[i] = &entity.ArticleIdea{
			ID:        uuid.New(),
			UserID:    userID,
			JobID:     jobID,
			Title:     mock.title,
			Summary:   mock.summary,
			Approved:  false,
			CreatedAt: now,
		}
	}

	return ideas
}

// processPublishArticles simula a publicação de artigos
func (q *MockQueue) processPublishArticles(payload map[string]interface{}) {
	jobIDStr, _ := payload["jobID"].(string)

	log.Info().
		Str("job_id", jobIDStr).
		Dur("delay", q.processingDelay).
		Msg("MockQueue: iniciando simulação de publicação de artigos")

	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		log.Error().Err(err).Msg("MockQueue: job_id inválido para publicação")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 1. Atualizar status para "processing"
	if err := q.articleJobRepo.UpdateStatus(ctx, jobID, entity.JobStatusProcessing, 10); err != nil {
		log.Error().Err(err).Msg("MockQueue: erro ao atualizar status de publicação")
		return
	}

	// 2. Simular delay
	time.Sleep(q.processingDelay / 2)
	_ = q.articleJobRepo.UpdateStatus(ctx, jobID, entity.JobStatusProcessing, 50)

	time.Sleep(q.processingDelay / 2)

	// 3. Atualizar status para "completed"
	if err := q.articleJobRepo.UpdateStatus(ctx, jobID, entity.JobStatusCompleted, 100); err != nil {
		log.Error().Err(err).Msg("MockQueue: erro ao completar publicação")
		return
	}

	log.Info().
		Str("job_id", jobIDStr).
		Msg("MockQueue: publicação simulada concluída com sucesso")
}

// ============================================
// MÉTODOS DA INTERFACE (não usados no mock)
// ============================================

func (q *MockQueue) ReceiveMessages(ctx context.Context, queueName string, maxMessages int) ([]*Message, error) {
	// Mock não recebe mensagens - processamento é inline
	return []*Message{}, nil
}

func (q *MockQueue) DeleteMessage(ctx context.Context, queueName string, receiptHandle string) error {
	return nil
}

func (q *MockQueue) ChangeMessageVisibility(ctx context.Context, queueName string, receiptHandle string, visibilityTimeout int) error {
	return nil
}

func (q *MockQueue) SendMessageBatch(ctx context.Context, queueName string, messages [][]byte) error {
	for _, msg := range messages {
		if err := q.SendMessage(ctx, queueName, msg); err != nil {
			return err
		}
	}
	return nil
}

func (q *MockQueue) GetQueueURL(ctx context.Context, queueName string) (string, error) {
	return "mock://" + queueName, nil
}

func (q *MockQueue) HealthCheck(ctx context.Context, queueName string) error {
	return nil
}
