// internal/domain/repository/plan_repository.go
package repository

import (
	"context"

	"github.com/google/uuid"
	"organiq/internal/domain/entity"
)

// PlanRepository define contrato para operações com planos
type PlanRepository interface {
	// Create insere novo plano
	Create(ctx context.Context, plan *entity.Plan) error

	// FindByID busca plano por ID
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Plan, error)

	// FindByName busca plano por nome único
	FindByName(ctx context.Context, name string) (*entity.Plan, error)

	// FindAll retorna todos os planos ativos
	FindAll(ctx context.Context) ([]*entity.Plan, error)

	// FindAllActive retorna apenas planos ativos
	FindAllActive(ctx context.Context) ([]*entity.Plan, error)

	// Update atualiza plano existente
	Update(ctx context.Context, plan *entity.Plan) error

	// Delete marca plano como inativo (soft delete)
	Delete(ctx context.Context, id uuid.UUID) error

	// Activate ativa um plano
	Activate(ctx context.Context, id uuid.UUID) error

	// Deactivate desativa um plano
	Deactivate(ctx context.Context, id uuid.UUID) error
}
