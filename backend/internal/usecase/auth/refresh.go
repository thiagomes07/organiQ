package auth

import (
	"context"
	"errors"

	"organiq/internal/domain/repository"
	"organiq/internal/util"
)

// RefreshAccessTokenInput dados para refresh de token.
type RefreshAccessTokenInput struct {
	RefreshToken string
}

// RefreshAccessTokenOutput novo access token.
type RefreshAccessTokenOutput struct {
	AccessToken string
}

// RefreshAccessTokenUseCase valida o refresh token e gera novo access token.
type RefreshAccessTokenUseCase struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	crypto           *util.CryptoService
}

// NewRefreshAccessTokenUseCase cria nova instância.
func NewRefreshAccessTokenUseCase(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	crypto *util.CryptoService,
) *RefreshAccessTokenUseCase {
	return &RefreshAccessTokenUseCase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		crypto:           crypto,
	}
}

// Execute valida o refresh token e retorna novo access token.
func (uc *RefreshAccessTokenUseCase) Execute(ctx context.Context, input RefreshAccessTokenInput) (*RefreshAccessTokenOutput, error) {
	if input.RefreshToken == "" {
		return nil, errors.New("refresh_token_required")
	}

	refreshHash := uc.crypto.HashRefreshToken(input.RefreshToken)
	storedToken, err := uc.refreshTokenRepo.FindByHash(ctx, refreshHash)
	if err != nil {
		return nil, err
	}
	if storedToken == nil {
		return nil, errors.New("invalid_refresh_token")
	}
	if storedToken.IsExpired() {
		return nil, errors.New("refresh_token_expired")
	}

	user, err := uc.userRepo.FindByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user_not_found")
	}

	storedToken.UpdateLastUsed()
	if err := uc.refreshTokenRepo.Update(ctx, storedToken); err != nil {
		return nil, err
	}

	accessToken, err := uc.crypto.GenerateAccessToken(user.ID, user.Email, user.HasCompletedOnboarding, user.OnboardingStep)
	if err != nil {
		return nil, err
	}

	return &RefreshAccessTokenOutput{AccessToken: accessToken}, nil
}
