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

// GetWizardDataOutput dados de saída
type GetWizardDataOutput struct {
	OnboardingStep int                 `json:"onboardingStep"`
	Business       *BusinessDataOutput `json:"business,omitempty"`
	Competitors    []string            `json:"competitors,omitempty"`
	HasIntegration bool                `json:"hasIntegration"`
}

// GetWizardDataUseCase implementa o caso de uso
type GetWizardDataUseCase struct {
	userRepo        repository.UserRepository
	businessRepo    repository.BusinessRepository
	integrationRepo repository.IntegrationRepository
}

// NewGetWizardDataUseCase cria nova instância
func NewGetWizardDataUseCase(
	userRepo repository.UserRepository,
	businessRepo repository.BusinessRepository,
	integrationRepo repository.IntegrationRepository,
) *GetWizardDataUseCase {
	return &GetWizardDataUseCase{
		userRepo:        userRepo,
		businessRepo:    businessRepo,
		integrationRepo: integrationRepo,
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

	log.Info().
		Str("user_id", input.UserID).
		Int("onboarding_step", output.OnboardingStep).
		Bool("has_business", output.Business != nil).
		Int("competitors_count", len(output.Competitors)).
		Bool("has_integration", output.HasIntegration).
		Msg("GetWizardDataUseCase bem-sucedido")

	return output, nil
}
