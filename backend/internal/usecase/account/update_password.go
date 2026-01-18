package account

import (
	"context"
	"time"

	"organiq/internal/domain/repository"
	"organiq/internal/util"

	"github.com/google/uuid"
)

// UpdatePasswordUseCase atualiza a senha do usuário.
type UpdatePasswordUseCase struct {
	userRepo  repository.UserRepository
	cryptoSvc *util.CryptoService
}

// NewUpdatePasswordUseCase instancia o caso de uso.
func NewUpdatePasswordUseCase(userRepo repository.UserRepository, cryptoSvc *util.CryptoService) *UpdatePasswordUseCase {
	return &UpdatePasswordUseCase{userRepo: userRepo, cryptoSvc: cryptoSvc}
}

// UpdatePasswordInput representa os dados de entrada.
type UpdatePasswordInput struct {
	UserID          string
	CurrentPassword string
	NewPassword     string
}

// Execute executa o caso de uso.
func (uc *UpdatePasswordUseCase) Execute(ctx context.Context, input UpdatePasswordInput) error {
	if len(input.NewPassword) < 6 {
		return ErrInvalidPassword
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		return ErrInvalidUserID
	}

	logger := loggerFromContext(ctx)

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		logger.Error().Err(err).Str("user_id", userID.String()).Msg("failed to fetch user")
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Verificar senha atual
	valid, err := uc.cryptoSvc.VerifyPassword(input.CurrentPassword, user.PasswordHash)
    if err != nil {
        logger.Error().Err(err).Msg("failed to verify password")
        return err // ou um erro genérico para não vazar detalhes
    }
	if !valid {
		return ErrIncorrectPassword
	}

	// Hash nova senha
	newHash, err := uc.cryptoSvc.HashPassword(input.NewPassword)
	if err != nil {
		logger.Error().Err(err).Msg("failed to hash new password")
		return err
	}

	user.PasswordHash = newHash
	user.UpdatedAt = time.Now()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		logger.Error().Err(err).Str("user_id", user.ID.String()).Msg("failed to update user password")
		return err
	}

	return nil
}
