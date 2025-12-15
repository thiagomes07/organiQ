// internal/infra/repository/postgres/integration_repository.go
package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"organiq/internal/domain/entity"
	"organiq/internal/domain/repository"
)

// IntegrationRepositoryPostgres implementa IntegrationRepository
type IntegrationRepositoryPostgres struct {
	db *gorm.DB
}

// NewIntegrationRepository cria nova instância
func NewIntegrationRepository(db *gorm.DB) repository.IntegrationRepository {
	return &IntegrationRepositoryPostgres{db: db}
}

// Create implementa repository.Create
func (r *IntegrationRepositoryPostgres) Create(ctx context.Context, integration *entity.Integration) error {
	if err := integration.Validate(); err != nil {
		log.Error().Err(err).Msg("IntegrationRepository Create: validação falhou")
		return err
	}

	log.Debug().
		Str("user_id", integration.UserID.String()).
		Str("type", string(integration.Type)).
		Msg("IntegrationRepository Create")

	if err := r.db.WithContext(ctx).Create(integration).Error; err != nil {
		log.Error().Err(err).Msg("IntegrationRepository Create erro no banco")
		return err
	}

	return nil
}

// FindByID implementa repository.FindByID
func (r *IntegrationRepositoryPostgres) FindByID(ctx context.Context, id uuid.UUID) (*entity.Integration, error) {
	var integration entity.Integration

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&integration).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		log.Error().Err(err).Msg("IntegrationRepository FindByID erro no banco")
		return nil, err
	}

	return &integration, nil
}

// FindByUserIDAndType implementa repository.FindByUserIDAndType
func (r *IntegrationRepositoryPostgres) FindByUserIDAndType(ctx context.Context, userID uuid.UUID, intType entity.IntegrationType) (*entity.Integration, error) {
	var integration entity.Integration

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, intType).
		First(&integration).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Debug().
			Str("user_id", userID.String()).
			Str("type", string(intType)).
			Msg("IntegrationRepository FindByUserIDAndType: não encontrado")
		return nil, nil
	}

	if err != nil {
		log.Error().Err(err).Msg("IntegrationRepository FindByUserIDAndType erro no banco")
		return nil, err
	}

	return &integration, nil
}

// FindByUserID implementa repository.FindByUserID
func (r *IntegrationRepositoryPostgres) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Integration, error) {
	log.Debug().Str("user_id", userID.String()).Msg("IntegrationRepository FindByUserID")

	var integrations []*entity.Integration

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&integrations).
		Error

	if err != nil {
		log.Error().Err(err).Msg("IntegrationRepository FindByUserID erro no banco")
		return nil, err
	}

	return integrations, nil
}

// FindEnabledByUserIDAndType implementa repository.FindEnabledByUserIDAndType
func (r *IntegrationRepositoryPostgres) FindEnabledByUserIDAndType(ctx context.Context, userID uuid.UUID, intType entity.IntegrationType) (*entity.Integration, error) {
	var integration entity.Integration

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ? AND enabled = ?", userID, intType, true).
		First(&integration).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Debug().
			Str("user_id", userID.String()).
			Str("type", string(intType)).
			Msg("IntegrationRepository FindEnabledByUserIDAndType: não encontrado")
		return nil, nil
	}

	if err != nil {
		log.Error().Err(err).Msg("IntegrationRepository FindEnabledByUserIDAndType erro no banco")
		return nil, err
	}

	return &integration, nil
}

// Update implementa repository.Update
func (r *IntegrationRepositoryPostgres) Update(ctx context.Context, integration *entity.Integration) error {
	if err := integration.Validate(); err != nil {
		log.Error().Err(err).Msg("IntegrationRepository Update: validação falhou")
		return err
	}

	log.Debug().Str("id", integration.ID.String()).Msg("IntegrationRepository Update")

	if err := r.db.WithContext(ctx).Save(integration).Error; err != nil {
		log.Error().Err(err).Msg("IntegrationRepository Update erro no banco")
		return err
	}

	return nil
}

// Delete implementa repository.Delete
func (r *IntegrationRepositoryPostgres) Delete(ctx context.Context, id uuid.UUID) error {
	log.Debug().Str("id", id.String()).Msg("IntegrationRepository Delete")

	if err := r.db.WithContext(ctx).Delete(&entity.Integration{}, "id = ?", id).Error; err != nil {
		log.Error().Err(err).Msg("IntegrationRepository Delete erro no banco")
		return err
	}

	return nil
}

// DeleteByUserIDAndType implementa repository.DeleteByUserIDAndType
func (r *IntegrationRepositoryPostgres) DeleteByUserIDAndType(ctx context.Context, userID uuid.UUID, intType entity.IntegrationType) error {
	log.Debug().
		Str("user_id", userID.String()).
		Str("type", string(intType)).
		Msg("IntegrationRepository DeleteByUserIDAndType")

	if err := r.db.WithContext(ctx).
		Delete(&entity.Integration{}, "user_id = ? AND type = ?", userID, intType).
		Error; err != nil {

		log.Error().Err(err).Msg("IntegrationRepository DeleteByUserIDAndType erro no banco")
		return err
	}

	return nil
}

// Enable implementa repository.Enable
func (r *IntegrationRepositoryPostgres) Enable(ctx context.Context, id uuid.UUID) error {
	log.Debug().Str("id", id.String()).Msg("IntegrationRepository Enable")

	if err := r.db.WithContext(ctx).
		Model(&entity.Integration{}).
		Where("id = ?", id).
		Update("enabled", true).
		Error; err != nil {

		log.Error().Err(err).Msg("IntegrationRepository Enable erro no banco")
		return err
	}

	return nil
}

// Disable implementa repository.Disable
func (r *IntegrationRepositoryPostgres) Disable(ctx context.Context, id uuid.UUID) error {
	log.Debug().Str("id", id.String()).Msg("IntegrationRepository Disable")

	if err := r.db.WithContext(ctx).
		Model(&entity.Integration{}).
		Where("id = ?", id).
		Update("enabled", false).
		Error; err != nil {

		log.Error().Err(err).Msg("IntegrationRepository Disable erro no banco")
		return err
	}

	return nil
}

// ExistsByUserIDAndType implementa repository.ExistsByUserIDAndType
func (r *IntegrationRepositoryPostgres) ExistsByUserIDAndType(ctx context.Context, userID uuid.UUID, intType entity.IntegrationType) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Table("integrations").
		Where("user_id = ? AND type = ?", userID, intType).
		Count(&count).
		Error

	if err != nil {
		log.Error().Err(err).Msg("IntegrationRepository ExistsByUserIDAndType erro no banco")
		return false, err
	}

	return count > 0, nil
}
