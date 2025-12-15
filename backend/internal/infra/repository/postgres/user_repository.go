// internal/infra/repository/postgres/user_repository.go
package postgres

import (
	"context"
	"errors"

	"organiq/internal/domain/entity"
	"organiq/internal/domain/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepositoryPostgres implementa UserRepository com PostgreSQL
type UserRepositoryPostgres struct {
	db *gorm.DB
}

// NewUserRepository cria nova instância
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &UserRepositoryPostgres{db: db}
}

// Create insere novo usuário no banco
func (r *UserRepositoryPostgres) Create(ctx context.Context, user *entity.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	return r.db.WithContext(ctx).Create(user).Error
}

// FindByID busca usuário por ID
func (r *UserRepositoryPostgres) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &user, err
}

// FindByEmail busca usuário por email
func (r *UserRepositoryPostgres) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &user, err
}

// Update atualiza usuário existente
func (r *UserRepositoryPostgres) Update(ctx context.Context, user *entity.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	return r.db.WithContext(ctx).Save(user).Error
}

// Delete remove usuário (soft delete poderia ser implementado)
func (r *UserRepositoryPostgres) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, "id = ?", id).Error
}

// ============================================
// REFRESH TOKEN REPOSITORY
// ============================================

// RefreshTokenRepositoryPostgres implementa RefreshTokenRepository com PostgreSQL
type RefreshTokenRepositoryPostgres struct {
	db *gorm.DB
}

// NewRefreshTokenRepository cria nova instância
func NewRefreshTokenRepository(db *gorm.DB) repository.RefreshTokenRepository {
	return &RefreshTokenRepositoryPostgres{db: db}
}

// Create insere novo refresh token
func (r *RefreshTokenRepositoryPostgres) Create(ctx context.Context, rt *entity.RefreshToken) error {
	return r.db.WithContext(ctx).Create(rt).Error
}

// FindByHash busca refresh token por hash
func (r *RefreshTokenRepositoryPostgres) FindByHash(ctx context.Context, hash string) (*entity.RefreshToken, error) {
	var rt entity.RefreshToken
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND expires_at > NOW()", hash).
		First(&rt).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &rt, err
}

// FindByUserID busca todos os refresh tokens de um usuário
func (r *RefreshTokenRepositoryPostgres) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.RefreshToken, error) {
	var tokens []*entity.RefreshToken
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND expires_at > NOW()", userID).
		Order("created_at DESC").
		Find(&tokens).Error

	return tokens, err
}

// Update atualiza refresh token (ex: last_used_at)
func (r *RefreshTokenRepositoryPostgres) Update(ctx context.Context, rt *entity.RefreshToken) error {
	return r.db.WithContext(ctx).Save(rt).Error
}

// Delete remove refresh token por ID
func (r *RefreshTokenRepositoryPostgres) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.RefreshToken{}, "id = ?", id).Error
}

// DeleteExpired remove tokens expirados (para limpeza periódica)
func (r *RefreshTokenRepositoryPostgres) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).Delete(&entity.RefreshToken{}, "expires_at < NOW()").Error
}
