// internal/domain/repository/user_repository.go
package repository

import (
	"context"

	"organiq/internal/domain/entity"

	"github.com/google/uuid"
)

// UserRepository define contrato para operações com usuários
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ============================================
// REFRESH TOKEN REPOSITORY
// ============================================

// RefreshTokenRepository define contrato para refresh tokens
type RefreshTokenRepository interface {
	Create(ctx context.Context, rt *entity.RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*entity.RefreshToken, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.RefreshToken, error)
	Update(ctx context.Context, rt *entity.RefreshToken) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error // Limpar tokens expirados
}
