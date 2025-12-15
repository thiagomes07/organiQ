package postgres

import (
	"context"
	"errors"

	"organiq/internal/domain/entity"
	"organiq/internal/domain/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PlanRepositoryPostgres implementa PlanRepository usando GORM/PostgreSQL.
type PlanRepositoryPostgres struct {
	db *gorm.DB
}

// NewPlanRepository cria nova instância do repositório de planos.
func NewPlanRepository(db *gorm.DB) repository.PlanRepository {
	return &PlanRepositoryPostgres{db: db}
}

// FindByID busca um plano ativo pelo ID.
func (r *PlanRepositoryPostgres) FindByID(ctx context.Context, id uuid.UUID) (*entity.Plan, error) {
	var plan entity.Plan
	err := r.db.WithContext(ctx).
		Where("id = ? AND active = ?", id, true).
		First(&plan).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &plan, err
}

// FindByName busca plano ativo pelo nome.
func (r *PlanRepositoryPostgres) FindByName(ctx context.Context, name string) (*entity.Plan, error) {
	var plan entity.Plan
	err := r.db.WithContext(ctx).
		Where("name = ? AND active = ?", name, true).
		First(&plan).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &plan, err
}

// FindAll retorna todos os planos (ativos ou não) ordenados por preço.
func (r *PlanRepositoryPostgres) FindAll(ctx context.Context) ([]*entity.Plan, error) {
	var plans []*entity.Plan
	err := r.db.WithContext(ctx).
		Order("price ASC, name ASC").
		Find(&plans).Error

	return plans, err
}

// FindAllActive retorna apenas planos ativos ordenados por preço.
func (r *PlanRepositoryPostgres) FindAllActive(ctx context.Context) ([]*entity.Plan, error) {
	var plans []*entity.Plan
	err := r.db.WithContext(ctx).
		Where("active = ?", true).
		Order("price ASC, name ASC").
		Find(&plans).Error

	return plans, err
}

// Create insere um novo plano.
func (r *PlanRepositoryPostgres) Create(ctx context.Context, plan *entity.Plan) error {
	if err := plan.Validate(); err != nil {
		return err
	}

	return r.db.WithContext(ctx).Create(plan).Error
}

// Update persiste alterações em um plano existente.
func (r *PlanRepositoryPostgres) Update(ctx context.Context, plan *entity.Plan) error {
	if err := plan.Validate(); err != nil {
		return err
	}

	return r.db.WithContext(ctx).Save(plan).Error
}

// Delete realiza soft delete, marcando o plano como inativo.
func (r *PlanRepositoryPostgres) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entity.Plan{}).
		Where("id = ?", id).
		Update("active", false).Error
}

// Activate reativa um plano previamente desativado.
func (r *PlanRepositoryPostgres) Activate(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entity.Plan{}).
		Where("id = ?", id).
		Update("active", true).Error
}

// Deactivate desativa um plano sem removê-lo.
func (r *PlanRepositoryPostgres) Deactivate(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entity.Plan{}).
		Where("id = ?", id).
		Update("active", false).Error
}
