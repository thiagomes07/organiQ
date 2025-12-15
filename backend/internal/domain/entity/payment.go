// internal/domain/entity/payment.go
package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Payment representa um pagamento/transação
type Payment struct {
	ID                  uuid.UUID     `gorm:"primaryKey" json:"id"`
	UserID              uuid.UUID     `gorm:"index;column:user_id" json:"userId"`
	PlanID              uuid.UUID     `gorm:"column:plan_id" json:"planId"`
	Provider            PaymentProvider `gorm:"column:provider" json:"provider"`
	ProviderSessionID   string        `gorm:"uniqueIndex;column:provider_session_id" json:"providerSessionId"`
	Status              PaymentStatus `gorm:"index;column:status" json:"status"`
	Amount              float64       `gorm:"column:amount" json:"amount"`
	CreatedAt           time.Time     `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt           time.Time     `gorm:"column:updated_at" json:"updatedAt"`
}

// PaymentProvider enum para provedores de pagamento
type PaymentProvider string

const (
	PaymentProviderStripe       PaymentProvider = "stripe"
	PaymentProviderMercadoPago  PaymentProvider = "mercadopago"
)

// IsValid verifica se o provedor é válido
func (pp PaymentProvider) IsValid() bool {
	return pp == PaymentProviderStripe || pp == PaymentProviderMercadoPago
}

// PaymentStatus enum para status do pagamento
type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusPaid    PaymentStatus = "paid"
	PaymentStatusFailed  PaymentStatus = "failed"
)

// IsValid verifica se o status é válido
func (ps PaymentStatus) IsValid() bool {
	return ps == PaymentStatusPending ||
		ps == PaymentStatusPaid ||
		ps == PaymentStatusFailed
}

// TableName especifica o nome da tabela
func (Payment) TableName() string {
	return "payments"
}

// Validate valida as regras de negócio
func (p *Payment) Validate() error {
	if p.ID == uuid.Nil {
		return errors.New("id é obrigatório")
	}

	if p.UserID == uuid.Nil {
		return errors.New("user_id é obrigatório")
	}

	if p.PlanID == uuid.Nil {
		return errors.New("plan_id é obrigatório")
	}

	if !p.Provider.IsValid() {
		return errors.New("provedor de pagamento inválido")
	}

	if len(p.ProviderSessionID) == 0 {
		return errors.New("provider_session_id é obrigatório")
	}

	if !p.Status.IsValid() {
		return errors.New("status de pagamento inválido")
	}

	if p.Amount < 0 {
		return errors.New("amount não pode ser negativo")
	}

	return nil
}

// SetPending marca pagamento como pendente
func (p *Payment) SetPending() {
	p.Status = PaymentStatusPending
	p.UpdatedAt = time.Now()
}

// SetPaid marca pagamento como pago
func (p *Payment) SetPaid() {
	p.Status = PaymentStatusPaid
	p.UpdatedAt = time.Now()
}

// SetFailed marca pagamento como falhado
func (p *Payment) SetFailed() {
	p.Status = PaymentStatusFailed
	p.UpdatedAt = time.Now()
}

// IsPending verifica se pagamento está pendente
func (p *Payment) IsPending() bool {
	return p.Status == PaymentStatusPending
}

// IsPaid verifica se pagamento foi feito
func (p *Payment) IsPaid() bool {
	return p.Status == PaymentStatusPaid
}

// IsFailed verifica se pagamento falhou
func (p *Payment) IsFailed() bool {
	return p.Status == PaymentStatusFailed
}

// IsExpired verifica se sessão de checkout expirou (24 horas)
func (p *Payment) IsExpired() bool {
	return p.IsPending() && time.Since(p.CreatedAt) > 24*time.Hour
}

// CanRetry verifica se pagamento pode ser retentado
func (p *Payment) CanRetry() bool {
	return p.IsFailed() || p.IsExpired()
}
