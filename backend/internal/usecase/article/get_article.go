// internal/usecase/article/get_article.go
package article

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"organiq/internal/domain/repository"
)

// GetArticleInput dados de entrada
type GetArticleInput struct {
	UserID    string // UUID como string do context
	ArticleID string // UUID do artigo
}

// GetArticleOutput dados de saída
type GetArticleOutput struct {
	ID           string
	Title        string
	Content      *string
	Status       string
	PostURL      *string
	ErrorMessage *string
	CreatedAt    string
	UpdatedAt    string
}

// GetArticleUseCase implementa o caso de uso
type GetArticleUseCase struct {
	articleRepo repository.ArticleRepository
}

// NewGetArticleUseCase cria nova instância
func NewGetArticleUseCase(
	articleRepo repository.ArticleRepository,
) *GetArticleUseCase {
	return &GetArticleUseCase{
		articleRepo: articleRepo,
	}
}

// Execute executa o caso de uso
func (uc *GetArticleUseCase) Execute(ctx context.Context, input GetArticleInput) (*GetArticleOutput, error) {
	log.Debug().
		Str("user_id", input.UserID).
		Str("article_id", input.ArticleID).
		Msg("GetArticleUseCase Execute iniciado")

	// 1. Parse IDs
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		log.Error().Err(err).Msg("GetArticleUseCase: user_id inválido")
		return nil, errors.New("invalid_user_id")
	}

	articleID, err := uuid.Parse(input.ArticleID)
	if err != nil {
		log.Error().Err(err).Msg("GetArticleUseCase: article_id inválido")
		return nil, errors.New("invalid_article_id")
	}

	// 2. Buscar artigo
	article, err := uc.articleRepo.FindByID(ctx, articleID)
	if err != nil {
		log.Error().Err(err).Str("article_id", input.ArticleID).Msg("GetArticleUseCase: erro ao buscar artigo")
		return nil, errors.New("erro ao buscar artigo")
	}

	if article == nil {
		log.Warn().Str("article_id", input.ArticleID).Msg("GetArticleUseCase: artigo não encontrado")
		return nil, errors.New("article_not_found")
	}

	// 3. Validar ownership
	if article.UserID != userID {
		log.Warn().
			Str("article_user_id", article.UserID.String()).
			Str("request_user_id", input.UserID).
			Msg("GetArticleUseCase: acesso negado")
		return nil, errors.New("access_denied")
	}

	// 4. Retornar detalhes
	log.Debug().
		Str("article_id", input.ArticleID).
		Str("status", string(article.Status)).
		Msg("GetArticleUseCase bem-sucedido")

	return &GetArticleOutput{
		ID:           article.ID.String(),
		Title:        article.Title,
		Content:      article.Content,
		Status:       string(article.Status),
		PostURL:      article.PostURL,
		ErrorMessage: article.ErrorMessage,
		CreatedAt:    article.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    article.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
