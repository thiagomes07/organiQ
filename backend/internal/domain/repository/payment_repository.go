// internal/domain/repository/payment_repository.go
package repository

import (
	"context"

	"github.com/google/uuid"
	"organiq/internal/domain/entity"
)

// PaymentRepository define contrato para operações com pagamentos
type PaymentRepository interface {
	// Create insere novo pagamento
	Create(ctx context.Context, payment *entity.Payment) error

	// FindByID busca pagamento por ID
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Payment, error)

	// FindByProviderSessionID busca pagamento por ID da sessão do provedor
	FindByProviderSessionID(ctx context.Context, sessionID string) (*entity.Payment, error)

	// FindByUserID retorna pagamentos de um usuário
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Payment, error)

	// FindLatestByUserID retorna último pagamento de um usuário
	FindLatestByUserID(ctx context.Context, userID uuid.UUID) (*entity.Payment, error)

	// FindPaidByUserID retorna pagamento pago mais recente de um usuário
	FindPaidByUserID(ctx context.Context, userID uuid.UUID) (*entity.Payment, error)

	// FindPendingByUserID retorna pagamentos pendentes de um usuário
	FindPendingByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Payment, error)

	// FindByStatus retorna pagamentos com status específico
	FindByStatus(ctx context.Context, status entity.PaymentStatus) ([]*entity.Payment, error)

	// FindExpiredPending retorna pagamentos pendentes expirados
	FindExpiredPending(ctx context.Context) ([]*entity.Payment, error)

	// Update atualiza pagamento existente
	Update(ctx context.Context, payment *entity.Payment) error

	// UpdateStatus atualiza apenas o status do pagamento
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.PaymentStatus) error

	// Delete deleta pagamento
	Delete(ctx context.Context, id uuid.UUID) error

	// ExistsPaidByUserAndPlan verifica se existe pagamento pago para usuário e plano
	ExistsPaidByUserAndPlan(ctx context.Context, userID, planID uuid.UUID) (bool, error)

	// CountPaidByPlan retorna quantidade de pagamentos bem-sucedidos por plano
	CountPaidByPlan(ctx context.Context, planID uuid.UUID) (int, error)

	// SumAmountByPlan retorna valor total de pagamentos bem-sucedidos por plano
	SumAmountByPlan(ctx context.Context, planID uuid.UUID) (float64, error)

	// CountByStatus retorna quantidade de pagamentos com status específico
	CountByStatus(ctx context.Context, status entity.PaymentStatus) (int, error)
}
