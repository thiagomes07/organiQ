// internal/worker/article_content_generator.go
package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"organiq/internal/domain/entity"
	"organiq/internal/domain/repository"
	"organiq/internal/infra/ai"
	"organiq/internal/infra/queue"
	"organiq/internal/util"
)

// ArticleContentGeneratorWorker consome mensagens da fila e gera conteúdo de artigos
type ArticleContentGeneratorWorker struct {
	queueService     queue.QueueService
	articleRepo      repository.ArticleRepository
	businessRepo     repository.BusinessRepository
	agentClient      *ai.AgentClient
	cryptoService    *util.CryptoService
	pollInterval     time.Duration
	maxRetries       int
	workerID         string
}

// NewArticleContentGeneratorWorker cria nova instância do worker
func NewArticleContentGeneratorWorker(
	queueService queue.QueueService,
	articleRepo repository.ArticleRepository,
	businessRepo repository.BusinessRepository,
	agentClient *ai.AgentClient,
	cryptoService *util.CryptoService,
	pollInterval time.Duration,
	maxRetries int,
) *ArticleContentGeneratorWorker {
	return &ArticleContentGeneratorWorker{
		queueService:    queueService,
		articleRepo:     articleRepo,
		businessRepo:    businessRepo,
		agentClient:     agentClient,
		cryptoService:   cryptoService,
		pollInterval:    pollInterval,
		maxRetries:      maxRetries,
		workerID:        generateWorkerID(),
	}
}

// Start inicia o worker em goroutine
func (w *ArticleContentGeneratorWorker) Start(ctx context.Context) error {
	log.Info().
		Str("worker_id", w.workerID).
		Dur("poll_interval", w.pollInterval).
		Msg("ArticleContentGeneratorWorker iniciando")

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Str("worker_id", w.workerID).Msg("ArticleContentGeneratorWorker parando gracefully")
			return ctx.Err()

		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

// processBatch consome e processa batch de mensagens
func (w *ArticleContentGeneratorWorker) processBatch(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	log.Debug().Str("worker_id", w.workerID).Msg("ArticleContentGeneratorWorker: iniciando processBatch")

	messages, err := w.queueService.ReceiveMessages(ctx, "article-publish-queue", 10)
	if err != nil {
		log.Error().
			Err(err).
			Str("worker_id", w.workerID).
			Msg("ArticleContentGeneratorWorker: erro ao receber mensagens")
		return
	}

	if len(messages) == 0 {
		log.Debug().Str("worker_id", w.workerID).Msg("ArticleContentGeneratorWorker: nenhuma mensagem disponível")
		return
	}

	log.Info().
		Str("worker_id", w.workerID).
		Int("message_count", len(messages)).
		Msg("ArticleContentGeneratorWorker: processando batch")

	for _, msg := range messages {
		w.processMessage(ctx, msg)
	}
}

// processMessage processa uma mensagem individual
func (w *ArticleContentGeneratorWorker) processMessage(ctx context.Context, message *queue.Message) {
	log.Debug().
		Str("worker_id", w.workerID).
		Str("message_id", message.ID).
		Msg("ArticleContentGeneratorWorker: processando mensagem")

	var payload map[string]interface{}
	if err := json.Unmarshal(message.Body, &payload); err != nil {
		log.Error().
			Err(err).
			Str("worker_id", w.workerID).
			Str("message_id", message.ID).
			Msg("ArticleContentGeneratorWorker: erro ao fazer parse da mensagem")
		_ = w.queueService.DeleteMessage(ctx, "article-publish-queue", message.ReceiptHandle)
		return
	}

	articleIDStr, ok := payload["articleId"].(string)
	if !ok {
		log.Error().
			Str("worker_id", w.workerID).
			Msg("ArticleContentGeneratorWorker: articleId não encontrado")
		_ = w.queueService.DeleteMessage(ctx, "article-publish-queue", message.ReceiptHandle)
		return
	}

	articleID, err := uuid.Parse(articleIDStr)
	if err != nil {
		log.Error().
			Err(err).
			Str("worker_id", w.workerID).
			Str("article_id", articleIDStr).
			Msg("ArticleContentGeneratorWorker: articleId inválido")
		_ = w.queueService.DeleteMessage(ctx, "article-publish-queue", message.ReceiptHandle)
		return
	}

	article, err := w.articleRepo.FindByID(ctx, articleID)
	if err != nil || article == nil {
		log.Error().
			Err(err).
			Str("article_id", articleIDStr).
			Msg("ArticleContentGeneratorWorker: erro ao buscar artigo")

		if message.ReceivedCount < 3 {
			return // Retry automático
		}

		_ = w.queueService.DeleteMessage(ctx, "article-publish-queue", message.ReceiptHandle)
		if article != nil {
			_ = w.articleRepo.UpdateStatusWithError(ctx, articleID, "erro ao buscar artigo após múltiplas tentativas")
		}
		return
	}

	success := false
	var lastErr error

	for attempt := 0; attempt < w.maxRetries; attempt++ {
		log.Debug().
			Str("worker_id", w.workerID).
			Str("article_id", articleIDStr).
			Int("attempt", attempt+1).
			Msg("ArticleContentGeneratorWorker: tentativa de processamento")

		_ = w.queueService.ChangeMessageVisibility(
			ctx,
			"article-publish-queue",
			message.ReceiptHandle,
			60*(attempt+1),
		)

		err := w.generateContent(ctx, article, payload)
		if err == nil {
			success = true
			break
		}

		lastErr = err
		log.Warn().
			Err(err).
			Str("article_id", articleIDStr).
			Int("attempt", attempt+1).
			Msg("ArticleContentGeneratorWorker: erro na tentativa")

		if attempt < w.maxRetries-1 {
			backoffDuration := time.Duration((1 << uint(attempt)) * 5) * time.Second
			log.Debug().
				Str("worker_id", w.workerID).
				Dur("backoff", backoffDuration).
				Msg("ArticleContentGeneratorWorker: aguardando antes de retry")

			select {
			case <-ctx.Done():
				return
			case <-time.After(backoffDuration):
			}
		}
	}

	if success {
		log.Info().
			Str("worker_id", w.workerID).
			Str("article_id", articleIDStr).
			Msg("ArticleContentGeneratorWorker: artigo gerado com sucesso")
	} else {
		log.Error().
			Err(lastErr).
			Str("article_id", articleIDStr).
			Msg("ArticleContentGeneratorWorker: falha ao gerar após retries")

		errorMsg := "erro ao gerar artigo"
		if lastErr != nil {
			errorMsg = lastErr.Error()
		}
		_ = w.articleRepo.UpdateStatusWithError(ctx, articleID, errorMsg)
	}

	if err := w.queueService.DeleteMessage(ctx, "article-publish-queue", message.ReceiptHandle); err != nil {
		log.Error().
			Err(err).
			Str("message_id", message.ID).
			Msg("ArticleContentGeneratorWorker: erro ao deletar mensagem")
	}
}

// generateContent executa a lógica de geração
func (w *ArticleContentGeneratorWorker) generateContent(
	ctx context.Context,
	article *entity.Article,
	payload map[string]interface{},
) error {
	log.Debug().
		Str("article_id", article.ID.String()).
		Msg("ArticleContentGeneratorWorker: iniciando generateContent")

	// 1. Verificar se já tem conteúdo (retry)
	isRetry := payload["isRetry"] != nil && payload["isRetry"].(bool)

	var content string
	if isRetry && article.Content != nil {
		content = *article.Content
		log.Debug().Str("article_id", article.ID.String()).Msg("Usando conteúdo existente (retry)")
	} else {
		// 2. Buscar dados de contexto
		businessProfile, err := w.businessRepo.FindProfileByUserID(ctx, article.UserID)
		if err != nil {
			return fmt.Errorf("erro ao buscar business profile: %w", err)
		}

		// 3. Extrair dados do payload
		title := article.Title
		summary := ""
		if s, ok := payload["summary"].(string); ok {
			summary = s
		}

		var feedback *string
		if f, ok := payload["feedback"].(string); ok && f != "" {
			feedback = &f
		}

		// 4. Montar contexto de negócio
		var businessInfo string
		var objectives string
		var location string
		var brandTone *string

		if businessProfile != nil {
			businessInfo = businessProfile.Description
			objectives = string(businessProfile.PrimaryObjective)
			if businessProfile.SecondaryObjective != nil {
				objectives += ", " + string(*businessProfile.SecondaryObjective)
			}

			location = fmt.Sprintf("%s, %s, %s",
				businessProfile.Location.City,
				businessProfile.Location.State,
				businessProfile.Location.Country,
			)

			// Tom da marca (análise do brandFile se existir)
			if businessProfile.BrandFileURL != nil {
				tone := "profissional e confiável"
				brandTone = &tone
			}
		}

		// 5. Gerar conteúdo com IA
		log.Debug().Str("article_id", article.ID.String()).Msg("Gerando conteúdo com IA")

		aiCtx, aiCancel := context.WithTimeout(ctx, 5*time.Minute)
		defer aiCancel()

		generatedContent, err := w.agentClient.GenerateArticle(
			aiCtx,
			title,
			summary,
			businessInfo,
			objectives,
			location,
			feedback,
			brandTone,
		)

		if err != nil {
			return fmt.Errorf("erro ao gerar conteúdo com IA: %w", err)
		}

		if len(generatedContent) == 0 {
			return errors.New("IA retornou conteúdo vazio")
		}

		content = generatedContent

		// 6. Salvar conteúdo gerado
		if err := w.articleRepo.SetContent(ctx, article.ID, content); err != nil {
			return fmt.Errorf("erro ao salvar conteúdo: %w", err)
		}

		log.Debug().
			Str("article_id", article.ID.String()).
			Int("content_length", len(content)).
			Msg("Conteúdo gerado e salvo")
	}

	// 7. Marcar artigo como GERADO (e não publicado)
	article.SetGenerated()

	// Usando UpdateStatus do repo
	if err := w.articleRepo.UpdateStatus(ctx, article.ID, entity.ArticleStatusGenerated); err != nil {
		return fmt.Errorf("erro ao atualizar status para generated: %w", err)
	}

	log.Info().
		Str("article_id", article.ID.String()).
		Msg("Artigo gerado e pronto para aprovação")

	return nil
}
