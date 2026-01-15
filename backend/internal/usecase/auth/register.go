// internal/usecase/auth/register.go
package auth

import (
	"context"
	"errors"
	"time"

	"organiq/internal/domain/entity"
	"organiq/internal/domain/repository"
	"organiq/internal/util"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// RegisterUserInput dados de entrada para registro
type RegisterUserInput struct {
	Name     string
	Email    string
	Password string
}

// RegisterUserOutput dados de saída do registro
type RegisterUserOutput struct {
	User         *entity.User
	AccessToken  string
	RefreshToken string
}

// RegisterUserUseCase implementa o caso de uso de registro
type RegisterUserUseCase struct {
	userRepo         repository.UserRepository
	planRepo         repository.PlanRepository
	refreshTokenRepo repository.RefreshTokenRepository
	crypto           *util.CryptoService
}

// NewRegisterUserUseCase cria nova instância
func NewRegisterUserUseCase(
	userRepo repository.UserRepository,
	planRepo repository.PlanRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	crypto *util.CryptoService,
) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		userRepo:         userRepo,
		planRepo:         planRepo,
		refreshTokenRepo: refreshTokenRepo,
		crypto:           crypto,
	}
}

// Execute executa o caso de uso
func (uc *RegisterUserUseCase) Execute(ctx context.Context, input RegisterUserInput) (*RegisterUserOutput, error) {
	// 1. Validar entrada
	if len(input.Name) < 2 || len(input.Name) > 100 {
		return nil, errors.New("nome deve ter entre 2 e 100 caracteres")
	}

	// Validar formato de email usando util.IsValidEmail (spec 3.7)
	if !util.IsValidEmail(input.Email) {
		return nil, errors.New("email inválido")
	}

	if len(input.Password) < 8 {
		return nil, errors.New("senha deve ter no mínimo 8 caracteres")
	}

	if !isValidPassword(input.Password) {
		return nil, errors.New("senha deve conter pelo menos 1 maiúscula, 1 minúscula e 1 número")
	}

	// 2. Verificar se email já existe
	existing, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, errors.New("email_already_exists")
	}

	// 3. Buscar plano Free (padrão)
	freePlan, err := uc.planRepo.FindByName(ctx, "Free")
	if err != nil {
		return nil, err
	}

	if freePlan == nil {
		return nil, errors.New("plano Free não encontrado")
	}

	// 4. Hash da senha com Argon2id
	passwordHash, err := uc.crypto.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	// 5. Criar entidade User
	user := &entity.User{
		ID:                     uuid.New(),
		Name:                   input.Name,
		Email:                  input.Email,
		PasswordHash:           passwordHash,
		PlanID:                 freePlan.ID,
		ArticlesUsed:           0,
		HasCompletedOnboarding: false,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	// 6. Validar regras de negócio
	if err := user.Validate(); err != nil {
		return nil, err
	}

	// 7. Persistir no banco
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 8. Gerar access token
	accessToken, err := uc.crypto.GenerateAccessToken(user.ID, user.Email, false) // Novo usuário não completou onboarding
	if err != nil {
		return nil, err
	}

	// 9. Gerar refresh token
	refreshTokenString, err := uc.crypto.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	tokenHash := uc.crypto.HashRefreshToken(refreshTokenString)
	refreshToken := &entity.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().AddDate(0, 0, 7), // 7 dias
		CreatedAt: time.Now(),
	}

	if err := uc.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, err
	}

	log.Info().
		Str("user_id", user.ID.String()).
		Str("email", user.Email).
		Msg("RegisterUserUseCase: usuário registrado com sucesso")

	return &RegisterUserOutput{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
	}, nil
}

// ============================================
// GET ME
// ============================================

// GetMeInput dados para buscar usuário autenticado
type GetMeInput struct {
	UserID string // UUID como string extraído do context
}

// GetMeOutput dados do usuário autenticado
type GetMeOutput struct {
	User *entity.User
}

// GetMeUseCase implementa busca do usuário autenticado
type GetMeUseCase struct {
	userRepo repository.UserRepository
}

// NewGetMeUseCase cria nova instância
func NewGetMeUseCase(userRepo repository.UserRepository) *GetMeUseCase {
	return &GetMeUseCase{
		userRepo: userRepo,
	}
}

// Execute executa o caso de uso
func (uc *GetMeUseCase) Execute(ctx context.Context, input GetMeInput) (*GetMeOutput, error) {
	// Parse UUID
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, errors.New("invalid_user_id")
	}

	// Buscar usuário
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user_not_found")
	}

	return &GetMeOutput{User: user}, nil
}

// ============================================
// HELPERS
// ============================================

func isValidPassword(password string) bool {
	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, r := range password {
		if r >= 'A' && r <= 'Z' {
			hasUpper = true
		}
		if r >= 'a' && r <= 'z' {
			hasLower = true
		}
		if r >= '0' && r <= '9' {
			hasDigit = true
		}
	}

	return hasUpper && hasLower && hasDigit
}
