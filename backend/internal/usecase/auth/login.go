package auth

import (
	"context"
	"errors"
	"time"

	"organiq/internal/domain/entity"
	"organiq/internal/domain/repository"
	"organiq/internal/util"

	"github.com/google/uuid"
)

// LoginUserInput dados de entrada para login.
type LoginUserInput struct {
	Email    string
	Password string
}

// LoginUserOutput dados de saída do login.
type LoginUserOutput struct {
	User         *entity.User
	AccessToken  string
	RefreshToken string
}

// LoginUserUseCase autentica o usuário e retorna novos tokens.
type LoginUserUseCase struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	crypto           *util.CryptoService
}

// NewLoginUserUseCase cria nova instância.
func NewLoginUserUseCase(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	crypto *util.CryptoService,
) *LoginUserUseCase {
	return &LoginUserUseCase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		crypto:           crypto,
	}
}

// Execute processa o login e emite tokens de acesso/refresh.
func (uc *LoginUserUseCase) Execute(ctx context.Context, input LoginUserInput) (*LoginUserOutput, error) {
	user, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid_credentials")
	}

	valid, err := uc.crypto.VerifyPassword(input.Password, user.PasswordHash)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, errors.New("invalid_credentials")
	}

	accessToken, err := uc.crypto.GenerateAccessToken(user.ID, user.Email, user.HasCompletedOnboarding, user.OnboardingStep)
	if err != nil {
		return nil, err
	}

	refreshTokenString, err := uc.crypto.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	refreshToken := &entity.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: uc.crypto.HashRefreshToken(refreshTokenString),
		ExpiresAt: time.Now().AddDate(0, 0, 7),
		CreatedAt: time.Now(),
	}
	if err := uc.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, err
	}

	return &LoginUserOutput{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
	}, nil
}
