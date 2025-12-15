// internal/usecase/article/republish_article.go
package article

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"organiq/internal/domain/entity"
	"organiq/internal/domain/repository"
	"organiq/internal/infra/queue"
)

// RepublishArticleInput dados de entrada
type RepublishArticleInput struct {
	UserID    string // UUID como string do context
	ArticleID string // UUID do artigo
}

// RepublishArticleOutput dados de saída
type RepublishArticleOutput struct {
	ArticleID string
	Status    string
	Message   string
}

// RepublishArticleUseCase implementa o caso de uso
type RepublishArticleUseCase struct {
	articleRepo     repository.ArticleRepository
	articleIdeaRepo repository.ArticleIdeaRepository
	queueService    queue.QueueService
}

// NewRepublishArticleUseCase cria nova instância
func NewRepublishArticleUseCase(
	articleRepo repository.ArticleRepository,
	articleIdeaRepo repository.ArticleIdeaRepository,
	queueService queue.QueueService,
) *RepublishArticleUseCase {
	return &RepublishArticleUseCase{
		articleRepo:     articleRepo,
		articleIdeaRepo: articleIdeaRepo,
		queueService:    queueService,
	}
}

// Execute executa o caso de uso
func (uc *RepublishArticleUseCase) Execute(ctx context.Context, input RepublishArticleInput) (*RepublishArticleOutput, error) {
	log.Debug().
		Str("user_id", input.UserID).
		Str("article_id", input.ArticleID).
		Msg("RepublishArticleUseCase Execute iniciado")

	// 1. Parse IDs
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		log.Error().Err(err).Msg("RepublishArticleUseCase: user_id inválido")
		return nil, errors.New("invalid_user_id")
	}

	articleID, err := uuid.Parse(input.ArticleID)
	if err != nil {
		log.Error().Err(err).Msg("RepublishArticleUseCase: article_id inválido")
		return nil, errors.New("invalid_article_id")
	}

	// 2. Buscar artigo
	article, err := uc.articleRepo.FindByID(ctx, articleID)
	if err != nil {
		log.Error().Err(err).Str("article_id", input.ArticleID).Msg("RepublishArticleUseCase: erro ao buscar artigo")
		return nil, errors.New("erro ao buscar artigo")
	}

	if article == nil {
		log.Warn().Str("article_id", input.ArticleID).Msg("RepublishArticleUseCase: artigo não encontrado")
		return nil, errors.New("article_not_found")
	}

	// 3. Validar ownership
	if article.UserID != userID {
		log.Warn().
			Str("article_user_id", article.UserID.String()).
			Str("request_user_id", input.UserID).
			Msg("RepublishArticleUseCase: acesso negado")
		return nil, errors.New("access_denied")
	}

	// 4. Validar que artigo está em erro e tem conteúdo
	if !article.CanRetry() {
		log.Warn().
			Str("article_id", input.ArticleID).
			Str("status", string(article.Status)).
			Bool("has_content", article.Content != nil).
			Msg("RepublishArticleUseCase: artigo não pode ser republicado")
		return nil, errors.New("artigo não pode ser republicado. Status deve ser 'error' e conteúdo deve existir")
	}

	// 5. Buscar ideia original (para pegar summary/feedback)
	var ideaSummary string
	var ideaFeedback *string

	if article.IdeaID != nil {
		idea, err := uc.articleIdeaRepo.FindByID(ctx, *article.IdeaID)
		if err != nil {
			log.Error().Err(err).Msg("RepublishArticleUseCase: erro ao buscar ideia")
			// Não falhar, apenas continuar sem summary
		} else if idea != nil {
			ideaSummary = idea.Summary
			ideaFeedback = idea.Feedback
		}
	}

	// 6. Atualizar status para publishing
	if err := uc.articleRepo.UpdateStatus(ctx, articleID, entity.ArticleStatusPublishing); err != nil {
		log.Error().Err(err).Msg("RepublishArticleUseCase: erro ao atualizar status")
		return nil, errors.New("erro ao atualizar status do artigo")
	}

	// 7. Enviar para fila (mesmo worker de publicação)
	queueMessage := map[string]interface{}{
		"articleId": articleID.String(),
		"userId":    userID.String(),
		"ideaId":    nil, // Pode ser nil em republish
		"title":     article.Title,
		"summary":   ideaSummary,
		"feedback":  ideaFeedback,
		"isRetry":   true, // Flag indicando que é retry
	}

	if article.IdeaID != nil {
		queueMessage["ideaId"] = article.IdeaID.String()
	}

	messageJSON, err := json.Marshal(queueMessage)
	if err != nil {
		log.Error().Err(err).Msg("RepublishArticleUseCase: erro ao serializar mensagem")
		_ = uc.articleRepo.UpdateStatusWithError(ctx, articleID, "erro ao enviar para fila")
		return nil, errors.New("erro ao processar republicação")
	}

	if err := uc.queueService.SendMessage(ctx, "article-publish-queue", messageJSON); err != nil {
		log.Error().Err(err).Msg("RepublishArticleUseCase: erro ao enviar para fila")
		_ = uc.articleRepo.UpdateStatusWithError(ctx, articleID, "erro ao enviar para fila de publicação")
		return nil, errors.New("erro ao enviar para fila de publicação")
	}

	log.Info().
		Str("article_id", input.ArticleID).
		Msg("RepublishArticleUseCase: republicação enfileirada")

	return &RepublishArticleOutput{
		ArticleID: articleID.String(),
		Status:    string(entity.ArticleStatusPublishing),
		Message:   "Republicação iniciada com sucesso",
	}, nil
}
