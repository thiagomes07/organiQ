// internal/handler/article_handler.go
package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"organiq/internal/middleware"
	"organiq/internal/usecase/article"
	"organiq/internal/util"
)

// ArticleHandler agrupa handlers de artigos
type ArticleHandler struct {
	listArticlesUC     *article.ListArticlesUseCase
	getArticleUC       *article.GetArticleUseCase
	republishArticleUC *article.RepublishArticleUseCase
	publishArticleUC   *article.PublishArticleUseCase
}

// NewArticleHandler cria nova instância
func NewArticleHandler(
	listArticlesUC *article.ListArticlesUseCase,
	getArticleUC *article.GetArticleUseCase,
	republishArticleUC *article.RepublishArticleUseCase,
	publishArticleUC *article.PublishArticleUseCase,
) *ArticleHandler {
	return &ArticleHandler{
		listArticlesUC:     listArticlesUC,
		getArticleUC:       getArticleUC,
		republishArticleUC: republishArticleUC,
		publishArticleUC:   publishArticleUC,
	}
}

// ============================================
// GET /api/articles
// ============================================

// ListArticlesResponse response da listagem
type ListArticlesResponse struct {
	Articles []*ArticleListItem `json:"articles"`
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	Limit    int                `json:"limit"`
}

// ArticleListItem item de artigo na listagem
type ArticleListItem struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	CreatedAt    string  `json:"createdAt"`
	Status       string  `json:"status"`
	PostURL      *string `json:"postUrl,omitempty"`
	ErrorMessage *string `json:"errorMessage,omitempty"`
}

// ListArticles implementa GET /api/articles
func (h *ArticleHandler) ListArticles(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("ArticleHandler ListArticles iniciado")

	// 1. Extrair user_id do context
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		log.Warn().Msg("ListArticles: usuário não autenticado")
		util.RespondError(w, http.StatusUnauthorized, "unauthorized", "Usuário não autenticado")
		return
	}

	// 2. Parse query params
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	statusFilter := r.URL.Query().Get("status")
	sortBy := r.URL.Query().Get("sort_by")
	order := r.URL.Query().Get("order")

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if statusFilter == "" {
		statusFilter = "all"
	}

	// 3. Executar use case
	input := article.ListArticlesInput{
		UserID:    userID,
		Page:      page,
		Limit:     limit,
		Status:    statusFilter,
		SortBy:    sortBy,
		SortOrder: order,
	}

	output, err := h.listArticlesUC.Execute(r.Context(), input)
	if err != nil {
		log.Error().Err(err).Msg("ListArticles: erro no use case")

		if err.Error() == "filtro de status inválido" {
			util.RespondError(w, http.StatusBadRequest, "invalid_filter", err.Error())
		} else {
			util.RespondError(w, http.StatusInternalServerError, "server_error", "Erro ao buscar artigos")
		}
		return
	}

	// 4. Converter para response
	items := make([]*ArticleListItem, len(output.Articles))
	for i, a := range output.Articles {
		items[i] = &ArticleListItem{
			ID:           a.ID,
			Title:        a.Title,
			CreatedAt:    a.CreatedAt,
			Status:       a.Status,
			PostURL:      a.PostURL,
			ErrorMessage: a.ErrorMessage,
		}
	}

	response := ListArticlesResponse{
		Articles: items,
		Total:    output.Total,
		Page:     output.Page,
		Limit:    output.Limit,
	}

	util.RespondJSON(w, http.StatusOK, response)
}

// ============================================
// GET /api/articles/:id
// ============================================

// GetArticleResponse response do artigo
type GetArticleResponse struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Content      *string `json:"content,omitempty"`
	Status       string  `json:"status"`
	PostURL      *string `json:"postUrl,omitempty"`
	ErrorMessage *string `json:"errorMessage,omitempty"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

// GetArticle implementa GET /api/articles/:id
func (h *ArticleHandler) GetArticle(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("ArticleHandler GetArticle iniciado")

	// 1. Extrair user_id do context
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		log.Warn().Msg("GetArticle: usuário não autenticado")
		util.RespondError(w, http.StatusUnauthorized, "unauthorized", "Usuário não autenticado")
		return
	}

	// 2. Extrair article_id do path
	articleID := chi.URLParam(r, "id")
	if articleID == "" {
		log.Warn().Msg("GetArticle: article_id não fornecido")
		util.RespondError(w, http.StatusBadRequest, "missing_param", "article_id é obrigatório")
		return
	}

	// 3. Executar use case
	input := article.GetArticleInput{
		UserID:    userID,
		ArticleID: articleID,
	}

	output, err := h.getArticleUC.Execute(r.Context(), input)
	if err != nil {
		log.Error().Err(err).Msg("GetArticle: erro no use case")

		if err.Error() == "article_not_found" {
			util.RespondError(w, http.StatusNotFound, "not_found", "Artigo não encontrado")
		} else if err.Error() == "access_denied" {
			util.RespondError(w, http.StatusForbidden, "forbidden", "Acesso negado")
		} else {
			util.RespondError(w, http.StatusInternalServerError, "server_error", "Erro ao buscar artigo")
		}
		return
	}

	// 4. Responder
	response := GetArticleResponse{
		ID:           output.ID,
		Title:        output.Title,
		Content:      output.Content,
		Status:       output.Status,
		PostURL:      output.PostURL,
		ErrorMessage: output.ErrorMessage,
		CreatedAt:    output.CreatedAt,
		UpdatedAt:    output.UpdatedAt,
	}

	util.RespondJSON(w, http.StatusOK, response)
}

// ============================================
// POST /api/articles/:id/republish
// ============================================

// RepublishArticleResponse response da republicação
type RepublishArticleResponse struct {
	ArticleID string `json:"articleId"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// RepublishArticle implementa POST /api/articles/:id/republish
func (h *ArticleHandler) RepublishArticle(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("ArticleHandler RepublishArticle iniciado")

	// 1. Extrair user_id do context
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		log.Warn().Msg("RepublishArticle: usuário não autenticado")
		util.RespondError(w, http.StatusUnauthorized, "unauthorized", "Usuário não autenticado")
		return
	}

	// 2. Extrair article_id do path
	articleID := chi.URLParam(r, "id")
	if articleID == "" {
		log.Warn().Msg("RepublishArticle: article_id não fornecido")
		util.RespondError(w, http.StatusBadRequest, "missing_param", "article_id é obrigatório")
		return
	}

	// 3. Executar use case
	input := article.RepublishArticleInput{
		UserID:    userID,
		ArticleID: articleID,
	}

	output, err := h.republishArticleUC.Execute(r.Context(), input)
	if err != nil {
		log.Error().Err(err).Msg("RepublishArticle: erro no use case")

		if err.Error() == "article_not_found" {
			util.RespondError(w, http.StatusNotFound, "not_found", "Artigo não encontrado")
		} else if err.Error() == "access_denied" {
			util.RespondError(w, http.StatusForbidden, "forbidden", "Acesso negado")
		} else if err.Error() == "artigo não pode ser republicado. Status deve ser 'error' e conteúdo deve existir" {
			util.RespondError(w, http.StatusUnprocessableEntity, "cannot_republish", err.Error())
		} else {
			util.RespondError(w, http.StatusInternalServerError, "server_error", "Erro ao republicar artigo")
		}
		return
	}

	// 4. Responder com 202 Accepted
	response := RepublishArticleResponse{
		ArticleID: output.ArticleID,
		Status:    output.Status,
		Message:   output.Message,
	}

	util.RespondJSON(w, http.StatusAccepted, response)
}

// ============================================
// POST /api/articles/:id/publish
// ============================================

// PublishArticleResponse response da publicação manual
type PublishArticleResponse struct {
	ArticleID string `json:"articleId"`
	Status    string `json:"status"`
	PostURL   string `json:"postUrl"`
}

// PublishArticle implementa POST /api/articles/:id/publish
func (h *ArticleHandler) PublishArticle(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("ArticleHandler PublishArticle iniciado")

	// 1. Extrair user_id do context
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		log.Warn().Msg("PublishArticle: usuário não autenticado")
		util.RespondError(w, http.StatusUnauthorized, "unauthorized", "Usuário não autenticado")
		return
	}

	// 2. Extrair article_id do path
	articleID := chi.URLParam(r, "id")
	if articleID == "" {
		log.Warn().Msg("PublishArticle: article_id não fornecido")
		util.RespondError(w, http.StatusBadRequest, "missing_param", "article_id é obrigatório")
		return
	}

	// 3. Executar use case
	input := article.PublishArticleInput{
		UserID:    userID,
		ArticleID: articleID,
	}

	output, err := h.publishArticleUC.Execute(r.Context(), input)
	if err != nil {
		log.Error().Err(err).Msg("PublishArticle: erro no use case")

		if err.Error() == "article_not_found" {
			util.RespondError(w, http.StatusNotFound, "not_found", "Artigo não encontrado")
		} else if err.Error() == "access_denied" {
			util.RespondError(w, http.StatusForbidden, "forbidden", "Acesso negado")
		} else if err.Error() == "integração WordPress não configurada" {
			util.RespondError(w, http.StatusPreconditionRequired, "integration_required", err.Error())
		} else if err.Error() == "artigo não está pronto para publicação (status incorreto)" {
			util.RespondError(w, http.StatusUnprocessableEntity, "invalid_status", err.Error())
		} else {
			util.RespondError(w, http.StatusInternalServerError, "server_error", "Erro ao publicar artigo")
		}
		return
	}

	// 4. Responder com 200 OK
	response := PublishArticleResponse{
		ArticleID: output.ArticleID,
		Status:    output.Status,
		PostURL:   output.PostURL,
	}

	util.RespondJSON(w, http.StatusOK, response)
}
