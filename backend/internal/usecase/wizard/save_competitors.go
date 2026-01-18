// internal/usecase/wizard/save_competitors.go
package wizard

import (
	"context"
	"errors"
	"net/url"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"organiq/internal/domain/repository"
)

// SaveCompetitorsInput dados de entrada
type SaveCompetitorsInput struct {
	UserID          string   // UUID como string do context
	CompetitorURLs  []string
}

// SaveCompetitorsOutput dados de saída
type SaveCompetitorsOutput struct {
	Success  bool
	Count    int
}

// SaveCompetitorsUseCase implementa o caso de uso
type SaveCompetitorsUseCase struct {
	businessRepo repository.BusinessRepository
	userRepo     repository.UserRepository
}

// NewSaveCompetitorsUseCase cria nova instância
func NewSaveCompetitorsUseCase(
	businessRepo repository.BusinessRepository,
	userRepo repository.UserRepository,
) *SaveCompetitorsUseCase {
	return &SaveCompetitorsUseCase{
		businessRepo: businessRepo,
		userRepo:     userRepo,
	}
}

// Execute executa o caso de uso
func (uc *SaveCompetitorsUseCase) Execute(ctx context.Context, input SaveCompetitorsInput) (*SaveCompetitorsOutput, error) {
	log.Debug().
		Str("user_id", input.UserID).
		Int("count", len(input.CompetitorURLs)).
		Msg("SaveCompetitorsUseCase Execute iniciado")

	// 1. Parse user_id
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		log.Error().Err(err).Msg("SaveCompetitorsUseCase: user_id inválido")
		return nil, errors.New("invalid_user_id")
	}

	// 2. Validar quantidade
	if len(input.CompetitorURLs) > 20 {
		log.Warn().Int("count", len(input.CompetitorURLs)).Msg("SaveCompetitorsUseCase: muitos concorrentes")
		return nil, errors.New("máximo 20 concorrentes permitidos")
	}

	// 3. Validar e processar inputs
	validItems := make([]string, 0, len(input.CompetitorURLs))
	itemMap := make(map[string]bool)

	for _, rawItem := range input.CompetitorURLs {
		trimmed := rawItem

		// Validar tamanho (2 a 200 caracteres)
		if len(trimmed) < 2 || len(trimmed) > 200 {
			log.Warn().Str("item", trimmed).Msg("SaveCompetitorsUseCase: item com tamanho inválido")
			return nil, errors.New("cada concorrente deve ter entre 2 e 200 caracteres")
		}

		// Se parece URL, tentar normalizar
		if isValidURL(trimmed) {
			trimmed = normalizeURL(trimmed)
		}

		// Deduplica
		if !itemMap[trimmed] {
			validItems = append(validItems, trimmed)
			itemMap[trimmed] = true
		}
	}

	// 4. Deletar competidores anteriores
	log.Debug().Str("user_id", input.UserID).Msg("SaveCompetitorsUseCase: deletando competidores anteriores")
	if err := uc.businessRepo.DeleteCompetitorsByUserID(ctx, userID); err != nil {
		log.Error().Err(err).Msg("SaveCompetitorsUseCase: erro ao deletar competidores anteriores")
		return nil, errors.New("erro ao processar dados anteriores")
	}

	// 5. Inserir novos competidores em lote
	if len(validItems) > 0 {
		log.Debug().Str("user_id", input.UserID).Int("count", len(validItems)).Msg("SaveCompetitorsUseCase: inserindo competidores")
		for _, itemStr := range validItems {
			if err := uc.businessRepo.CreateCompetitor(ctx, userID, itemStr); err != nil {
				log.Error().Err(err).Str("item", itemStr).Msg("SaveCompetitorsUseCase: erro ao criar competidor")
				return nil, errors.New("erro ao salvar concorrente: " + itemStr)
			}
		}
	}

	log.Info().
		Str("user_id", input.UserID).
		Int("count", len(validItems)).
		Msg("SaveCompetitorsUseCase bem-sucedido")

	// 6. Atualizar onboarding_step do usuário para 3 (competitors completo)
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err == nil && user != nil && user.OnboardingStep < 3 {
		user.OnboardingStep = 3
		if err := uc.userRepo.Update(ctx, user); err != nil {
			log.Warn().Err(err).Msg("SaveCompetitorsUseCase: erro ao atualizar onboarding_step")
		}
	}

	return &SaveCompetitorsOutput{
		Success: true,
		Count:   len(validItems),
	}, nil
}

// ============================================
// HELPERS
// ============================================

func isValidURL(urlString string) bool {
	// Adicionar schema se não tiver
	if len(urlString) == 0 {
		return false
	}

	// Se não tiver schema, adicionar https://
	if len(urlString) > 0 && urlString[0:1] != "h" {
		urlString = "https://" + urlString
	}

	// Tentar fazer parse
	u, err := url.Parse(urlString)
	if err != nil {
		return false
	}

	// Validar scheme e host
	if u.Scheme == "" || u.Host == "" {
		return false
	}

	// Validar scheme (apenas http e https)
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	return true
}

func normalizeURL(urlString string) string {
	// Se não tiver schema, adicionar https://
	if len(urlString) > 0 && (urlString[0:1] != "h" && urlString[0:2] != "ht") {
		urlString = "https://" + urlString
	}

	// Parse para normalizar
	u, err := url.Parse(urlString)
	if err != nil {
		return urlString
	}

	// Reconstruir URL normalizada (remover trailing slash)
	normalized := u.Scheme + "://" + u.Host + u.Path
	if u.Path != "" && len(u.Path) > 1 {
		// Remover trailing slash
		for len(normalized) > 0 && normalized[len(normalized)-1] == '/' {
			normalized = normalized[:len(normalized)-1]
		}
	}

	return normalized
}
