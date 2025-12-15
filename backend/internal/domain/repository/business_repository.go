// internal/domain/repository/business_repository.go
package repository

import (
	"context"

	"github.com/google/uuid"
	"organiq/internal/domain/entity"
)

// BusinessRepository define contrato para operações com perfil de negócio
type BusinessRepository interface {
	// CreateProfile cria novo perfil de negócio para um usuário
	CreateProfile(ctx context.Context, profile *entity.BusinessProfile) error

	// FindProfileByUserID busca perfil de negócio por ID do usuário
	FindProfileByUserID(ctx context.Context, userID uuid.UUID) (*entity.BusinessProfile, error)

	// UpdateProfile atualiza perfil de negócio existente
	UpdateProfile(ctx context.Context, profile *entity.BusinessProfile) error

	// DeleteProfile deleta perfil de negócio
	DeleteProfile(ctx context.Context, userID uuid.UUID) error

	// CreateCompetitor adiciona URL de concorrente
	CreateCompetitor(ctx context.Context, userID uuid.UUID, url string) error

	// FindCompetitorsByUserID retorna todas as URLs de concorrentes do usuário
	FindCompetitorsByUserID(ctx context.Context, userID uuid.UUID) ([]string, error)

	// DeleteCompetitor remove uma URL de concorrente específica
	DeleteCompetitor(ctx context.Context, userID uuid.UUID, url string) error

	// DeleteCompetitorsByUserID remove todos os concorrentes de um usuário
	DeleteCompetitorsByUserID(ctx context.Context, userID uuid.UUID) error

	// CompetitorCount retorna número de concorrentes de um usuário
	CompetitorCount(ctx context.Context, userID uuid.UUID) (int, error)
}

// Competitor representa um concorrente armazenado (para referência)
type Competitor struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	URL       string
	CreatedAt interface{} // time.Time
}
