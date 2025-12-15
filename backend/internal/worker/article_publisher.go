// internal/worker/article_publisher.go
package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"organiq/internal/domain/entity"
	"organiq/internal/domain/repository"
	"organiq/internal/infra/ai"
	"organiq/internal/infra/queue"
	"organiq/internal/infra/wordpress"
	"organiq/internal/util"
)

// ArticlePublisherWorker consome mensagens da fila de publicação
type ArticlePublisherWorker struct {
	queueService     queue.QueueService
	articleRepo      repository.ArticleRepository
	businessRepo     repository.BusinessRepository
	integrationRepo  repository.IntegrationRepository
	agentClient      *ai.AgentClient
	cryptoService    *util.CryptoService
	pollInterval     time.Duration
	maxRetries       int
	workerID         string
}

// NewArticlePublisherWorker cria nova instância do worker
func NewArticlePublisherWorker(
	queueService queue.QueueService,
	articleRepo repository.ArticleRepository,
	businessRepo repository.BusinessRepository,
	integrationRepo repository.IntegrationRepository,
	agentClient *ai.AgentClient,
	cryptoService *util.CryptoService,
	pollInterval time.Duration,
	maxRetries int,
) *ArticlePublisherWorker {
	return &ArticlePublisherWorker{
		queueService:    queueService,
		articleRepo:     articleRepo,
		businessRepo:    businessRepo,
		integrationRepo: integrationRepo,
		agentClient:     agentClient,
		cryptoService:   cryptoService,
		pollInterval:    pollInterval,
		maxRetries:      maxRetries,
		workerID:        generateWorkerID(),
	}
}

// Start inicia o worker em goroutine
func (w *ArticlePublisherWorker) Start(ctx context.Context) error {
	log.Info().
		Str("worker_id", w.workerID).
		Dur("poll_interval", w.pollInterval).
		Msg("ArticlePublisherWorker iniciando")

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Str("worker_id", w.workerID).Msg("ArticlePublisherWorker parando gracefully")
			return ctx.Err()

		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

// processBatch consome e processa batch de mensagens
func (w *ArticlePublisherWorker) processBatch(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	log.Debug().Str("worker_id", w.workerID).Msg("ArticlePublisherWorker: iniciando processBatch")

	messages, err := w.queueService.ReceiveMessages(ctx, "article-publish-queue", 10)
	if err != nil {
		log.Error().
			Err(err).
			Str("worker_id", w.workerID).
			Msg("ArticlePublisherWorker: erro ao receber mensagens")
		return
	}

	if len(messages) == 0 {
		log.Debug().Str("worker_id", w.workerID).Msg("ArticlePublisherWorker: nenhuma mensagem disponível")
		return
	}

	log.Info().
		Str("worker_id", w.workerID).
		Int("message_count", len(messages)).
		Msg("ArticlePublisherWorker: processando batch")

	for _, msg := range messages {
		w.processMessage(ctx, msg)
	}
}

// processMessage processa uma mensagem individual
func (w *ArticlePublisherWorker) processMessage(ctx context.Context, message *queue.Message) {
	log.Debug().
		Str("worker_id", w.workerID).
		Str("message_id", message.ID).
		Msg("ArticlePublisherWorker: processando mensagem")

	var payload map[string]interface{}
	if err := json.Unmarshal(message.Body, &payload); err != nil {
		log.Error().
			Err(err).
			Str("worker_id", w.workerID).
			Str("message_id", message.ID).
			Msg("ArticlePublisherWorker: erro ao fazer parse da mensagem")
		_ = w.queueService.DeleteMessage(ctx, "article-publish-queue", message.ReceiptHandle)
		return
	}

	articleIDStr, ok := payload["articleId"].(string)
	if !ok {
		log.Error().
			Str("worker_id", w.workerID).
			Msg("ArticlePublisherWorker: articleId não encontrado")
		_ = w.queueService.DeleteMessage(ctx, "article-publish-queue", message.ReceiptHandle)
		return
	}

	articleID, err := uuid.Parse(articleIDStr)
	if err != nil {
		log.Error().
			Err(err).
			Str("worker_id", w.workerID).
			Str("article_id", articleIDStr).
			Msg("ArticlePublisherWorker: articleId inválido")
		_ = w.queueService.DeleteMessage(ctx, "article-publish-queue", message.ReceiptHandle)
		return
	}

	article, err := w.articleRepo.FindByID(ctx, articleID)
	if err != nil || article == nil {
		log.Error().
			Err(err).
			Str("article_id", articleIDStr).
			Msg("ArticlePublisherWorker: erro ao buscar artigo")

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
			Msg("ArticlePublisherWorker: tentativa de processamento")

		_ = w.queueService.ChangeMessageVisibility(
			ctx,
			"article-publish-queue",
			message.ReceiptHandle,
			60*(attempt+1),
		)

		err := w.publishArticle(ctx, article, payload)
		if err == nil {
			success = true
			break
		}

		lastErr = err
		log.Warn().
			Err(err).
			Str("article_id", articleIDStr).
			Int("attempt", attempt+1).
			Msg("ArticlePublisherWorker: erro na tentativa")

		if attempt < w.maxRetries-1 {
			backoffDuration := time.Duration((1 << uint(attempt)) * 5) * time.Second
			log.Debug().
				Str("worker_id", w.workerID).
				Dur("backoff", backoffDuration).
				Msg("ArticlePublisherWorker: aguardando antes de retry")

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
			Msg("ArticlePublisherWorker: artigo publicado com sucesso")
	} else {
		log.Error().
			Err(lastErr).
			Str("article_id", articleIDStr).
			Msg("ArticlePublisherWorker: falha ao publicar após retries")

		errorMsg := "erro ao publicar artigo"
		if lastErr != nil {
			errorMsg = lastErr.Error()
		}
		_ = w.articleRepo.UpdateStatusWithError(ctx, articleID, errorMsg)
	}

	if err := w.queueService.DeleteMessage(ctx, "article-publish-queue", message.ReceiptHandle); err != nil {
		log.Error().
			Err(err).
			Str("message_id", message.ID).
			Msg("ArticlePublisherWorker: erro ao deletar mensagem")
	}
}

// publishArticle executa a lógica de publicação
func (w *ArticlePublisherWorker) publishArticle(
	ctx context.Context,
	article *entity.Article,
	payload map[string]interface{},
) error {
	log.Debug().
		Str("article_id", article.ID.String()).
		Msg("ArticlePublisherWorker: iniciando publishArticle")

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

	// 7. Atualizar status para publishing
	if err := w.articleRepo.UpdateStatus(ctx, article.ID, entity.ArticleStatusPublishing); err != nil {
		return fmt.Errorf("erro ao atualizar status: %w", err)
	}

	// 8. Buscar integração WordPress
	wpIntegration, err := w.integrationRepo.FindEnabledByUserIDAndType(
		ctx,
		article.UserID,
		entity.IntegrationTypeWordPress,
	)

	if err != nil {
		return fmt.Errorf("erro ao buscar integração WordPress: %w", err)
	}

	if wpIntegration == nil {
		return errors.New("integração WordPress não configurada")
	}

	// 9. Extrair credenciais WordPress
	wpConfig, err := wpIntegration.GetWordPressConfig()
	if err != nil {
		return fmt.Errorf("erro ao extrair config WordPress: %w", err)
	}

	// Descriptografar appPassword
	decryptedPassword, err := w.cryptoService.DecryptAES(wpConfig.AppPassword)
	if err != nil {
		return fmt.Errorf("erro ao descriptografar senha: %w", err)
	}

	// 10. Converter Markdown para HTML (simplificado)
	htmlContent := markdownToHTML(content)

	// 11. Publicar no WordPress
	log.Debug().
		Str("article_id", article.ID.String()).
		Str("wp_site", wpConfig.SiteURL).
		Msg("Publicando no WordPress")

	wpClient := wordpress.NewClient(wpConfig.SiteURL, wpConfig.Username, decryptedPassword)

	wpPost := &wordpress.Post{
		Title:   article.Title,
		Content: htmlContent,
		Status:  "publish",
	}

	if err := wpPost.Validate(); err != nil {
		return fmt.Errorf("post WordPress inválido: %w", err)
	}

	wpResponse, err := wpClient.CreatePost(ctx, wpPost)
	if err != nil {
		return fmt.Errorf("erro ao publicar no WordPress: %w", err)
	}

	// 12. Salvar URL do post e marcar como publicado
	if err := w.articleRepo.SetPublished(ctx, article.ID, wpResponse.Link); err != nil {
		return fmt.Errorf("erro ao marcar como publicado: %w", err)
	}

	log.Info().
		Str("article_id", article.ID.String()).
		Str("post_url", wpResponse.Link).
		Msg("Artigo publicado com sucesso")

	return nil
}

// markdownToHTML converte Markdown para HTML (simplificado)
func markdownToHTML(markdown string) string {
	html := markdown

	// Headers
	html = strings.ReplaceAll(html, "### ", "<h3>")
	html = strings.ReplaceAll(html, "## ", "<h2>")
	html = strings.ReplaceAll(html, "# ", "<h1>")

	// Bold/Italic
	html = strings.ReplaceAll(html, "**", "<strong>")
	html = strings.ReplaceAll(html, "*", "<em>")

	// Quebras de linha
	html = strings.ReplaceAll(html, "\n\n", "</p><p>")
	html = "<p>" + html + "</p>"

	return html
}
