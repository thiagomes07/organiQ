// internal/usecase/wizard/get_wizard_data.go
package wizard

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"organiq/internal/domain/entity"
	"organiq/internal/domain/repository"
)

// GetWizardDataInput dados de entrada
type GetWizardDataInput struct {
	UserID string // UUID como string do context
}

// BusinessDataOutput dados do negócio
type BusinessDataOutput struct {
	Description        string           `json:"description"`
	PrimaryObjective   string           `json:"primaryObjective"`
	SecondaryObjective *string          `json:"secondaryObjective,omitempty"`
	Location           *entity.Location `json:"location,omitempty"`
	SiteURL            *string          `json:"siteUrl,omitempty"`
	HasBlog            bool             `json:"hasBlog"`
	BlogURLs           []string         `json:"blogUrls,omitempty"`
	BrandFileURL       *string          `json:"brandFileUrl,omitempty"`
}

// PendingIdeaOutput dados de uma ideia pendente
type PendingIdeaOutput struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Summary  string  `json:"summary"`
	Approved bool    `json:"approved"`
	Feedback *string `json:"feedback,omitempty"`
}

// GetWizardDataOutput dados de saída
type GetWizardDataOutput struct {
	OnboardingStep         int                 `json:"onboardingStep"`
	Business               *BusinessDataOutput `json:"business,omitempty"`
	Competitors            []string            `json:"competitors,omitempty"`
	HasIntegration         bool                `json:"hasIntegration"`
	PendingIdeas           []PendingIdeaOutput `json:"pendingIdeas,omitempty"`
	HasGeneratedIdeas      bool                `json:"hasGeneratedIdeas"`
	TotalIdeasCount        int                 `json:"totalIdeasCount"`
	ApprovedIdeasCount     int                 `json:"approvedIdeasCount"`
	RegenerationsRemaining int                 `json:"regenerationsRemaining"`
	RegenerationsLimit     int                 `json:"regenerationsLimit"`
	NextRegenerationAt     *string             `json:"nextRegenerationAt,omitempty"`
}

// GetWizardDataUseCase implementa o caso de uso
type GetWizardDataUseCase struct {
	userRepo        repository.UserRepository
	planRepo        repository.PlanRepository
	businessRepo    repository.BusinessRepository
	integrationRepo repository.IntegrationRepository
	articleIdeaRepo repository.ArticleIdeaRepository
}

// NewGetWizardDataUseCase cria nova instância
func NewGetWizardDataUseCase(
	userRepo repository.UserRepository,
	planRepo repository.PlanRepository,
	businessRepo repository.BusinessRepository,
	integrationRepo repository.IntegrationRepository,
	articleIdeaRepo repository.ArticleIdeaRepository,
) *GetWizardDataUseCase {
	return &GetWizardDataUseCase{
		userRepo:        userRepo,
		planRepo:        planRepo,
		businessRepo:    businessRepo,
		integrationRepo: integrationRepo,
		articleIdeaRepo: articleIdeaRepo,
	}
}

// Execute executa o caso de uso
func (uc *GetWizardDataUseCase) Execute(ctx context.Context, input GetWizardDataInput) (*GetWizardDataOutput, error) {
	log.Debug().Str("user_id", input.UserID).Msg("GetWizardDataUseCase Execute iniciado")

	// 1. Parse user_id
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		log.Error().Err(err).Msg("GetWizardDataUseCase: user_id inválido")
		return nil, errors.New("invalid_user_id")
	}

	// 2. Buscar usuário para obter onboarding_step
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("GetWizardDataUseCase: erro ao buscar usuário")
		return nil, errors.New("user_not_found")
	}

	if user == nil {
		log.Warn().Str("user_id", input.UserID).Msg("GetWizardDataUseCase: usuário não encontrado")
		return nil, errors.New("user_not_found")
	}

	// Buscar plano do usuário para limites
	plan, err := uc.planRepo.FindByID(ctx, user.PlanID)
	if err != nil {
		log.Error().Err(err).Msg("GetWizardDataUseCase: erro ao buscar plano")
		return nil, errors.New("erro ao buscar plano")
	}

	output := &GetWizardDataOutput{
		OnboardingStep: user.OnboardingStep,
	}

	// 3. Buscar dados do negócio
	businessProfile, err := uc.businessRepo.FindProfileByUserID(ctx, userID)
	if err == nil && businessProfile != nil {
		var secondaryObj *string
		if businessProfile.SecondaryObjective != nil {
			s := string(*businessProfile.SecondaryObjective)
			secondaryObj = &s
		}

		output.Business = &BusinessDataOutput{
			Description:        businessProfile.Description,
			PrimaryObjective:   string(businessProfile.PrimaryObjective),
			SecondaryObjective: secondaryObj,
			Location:           &businessProfile.Location,
			SiteURL:            businessProfile.SiteURL,
			HasBlog:            businessProfile.HasBlog,
			BlogURLs:           businessProfile.BlogURLs,
			BrandFileURL:       businessProfile.BrandFileURL,
		}
	}

	// 4. Buscar competitors
	competitors, err := uc.businessRepo.FindCompetitorsByUserID(ctx, userID)
	if err == nil && len(competitors) > 0 {
		output.Competitors = competitors
	}

	// 5. Verificar se tem integração WordPress
	wpIntegration, err := uc.integrationRepo.FindByUserIDAndType(ctx, userID, entity.IntegrationTypeWordPress)
	output.HasIntegration = err == nil && wpIntegration != nil && wpIntegration.Enabled

	// 6. Dados de ideias e regeneração
	// Contar gerações na última hora
	gensInLastHour, _ := uc.articleIdeaRepo.CountGenerationsInLastHour(ctx, userID)
	output.RegenerationsLimit = plan.MaxIdeaRegenerationsPerHour
	output.RegenerationsRemaining = (plan.MaxIdeaRegenerationsPerHour + 1) - gensInLastHour
	if output.RegenerationsRemaining < 0 {
		output.RegenerationsRemaining = 0
	}

	// Contar ideias
	totalIdeas, _ := uc.articleIdeaRepo.CountByUserID(ctx, userID)
	approvedIdeas, _ := uc.articleIdeaRepo.CountApprovedByUserID(ctx, userID)
	
	output.TotalIdeasCount = totalIdeas
	output.ApprovedIdeasCount = approvedIdeas
	output.HasGeneratedIdeas = totalIdeas > 0

	// 7. Buscar ideias de artigos pendentes (se o usuário estiver no step 4)
	if user.OnboardingStep >= 4 {
		ideas, err := uc.articleIdeaRepo.FindByUserID(ctx, userID)
		if err == nil && len(ideas) > 0 {
			output.PendingIdeas = make([]PendingIdeaOutput, 0, len(ideas))
			for _, idea := range ideas {
				var feedback *string
				if idea.Feedback != nil {
					feedback = idea.Feedback
				}
				output.PendingIdeas = append(output.PendingIdeas, PendingIdeaOutput{
					ID:       idea.ID.String(),
					Title:    idea.Title,
					Summary:  idea.Summary,
					Approved: idea.Approved,
					Feedback: feedback,
				})
			}
			log.Debug().Int("ideas_count", len(output.PendingIdeas)).Msg("GetWizardDataUseCase: ideias pendentes encontradas")
		}
	}

	log.Info().
		Str("user_id", input.UserID).
		Int("onboarding_step", output.OnboardingStep).
		Bool("has_business", output.Business != nil).
		Int("competitors_count", len(output.Competitors)).
		Bool("has_integration", output.HasIntegration).
		Int("pending_ideas_count", len(output.PendingIdeas)).
		Int("regenerations_remaining", output.RegenerationsRemaining).
		Msg("GetWizardDataUseCase bem-sucedido")

	return output, nil
}
