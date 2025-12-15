// internal/infra/repository/postgres/payment_repository.go
package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"organiq/internal/domain/entity"
)

// PaymentRepositoryPostgres implementa PaymentRepository com PostgreSQL
type PaymentRepositoryPostgres struct {
	db *gorm.DB
}

// NewPaymentRepository cria nova instância
func NewPaymentRepository(db *gorm.DB) *PaymentRepositoryPostgres {
	return &PaymentRepositoryPostgres{db: db}
}

// Create insere novo pagamento
func (r *PaymentRepositoryPostgres) Create(ctx context.Context, payment *entity.Payment) error {
	log.Debug().Str("payment_id", payment.ID.String()).Msg("PaymentRepository Create")
	return r.db.WithContext(ctx).Create(payment).Error
}

// FindByID busca pagamento por ID
func (r *PaymentRepositoryPostgres) FindByID(ctx context.Context, id uuid.UUID) (*entity.Payment, error) {
	log.Debug().Str("id", id.String()).Msg("PaymentRepository FindByID")
	var payment entity.Payment
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&payment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

// FindByProviderSessionID busca pagamento por ID da sessão do provedor
func (r *PaymentRepositoryPostgres) FindByProviderSessionID(ctx context.Context, sessionID string) (*entity.Payment, error) {
	log.Debug().Str("session_id", sessionID).Msg("PaymentRepository FindByProviderSessionID")
	var payment entity.Payment
	err := r.db.WithContext(ctx).Where("provider_session_id = ?", sessionID).First(&payment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

// FindByUserID retorna pagamentos de um usuário
func (r *PaymentRepositoryPostgres) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Payment, error) {
	log.Debug().Str("user_id", userID.String()).Msg("PaymentRepository FindByUserID")
	var payments []*entity.Payment
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&payments).Error
	return payments, err
}

// FindLatestByUserID retorna último pagamento de um usuário
func (r *PaymentRepositoryPostgres) FindLatestByUserID(ctx context.Context, userID uuid.UUID) (*entity.Payment, error) {
	log.Debug().Str("user_id", userID.String()).Msg("PaymentRepository FindLatestByUserID")
	var payment entity.Payment
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&payment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

// FindPaidByUserID retorna pagamento pago mais recente de um usuário
func (r *PaymentRepositoryPostgres) FindPaidByUserID(ctx context.Context, userID uuid.UUID) (*entity.Payment, error) {
	log.Debug().Str("user_id", userID.String()).Msg("PaymentRepository FindPaidByUserID")
	var payment entity.Payment
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, entity.PaymentStatusPaid).
		Order("created_at DESC").
		First(&payment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

// FindPendingByUserID retorna pagamentos pendentes de um usuário
func (r *PaymentRepositoryPostgres) FindPendingByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Payment, error) {
	log.Debug().Str("user_id", userID.String()).Msg("PaymentRepository FindPendingByUserID")
	var payments []*entity.Payment
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, entity.PaymentStatusPending).
		Order("created_at DESC").
		Find(&payments).Error
	return payments, err
}

// FindByStatus retorna pagamentos com status específico
func (r *PaymentRepositoryPostgres) FindByStatus(ctx context.Context, status entity.PaymentStatus) ([]*entity.Payment, error) {
	log.Debug().Str("status", string(status)).Msg("PaymentRepository FindByStatus")
	var payments []*entity.Payment
	err := r.db.WithContext(ctx).Where("status = ?", status).Find(&payments).Error
	return payments, err
}

// FindExpiredPending retorna pagamentos pendentes expirados (mais de 24h)
func (r *PaymentRepositoryPostgres) FindExpiredPending(ctx context.Context) ([]*entity.Payment, error) {
	log.Debug().Msg("PaymentRepository FindExpiredPending")
	var payments []*entity.Payment
	err := r.db.WithContext(ctx).
		Where("status = ? AND created_at < NOW() - INTERVAL '24 hours'", entity.PaymentStatusPending).
		Find(&payments).Error
	return payments, err
}

// Update atualiza pagamento existente
func (r *PaymentRepositoryPostgres) Update(ctx context.Context, payment *entity.Payment) error {
	log.Debug().Str("payment_id", payment.ID.String()).Msg("PaymentRepository Update")
	return r.db.WithContext(ctx).Save(payment).Error
}

// UpdateStatus atualiza apenas o status do pagamento
func (r *PaymentRepositoryPostgres) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.PaymentStatus) error {
	log.Debug().Str("payment_id", id.String()).Str("status", string(status)).Msg("PaymentRepository UpdateStatus")
	return r.db.WithContext(ctx).
		Model(&entity.Payment{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// Delete deleta pagamento
func (r *PaymentRepositoryPostgres) Delete(ctx context.Context, id uuid.UUID) error {
	log.Debug().Str("payment_id", id.String()).Msg("PaymentRepository Delete")
	return r.db.WithContext(ctx).Delete(&entity.Payment{}, "id = ?", id).Error
}

// ExistsPaidByUserAndPlan verifica se existe pagamento pago para usuário e plano
func (r *PaymentRepositoryPostgres) ExistsPaidByUserAndPlan(ctx context.Context, userID, planID uuid.UUID) (bool, error) {
	log.Debug().
		Str("user_id", userID.String()).
		Str("plan_id", planID.String()).
		Msg("PaymentRepository ExistsPaidByUserAndPlan")
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Payment{}).
		Where("user_id = ? AND plan_id = ? AND status = ?", userID, planID, entity.PaymentStatusPaid).
		Count(&count).Error
	return count > 0, err
}

// CountPaidByPlan retorna quantidade de pagamentos bem-sucedidos por plano
func (r *PaymentRepositoryPostgres) CountPaidByPlan(ctx context.Context, planID uuid.UUID) (int, error) {
	log.Debug().Str("plan_id", planID.String()).Msg("PaymentRepository CountPaidByPlan")
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Payment{}).
		Where("plan_id = ? AND status = ?", planID, entity.PaymentStatusPaid).
		Count(&count).Error
	return int(count), err
}

// SumAmountByPlan retorna valor total de pagamentos bem-sucedidos por plano
func (r *PaymentRepositoryPostgres) SumAmountByPlan(ctx context.Context, planID uuid.UUID) (float64, error) {
	log.Debug().Str("plan_id", planID.String()).Msg("PaymentRepository SumAmountByPlan")
	var sum float64
	err := r.db.WithContext(ctx).
		Model(&entity.Payment{}).
		Where("plan_id = ? AND status = ?", planID, entity.PaymentStatusPaid).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&sum).Error
	return sum, err
}

// CountByStatus retorna quantidade de pagamentos com status específico
func (r *PaymentRepositoryPostgres) CountByStatus(ctx context.Context, status entity.PaymentStatus) (int, error) {
	log.Debug().Str("status", string(status)).Msg("PaymentRepository CountByStatus")
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Payment{}).
		Where("status = ?", status).
		Count(&count).Error
	return int(count), err
}
