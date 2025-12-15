// internal/worker/article_generator.go
package worker

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"organiq/internal/domain/entity"
	"organiq/internal/domain/repository"
	"organiq/internal/infra/ai"
	"organiq/internal/infra/queue"
)

// ArticleGeneratorWorker consome mensagens da fila e gera ideias de artigos
type ArticleGeneratorWorker struct {
	queueService     queue.QueueService
	articleJobRepo   repository.ArticleJobRepository
	articleIdeaRepo  repository.ArticleIdeaRepository
	businessRepo     repository.BusinessRepository
	agentClient      *ai.AgentClient
	pollInterval     time.Duration
	maxRetries       int
	workerID         string // Para logging
}

// NewArticleGeneratorWorker cria nova instância do worker
func NewArticleGeneratorWorker(
	queueService queue.QueueService,
	articleJobRepo repository.ArticleJobRepository,
	articleIdeaRepo repository.ArticleIdeaRepository,
	businessRepo repository.BusinessRepository,
	agentClient *ai.AgentClient,
	pollInterval time.Duration,
	maxRetries int,
) *ArticleGeneratorWorker {
	return &ArticleGeneratorWorker{
		queueService:    queueService,
		articleJobRepo:  articleJobRepo,
		articleIdeaRepo: articleIdeaRepo,
		businessRepo:    businessRepo,
		agentClient:     agentClient,
		pollInterval:    pollInterval,
		maxRetries:      maxRetries,
		workerID:        generateWorkerID(),
	}
}

// Start inicia o worker em uma goroutine, consumindo mensagens até context ser cancelado
func (w *ArticleGeneratorWorker) Start(ctx context.Context) error {
	log.Info().
		Str("worker_id", w.workerID).
		Dur("poll_interval", w.pollInterval).
		Msg("ArticleGeneratorWorker iniciando")

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Str("worker_id", w.workerID).Msg("ArticleGeneratorWorker parando gracefully")
			return ctx.Err()

		case <-ticker.C:
			// Processar mensagens disponíveis
			w.processBatch(ctx)
		}
	}
}

// ============================================
// PRIVATE METHODS
// ============================================

// processBatch consome e processa um lote de mensagens
func (w *ArticleGeneratorWorker) processBatch(ctx context.Context) {
	// Timeout de 30 segundos para não bloquear indefinidamente
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	log.Debug().Str("worker_id", w.workerID).Msg("ArticleGeneratorWorker: iniciando processBatch")

	// Receber até 10 mensagens por batch
	messages, err := w.queueService.ReceiveMessages(ctx, "article-generation-queue", 10)
	if err != nil {
		log.Error().
			Err(err).
			Str("worker_id", w.workerID).
			Msg("ArticleGeneratorWorker: erro ao receber mensagens")
		return
	}

	if len(messages) == 0 {
		log.Debug().Str("worker_id", w.workerID).Msg("ArticleGeneratorWorker: nenhuma mensagem disponível")
		return
	}

	log.Info().
		Str("worker_id", w.workerID).
		Int("message_count", len(messages)).
		Msg("ArticleGeneratorWorker: processando batch")

	// Processar cada mensagem
	for _, msg := range messages {
		// Não deixar uma falha bloquear o batch
		w.processMessage(ctx, msg)
	}
}

// processMessage processa uma mensagem individual
func (w *ArticleGeneratorWorker) processMessage(ctx context.Context, message *queue.Message) {
	log.Debug().
		Str("worker_id", w.workerID).
		Str("message_id", message.ID).
		Msg("ArticleGeneratorWorker: processando mensagem")

	// 1. Parse da mensagem
	var payload map[string]interface{}
	if err := json.Unmarshal(message.Body, &payload); err != nil {
		log.Error().
			Err(err).
			Str("worker_id", w.workerID).
			Str("message_id", message.ID).
			Msg("ArticleGeneratorWorker: erro ao fazer parse da mensagem")

		// Tentar deletar mensagem da fila (mesmo com erro)
		_ = w.queueService.DeleteMessage(ctx, "article-generation-queue", message.ReceiptHandle)
		return
	}

	// Extrair jobID
	jobIDStr, ok := payload["jobID"].(string)
	if !ok {
		log.Error().
			Str("worker_id", w.workerID).
			Msg("ArticleGeneratorWorker: jobID não encontrado na mensagem")
		_ = w.queueService.DeleteMessage(ctx, "article-generation-queue", message.ReceiptHandle)
		return
	}

	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		log.Error().
			Err(err).
			Str("worker_id", w.workerID).
			Str("job_id", jobIDStr).
			Msg("ArticleGeneratorWorker: jobID inválido")
		_ = w.queueService.DeleteMessage(ctx, "article-generation-queue", message.ReceiptHandle)
		return
	}

	// 2. Buscar job do banco
	job, err := w.articleJobRepo.FindByID(ctx, jobID)
	if err != nil {
		log.Error().
			Err(err).
			Str("worker_id", w.workerID).
			Str("job_id", jobIDStr).
			Msg("ArticleGeneratorWorker: erro ao buscar job")

		// Não deletar da fila, tentar novamente depois
		if message.ReceivedCount < 3 {
			return // Retry automático do SQS
		}

		// Após 3 tentativas, deletar e marcar como falhado
		_ = w.queueService.DeleteMessage(ctx, "article-generation-queue", message.ReceiptHandle)
		_ = w.articleJobRepo.UpdateError(ctx, jobID, "erro ao buscar job após múltiplas tentativas")
		return
	}

	if job == nil {
		log.Warn().
			Str("worker_id", w.workerID).
			Str("job_id", jobIDStr).
			Msg("ArticleGeneratorWorker: job não encontrado")

		_ = w.queueService.DeleteMessage(ctx, "article-generation-queue", message.ReceiptHandle)
		return
	}

	// 3. Processar com retry
	success := false
	var lastErr error

	for attempt := 0; attempt < w.maxRetries; attempt++ {
		log.Debug().
			Str("worker_id", w.workerID).
			Str("job_id", jobIDStr).
			Int("attempt", attempt+1).
			Msg("ArticleGeneratorWorker: tentativa de processamento")

		// Aumentar visibilidade timeout na fila a cada tentativa
		_ = w.queueService.ChangeMessageVisibility(
			ctx,
			"article-generation-queue",
			message.ReceiptHandle,
			60*(attempt+1), // 60s na primeira, 120s na segunda, etc
		)

		err := w.processJob(ctx, job)
		if err == nil {
			success = true
			break
		}

		lastErr = err
		log.Warn().
			Err(err).
			Str("worker_id", w.workerID).
			Str("job_id", jobIDStr).
			Int("attempt", attempt+1).
			Msg("ArticleGeneratorWorker: erro na tentativa")

		// Exponential backoff entre tentativas
		if attempt < w.maxRetries-1 {
			backoffDuration := time.Duration((1 << uint(attempt)) * 5) * time.Second // 5s, 10s, 20s...
			log.Debug().
				Str("worker_id", w.workerID).
				Dur("backoff", backoffDuration).
				Msg("ArticleGeneratorWorker: aguardando antes de retry")

			select {
			case <-ctx.Done():
				// Contexto cancelado, sair do retry loop
				return
			case <-time.After(backoffDuration):
				// Continuar para próxima tentativa
			}
		}
	}

	// 4. Atualizar job com resultado
	if success {
		log.Info().
			Str("worker_id", w.workerID).
			Str("job_id", jobIDStr).
			Msg("ArticleGeneratorWorker: job processado com sucesso")

		_ = w.articleJobRepo.Update(ctx, job)
	} else {
		log.Error().
			Err(lastErr).
			Str("worker_id", w.workerID).
			Str("job_id", jobIDStr).
			Msg("ArticleGeneratorWorker: falha ao processar job após retries")

		errorMsg := "erro ao gerar ideias: "
		if lastErr != nil {
			errorMsg += lastErr.Error()
		} else {
			errorMsg += "erro desconhecido"
		}

		_ = w.articleJobRepo.UpdateError(ctx, jobID, errorMsg)
	}

	// 5. Deletar mensagem da fila (quer tenha sucesso ou não após retries)
	if err := w.queueService.DeleteMessage(ctx, "article-generation-queue", message.ReceiptHandle); err != nil {
		log.Error().
			Err(err).
			Str("worker_id", w.workerID).
			Str("message_id", message.ID).
			Msg("ArticleGeneratorWorker: erro ao deletar mensagem da fila")
	}
}

// processJob executa a lógica de geração de ideias para um job
func (w *ArticleGeneratorWorker) processJob(ctx context.Context, job *entity.ArticleJob) error {
	log.Debug().
		Str("worker_id", w.workerID).
		Str("job_id", job.ID.String()).
		Msg("ArticleGeneratorWorker: iniciando processJob")

	// 1. Atualizar status para processing
	if err := w.articleJobRepo.UpdateStatus(ctx, job.ID, entity.JobStatusProcessing, 10); err != nil {
		log.Error().Err(err).Msg("ArticleGeneratorWorker: erro ao atualizar status para processing")
		return errors.New("erro ao atualizar status de job")
	}

	// 2. Extrair dados do payload
	businessInfo, err := extractBusinessInfo(job.Payload)
	if err != nil {
		log.Error().Err(err).Msg("ArticleGeneratorWorker: erro ao extrair business info")
		return err
	}

	competitors, _ := extractCompetitors(job.Payload)
	objectives, _ := extractObjectives(businessInfo)
	location, _ := extractLocation(businessInfo)

	log.Debug().
		Str("job_id", job.ID.String()).
		Str("business_info", businessInfo).
		Int("competitors_count", len(competitors)).
		Msg("ArticleGeneratorWorker: dados extraídos do payload")

	// 3. Atualizar progress
	_ = w.articleJobRepo.UpdateStatus(ctx, job.ID, entity.JobStatusProcessing, 30)

	// 4. Chamar agente de IA para gerar ideias
	log.Debug().
		Str("job_id", job.ID.String()).
		Msg("ArticleGeneratorWorker: chamando agente de IA")

	// Timeout de 5 minutos para chamada de IA
	aiCtx, aiCancel := context.WithTimeout(ctx, 5*time.Minute)
	defer aiCancel()

	ideas, err := w.agentClient.GenerateIdeas(
		aiCtx,
		businessInfo,
		competitors,
		5, // 5 ideias padrão
		objectives,
		location,
	)

	if err != nil {
		log.Error().
			Err(err).
			Str("job_id", job.ID.String()).
			Msg("ArticleGeneratorWorker: erro ao gerar ideias com IA")
		return errors.New("erro ao chamar agente de IA")
	}

	if len(ideas) == 0 {
		log.Warn().Str("job_id", job.ID.String()).Msg("ArticleGeneratorWorker: agente não retornou ideias")
		return errors.New("agente de IA não gerou ideias")
	}

	log.Info().
		Str("job_id", job.ID.String()).
		Int("ideas_count", len(ideas)).
		Msg("ArticleGeneratorWorker: ideias geradas com sucesso")

	// 5. Atualizar progress
	_ = w.articleJobRepo.UpdateStatus(ctx, job.ID, entity.JobStatusProcessing, 60)

	// 6. Salvar ideias no banco em lote
	articleIdeas := make([]*entity.ArticleIdea, len(ideas))
	for i, title := range ideas {
		// Usar o agente para gerar um resumo baseado no título
		// Por enquanto, usar título como resumo truncado
		summary := truncateString(title, 200)

		articleIdeas[i] = &entity.ArticleIdea{
			ID:        uuid.New(),
			UserID:    job.UserID,
			JobID:     job.ID,
			Title:     title,
			Summary:   summary,
			Approved:  false,
			CreatedAt: time.Now(),
		}
	}

	if err := w.articleIdeaRepo.CreateBatch(ctx, articleIdeas); err != nil {
		log.Error().Err(err).Msg("ArticleGeneratorWorker: erro ao salvar ideias no banco")
		return errors.New("erro ao salvar ideias")
	}

	log.Debug().
		Str("job_id", job.ID.String()).
		Int("saved_ideas", len(articleIdeas)).
		Msg("ArticleGeneratorWorker: ideias salvas no banco")

	// 7. Atualizar status para completed
	_ = w.articleJobRepo.UpdateStatus(ctx, job.ID, entity.JobStatusCompleted, 100)

	log.Info().
		Str("worker_id", w.workerID).
		Str("job_id", job.ID.String()).
		Msg("ArticleGeneratorWorker: processJob completado com sucesso")

	return nil
}

// ============================================
// HELPERS
// ============================================

func generateWorkerID() string {
	return "worker-" + uuid.New().String()[:8]
}

func extractBusinessInfo(payload map[string]interface{}) (string, error) {
	if businessData, ok := payload["businessProfile"].(map[string]interface{}); ok {
		if desc, ok := businessData["description"].(string); ok {
			return desc, nil
		}
	}

	return "", errors.New("business profile description não encontrado")
}

func extractCompetitors(payload map[string]interface{}) ([]string, error) {
	if competitors, ok := payload["competitors"].([]interface{}); ok {
		urls := make([]string, len(competitors))
		for i, comp := range competitors {
			if url, ok := comp.(string); ok {
				urls[i] = url
			}
		}
		return urls, nil
	}

	return []string{}, nil
}

func extractObjectives(businessInfo string) (string, error) {
	// Simplificado: retornar informação genérica
	// Em produção, extrair do payload businessProfile
	return "Aumentar leads e visibilidade online", nil
}

func extractLocation(businessInfo string) (string, error) {
	// Simplificado
	return "Brasil, São Paulo", nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	// Truncar no último espaço antes do maxLen
	truncated := s[:maxLen]
	for i := len(truncated) - 1; i >= 0; i-- {
		if truncated[i] == ' ' {
			return truncated[:i]
		}
	}

	return truncated
}
