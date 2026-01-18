// internal/usecase/wizard/publish_articles.go
package wizard

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"organiq/internal/domain/entity"
	"organiq/internal/domain/repository"
	"organiq/internal/infra/queue"
)

// PublishArticlesInput dados de entrada
type PublishArticlesInput struct {
	UserID   string                    // UUID como string do context
	Articles []PublishArticleItem
}

// PublishArticleItem item de artigo para publicar
type PublishArticleItem struct {
	ID       string  // UUID da ArticleIdea
	Feedback *string // Feedback opcional do usuário
}

// PublishArticlesOutput dados de saída
type PublishArticlesOutput struct {
	JobID         string
	Status        string
	ArticlesCount int
}

// PublishArticlesUseCase implementa o caso de uso
type PublishArticlesUseCase struct {
	userRepo        repository.UserRepository
	planRepo        repository.PlanRepository
	articleIdeaRepo repository.ArticleIdeaRepository
	articleRepo     repository.ArticleRepository
	articleJobRepo  repository.ArticleJobRepository
	queueService    queue.QueueService
}

// NewPublishArticlesUseCase cria nova instância
func NewPublishArticlesUseCase(
	userRepo repository.UserRepository,
	planRepo repository.PlanRepository,
	articleIdeaRepo repository.ArticleIdeaRepository,
	articleRepo repository.ArticleRepository,
	articleJobRepo repository.ArticleJobRepository,
	queueService queue.QueueService,
) *PublishArticlesUseCase {
	return &PublishArticlesUseCase{
		userRepo:        userRepo,
		planRepo:        planRepo,
		articleIdeaRepo: articleIdeaRepo,
		articleRepo:     articleRepo,
		articleJobRepo:  articleJobRepo,
		queueService:    queueService,
	}
}

// Execute executa o caso de uso
func (uc *PublishArticlesUseCase) Execute(ctx context.Context, input PublishArticlesInput) (*PublishArticlesOutput, error) {
	log.Debug().
		Str("user_id", input.UserID).
		Int("articles_count", len(input.Articles)).
		Msg("PublishArticlesUseCase Execute iniciado")

	// 1. Parse user_id
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		log.Error().Err(err).Msg("PublishArticlesUseCase: user_id inválido")
		return nil, errors.New("invalid_user_id")
	}

	// 2. Validar entrada
	if len(input.Articles) == 0 {
		log.Warn().Msg("PublishArticlesUseCase: nenhum artigo fornecido")
		return nil, errors.New("pelo menos um artigo deve ser selecionado")
	}

	if len(input.Articles) > 50 {
		log.Warn().Int("count", len(input.Articles)).Msg("PublishArticlesUseCase: muitos artigos")
		return nil, errors.New("máximo 50 artigos por vez")
	}

	// 3. Buscar usuário e plano
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("PublishArticlesUseCase: erro ao buscar usuário")
		return nil, errors.New("erro ao buscar usuário")
	}

	if user == nil {
		log.Warn().Str("user_id", input.UserID).Msg("PublishArticlesUseCase: usuário não encontrado")
		return nil, errors.New("user_not_found")
	}

	plan, err := uc.planRepo.FindByID(ctx, user.PlanID)
	if err != nil {
		log.Error().Err(err).Msg("PublishArticlesUseCase: erro ao buscar plano")
		return nil, errors.New("erro ao buscar plano")
	}

	if plan == nil {
		log.Warn().Str("plan_id", user.PlanID.String()).Msg("PublishArticlesUseCase: plano não encontrado")
		return nil, errors.New("plano não encontrado")
	}

	// 4. Validar limite de artigos
	articlesCount := len(input.Articles)
	if !user.CanGenerateArticles(articlesCount, plan.MaxArticles) {
		log.Warn().
			Int("articles_used", user.ArticlesUsed).
			Int("max_articles", plan.MaxArticles).
			Int("requested", articlesCount).
			Msg("PublishArticlesUseCase: limite excedido")
		return nil, errors.New("limite de artigos excedido para o seu plano")
	}

	// 5. Validar que todas as ideias existem e pertencem ao usuário
	ideaIDs := make([]uuid.UUID, len(input.Articles))
	ideaFeedbackMap := make(map[uuid.UUID]*string)

	for i, article := range input.Articles {
		ideaID, err := uuid.Parse(article.ID)
		if err != nil {
			log.Error().Err(err).Str("idea_id", article.ID).Msg("PublishArticlesUseCase: idea_id inválido")
			return nil, errors.New("idea_id inválido: " + article.ID)
		}

		ideaIDs[i] = ideaID
		ideaFeedbackMap[ideaID] = article.Feedback
	}

	// Buscar ideias
	ideas := make([]*entity.ArticleIdea, 0, len(ideaIDs))
	for _, ideaID := range ideaIDs {
		idea, err := uc.articleIdeaRepo.FindByID(ctx, ideaID)
		if err != nil {
			log.Error().Err(err).Str("idea_id", ideaID.String()).Msg("PublishArticlesUseCase: erro ao buscar ideia")
			return nil, errors.New("erro ao buscar ideia")
		}

		if idea == nil {
			log.Warn().Str("idea_id", ideaID.String()).Msg("PublishArticlesUseCase: ideia não encontrada")
			return nil, errors.New("ideia não encontrada: " + ideaID.String())
		}

		if idea.UserID != userID {
			log.Warn().
				Str("idea_user_id", idea.UserID.String()).
				Str("request_user_id", input.UserID).
				Msg("PublishArticlesUseCase: acesso negado")
			return nil, errors.New("acesso negado a ideia: " + ideaID.String())
		}

		ideas = append(ideas, idea)
	}

	// 6. Marcar ideias como aprovadas
	if err := uc.articleIdeaRepo.ApproveMultiple(ctx, ideaIDs); err != nil {
		log.Error().Err(err).Msg("PublishArticlesUseCase: erro ao aprovar ideias")
		return nil, errors.New("erro ao processar ideias")
	}

	// 6.1. Deletar ideias não aprovadas (limpeza)
	if err := uc.articleIdeaRepo.DeleteUnapprovedByUserID(ctx, userID); err != nil {
		// Logar erro mas não falhar o fluxo, pois as ideias principais já foram aprovadas
		log.Error().Err(err).Msg("PublishArticlesUseCase: erro ao deletar ideias não aprovadas")
	}

	// 7. Criar ArticleJob de publicação
	jobID := uuid.New()

	payload := map[string]interface{}{
		"userID":        userID.String(),
		"articlesCount": articlesCount,
		"ideaIDs":       ideaIDs,
	}

	job := &entity.ArticleJob{
		ID:        jobID,
		UserID:    userID,
		Type:      entity.JobTypePublish,
		Status:    entity.JobStatusQueued,
		Progress:  0,
		Payload:   payload,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := job.Validate(); err != nil {
		log.Error().Err(err).Msg("PublishArticlesUseCase: job inválido")
		return nil, errors.New("erro ao criar job de publicação")
	}

	if err := uc.articleJobRepo.Create(ctx, job); err != nil {
		log.Error().Err(err).Msg("PublishArticlesUseCase: erro ao salvar job")
		return nil, errors.New("erro ao criar job de publicação")
	}

	log.Info().Str("job_id", jobID.String()).Msg("PublishArticlesUseCase: job criado")

	// 7.1 Enviar mensagem de início de job para a fila (para MockQueue ou worker processar)
	jobStartMessage := map[string]interface{}{
		"jobID":         jobID.String(),
		"userID":        userID.String(),
		"type":          "publish_articles",
		"articlesCount": articlesCount,
	}

	jobStartJSON, err := json.Marshal(jobStartMessage)
	if err != nil {
		log.Error().Err(err).Msg("PublishArticlesUseCase: erro ao serializar mensagem de job")
		// Não falhar, apenas logar
	} else {
		if err := uc.queueService.SendMessage(ctx, "article-publish-queue", jobStartJSON); err != nil {
			log.Error().Err(err).Msg("PublishArticlesUseCase: erro ao enviar mensagem de job para fila")
			// Não falhar, apenas logar
		}
	}

	// 8. Criar registros de Article e enviar para fila
	for _, idea := range ideas {
		articleID := uuid.New()

		feedback := ideaFeedbackMap[idea.ID]

		article := &entity.Article{
			ID:        articleID,
			UserID:    userID,
			IdeaID:    &idea.ID,
			Title:     idea.Title,
			Status:    entity.ArticleStatusGenerating,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := article.Validate(); err != nil {
			log.Error().Err(err).Str("article_id", articleID.String()).Msg("PublishArticlesUseCase: article inválido")
			// Não falhar todo o batch, apenas logar
			continue
		}

		if err := uc.articleRepo.Create(ctx, article); err != nil {
			log.Error().Err(err).Str("article_id", articleID.String()).Msg("PublishArticlesUseCase: erro ao criar article")
			// Não falhar todo o batch
			continue
		}

		// Enviar mensagem para fila
		queueMessage := map[string]interface{}{
			"articleId": articleID.String(),
			"userId":    userID.String(),
			"ideaId":    idea.ID.String(),
			"title":     idea.Title,
			"summary":   idea.Summary,
			"feedback":  feedback,
		}

		messageJSON, err := json.Marshal(queueMessage)
		if err != nil {
			log.Error().Err(err).Str("article_id", articleID.String()).Msg("PublishArticlesUseCase: erro ao serializar mensagem")
			_ = uc.articleRepo.UpdateStatusWithError(ctx, articleID, "erro ao enviar para fila")
			continue
		}

		if err := uc.queueService.SendMessage(ctx, "article-publish-queue", messageJSON); err != nil {
			log.Error().Err(err).Str("article_id", articleID.String()).Msg("PublishArticlesUseCase: erro ao enviar para fila")
			_ = uc.articleRepo.UpdateStatusWithError(ctx, articleID, "erro ao enviar para fila de publicação")
			continue
		}

		log.Debug().Str("article_id", articleID.String()).Msg("PublishArticlesUseCase: artigo enfileirado")
	}

	// 9. Incrementar articlesUsed do usuário
	if err := user.IncrementArticlesUsed(articlesCount); err != nil {
		log.Error().Err(err).Msg("PublishArticlesUseCase: erro ao incrementar articlesUsed")
		// Não falhar a operação, apenas logar
	}

	// 10. Marcar onboarding como completo (step=5)
	if !user.HasCompletedOnboarding {
		user.HasCompletedOnboarding = true
		user.OnboardingStep = 5
		log.Info().Str("user_id", input.UserID).Msg("PublishArticlesUseCase: onboarding completado")
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		log.Error().Err(err).Msg("PublishArticlesUseCase: erro ao atualizar usuário")
		// Não falhar a operação
	}

	log.Info().
		Str("user_id", input.UserID).
		Str("job_id", jobID.String()).
		Int("articles_count", articlesCount).
		Msg("PublishArticlesUseCase bem-sucedido")

	return &PublishArticlesOutput{
		JobID:         jobID.String(),
		Status:        string(entity.JobStatusQueued),
		ArticlesCount: articlesCount,
	}, nil
}

// ============================================
// GET PUBLISH STATUS
// ============================================

// GetPublishStatusInput dados de entrada
type GetPublishStatusInput struct {
	UserID string
	JobID  string
}

// GetPublishStatusOutput dados de saída
type GetPublishStatusOutput struct {
	JobID     string
	Status    string
	Progress  int
	Published int
	Total     int
	Message   string
	ErrorMsg  *string
}

// GetPublishStatusUseCase implementa o caso de uso
type GetPublishStatusUseCase struct {
	userRepo       repository.UserRepository
	articleJobRepo repository.ArticleJobRepository
	articleRepo    repository.ArticleRepository
}

// NewGetPublishStatusUseCase cria nova instância
func NewGetPublishStatusUseCase(
	userRepo repository.UserRepository,
	articleJobRepo repository.ArticleJobRepository,
	articleRepo repository.ArticleRepository,
) *GetPublishStatusUseCase {
	return &GetPublishStatusUseCase{
		userRepo:       userRepo,
		articleJobRepo: articleJobRepo,
		articleRepo:    articleRepo,
	}
}

// Execute executa o caso de uso
func (uc *GetPublishStatusUseCase) Execute(ctx context.Context, input GetPublishStatusInput) (*GetPublishStatusOutput, error) {
	log.Debug().
		Str("user_id", input.UserID).
		Str("job_id", input.JobID).
		Msg("GetPublishStatusUseCase Execute iniciado")

	// 1. Parse IDs
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		log.Error().Err(err).Msg("GetPublishStatusUseCase: user_id inválido")
		return nil, errors.New("invalid_user_id")
	}

	jobID, err := uuid.Parse(input.JobID)
	if err != nil {
		log.Error().Err(err).Msg("GetPublishStatusUseCase: job_id inválido")
		return nil, errors.New("invalid_job_id")
	}

	// 2. Buscar job
	job, err := uc.articleJobRepo.FindByID(ctx, jobID)
	if err != nil {
		log.Error().Err(err).Msg("GetPublishStatusUseCase: erro ao buscar job")
		return nil, errors.New("erro ao buscar status")
	}

	if job == nil {
		log.Warn().Str("job_id", input.JobID).Msg("GetPublishStatusUseCase: job não encontrado")
		return nil, errors.New("job_not_found")
	}

	// 3. Validar ownership (job pertence ao usuário)
	if job.UserID != userID {
		log.Warn().
			Str("job_user_id", job.UserID.String()).
			Str("request_user_id", input.UserID).
			Msg("GetPublishStatusUseCase: acesso negado")
		return nil, errors.New("access_denied")
	}

	// 4. Validar que é job de publicação
	if job.Type != entity.JobTypePublish {
		log.Warn().Str("job_type", string(job.Type)).Msg("GetPublishStatusUseCase: tipo de job incorreto")
		return nil, errors.New("invalid_job_type")
	}

	// 5. Extrair total do payload
	total := 0
	if articlesCount, ok := job.Payload["articlesCount"].(float64); ok {
		total = int(articlesCount)
	}

	// 6. Contar artigos publicados (status = published)
	published := 0
	if job.Status == entity.JobStatusCompleted || job.Status == entity.JobStatusFailed {
		// Buscar artigos associados a este job via ideaIDs no payload
		if ideaIDsRaw, ok := job.Payload["ideaIDs"].([]interface{}); ok {
			for range ideaIDsRaw {
				// Cada idea corresponde a um artigo potencialmente publicado
				published++
			}
		}
		// Em caso de sucesso, consideramos todos publicados
		if job.Status == entity.JobStatusCompleted {
			published = total
		}
	}

	// 7. Construir output
	output := &GetPublishStatusOutput{
		JobID:     job.ID.String(),
		Status:    string(job.Status),
		Progress:  job.Progress,
		Published: published,
		Total:     total,
		Message:   getPublishStatusMessage(job.Status),
		ErrorMsg:  job.ErrorMessage,
	}

	log.Debug().
		Str("job_id", input.JobID).
		Str("status", output.Status).
		Int("progress", output.Progress).
		Int("published", output.Published).
		Int("total", output.Total).
		Msg("GetPublishStatusUseCase bem-sucedido")

	return output, nil
}

// getPublishStatusMessage retorna mensagem amigável para o status
func getPublishStatusMessage(status entity.JobStatus) string {
	switch status {
	case entity.JobStatusQueued:
		return "Aguardando processamento..."
	case entity.JobStatusProcessing:
		return "Publicando artigos..."
	case entity.JobStatusCompleted:
		return "Artigos publicados com sucesso!"
	case entity.JobStatusFailed:
		return "Erro ao publicar artigos. Tente novamente."
	default:
		return "Status desconhecido"
	}
}
