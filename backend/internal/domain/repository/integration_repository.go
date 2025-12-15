// internal/domain/repository/integration_repository.go
package repository

import (
	"context"

	"github.com/google/uuid"
	"organiq/internal/domain/entity"
)

// IntegrationRepository define contrato para operações com integrações
type IntegrationRepository interface {
	// Create insere nova integração
	Create(ctx context.Context, integration *entity.Integration) error

	// FindByID busca integração por ID
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Integration, error)

	// FindByUserIDAndType busca integração específica de um usuário
	FindByUserIDAndType(ctx context.Context, userID uuid.UUID, intType entity.IntegrationType) (*entity.Integration, error)

	// FindByUserID retorna todas as integrações de um usuário
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Integration, error)

	// FindEnabledByUserIDAndType busca integração habilitada específica
	FindEnabledByUserIDAndType(ctx context.Context, userID uuid.UUID, intType entity.IntegrationType) (*entity.Integration, error)

	// Update atualiza integração existente
	Update(ctx context.Context, integration *entity.Integration) error

	// Delete deleta integração
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByUserIDAndType deleta integração específica de um usuário
	DeleteByUserIDAndType(ctx context.Context, userID uuid.UUID, intType entity.IntegrationType) error

	// Enable ativa uma integração
	Enable(ctx context.Context, id uuid.UUID) error

	// Disable desativa uma integração
	Disable(ctx context.Context, id uuid.UUID) error

	// ExistsByUserIDAndType verifica se integração existe
	ExistsByUserIDAndType(ctx context.Context, userID uuid.UUID, intType entity.IntegrationType) (bool, error)
}
