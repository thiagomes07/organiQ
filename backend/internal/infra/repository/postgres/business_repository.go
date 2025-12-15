// internal/infra/repository/postgres/business_repository.go
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

// BusinessRepositoryPostgres implementa BusinessRepository com PostgreSQL
type BusinessRepositoryPostgres struct {
	db *gorm.DB
}

// NewBusinessRepository cria nova instância
func NewBusinessRepository(db *gorm.DB) repository.BusinessRepository {
	return &BusinessRepositoryPostgres{db: db}
}

// ============================================
// BUSINESS PROFILE
// ============================================

// CreateProfile implementa repository.CreateProfile
func (r *BusinessRepositoryPostgres) CreateProfile(ctx context.Context, profile *entity.BusinessProfile) error {
	if err := profile.Validate(); err != nil {
		log.Error().Err(err).Msg("BusinessRepository CreateProfile: validação falhou")
		return err
	}

	log.Debug().Str("user_id", profile.UserID.String()).Msg("BusinessRepository CreateProfile")

	if err := r.db.WithContext(ctx).Create(profile).Error; err != nil {
		log.Error().Err(err).Str("user_id", profile.UserID.String()).Msg("BusinessRepository CreateProfile erro no banco")
		return err
	}

	return nil
}

// FindProfileByUserID implementa repository.FindProfileByUserID
func (r *BusinessRepositoryPostgres) FindProfileByUserID(ctx context.Context, userID uuid.UUID) (*entity.BusinessProfile, error) {
	var profile entity.BusinessProfile

	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Debug().Str("user_id", userID.String()).Msg("BusinessRepository FindProfileByUserID: não encontrado")
		return nil, nil
	}

	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("BusinessRepository FindProfileByUserID erro no banco")
		return nil, err
	}

	return &profile, nil
}

// UpdateProfile implementa repository.UpdateProfile
func (r *BusinessRepositoryPostgres) UpdateProfile(ctx context.Context, profile *entity.BusinessProfile) error {
	if err := profile.Validate(); err != nil {
		log.Error().Err(err).Msg("BusinessRepository UpdateProfile: validação falhou")
		return err
	}

	log.Debug().Str("user_id", profile.UserID.String()).Msg("BusinessRepository UpdateProfile")

	if err := r.db.WithContext(ctx).Save(profile).Error; err != nil {
		log.Error().Err(err).Msg("BusinessRepository UpdateProfile erro no banco")
		return err
	}

	return nil
}

// DeleteProfile implementa repository.DeleteProfile
func (r *BusinessRepositoryPostgres) DeleteProfile(ctx context.Context, userID uuid.UUID) error {
	log.Debug().Str("user_id", userID.String()).Msg("BusinessRepository DeleteProfile")

	if err := r.db.WithContext(ctx).Delete(&entity.BusinessProfile{}, "user_id = ?", userID).Error; err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("BusinessRepository DeleteProfile erro no banco")
		return err
	}

	return nil
}

// ============================================
// COMPETITORS
// ============================================

// CreateCompetitor implementa repository.CreateCompetitor
func (r *BusinessRepositoryPostgres) CreateCompetitor(ctx context.Context, userID uuid.UUID, url string) error {
	if userID == uuid.Nil {
		log.Error().Msg("BusinessRepository CreateCompetitor: user_id inválido")
		return errors.New("user_id não pode ser nil")
	}

	if len(url) == 0 {
		log.Error().Msg("BusinessRepository CreateCompetitor: url não pode estar vazio")
		return errors.New("url não pode estar vazio")
	}

	log.Debug().Str("user_id", userID.String()).Str("url", url).Msg("BusinessRepository CreateCompetitor")

	competitor := struct {
		ID        uuid.UUID
		UserID    uuid.UUID
		URL       string
	}{
		ID:     uuid.New(),
		UserID: userID,
		URL:    url,
	}

	if err := r.db.WithContext(ctx).Table("competitors").Create(competitor).Error; err != nil {
		log.Error().Err(err).Msg("BusinessRepository CreateCompetitor erro no banco")
		return err
	}

	return nil
}

// FindCompetitorsByUserID implementa repository.FindCompetitorsByUserID
func (r *BusinessRepositoryPostgres) FindCompetitorsByUserID(ctx context.Context, userID uuid.UUID) ([]string, error) {
	log.Debug().Str("user_id", userID.String()).Msg("BusinessRepository FindCompetitorsByUserID")

	var competitors []struct {
		URL string
	}

	err := r.db.WithContext(ctx).
		Table("competitors").
		Where("user_id = ?", userID).
		Select("url").
		Scan(&competitors).Error

	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("BusinessRepository FindCompetitorsByUserID erro no banco")
		return nil, err
	}

	urls := make([]string, len(competitors))
	for i, c := range competitors {
		urls[i] = c.URL
	}

	return urls, nil
}

// DeleteCompetitor implementa repository.DeleteCompetitor
func (r *BusinessRepositoryPostgres) DeleteCompetitor(ctx context.Context, userID uuid.UUID, url string) error {
	if userID == uuid.Nil || len(url) == 0 {
		log.Error().Msg("BusinessRepository DeleteCompetitor: parâmetros inválidos")
		return errors.New("user_id e url são obrigatórios")
	}

	log.Debug().Str("user_id", userID.String()).Str("url", url).Msg("BusinessRepository DeleteCompetitor")

	if err := r.db.WithContext(ctx).
		Table("competitors").
		Where("user_id = ? AND url = ?", userID, url).
		Delete(&struct{}{}).
		Error; err != nil {

		log.Error().Err(err).Msg("BusinessRepository DeleteCompetitor erro no banco")
		return err
	}

	return nil
}

// DeleteCompetitorsByUserID implementa repository.DeleteCompetitorsByUserID
func (r *BusinessRepositoryPostgres) DeleteCompetitorsByUserID(ctx context.Context, userID uuid.UUID) error {
	log.Debug().Str("user_id", userID.String()).Msg("BusinessRepository DeleteCompetitorsByUserID")

	if err := r.db.WithContext(ctx).
		Table("competitors").
		Where("user_id = ?", userID).
		Delete(&struct{}{}).
		Error; err != nil {

		log.Error().Err(err).Msg("BusinessRepository DeleteCompetitorsByUserID erro no banco")
		return err
	}

	return nil
}

// CompetitorCount implementa repository.CompetitorCount
func (r *BusinessRepositoryPostgres) CompetitorCount(ctx context.Context, userID uuid.UUID) (int, error) {
	log.Debug().Str("user_id", userID.String()).Msg("BusinessRepository CompetitorCount")

	var count int64
	if err := r.db.WithContext(ctx).
		Table("competitors").
		Where("user_id = ?", userID).
		Count(&count).
		Error; err != nil {

		log.Error().Err(err).Msg("BusinessRepository CompetitorCount erro no banco")
		return 0, err
	}

	return int(count), nil
}
