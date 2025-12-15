package account

import (
	"context"
	"strings"
	"time"

	"organiq/internal/domain/entity"
	"organiq/internal/domain/repository"
	"organiq/internal/util"

	"github.com/google/uuid"
)

// UpdateIntegrationsUseCase gerencia configurações de integrações.
type UpdateIntegrationsUseCase struct {
	integrationRepo repository.IntegrationRepository
	crypto          *util.CryptoService
}

// NewUpdateIntegrationsUseCase instancia o caso de uso.
func NewUpdateIntegrationsUseCase(
	integrationRepo repository.IntegrationRepository,
	crypto *util.CryptoService,
) *UpdateIntegrationsUseCase {
	return &UpdateIntegrationsUseCase{
		integrationRepo: integrationRepo,
		crypto:          crypto,
	}
}

// UpdateIntegrationsInput representa o payload de atualização.
type UpdateIntegrationsInput struct {
	UserID    string
	WordPress *WordPressIntegrationInput
	Analytics *AnalyticsIntegrationInput
}

// WordPressIntegrationInput contém dados do WordPress.
type WordPressIntegrationInput struct {
	SiteURL     string
	Username    string
	AppPassword string
	Enabled     bool
}

// AnalyticsIntegrationInput contém dados do GA4.
type AnalyticsIntegrationInput struct {
	MeasurementID string
	Enabled       bool
}

// UpdateIntegrationsOutput retorna integrações atualizadas.
type UpdateIntegrationsOutput struct {
	Integrations []*entity.Integration
}

// Execute executa o caso de uso.
func (uc *UpdateIntegrationsUseCase) Execute(ctx context.Context, input UpdateIntegrationsInput) (*UpdateIntegrationsOutput, error) {
	if input.UserID == "" {
		return nil, ErrInvalidUserID
	}

	if input.WordPress == nil && input.Analytics == nil {
		return nil, ErrNoIntegrationPayload
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	var updated []*entity.Integration

	if input.WordPress != nil {
		integration, err := uc.handleWordPress(ctx, userID, input.WordPress)
		if err != nil {
			return nil, err
		}
		updated = append(updated, integration)
	}

	if input.Analytics != nil {
		integration, err := uc.handleAnalytics(ctx, userID, input.Analytics)
		if err != nil {
			return nil, err
		}
		updated = append(updated, integration)
	}

	return &UpdateIntegrationsOutput{Integrations: updated}, nil
}

func (uc *UpdateIntegrationsUseCase) handleWordPress(ctx context.Context, userID uuid.UUID, input *WordPressIntegrationInput) (*entity.Integration, error) {
	logger := loggerFromContext(ctx)

	integration, err := uc.integrationRepo.FindByUserIDAndType(ctx, userID, entity.IntegrationTypeWordPress)
	if err != nil {
		logger.Error().Err(err).Str("user_id", userID.String()).Msg("failed to fetch wordpress integration")
		return nil, err
	}

	isNew := false
	if integration == nil {
		integration = &entity.Integration{
			ID:        uuid.New(),
			UserID:    userID,
			Type:      entity.IntegrationTypeWordPress,
			Config:    entity.IntegrationConfig{},
			Enabled:   input.Enabled,
			CreatedAt: time.Now(),
		}
		isNew = true
	}

	config := cloneConfig(integration.Config)

	if site := strings.TrimSpace(input.SiteURL); site != "" {
		config["siteUrl"] = site
	}

	if username := strings.TrimSpace(input.Username); username != "" {
		config["username"] = username
	}

	if pass := strings.TrimSpace(input.AppPassword); pass != "" {
		encrypted, err := uc.crypto.EncryptAES(pass)
		if err != nil {
			logger.Error().Err(err).Msg("failed to encrypt wordpress password")
			return nil, err
		}
		config["appPassword"] = encrypted
	}

	if strings.TrimSpace(getString(config, "siteUrl")) == "" ||
		strings.TrimSpace(getString(config, "username")) == "" ||
		strings.TrimSpace(getString(config, "appPassword")) == "" {
		return nil, ErrWordPressConfigIncomplete
	}

	integration.Config = config
	integration.Enabled = input.Enabled
	integration.Type = entity.IntegrationTypeWordPress
	integration.UserID = userID
	integration.UpdatedAt = time.Now()

	if err := integration.Validate(); err != nil {
		return nil, err
	}

	if isNew {
		if err := uc.integrationRepo.Create(ctx, integration); err != nil {
			logger.Error().Err(err).Str("integration_id", integration.ID.String()).Msg("failed to create wordpress integration")
			return nil, err
		}
	} else {
		if err := uc.integrationRepo.Update(ctx, integration); err != nil {
			logger.Error().Err(err).Str("integration_id", integration.ID.String()).Msg("failed to update wordpress integration")
			return nil, err
		}
	}

	return integration, nil
}

func (uc *UpdateIntegrationsUseCase) handleAnalytics(ctx context.Context, userID uuid.UUID, input *AnalyticsIntegrationInput) (*entity.Integration, error) {
	logger := loggerFromContext(ctx)

	integration, err := uc.integrationRepo.FindByUserIDAndType(ctx, userID, entity.IntegrationTypeAnalytics)
	if err != nil {
		logger.Error().Err(err).Str("user_id", userID.String()).Msg("failed to fetch analytics integration")
		return nil, err
	}

	isNew := false
	if integration == nil {
		integration = &entity.Integration{
			ID:        uuid.New(),
			UserID:    userID,
			Type:      entity.IntegrationTypeAnalytics,
			Config:    entity.IntegrationConfig{},
			Enabled:   input.Enabled,
			CreatedAt: time.Now(),
		}
		isNew = true
	}

	config := cloneConfig(integration.Config)

	if measurement := strings.TrimSpace(input.MeasurementID); measurement != "" {
		config["measurementId"] = measurement
	}

	if strings.TrimSpace(getString(config, "measurementId")) == "" {
		return nil, ErrAnalyticsConfigIncomplete
	}

	integration.Config = config
	integration.Enabled = input.Enabled
	integration.Type = entity.IntegrationTypeAnalytics
	integration.UserID = userID
	integration.UpdatedAt = time.Now()

	if err := integration.Validate(); err != nil {
		return nil, err
	}

	if isNew {
		if err := uc.integrationRepo.Create(ctx, integration); err != nil {
			logger.Error().Err(err).Str("integration_id", integration.ID.String()).Msg("failed to create analytics integration")
			return nil, err
		}
	} else {
		if err := uc.integrationRepo.Update(ctx, integration); err != nil {
			logger.Error().Err(err).Str("integration_id", integration.ID.String()).Msg("failed to update analytics integration")
			return nil, err
		}
	}

	return integration, nil
}

func cloneConfig(original entity.IntegrationConfig) entity.IntegrationConfig {
	clone := entity.IntegrationConfig{}
	if original == nil {
		return clone
	}
	for k, v := range original {
		clone[k] = v
	}
	return clone
}

func getString(config entity.IntegrationConfig, key string) string {
	if config == nil {
		return ""
	}
	if value, ok := config[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}
