package account

import (
	"context"

	"organiq/internal/domain/entity"
	"organiq/internal/domain/repository"

	"github.com/google/uuid"
)

// GetPlanUseCase retorna detalhes do plano e uso atual.
type GetPlanUseCase struct {
	userRepo repository.UserRepository
	planRepo repository.PlanRepository
}

// NewGetPlanUseCase instancia o caso de uso.
func NewGetPlanUseCase(userRepo repository.UserRepository, planRepo repository.PlanRepository) *GetPlanUseCase {
	return &GetPlanUseCase{userRepo: userRepo, planRepo: planRepo}
}

// GetPlanInput define o payload de entrada.
type GetPlanInput struct {
	UserID string
}

// GetPlanOutput consolida dados do plano.
type GetPlanOutput struct {
	Plan         *entity.Plan
	ArticlesUsed int
}

// Execute executa o caso de uso.
func (uc *GetPlanUseCase) Execute(ctx context.Context, input GetPlanInput) (*GetPlanOutput, error) {
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

	return &GetPlanOutput{
		Plan:         plan,
		ArticlesUsed: user.ArticlesUsed,
	}, nil
}
