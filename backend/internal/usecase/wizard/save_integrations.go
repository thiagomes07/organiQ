// internal/usecase/wizard/save_integrations.go
package wizard

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"organiq/internal/domain/entity"
	"organiq/internal/domain/repository"
	"organiq/internal/infra/wordpress"
	"organiq/internal/util"
)

// SaveIntegrationsInput dados de entrada
type SaveIntegrationsInput struct {
	UserID        string
	WordPress     *WordPressIntegrationInput
	SearchConsole *SearchConsoleIntegrationInput
	Analytics     *AnalyticsIntegrationInput
}

// WordPressIntegrationInput configuração do WordPress
type WordPressIntegrationInput struct {
	SiteURL     string
	Username    string
	AppPassword string
}

// SearchConsoleIntegrationInput configuração do Search Console
type SearchConsoleIntegrationInput struct {
	PropertyURL string
}

// AnalyticsIntegrationInput configuração do Google Analytics
type AnalyticsIntegrationInput struct {
	MeasurementID string
}

// SaveIntegrationsOutput dados de saída
type SaveIntegrationsOutput struct {
	Success                bool
	WordPressConnected     bool
	SearchConsoleConnected bool
	AnalyticsConnected     bool
	Errors                 map[string]string // tipo -> erro
}

// SaveIntegrationsUseCase implementa o caso de uso
type SaveIntegrationsUseCase struct {
	integrationRepo repository.IntegrationRepository
	cryptoService   *util.CryptoService
	wpClient        *wordpress.Client // Será nil, criamos no execute
}

// NewSaveIntegrationsUseCase cria nova instância
func NewSaveIntegrationsUseCase(
	integrationRepo repository.IntegrationRepository,
	cryptoService *util.CryptoService,
) *SaveIntegrationsUseCase {
	return &SaveIntegrationsUseCase{
		integrationRepo: integrationRepo,
		cryptoService:   cryptoService,
	}
}

// Execute executa o caso de uso
func (uc *SaveIntegrationsUseCase) Execute(ctx context.Context, input SaveIntegrationsInput) (*SaveIntegrationsOutput, error) {
	log.Debug().Str("user_id", input.UserID).Msg("SaveIntegrationsUseCase Execute iniciado")

	// 1. Parse user_id
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		log.Error().Err(err).Msg("SaveIntegrationsUseCase: user_id inválido")
		return nil, errors.New("invalid_user_id")
	}

	// 2. Validar que pelo menos uma integração foi fornecida
	if input.WordPress == nil && input.SearchConsole == nil && input.Analytics == nil {
		log.Warn().Msg("SaveIntegrationsUseCase: nenhuma integração fornecida")
		return nil, errors.New("pelo menos uma integração deve ser configurada")
	}

	output := &SaveIntegrationsOutput{
		Errors: make(map[string]string),
	}

	// 3. Processar WordPress
	if input.WordPress != nil {
		log.Debug().Str("user_id", input.UserID).Msg("SaveIntegrationsUseCase: processando WordPress")
		if err := uc.processWordPress(ctx, userID, input.WordPress, output); err != nil {
			log.Error().Err(err).Msg("SaveIntegrationsUseCase: erro ao processar WordPress")
		}
	}

	// 4. Processar Search Console
	if input.SearchConsole != nil {
		log.Debug().Str("user_id", input.UserID).Msg("SaveIntegrationsUseCase: processando Search Console")
		if err := uc.processSearchConsole(ctx, userID, input.SearchConsole, output); err != nil {
			log.Error().Err(err).Msg("SaveIntegrationsUseCase: erro ao processar Search Console")
		}
	}

	// 5. Processar Analytics
	if input.Analytics != nil {
		log.Debug().Str("user_id", input.UserID).Msg("SaveIntegrationsUseCase: processando Analytics")
		if err := uc.processAnalytics(ctx, userID, input.Analytics, output); err != nil {
			log.Error().Err(err).Msg("SaveIntegrationsUseCase: erro ao processar Analytics")
		}
	}

	// 6. Retornar resultado
	output.Success = len(output.Errors) == 0 || (output.WordPressConnected || output.SearchConsoleConnected || output.AnalyticsConnected)

	log.Info().
		Str("user_id", input.UserID).
		Bool("success", output.Success).
		Msg("SaveIntegrationsUseCase bem-sucedido")

	return output, nil
}

// ============================================
// PRIVATE METHODS
// ============================================

func (uc *SaveIntegrationsUseCase) processWordPress(
	ctx context.Context,
	userID uuid.UUID,
	input *WordPressIntegrationInput,
	output *SaveIntegrationsOutput,
) error {
	log.Debug().Msg("processWordPress iniciado")

	// 1. Validar entrada
	if len(input.SiteURL) == 0 {
		output.Errors["wordpress"] = "siteUrl é obrigatório"
		return errors.New("siteUrl é obrigatório")
	}

	if len(input.Username) == 0 {
		output.Errors["wordpress"] = "username é obrigatório"
		return errors.New("username é obrigatório")
	}

	if len(input.AppPassword) == 0 {
		output.Errors["wordpress"] = "appPassword é obrigatório"
		return errors.New("appPassword é obrigatório")
	}

	// 2. Testar conexão com WordPress
	log.Debug().Str("siteUrl", input.SiteURL).Msg("Testando conexão com WordPress")
	wpClient := wordpress.NewClient(input.SiteURL, input.Username, input.AppPassword)

	if err := wpClient.TestConnection(ctx); err != nil {
		log.Error().Err(err).Msg("Erro ao conectar com WordPress")
		output.Errors["wordpress"] = "Falha ao conectar com WordPress. Verifique as credenciais."
		return err
	}

	// 3. Criptografar app password
	encryptedPassword, err := uc.cryptoService.EncryptAES(input.AppPassword)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao criptografar password")
		output.Errors["wordpress"] = "Erro ao processar credenciais"
		return err
	}

	// 4. Criar/atualizar integração
	config := &entity.WordPressConfig{
		SiteURL:     input.SiteURL,
		Username:    input.Username,
		AppPassword: encryptedPassword, // Armazenar criptografado
	}

	integration := &entity.Integration{
		ID:      uuid.New(),
		UserID:  userID,
		Type:    entity.IntegrationTypeWordPress,
		Enabled: true,
	}

	if err := integration.SetWordPressConfig(config); err != nil {
		log.Error().Err(err).Msg("Erro ao configurar WordPress")
		output.Errors["wordpress"] = "Erro ao configurar integração"
		return err
	}

	// Deletar integração anterior se existir
	_ = uc.integrationRepo.DeleteByUserIDAndType(ctx, userID, entity.IntegrationTypeWordPress)

	// Salvar nova integração
	if err := uc.integrationRepo.Create(ctx, integration); err != nil {
		log.Error().Err(err).Msg("Erro ao salvar integração WordPress")
		output.Errors["wordpress"] = "Erro ao salvar integração"
		return err
	}

	log.Info().Str("user_id", userID.String()).Msg("WordPress integrado com sucesso")
	output.WordPressConnected = true
	return nil
}

func (uc *SaveIntegrationsUseCase) processSearchConsole(
	ctx context.Context,
	userID uuid.UUID,
	input *SearchConsoleIntegrationInput,
	output *SaveIntegrationsOutput,
) error {
	log.Debug().Msg("processSearchConsole iniciado")

	// 1. Validar entrada
	if len(input.PropertyURL) == 0 {
		output.Errors["search_console"] = "propertyUrl é obrigatório"
		return errors.New("propertyUrl é obrigatório")
	}

	if !isValidURL(input.PropertyURL) {
		output.Errors["search_console"] = "propertyUrl inválida"
		return errors.New("propertyUrl inválida")
	}

	// 2. Normalizar URL
	normalizedURL := normalizeURL(input.PropertyURL)

	// 3. Criar integração
	integration := &entity.Integration{
		ID:      uuid.New(),
		UserID:  userID,
		Type:    entity.IntegrationTypeSearchConsole,
		Enabled: true,
	}

	config := &entity.SearchConsoleConfig{
		PropertyURL: normalizedURL,
	}

	if err := integration.SetSearchConsoleConfig(config); err != nil {
		log.Error().Err(err).Msg("Erro ao configurar Search Console")
		output.Errors["search_console"] = "Erro ao configurar integração"
		return err
	}

	// Deletar integração anterior se existir
	_ = uc.integrationRepo.DeleteByUserIDAndType(ctx, userID, entity.IntegrationTypeSearchConsole)

	// Salvar
	if err := uc.integrationRepo.Create(ctx, integration); err != nil {
		log.Error().Err(err).Msg("Erro ao salvar integração Search Console")
		output.Errors["search_console"] = "Erro ao salvar integração"
		return err
	}

	// TODO: Em produção, fazer OAuth flow com Google
	// Por enquanto, apenas armazenar URL como placeholder

	log.Info().Str("user_id", userID.String()).Msg("Search Console integrado com sucesso")
	output.SearchConsoleConnected = true
	return nil
}

func (uc *SaveIntegrationsUseCase) processAnalytics(
	ctx context.Context,
	userID uuid.UUID,
	input *AnalyticsIntegrationInput,
	output *SaveIntegrationsOutput,
) error {
	log.Debug().Msg("processAnalytics iniciado")

	// 1. Validar entrada
	if len(input.MeasurementID) == 0 {
		output.Errors["analytics"] = "measurementId é obrigatório"
		return errors.New("measurementId é obrigatório")
	}

	// Validar formato de Measurement ID (G-XXXXXXXXXX)
	if !isValidMeasurementID(input.MeasurementID) {
		output.Errors["analytics"] = "measurementId em formato inválido (deve ser G-XXXXXXXXXX)"
		return errors.New("measurementId inválido")
	}

	// 2. Criar integração
	integration := &entity.Integration{
		ID:      uuid.New(),
		UserID:  userID,
		Type:    entity.IntegrationTypeAnalytics,
		Enabled: true,
	}

	config := &entity.AnalyticsConfig{
		MeasurementID: input.MeasurementID,
	}

	if err := integration.SetAnalyticsConfig(config); err != nil {
		log.Error().Err(err).Msg("Erro ao configurar Analytics")
		output.Errors["analytics"] = "Erro ao configurar integração"
		return err
	}

	// Deletar integração anterior se existir
	_ = uc.integrationRepo.DeleteByUserIDAndType(ctx, userID, entity.IntegrationTypeAnalytics)

	// Salvar
	if err := uc.integrationRepo.Create(ctx, integration); err != nil {
		log.Error().Err(err).Msg("Erro ao salvar integração Analytics")
		output.Errors["analytics"] = "Erro ao salvar integração"
		return err
	}

	log.Info().Str("user_id", userID.String()).Msg("Analytics integrado com sucesso")
	output.AnalyticsConnected = true
	return nil
}

// ============================================
// HELPERS
// ============================================

func isValidMeasurementID(id string) bool {
	// Formato esperado: G-XXXXXXXXXX
	if len(id) < 3 || id[0] != 'G' || id[1] != '-' {
		return false
	}

	// Validar caracteres restantes (alfanuméricos)
	for i := 2; i < len(id); i++ {
		c := id[i]
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')) {
			return false
		}
	}

	return true
}
