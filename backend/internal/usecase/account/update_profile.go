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

// UpdateProfileUseCase atualiza dados básicos do usuário.
type UpdateProfileUseCase struct {
	userRepo repository.UserRepository
}

// NewUpdateProfileUseCase instancia o caso de uso.
func NewUpdateProfileUseCase(userRepo repository.UserRepository) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{userRepo: userRepo}
}

// UpdateProfileInput representa os dados de entrada.
type UpdateProfileInput struct {
	UserID string
	Name   string
	Email  string
}

// UpdateProfileOutput representa os dados atualizados.
type UpdateProfileOutput struct {
	User *entity.User
}

// Execute executa o caso de uso.
func (uc *UpdateProfileUseCase) Execute(ctx context.Context, input UpdateProfileInput) (*UpdateProfileOutput, error) {
	name := strings.TrimSpace(input.Name)
	email := strings.TrimSpace(strings.ToLower(input.Email))

	if len(name) < 2 || len(name) > 100 {
		return nil, ErrInvalidName
	}

	if email == "" || !util.IsValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	logger := loggerFromContext(ctx)

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		logger.Error().Err(err).Str("user_id", userID.String()).Msg("failed to fetch user")
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if !strings.EqualFold(user.Email, email) {
		existing, err := uc.userRepo.FindByEmail(ctx, email)
		if err != nil {
			logger.Error().Err(err).Str("email", email).Msg("failed to check email availability")
			return nil, err
		}
		if existing != nil && existing.ID != user.ID {
			return nil, ErrEmailAlreadyExists
		}
	}

	user.Name = name
	user.Email = email
	user.UpdatedAt = time.Now()

	if err := user.Validate(); err != nil {
		return nil, err
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		logger.Error().Err(err).Str("user_id", user.ID.String()).Msg("failed to update user")
		return nil, err
	}

	return &UpdateProfileOutput{User: user}, nil
}
