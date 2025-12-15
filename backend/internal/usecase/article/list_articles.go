// internal/usecase/article/list_articles.go
package article

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"organiq/internal/domain/entity"
	"organiq/internal/domain/repository"
)

// ListArticlesInput dados de entrada
type ListArticlesInput struct {
	UserID string // UUID como string do context
	Page   int    // Página (mínimo 1)
	Limit  int    // Itens por página (mínimo 1, máximo 100)
	Status string // "all", "generating", "publishing", "published", "error"
}

// ListArticlesOutput dados de saída
type ListArticlesOutput struct {
	Articles   []*ArticleListItem
	Total      int
	Page       int
	Limit      int
	TotalPages int
}

// ArticleListItem item de artigo na listagem
type ArticleListItem struct {
	ID           string
	Title        string
	CreatedAt    string
	Status       string
	PostURL      *string
	ErrorMessage *string
}

// ListArticlesUseCase implementa o caso de uso
type ListArticlesUseCase struct {
	articleRepo repository.ArticleRepository
}

// NewListArticlesUseCase cria nova instância
func NewListArticlesUseCase(
	articleRepo repository.ArticleRepository,
) *ListArticlesUseCase {
	return &ListArticlesUseCase{
		articleRepo: articleRepo,
	}
}

// Execute executa o caso de uso
func (uc *ListArticlesUseCase) Execute(ctx context.Context, input ListArticlesInput) (*ListArticlesOutput, error) {
	log.Debug().
		Str("user_id", input.UserID).
		Int("page", input.Page).
		Int("limit", input.Limit).
		Str("status", input.Status).
		Msg("ListArticlesUseCase Execute iniciado")

	// 1. Parse user_id
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		log.Error().Err(err).Msg("ListArticlesUseCase: user_id inválido")
		return nil, errors.New("invalid_user_id")
	}

	// 2. Validar paginação
	page := input.Page
	if page < 1 {
		page = 1
	}

	limit := input.Limit
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	// 3. Buscar artigos baseado no filtro de status
	var articles []*entity.Article
	var total int

	statusFilter := input.Status
	if statusFilter == "" {
		statusFilter = "all"
	}

	switch statusFilter {
	case "all":
		articles, err = uc.articleRepo.FindByUserID(ctx, userID, limit, offset)
		if err != nil {
			log.Error().Err(err).Msg("ListArticlesUseCase: erro ao buscar artigos")
			return nil, errors.New("erro ao buscar artigos")
		}

		total, err = uc.articleRepo.CountByUserID(ctx, userID)
		if err != nil {
			log.Error().Err(err).Msg("ListArticlesUseCase: erro ao contar artigos")
			return nil, errors.New("erro ao contar artigos")
		}

	case "generating":
		articles, err = uc.articleRepo.FindByUserIDAndStatus(ctx, userID, entity.ArticleStatusGenerating, limit, offset)
		if err != nil {
			log.Error().Err(err).Msg("ListArticlesUseCase: erro ao buscar artigos generating")
			return nil, errors.New("erro ao buscar artigos")
		}

		total, err = uc.articleRepo.CountByUserIDAndStatus(ctx, userID, entity.ArticleStatusGenerating)
		if err != nil {
			log.Error().Err(err).Msg("ListArticlesUseCase: erro ao contar artigos generating")
			return nil, errors.New("erro ao contar artigos")
		}

	case "publishing":
		articles, err = uc.articleRepo.FindByUserIDAndStatus(ctx, userID, entity.ArticleStatusPublishing, limit, offset)
		if err != nil {
			log.Error().Err(err).Msg("ListArticlesUseCase: erro ao buscar artigos publishing")
			return nil, errors.New("erro ao buscar artigos")
		}

		total, err = uc.articleRepo.CountByUserIDAndStatus(ctx, userID, entity.ArticleStatusPublishing)
		if err != nil {
			log.Error().Err(err).Msg("ListArticlesUseCase: erro ao contar artigos publishing")
			return nil, errors.New("erro ao contar artigos")
		}

	case "published":
		articles, err = uc.articleRepo.FindPublishedByUserID(ctx, userID, limit, offset)
		if err != nil {
			log.Error().Err(err).Msg("ListArticlesUseCase: erro ao buscar artigos published")
			return nil, errors.New("erro ao buscar artigos")
		}

		total, err = uc.articleRepo.CountPublishedByUserID(ctx, userID)
		if err != nil {
			log.Error().Err(err).Msg("ListArticlesUseCase: erro ao contar artigos published")
			return nil, errors.New("erro ao contar artigos")
		}

	case "error":
		articles, err = uc.articleRepo.FindErrorsByUserID(ctx, userID, limit, offset)
		if err != nil {
			log.Error().Err(err).Msg("ListArticlesUseCase: erro ao buscar artigos error")
			return nil, errors.New("erro ao buscar artigos")
		}

		total, err = uc.articleRepo.CountErrorsByUserID(ctx, userID)
		if err != nil {
			log.Error().Err(err).Msg("ListArticlesUseCase: erro ao contar artigos error")
			return nil, errors.New("erro ao contar artigos")
		}

	default:
		log.Warn().Str("status", statusFilter).Msg("ListArticlesUseCase: filtro de status inválido")
		return nil, errors.New("filtro de status inválido")
	}

	// 4. Converter para output
	items := make([]*ArticleListItem, len(articles))
	for i, article := range articles {
		items[i] = &ArticleListItem{
			ID:           article.ID.String(),
			Title:        article.Title,
			CreatedAt:    article.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Status:       string(article.Status),
			PostURL:      article.PostURL,
			ErrorMessage: article.ErrorMessage,
		}
	}

	// 5. Calcular total de páginas
	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}

	log.Debug().
		Str("user_id", input.UserID).
		Int("total", total).
		Int("page", page).
		Int("total_pages", totalPages).
		Msg("ListArticlesUseCase bem-sucedido")

	return &ListArticlesOutput{
		Articles:   items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}
