// internal/usecase/wizard/generate_ideas.go
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

// GenerateIdeasInput dados de entrada
type GenerateIdeasInput struct {
	UserID         string // UUID como string do context
	IsRegeneration bool   // Indica se é uma regeneração de ideias
}

// GenerateIdeasOutput dados de saída
type GenerateIdeasOutput struct {
	JobID                  string
	Status                 string
	RegenerationsRemaining int
	RegenerationsLimit     int
	NextRegenerationAt     *string
}

// GenerateIdeasUseCase implementa o caso de uso
type GenerateIdeasUseCase struct {
	userRepo        repository.UserRepository
	planRepo        repository.PlanRepository
	businessRepo    repository.BusinessRepository
	articleJobRepo  repository.ArticleJobRepository
	articleIdeaRepo repository.ArticleIdeaRepository
	queueService    queue.QueueService
}

// NewGenerateIdeasUseCase cria nova instância
func NewGenerateIdeasUseCase(
	userRepo repository.UserRepository,
	planRepo repository.PlanRepository,
	businessRepo repository.BusinessRepository,
	articleJobRepo repository.ArticleJobRepository,
	articleIdeaRepo repository.ArticleIdeaRepository,
	queueService queue.QueueService,
) *GenerateIdeasUseCase {
	return &GenerateIdeasUseCase{
		userRepo:        userRepo,
		planRepo:        planRepo,
		businessRepo:    businessRepo,
		articleJobRepo:  articleJobRepo,
		articleIdeaRepo: articleIdeaRepo,
		queueService:    queueService,
	}
}

// Execute executa o caso de uso
func (uc *GenerateIdeasUseCase) Execute(ctx context.Context, input GenerateIdeasInput) (*GenerateIdeasOutput, error) {
	log.Debug().Str("user_id", input.UserID).Bool("is_regeneration", input.IsRegeneration).Msg("GenerateIdeasUseCase Execute iniciado")

	// 1. Parse user_id
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		log.Error().Err(err).Msg("GenerateIdeasUseCase: user_id inválido")
		return nil, errors.New("invalid_user_id")
	}

	// 2. Buscar usuário (validar que existe)
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("GenerateIdeasUseCase: erro ao buscar usuário")
		return nil, errors.New("erro ao buscar usuário")
	}

	if user == nil {
		log.Warn().Str("user_id", input.UserID).Msg("GenerateIdeasUseCase: usuário não encontrado")
		return nil, errors.New("user_not_found")
	}

	// 3. Buscar plano do usuário
	plan, err := uc.planRepo.FindByID(ctx, user.PlanID)
	if err != nil {
		log.Error().Err(err).Msg("GenerateIdeasUseCase: erro ao buscar plano")
		return nil, errors.New("erro ao buscar plano")
	}

	if plan == nil {
		log.Warn().Str("plan_id", user.PlanID.String()).Msg("GenerateIdeasUseCase: plano não encontrado")
		return nil, errors.New("plano não encontrado")
	}

	// 4. Se for regeneração, validar limites e limpar ideias anteriores
	articleCount := 5 // Default inicial

	if input.IsRegeneration {
		// Validar limite de regenerações por hora
		generationsInLastHour, err := uc.articleIdeaRepo.CountGenerationsInLastHour(ctx, userID)
		if err != nil {
			log.Error().Err(err).Msg("GenerateIdeasUseCase: erro ao contar gerações")
			return nil, errors.New("erro ao verificar limites")
		}

		if generationsInLastHour > plan.MaxIdeaRegenerationsPerHour {
			log.Warn().
				Int("generations", generationsInLastHour).
				Int("limit", plan.MaxIdeaRegenerationsPerHour).
				Msg("GenerateIdeasUseCase: limite de regeneração excedido")
			
			// Calcular quando poderá regenerar novamente (simples aproximacao: 1 hora)
			nextTime := time.Now().Add(1 * time.Hour).Format(time.RFC3339)
			return &GenerateIdeasOutput{
				Status:                 "limit_exceeded",
				RegenerationsRemaining: 0,
				RegenerationsLimit:     plan.MaxIdeaRegenerationsPerHour,
				NextRegenerationAt:     &nextTime,
			}, errors.New("limite de regeneração por hora excedido")
		}

		// Contar quantas já estão aprovadas
		approvedCount, err := uc.articleIdeaRepo.CountApprovedByUserID(ctx, userID)
		if err != nil {
			log.Error().Err(err).Msg("GenerateIdeasUseCase: erro ao contar aprovadas")
			return nil, errors.New("erro ao processar regeneração")
		}

		// Calcular quantas novas gerar: Total Inicial (5) - Aprovadas
		// Se já aprovou 5 ou mais, não deveria estar regenerando, mas protegemos
		if approvedCount >= 5 {
			log.Warn().Int("approved", approvedCount).Msg("GenerateIdeasUseCase: todas as ideias já aprovadas")
			return nil, errors.New("todas as ideias já foram aprovadas")
		}
		articleCount = 5 - approvedCount

		// Deletar ideias não aprovadas do usuário antes de gerar novas
		if err := uc.articleIdeaRepo.DeleteUnapprovedByUserID(ctx, userID); err != nil {
			log.Error().Err(err).Msg("GenerateIdeasUseCase: erro ao deletar ideias antigas")
			return nil, errors.New("erro ao limpar ideias antigas")
		}

		log.Info().Int("count", articleCount).Msg("GenerateIdeasUseCase: regenerando ideias")
	}

	// 5. Buscar perfil de negócio (validar que foi preenchido)
	businessProfile, err := uc.businessRepo.FindProfileByUserID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("GenerateIdeasUseCase: erro ao buscar business profile")
		return nil, errors.New("erro ao buscar perfil de negócio")
	}

	if businessProfile == nil {
		log.Warn().Str("user_id", input.UserID).Msg("GenerateIdeasUseCase: business profile não preenchido")
		return nil, errors.New("business_profile_not_found")
	}

	if err := businessProfile.Validate(); err != nil {
		log.Warn().Err(err).Msg("GenerateIdeasUseCase: business profile inválido")
		return nil, errors.New("business_profile_incomplete")
	}

	// 6. Buscar competidores
	competitors, err := uc.businessRepo.FindCompetitorsByUserID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("GenerateIdeasUseCase: erro ao buscar competidores")
		return nil, errors.New("erro ao buscar competidores")
	}

	log.Debug().Str("user_id", input.UserID).Int("competitors", len(competitors)).Int("article_count", articleCount).Msg("GenerateIdeasUseCase: dados preparados")

	// 7. Criar ArticleJob
	jobID := uuid.New()

	// Payload contém os dados necessários para o worker
	payload := map[string]interface{}{
		"userID": userID.String(),
		"businessProfile": map[string]interface{}{
			"description":       businessProfile.Description,
			"primaryObjective":  businessProfile.PrimaryObjective,
			"location":          businessProfile.Location,
			"siteURL":           businessProfile.SiteURL,
			"hasBlog":           businessProfile.HasBlog,
			"blogURLs":          businessProfile.BlogURLs,
		},
		"competitors":  competitors,
		"articleCount": articleCount,
	}

	job := &entity.ArticleJob{
		ID:        jobID,
		UserID:    userID,
		Type:      entity.JobTypeGenerateIdeas,
		Status:    entity.JobStatusQueued,
		Progress:  0,
		Payload:   payload,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := job.Validate(); err != nil {
		log.Error().Err(err).Msg("GenerateIdeasUseCase: job inválido")
		return nil, errors.New("erro ao criar job")
	}

	// 8. Salvar job no banco (com status queued)
	if err := uc.articleJobRepo.Create(ctx, job); err != nil {
		log.Error().Err(err).Msg("GenerateIdeasUseCase: erro ao salvar job no banco")
		return nil, errors.New("erro ao criar job de geração")
	}

	log.Info().Str("job_id", jobID.String()).Msg("GenerateIdeasUseCase: job criado com sucesso")

	// 9. Enviar mensagem para fila SQS
	// Estrutura da mensagem que o worker vai processar
	queueMessage := map[string]interface{}{
		"jobID":   jobID.String(),
		"userID":  userID.String(),
		"type":    "generate_ideas",
		"payload": payload,
	}

	messageJSON, err := json.Marshal(queueMessage)
	if err != nil {
		log.Error().Err(err).Msg("GenerateIdeasUseCase: erro ao serializar mensagem")
		_ = uc.articleJobRepo.UpdateError(ctx, jobID, "erro ao enviar mensagem para fila")
		return nil, errors.New("erro ao processar geração")
	}

	// Enviar para fila
	if err := uc.queueService.SendMessage(ctx, "article-generation-queue", messageJSON); err != nil {
		log.Error().Err(err).Str("job_id", jobID.String()).Msg("GenerateIdeasUseCase: erro ao enviar mensagem para fila")
		_ = uc.articleJobRepo.UpdateError(ctx, jobID, "erro ao enviar para fila de processamento")
		return nil, errors.New("erro ao iniciar processamento")
	}

	log.Info().Str("job_id", jobID.String()).Msg("GenerateIdeasUseCase: mensagem enviada para fila")

	// Calcular regenerações restantes para retorno
	// generationsInLastHour conta TODAS as gerações (inclusive inicial)
	// Limite "Regenerations" = Total (Max + 1) - Usadas (Count)
	// Se Max é 3, Total permitido é 4.
	// Se Count é 1 (só inicial), Restante = (3+1) - 1 = 3. OK.
	// Se Count é 4 (inicial + 3 regen), Restante = (3+1) - 4 = 0. OK.
	
	count, _ := uc.articleIdeaRepo.CountGenerationsInLastHour(ctx, userID)
	
	// Se estamos numa regeneração, o count já incluiu (se o job foi rápido) ou ainda não.
	// O job acabou de ser criado e article_ideas ainda não foram criadas (worker faz isso).
	// Portanto, o CountGenerationsInLastHour NÃO deve ter mudado ainda para ESTE job.
	// Então, se foi regeneração, o uso efetivo será count + 1.
	// MAS se foi a primeira geração, também será count + 1.
	// Na verdade, queremos o "saldo após esta operação".
	// currentUsage = count + 1 (este job).
	
	currentUsage := count + 1
	remaining := (plan.MaxIdeaRegenerationsPerHour + 1) - currentUsage
	if remaining < 0 {
		remaining = 0
	}

	return &GenerateIdeasOutput{
		JobID:                  jobID.String(),
		Status:                 string(entity.JobStatusQueued),
		RegenerationsRemaining: remaining,
		RegenerationsLimit:     plan.MaxIdeaRegenerationsPerHour,
	}, nil
}

// ============================================
// GET IDEAS STATUS
// ============================================

// GetIdeasStatusInput dados de entrada
type GetIdeasStatusInput struct {
	UserID string
	JobID  string
}

// GetIdeasStatusOutput dados de saída
type GetIdeasStatusOutput struct {
	JobID    string
	Status   string
	Progress int
	Message  string
	Ideas    []*IdeaResponse
	ErrorMsg *string
}

// IdeaResponse resposta de uma ideia
type IdeaResponse struct {
	ID       string
	Title    string
	Summary  string
	Approved bool
	Feedback *string
}

// GetIdeasStatusUseCase implementa o caso de uso
type GetIdeasStatusUseCase struct {
	userRepo       repository.UserRepository
	articleJobRepo repository.ArticleJobRepository
	articleIdeaRepo repository.ArticleIdeaRepository
}

// NewGetIdeasStatusUseCase cria nova instância
func NewGetIdeasStatusUseCase(
	userRepo repository.UserRepository,
	articleJobRepo repository.ArticleJobRepository,
	articleIdeaRepo repository.ArticleIdeaRepository,
) *GetIdeasStatusUseCase {
	return &GetIdeasStatusUseCase{
		userRepo:        userRepo,
		articleJobRepo:  articleJobRepo,
		articleIdeaRepo: articleIdeaRepo,
	}
}

// Execute executa o caso de uso
func (uc *GetIdeasStatusUseCase) Execute(ctx context.Context, input GetIdeasStatusInput) (*GetIdeasStatusOutput, error) {
	log.Debug().
		Str("user_id", input.UserID).
		Str("job_id", input.JobID).
		Msg("GetIdeasStatusUseCase Execute iniciado")

	// 1. Parse IDs
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		log.Error().Err(err).Msg("GetIdeasStatusUseCase: user_id inválido")
		return nil, errors.New("invalid_user_id")
	}

	jobID, err := uuid.Parse(input.JobID)
	if err != nil {
		log.Error().Err(err).Msg("GetIdeasStatusUseCase: job_id inválido")
		return nil, errors.New("invalid_job_id")
	}

	// 2. Buscar job
	job, err := uc.articleJobRepo.FindByID(ctx, jobID)
	if err != nil {
		log.Error().Err(err).Msg("GetIdeasStatusUseCase: erro ao buscar job")
		return nil, errors.New("erro ao buscar status")
	}

	if job == nil {
		log.Warn().Str("job_id", input.JobID).Msg("GetIdeasStatusUseCase: job não encontrado")
		return nil, errors.New("job_not_found")
	}

	// 3. Validar ownership (job pertence ao usuário)
	if job.UserID != userID {
		log.Warn().
			Str("job_user_id", job.UserID.String()).
			Str("request_user_id", input.UserID).
			Msg("GetIdeasStatusUseCase: acesso negado")
		return nil, errors.New("access_denied")
	}

	// 4. Validar que é job de geração de ideias
	if job.Type != entity.JobTypeGenerateIdeas {
		log.Warn().Str("job_type", string(job.Type)).Msg("GetIdeasStatusUseCase: tipo de job incorreto")
		return nil, errors.New("invalid_job_type")
	}

	// 5. Construir output
	output := &GetIdeasStatusOutput{
		JobID:    job.ID.String(),
		Status:   string(job.Status),
		Progress: job.Progress,
		Message:  getStatusMessage(job.Status),
		ErrorMsg: job.ErrorMessage,
	}

	// 6. Se job completado ou falhado, buscar ideias
	if job.Status == entity.JobStatusCompleted || job.Status == entity.JobStatusFailed {
		ideas, err := uc.articleIdeaRepo.FindByJobID(ctx, jobID)
		if err != nil {
			log.Error().Err(err).Msg("GetIdeasStatusUseCase: erro ao buscar ideias")
			return nil, errors.New("erro ao buscar ideias")
		}

		// Converter para response
		output.Ideas = make([]*IdeaResponse, 0, len(ideas))
		for _, idea := range ideas {
			output.Ideas = append(output.Ideas, &IdeaResponse{
				ID:       idea.ID.String(),
				Title:    idea.Title,
				Summary:  idea.Summary,
				Approved: idea.Approved,
				Feedback: idea.Feedback,
			})
		}
	}

	log.Debug().
		Str("job_id", input.JobID).
		Str("status", output.Status).
		Int("progress", output.Progress).
		Msg("GetIdeasStatusUseCase bem-sucedido")

	return output, nil
}

// ============================================
// HELPERS
// ============================================

func getStatusMessage(status entity.JobStatus) string {
	switch status {
	case entity.JobStatusQueued:
		return "Aguardando processamento..."
	case entity.JobStatusProcessing:
		return "Gerando ideias de artigos..."
	case entity.JobStatusCompleted:
		return "Ideias geradas com sucesso!"
	case entity.JobStatusFailed:
		return "Erro ao gerar ideias. Tente novamente."
	default:
		return "Status desconhecido"
	}
}
