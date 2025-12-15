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
	UserID string // UUID como string do context
}

// GenerateIdeasOutput dados de saída
type GenerateIdeasOutput struct {
	JobID  string
	Status string
}

// GenerateIdeasUseCase implementa o caso de uso
type GenerateIdeasUseCase struct {
	userRepo       repository.UserRepository
	businessRepo   repository.BusinessRepository
	articleJobRepo repository.ArticleJobRepository
	queueService   queue.QueueService
}

// NewGenerateIdeasUseCase cria nova instância
func NewGenerateIdeasUseCase(
	userRepo repository.UserRepository,
	businessRepo repository.BusinessRepository,
	articleJobRepo repository.ArticleJobRepository,
	queueService queue.QueueService,
) *GenerateIdeasUseCase {
	return &GenerateIdeasUseCase{
		userRepo:       userRepo,
		businessRepo:   businessRepo,
		articleJobRepo: articleJobRepo,
		queueService:   queueService,
	}
}

// Execute executa o caso de uso
func (uc *GenerateIdeasUseCase) Execute(ctx context.Context, input GenerateIdeasInput) (*GenerateIdeasOutput, error) {
	log.Debug().Str("user_id", input.UserID).Msg("GenerateIdeasUseCase Execute iniciado")

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

	// 3. Buscar perfil de negócio (validar que foi preenchido)
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

	// 4. Buscar competidores
	competitors, err := uc.businessRepo.FindCompetitorsByUserID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("GenerateIdeasUseCase: erro ao buscar competidores")
		return nil, errors.New("erro ao buscar competidores")
	}

	log.Debug().Str("user_id", input.UserID).Int("competitors", len(competitors)).Msg("GenerateIdeasUseCase: competidores carregados")

	// 5. Criar ArticleJob
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
		"competitors": competitors,
		"articleCount": 5, // Fixo por enquanto
	}

	job := &entity.ArticleJob{
		ID:      jobID,
		UserID:  userID,
		Type:    entity.JobTypeGenerateIdeas,
		Status:  entity.JobStatusQueued,
		Progress: 0,
		Payload: payload,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := job.Validate(); err != nil {
		log.Error().Err(err).Msg("GenerateIdeasUseCase: job inválido")
		return nil, errors.New("erro ao criar job")
	}

	// 6. Salvar job no banco (com status queued)
	if err := uc.articleJobRepo.Create(ctx, job); err != nil {
		log.Error().Err(err).Msg("GenerateIdeasUseCase: erro ao salvar job no banco")
		return nil, errors.New("erro ao criar job de geração")
	}

	log.Info().Str("job_id", jobID.String()).Msg("GenerateIdeasUseCase: job criado com sucesso")

	// 7. Enviar mensagem para fila SQS
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
		// Não retornar erro aqui, pois o job já foi criado
		// O sistema deve dar retry
		// Mas atualizar status para falhado
		_ = uc.articleJobRepo.UpdateError(ctx, jobID, "erro ao enviar mensagem para fila")
		return nil, errors.New("erro ao processar geração")
	}

	// Enviar para fila
	if err := uc.queueService.SendMessage(ctx, "article-generation-queue", messageJSON); err != nil {
		log.Error().Err(err).Str("job_id", jobID.String()).Msg("GenerateIdeasUseCase: erro ao enviar mensagem para fila")
		// Atualizar job status para falhado
		_ = uc.articleJobRepo.UpdateError(ctx, jobID, "erro ao enviar para fila de processamento")
		return nil, errors.New("erro ao iniciar processamento")
	}

	log.Info().Str("job_id", jobID.String()).Msg("GenerateIdeasUseCase: mensagem enviada para fila")

	return &GenerateIdeasOutput{
		JobID:  jobID.String(),
		Status: string(entity.JobStatusQueued),
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
