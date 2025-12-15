package account

import (
	"context"

	"organiq/internal/domain/entity"
	"organiq/internal/domain/repository"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// GetAccountUseCase agrega informações do usuário com plano e integrações.
type GetAccountUseCase struct {
	userRepo        repository.UserRepository
	planRepo        repository.PlanRepository
	integrationRepo repository.IntegrationRepository
}

// NewGetAccountUseCase instancia o caso de uso.
func NewGetAccountUseCase(
	userRepo repository.UserRepository,
	planRepo repository.PlanRepository,
	integrationRepo repository.IntegrationRepository,
) *GetAccountUseCase {
	return &GetAccountUseCase{
		userRepo:        userRepo,
		planRepo:        planRepo,
		integrationRepo: integrationRepo,
	}
}

// GetAccountInput define os dados necessários para obter o agregado de conta.
type GetAccountInput struct {
	UserID string
}

// GetAccountOutput retorna o agregado de conta completo.
type GetAccountOutput struct {
	User         *entity.User
	Plan         *entity.Plan
	Integrations []*entity.Integration
}

// Execute executa o caso de uso.
func (uc *GetAccountUseCase) Execute(ctx context.Context, input GetAccountInput) (*GetAccountOutput, error) {
	if input.UserID == "" {
		return nil, ErrInvalidUserID
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

	plan, err := uc.planRepo.FindByID(ctx, user.PlanID)
	if err != nil {
		logger.Error().Err(err).Str("plan_id", user.PlanID.String()).Msg("failed to fetch plan")
		return nil, err
	}
	if plan == nil {
		return nil, ErrPlanNotFound
	}

	integrations, err := uc.integrationRepo.FindByUserID(ctx, user.ID)
	if err != nil {
		logger.Error().Err(err).Str("user_id", user.ID.String()).Msg("failed to fetch integrations")
		return nil, err
	}
	if integrations == nil {
		integrations = []*entity.Integration{}
	}

	return &GetAccountOutput{
		User:         user,
		Plan:         plan,
		Integrations: integrations,
	}, nil
}

func loggerFromContext(ctx context.Context) *zerolog.Logger {
	if logger := log.Ctx(ctx); logger != nil {
		return logger
	}
	l := log.Logger
	return &l
}
