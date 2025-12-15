package payment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"organiq/internal/domain/entity"
	"organiq/internal/domain/repository"
)

// ============================================
// CREATE CHECKOUT
// ============================================

// CreateCheckoutInput dados de entrada
type CreateCheckoutInput struct {
	UserID string // UUID como string
	PlanID string // UUID como string
}

// CreateCheckoutOutput dados de saída
type CreateCheckoutOutput struct {
	CheckoutURL string
	SessionID   string
	Provider    string
}

// CreateCheckout executa o caso de uso de criar checkout
func CreateCheckout(
	ctx context.Context,
	input CreateCheckoutInput,
	userRepo repository.UserRepository,
	planRepo repository.PlanRepository,
	paymentRepo repository.PaymentRepository,
) (*CreateCheckoutOutput, error) {
	log.Debug().
		Str("user_id", input.UserID).
		Str("plan_id", input.PlanID).
		Msg("CreateCheckout iniciado")

	// 1. Parse e validar IDs
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		log.Error().Err(err).Msg("CreateCheckout: user_id inválido")
		return nil, errors.New("invalid_user_id")
	}

	planID, err := uuid.Parse(input.PlanID)
	if err != nil {
		log.Error().Err(err).Msg("CreateCheckout: plan_id inválido")
		return nil, errors.New("invalid_plan_id")
	}

	// 2. Buscar usuário
	user, err := userRepo.FindByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("CreateCheckout: erro ao buscar usuário")
		return nil, errors.New("user_not_found")
	}

	if user == nil {
		log.Warn().Str("user_id", input.UserID).Msg("CreateCheckout: usuário não encontrado")
		return nil, errors.New("user_not_found")
	}

	// 3. Buscar plano
	plan, err := planRepo.FindByID(ctx, planID)
	if err != nil {
		log.Error().Err(err).Msg("CreateCheckout: erro ao buscar plano")
		return nil, errors.New("plan_not_found")
	}

	if plan == nil {
		log.Warn().Str("plan_id", input.PlanID).Msg("CreateCheckout: plano não encontrado")
		return nil, errors.New("plan_not_found")
	}

	// 4. Validar que plano é diferente do atual
	if user.PlanID == planID {
		log.Warn().
			Str("user_id", input.UserID).
			Str("plan_id", input.PlanID).
			Msg("CreateCheckout: usuário já tem este plano")
		return nil, errors.New("user_already_has_plan")
	}

	// 5. Validar que plano está ativo
	if !plan.Active {
		log.Warn().
			Str("plan_id", input.PlanID).
			Msg("CreateCheckout: plano não está ativo")
		return nil, errors.New("plan_not_active")
	}

	// 6. Validar que plano não é Free
	if plan.Name == "Free" {
		log.Warn().Msg("CreateCheckout: não é possível fazer checkout para plano Free")
		return nil, errors.New("free_plan_not_allowed")
	}

	// 7. Criar registro de pagamento com status 'pending'
	paymentID := uuid.New()
	sessionID := generateSessionID()
	provider := entity.PaymentProviderStripe // Default Stripe por enquanto

	payment := &entity.Payment{
		ID:                paymentID,
		UserID:            userID,
		PlanID:            planID,
		Provider:          provider,
		ProviderSessionID: sessionID,
		Status:            entity.PaymentStatusPending,
		Amount:            plan.Price,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := payment.Validate(); err != nil {
		log.Error().Err(err).Msg("CreateCheckout: payment inválido")
		return nil, errors.New("payment_validation_failed")
	}

	if err := paymentRepo.Create(ctx, payment); err != nil {
		log.Error().Err(err).Msg("CreateCheckout: erro ao criar payment no banco")
		return nil, errors.New("failed_to_create_payment")
	}

	log.Debug().
		Str("payment_id", paymentID.String()).
		Str("session_id", sessionID).
		Msg("CreateCheckout: payment criado")

	// 8. Gerar URL de checkout
	// TODO: Stripe integration
	// Em produção:
	// 1. Chamar Stripe API para criar Checkout Session
	// 2. Passar metadata com userID, planID, paymentID
	// 3. Retornar URL da sessão
	checkoutURL := generateMockCheckoutURL(sessionID, plan.Name, plan.Price)

	log.Info().
		Str("user_id", input.UserID).
		Str("plan_id", input.PlanID).
		Str("session_id", sessionID).
		Msg("CreateCheckout bem-sucedido")

	return &CreateCheckoutOutput{
		CheckoutURL: checkoutURL,
		SessionID:   sessionID,
		Provider:    string(provider),
	}, nil
}

// ============================================
// PROCESS WEBHOOK
// ============================================

// ProcessWebhookInput dados de entrada do webhook
type ProcessWebhookInput struct {
	SessionID   string                      // ID da sessão de pagamento
	UserID      string                      // UUID como string
	PlanID      string                      // UUID como string
	Status      string                      // "paid" ou "failed"
	Provider    entity.PaymentProvider      // Provider do pagamento
	ProviderRef string                      // ID do pagamento no provedor
}

// ProcessWebhook processa webhook de confirmação de pagamento
func ProcessWebhook(
	ctx context.Context,
	input ProcessWebhookInput,
	userRepo repository.UserRepository,
	paymentRepo repository.PaymentRepository,
) error {
	log.Debug().
		Str("session_id", input.SessionID).
		Str("user_id", input.UserID).
		Str("plan_id", input.PlanID).
		Str("status", input.Status).
		Msg("ProcessWebhook iniciado")

	// 1. Parse IDs
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		log.Error().Err(err).Msg("ProcessWebhook: user_id inválido")
		return errors.New("invalid_user_id")
	}

	planID, err := uuid.Parse(input.PlanID)
	if err != nil {
		log.Error().Err(err).Msg("ProcessWebhook: plan_id inválido")
		return errors.New("invalid_plan_id")
	}

	// 2. Buscar pagamento pela sessão
	payment, err := paymentRepo.FindByProviderSessionID(ctx, input.SessionID)
	if err != nil {
		log.Error().Err(err).Msg("ProcessWebhook: erro ao buscar pagamento")
		return errors.New("payment_not_found")
	}

	if payment == nil {
		log.Warn().Str("session_id", input.SessionID).Msg("ProcessWebhook: pagamento não encontrado")
		return errors.New("payment_not_found")
	}

	// 3. Validar que pagamento pertence ao usuário e plano
	if payment.UserID != userID || payment.PlanID != planID {
		log.Error().
			Str("payment_user_id", payment.UserID.String()).
			Str("request_user_id", input.UserID).
			Str("payment_plan_id", payment.PlanID.String()).
			Str("request_plan_id", input.PlanID).
			Msg("ProcessWebhook: dados do pagamento não correspondem")
		return errors.New("payment_mismatch")
	}

	// 4. Processar baseado no status
	switch input.Status {
	case "paid":
		return processPaymentSucceeded(ctx, payment, planID, userID, userRepo, paymentRepo)
	case "failed":
		return processPaymentFailed(ctx, payment, paymentRepo)
	default:
		log.Warn().Str("status", input.Status).Msg("ProcessWebhook: status desconhecido")
		return errors.New("unknown_status")
	}
}

// processPaymentSucceeded processa pagamento bem-sucedido
func processPaymentSucceeded(
	ctx context.Context,
	payment *entity.Payment,
	planID uuid.UUID,
	userID uuid.UUID,
	userRepo repository.UserRepository,
	paymentRepo repository.PaymentRepository,
) error {
	log.Debug().
		Str("payment_id", payment.ID.String()).
		Str("user_id", userID.String()).
		Str("plan_id", planID.String()).
		Msg("processPaymentSucceeded iniciado")

	// 1. Atualizar status do pagamento para 'paid'
	payment.SetPaid()
	if err := paymentRepo.Update(ctx, payment); err != nil {
		log.Error().
			Err(err).
			Str("payment_id", payment.ID.String()).
			Msg("processPaymentSucceeded: erro ao atualizar payment")
		return errors.New("failed_to_update_payment")
	}

	log.Debug().
		Str("payment_id", payment.ID.String()).
		Msg("processPaymentSucceeded: payment marcado como paid")

	// 2. Buscar usuário
	user, err := userRepo.FindByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("processPaymentSucceeded: erro ao buscar usuário")
		return errors.New("user_not_found")
	}

	if user == nil {
		log.Warn().Str("user_id", userID.String()).Msg("processPaymentSucceeded: usuário não encontrado")
		return errors.New("user_not_found")
	}

	// 3. Atualizar plano do usuário
	user.UpdatePlan(planID)
	if err := user.Validate(); err != nil {
		log.Error().Err(err).Msg("processPaymentSucceeded: usuário inválido")
		return errors.New("user_validation_failed")
	}

	// 4. Resetar contador de artigos usados
	user.ArticlesUsed = 0

	// 5. Resetar onboarding flag (usuário pode refazer wizard se necessário)
	user.HasCompletedOnboarding = false

	// 6. Salvar mudanças do usuário
	if err := userRepo.Update(ctx, user); err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID.String()).
			Msg("processPaymentSucceeded: erro ao atualizar usuário")
		return errors.New("failed_to_update_user")
	}

	log.Info().
		Str("payment_id", payment.ID.String()).
		Str("user_id", userID.String()).
		Str("old_plan_id", user.PlanID.String()).
		Str("new_plan_id", planID.String()).
		Msg("processPaymentSucceeded bem-sucedido")

	// 7. TODO: Enviar email de confirmação
	// sendConfirmationEmail(user.Email, planName)

	// 8. TODO: Registrar evento no analytics
	// trackPaymentEvent("payment_succeeded", user.ID, planID)

	return nil
}

// processPaymentFailed processa pagamento falhado
func processPaymentFailed(
	ctx context.Context,
	payment *entity.Payment,
	paymentRepo repository.PaymentRepository,
) error {
	log.Debug().
		Str("payment_id", payment.ID.String()).
		Msg("processPaymentFailed iniciado")

	// 1. Atualizar status do pagamento para 'failed'
	payment.SetFailed()
	if err := paymentRepo.Update(ctx, payment); err != nil {
		log.Error().
			Err(err).
			Str("payment_id", payment.ID.String()).
			Msg("processPaymentFailed: erro ao atualizar payment")
		return errors.New("failed_to_update_payment")
	}

	log.Info().
		Str("payment_id", payment.ID.String()).
		Msg("processPaymentFailed bem-sucedido")

	// 2. TODO: Enviar email de falha de pagamento
	// sendPaymentFailedEmail(payment.UserID)

	return nil
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// generateSessionID gera ID único para sessão de pagamento
func generateSessionID() string {
	return uuid.New().String()
}

// generateMockCheckoutURL gera URL mockada de checkout
func generateMockCheckoutURL(sessionID string, planName string, price float64) string {
	return fmt.Sprintf(
		"https://checkout.mock.local/session/%s?plan=%s&price=%.2f",
		sessionID,
		planName,
		price,
	)
}
