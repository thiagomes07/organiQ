package auth

import (
	"context"

	"organiq/internal/domain/repository"
	"organiq/internal/util"
)

// LogoutUserInput dados necessários para logout.
type LogoutUserInput struct {
	UserID       string
	RefreshToken string
}

// LogoutUserUseCase invalida o refresh token informado.
type LogoutUserUseCase struct {
	refreshTokenRepo repository.RefreshTokenRepository
	crypto           *util.CryptoService
}

// NewLogoutUserUseCase cria nova instância.
func NewLogoutUserUseCase(
	refreshTokenRepo repository.RefreshTokenRepository,
	crypto *util.CryptoService,
) *LogoutUserUseCase {
	return &LogoutUserUseCase{
		refreshTokenRepo: refreshTokenRepo,
		crypto:           crypto,
	}
}

// Execute remove o refresh token do usuário.
func (uc *LogoutUserUseCase) Execute(ctx context.Context, input LogoutUserInput) error {
	hash := uc.crypto.HashRefreshToken(input.RefreshToken)
	token, err := uc.refreshTokenRepo.FindByHash(ctx, hash)
	if err != nil {
		return err
	}
	if token == nil {
		return nil
	}
	return uc.refreshTokenRepo.Delete(ctx, token.ID)
}
