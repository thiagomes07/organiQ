package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"organiq/internal/domain/entity"
	"organiq/internal/domain/repository"
	"organiq/internal/middleware"
	"organiq/internal/usecase/payment"
	"organiq/internal/util"
)

// PaymentHandler agrupa handlers de pagamentos
type PaymentHandler struct {
	userRepo          repository.UserRepository
	planRepo          repository.PlanRepository
	paymentRepo       repository.PaymentRepository
	cryptoService     *util.CryptoService
	stripeSecretKey   string
	stripePubKey      string
	stripeWebhookSec  string
	mercadoPagoToken  string
	mercadoPagoSecret string
}

// NewPaymentHandler cria nova instância
func NewPaymentHandler(
	userRepo repository.UserRepository,
	planRepo repository.PlanRepository,
	paymentRepo repository.PaymentRepository,
	cryptoService *util.CryptoService,
	stripeSecretKey string,
	stripePubKey string,
	stripeWebhookSec string,
	mercadoPagoToken string,
	mercadoPagoSecret string,
) *PaymentHandler {
	return &PaymentHandler{
		userRepo:          userRepo,
		planRepo:          planRepo,
		paymentRepo:       paymentRepo,
		cryptoService:     cryptoService,
		stripeSecretKey:   stripeSecretKey,
		stripePubKey:      stripePubKey,
		stripeWebhookSec:  stripeWebhookSec,
		mercadoPagoToken:  mercadoPagoToken,
		mercadoPagoSecret: mercadoPagoSecret,
	}
}

// ============================================
// POST /api/payments/create-checkout
// ============================================

// CreateCheckoutRequest request body
type CreateCheckoutRequest struct {
	PlanID string `json:"planId" validate:"required"`
}

// CreateCheckoutResponse response body
type CreateCheckoutResponse struct {
	CheckoutURL string `json:"checkoutUrl"`
	SessionID   string `json:"sessionId"`
	Provider    string `json:"provider"`
}

// CreateCheckout cria sessão de checkout para assinatura de plano
func (h *PaymentHandler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("PaymentHandler CreateCheckout iniciado")

	// 1. Extrair user_id do context
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		log.Warn().Msg("CreateCheckout: usuário não autenticado")
		util.RespondError(w, http.StatusUnauthorized, "unauthorized", "Usuário não autenticado")
		return
	}

	// 2. Parse request body
	var req CreateCheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("CreateCheckout: erro ao decodificar JSON")
		util.RespondError(w, http.StatusBadRequest, "invalid_json", "JSON inválido")
		return
	}

	// 3. Validar entrada
	if len(req.PlanID) == 0 {
		log.Warn().Msg("CreateCheckout: planId não fornecido")
		util.RespondError(w, http.StatusBadRequest, "missing_plan", "planId é obrigatório")
		return
	}

	// 4. Executar use case
	input := payment.CreateCheckoutInput{
		UserID: userID,
		PlanID: req.PlanID,
	}

	output, err := payment.CreateCheckout(
		r.Context(),
		input,
		h.userRepo,
		h.planRepo,
		h.paymentRepo,
	)

	if err != nil {
		log.Error().Err(err).Msg("CreateCheckout: erro no use case")

		// Mapear erros de negócio para HTTP
		if err.Error() == "plan_not_found" {
			util.RespondError(w, http.StatusNotFound, "plan_not_found", "Plano não encontrado")
		} else if err.Error() == "user_not_found" {
			util.RespondError(w, http.StatusNotFound, "user_not_found", "Usuário não encontrado")
		} else if err.Error() == "invalid_plan_id" {
			util.RespondError(w, http.StatusBadRequest, "invalid_plan", "Plano inválido")
		} else if err.Error() == "user_already_has_plan" {
			// Usuário já possui o plano solicitado — responder de forma amigável
			log.Warn().Err(err).Msg("CreateCheckout: usuário já possui este plano")
			util.RespondError(w, http.StatusBadRequest, "user_already_has_plan", "Usuário já possui este plano")
		} else {
			util.RespondError(w, http.StatusInternalServerError, "server_error", "Erro ao criar checkout")
		}
		return
	}

	// 5. Responder
	response := CreateCheckoutResponse{
		CheckoutURL: output.CheckoutURL,
		SessionID:   output.SessionID,
		Provider:    output.Provider,
	}

	util.RespondJSON(w, http.StatusCreated, response)
}

// ============================================
// POST /api/payments/webhook
// ============================================

// WebhookPayload payload genérico de webhook
type WebhookPayload struct {
	Type   string          `json:"type"`
	Data   json.RawMessage `json:"data"`
	Source string          `json:"source"` // "stripe" ou "mercadopago"
}

// StripeWebhookPayload payload do Stripe
type StripeWebhookPayload struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Data struct {
		Object struct {
			ID       string `json:"id"`
			Status   string `json:"status"`
			Metadata struct {
				SessionID string `json:"session_id"`
				UserID    string `json:"user_id"`
				PlanID    string `json:"plan_id"`
			} `json:"metadata"`
		} `json:"object"`
	} `json:"data"`
}

// MercadoPagoWebhookPayload payload do Mercado Pago
type MercadoPagoWebhookPayload struct {
	ID     int64  `json:"id"`
	Type   string `json:"type"`
	Action string `json:"action"`
	Data   struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	} `json:"data"`
}

// Webhook endpoint público para webhooks de pagamento
func (h *PaymentHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("PaymentHandler Webhook iniciado")

	// 1. Ler body completo para validação de assinatura
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Webhook: erro ao ler body")
		util.RespondError(w, http.StatusBadRequest, "invalid_body", "Erro ao ler body")
		return
	}

	// 2. Detectar provedor pelo header
	provider := h.detectProvider(r)
	if provider == "" {
		log.Warn().Msg("Webhook: provedor não detectado")
		util.RespondError(w, http.StatusBadRequest, "unknown_provider", "Provedor não identificado")
		return
	}

	log.Debug().Str("provider", provider).Msg("Webhook: provedor detectado")

	// 3. Validar assinatura
	if !h.validateWebhookSignature(r, bodyBytes, provider) {
		log.Error().
			Str("provider", provider).
			Msg("Webhook: assinatura inválida")
		util.RespondError(w, http.StatusUnauthorized, "invalid_signature", "Assinatura inválida")
		return
	}

	// 4. Processar webhook baseado no provedor
	switch provider {
	case "stripe":
		h.handleStripeWebhook(w, r, bodyBytes)
	case "mercadopago":
		h.handleMercadoPagoWebhook(w, r, bodyBytes)
	default:
		util.RespondError(w, http.StatusBadRequest, "unknown_provider", "Provedor não suportado")
	}
}

// ============================================
// POST /api/payments/create-portal-session
// ============================================

// CreatePortalSessionResponse response body
type CreatePortalSessionResponse struct {
	URL string `json:"url"`
}

// CreatePortalSession cria sessão do portal de gerenciamento de assinatura
func (h *PaymentHandler) CreatePortalSession(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("PaymentHandler CreatePortalSession iniciado")

	// 1. Extrair user_id do context
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		log.Warn().Msg("CreatePortalSession: usuário não autenticado")
		util.RespondError(w, http.StatusUnauthorized, "unauthorized", "Usuário não autenticado")
		return
	}

	// 2. Buscar usuário
	user, err := h.userRepo.FindByID(r.Context(), util.MustParseUUID(userID))
	if err != nil || user == nil {
		log.Error().Err(err).Msg("CreatePortalSession: erro ao buscar usuário")
		util.RespondError(w, http.StatusNotFound, "user_not_found", "Usuário não encontrado")
		return
	}

	// 3. Validar que usuário tem pagamento ativo
	payment, err := h.paymentRepo.FindPaidByUserID(r.Context(), user.ID)
	if err != nil || payment == nil {
		log.Warn().Str("user_id", userID).Msg("CreatePortalSession: usuário sem pagamento ativo")
		util.RespondError(w, http.StatusBadRequest, "no_active_payment", "Nenhuma assinatura ativa")
		return
	}

	// 4. Gerar portal URL baseado no provedor
	portalURL := h.generatePortalURL(userID, payment.Provider)

	log.Info().
		Str("user_id", userID).
		Str("provider", string(payment.Provider)).
		Msg("CreatePortalSession bem-sucedido")

	// 5. Responder
	response := CreatePortalSessionResponse{
		URL: portalURL,
	}

	util.RespondJSON(w, http.StatusOK, response)
}

// ============================================
// POST /api/payments/confirm-free-plan
// ============================================

// ConfirmFreePlanResponse response body
type ConfirmFreePlanResponse struct {
	Success        bool   `json:"success"`
	OnboardingStep int    `json:"onboardingStep"`
	Message        string `json:"message"`
}

// ConfirmFreePlan confirma seleção do plano Free e avança para onboarding
// Este endpoint existe porque o plano Free não passa pelo checkout de pagamento,
// então precisamos de uma forma de atualizar o onboarding_step manualmente.
func (h *PaymentHandler) ConfirmFreePlan(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("PaymentHandler ConfirmFreePlan iniciado")

	// 1. Extrair user_id do context
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		log.Warn().Msg("ConfirmFreePlan: usuário não autenticado")
		util.RespondError(w, http.StatusUnauthorized, "unauthorized", "Usuário não autenticado")
		return
	}

	// 2. Buscar usuário
	user, err := h.userRepo.FindByID(r.Context(), util.MustParseUUID(userID))
	if err != nil || user == nil {
		log.Error().Err(err).Msg("ConfirmFreePlan: erro ao buscar usuário")
		util.RespondError(w, http.StatusNotFound, "user_not_found", "Usuário não encontrado")
		return
	}

	// 3. Verificar se o usuário tem um plano Free (pela lógica, novo usuário começa com Free)
	plan, err := h.planRepo.FindByID(r.Context(), user.PlanID)
	if err != nil || plan == nil {
		log.Error().Err(err).Msg("ConfirmFreePlan: erro ao buscar plano")
		util.RespondError(w, http.StatusInternalServerError, "plan_not_found", "Erro ao buscar plano")
		return
	}

	// 4. Atualizar onboarding_step para 1 se ainda estiver em 0
	if user.OnboardingStep < 1 {
		user.OnboardingStep = 1
		if err := h.userRepo.Update(r.Context(), user); err != nil {
			log.Error().Err(err).Msg("ConfirmFreePlan: erro ao atualizar usuário")
			util.RespondError(w, http.StatusInternalServerError, "update_failed", "Erro ao atualizar usuário")
			return
		}
		log.Info().Str("user_id", userID).Msg("ConfirmFreePlan: onboarding_step atualizado para 1")
	}

	// 5. Responder
	response := ConfirmFreePlanResponse{
		Success:        true,
		OnboardingStep: user.OnboardingStep,
		Message:        "Plano Free confirmado. Pronto para iniciar o onboarding.",
	}

	util.RespondJSON(w, http.StatusOK, response)
}

// ============================================
// GET /api/payments/status/{sessionId}
// ============================================

// GetStatusResponse response body
type GetStatusResponse struct {
	SessionID string `json:"sessionId"`
	Status    string `json:"status"`
	PlanID    string `json:"planId,omitempty"`
	Provider  string `json:"provider,omitempty"`
	PaidAt    string `json:"paidAt,omitempty"`
}

// GetStatus retorna status de uma sessão de pagamento
func (h *PaymentHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("PaymentHandler GetStatus iniciado")

	// 1. Extrair user_id do context
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		log.Warn().Msg("GetStatus: usuário não autenticado")
		util.RespondError(w, http.StatusUnauthorized, "unauthorized", "Usuário não autenticado")
		return
	}

	// 2. Extrair sessionId do path
	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		log.Warn().Msg("GetStatus: sessionId não fornecido")
		util.RespondError(w, http.StatusBadRequest, "missing_param", "sessionId é obrigatório")
		return
	}

	// 3. Buscar pagamento pela sessionID
	payment, err := h.paymentRepo.FindByProviderSessionID(r.Context(), sessionID)
	if err != nil {
		log.Error().Err(err).Str("session_id", sessionID).Msg("GetStatus: erro ao buscar pagamento")
		util.RespondError(w, http.StatusInternalServerError, "database_error", "Erro ao buscar status")
		return
	}

	if payment == nil {
		log.Warn().Str("session_id", sessionID).Msg("GetStatus: sessão não encontrada")
		util.RespondError(w, http.StatusNotFound, "session_not_found", "Sessão de pagamento não encontrada")
		return
	}

	// 4. Validar que pagamento pertence ao usuário
	if payment.UserID.String() != userID {
		log.Warn().
			Str("session_id", sessionID).
			Str("payment_user", payment.UserID.String()).
			Str("request_user", userID).
			Msg("GetStatus: acesso negado")
		util.RespondError(w, http.StatusForbidden, "forbidden", "Acesso negado")
		return
	}

	// 5. Construir response
	response := GetStatusResponse{
		SessionID: sessionID,
		Status:    string(payment.Status),
		PlanID:    payment.PlanID.String(),
		Provider:  string(payment.Provider),
	}

	// Se pagamento foi concluído, usar UpdatedAt como data de pagamento
	if payment.IsPaid() {
		response.PaidAt = payment.UpdatedAt.Format("2006-01-02T15:04:05Z")
	}

	log.Debug().
		Str("session_id", sessionID).
		Str("status", response.Status).
		Msg("GetStatus bem-sucedido")

	util.RespondJSON(w, http.StatusOK, response)
}

// ============================================
// PRIVATE METHODS - WEBHOOKS
// ============================================

// detectProvider detecta o provedor pelo header ou conteúdo
func (h *PaymentHandler) detectProvider(r *http.Request) string {
	// Tentar identificar por header customizado
	if provider := r.Header.Get("X-Payment-Provider"); provider != "" {
		return provider
	}

	// Tentar identificar por Stripe-Signature header
	if r.Header.Get("Stripe-Signature") != "" {
		return "stripe"
	}

	// Tentar identificar por User-Agent do Mercado Pago
	if strings.Contains(r.Header.Get("User-Agent"), "Mercado-Pago") {
		return "mercadopago"
	}

	return ""
}

// validateWebhookSignature valida assinatura do webhook
func (h *PaymentHandler) validateWebhookSignature(r *http.Request, body []byte, provider string) bool {
	switch provider {
	case "stripe":
		return h.validateStripeSignature(r, body)
	case "mercadopago":
		return h.validateMercadoPagoSignature(r, body)
	default:
		return false
	}
}

// validateStripeSignature valida assinatura do Stripe
func (h *PaymentHandler) validateStripeSignature(r *http.Request, body []byte) bool {
	log.Debug().Msg("validateStripeSignature iniciado")

	// TODO: Stripe integration
	// Em produção:
	// 1. Obter header "Stripe-Signature"
	// 2. Extrair timestamp e signature
	// 3. Recriar signed content com timestamp + payload
	// 4. Validar HMAC-SHA256 com webhook secret

	signature := r.Header.Get("Stripe-Signature")
	if signature == "" {
		log.Warn().Msg("validateStripeSignature: header Stripe-Signature ausente")
		return false
	}

	// Validação mockada por enquanto
	log.Debug().Msg("validateStripeSignature: assinatura aceita (mock)")
	return true
}

// validateMercadoPagoSignature valida assinatura do Mercado Pago
func (h *PaymentHandler) validateMercadoPagoSignature(r *http.Request, body []byte) bool {
	log.Debug().Msg("validateMercadoPagoSignature iniciado")

	// TODO: Mercado Pago integration
	// Em produção:
	// 1. Obter header "x-signature"
	// 2. Extrair timestamp e hash
	// 3. Recriar signed content com timestamp + payload
	// 4. Validar HMAC-SHA256 com webhook secret

	xSignature := r.Header.Get("x-signature")
	if xSignature == "" {
		log.Warn().Msg("validateMercadoPagoSignature: header x-signature ausente")
		return false
	}

	// Validação mockada por enquanto
	log.Debug().Msg("validateMercadoPagoSignature: assinatura aceita (mock)")
	return true
}

// handleStripeWebhook processa webhook do Stripe
func (h *PaymentHandler) handleStripeWebhook(w http.ResponseWriter, r *http.Request, bodyBytes []byte) {
	log.Debug().Msg("handleStripeWebhook iniciado")

	// 1. Parse webhook payload
	var payload StripeWebhookPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		log.Error().Err(err).Msg("handleStripeWebhook: erro ao fazer parse")
		util.RespondError(w, http.StatusBadRequest, "invalid_payload", "Payload inválido")
		return
	}

	// 2. Filtrar por tipo de evento (apenas checkout.session.completed)
	if !strings.Contains(payload.Type, "checkout.session.completed") &&
		!strings.Contains(payload.Type, "invoice.payment_succeeded") {
		log.Debug().Str("event_type", payload.Type).Msg("handleStripeWebhook: evento ignorado")
		w.WriteHeader(http.StatusOK)
		return
	}

	// 3. Extrair informações da sessão
	sessionID := payload.Data.Object.Metadata.SessionID
	userIDStr := payload.Data.Object.Metadata.UserID
	planIDStr := payload.Data.Object.Metadata.PlanID

	log.Debug().
		Str("session_id", sessionID).
		Str("user_id", userIDStr).
		Str("plan_id", planIDStr).
		Msg("handleStripeWebhook: processando pagamento")

	// 4. Executar use case de webhook
	input := payment.ProcessWebhookInput{
		SessionID:   sessionID,
		UserID:      userIDStr,
		PlanID:      planIDStr,
		Status:      "paid",
		Provider:    entity.PaymentProviderStripe,
		ProviderRef: payload.Data.Object.ID,
	}

	if err := payment.ProcessWebhook(r.Context(), input, h.userRepo, h.paymentRepo); err != nil {
		log.Error().Err(err).Msg("handleStripeWebhook: erro no use case")
		util.RespondError(w, http.StatusInternalServerError, "processing_error", "Erro ao processar pagamento")
		return
	}

	log.Info().
		Str("user_id", userIDStr).
		Str("plan_id", planIDStr).
		Msg("handleStripeWebhook bem-sucedido")

	// 5. Responder com 200 OK para Stripe reconhecer recebimento
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}

// handleMercadoPagoWebhook processa webhook do Mercado Pago
func (h *PaymentHandler) handleMercadoPagoWebhook(w http.ResponseWriter, r *http.Request, bodyBytes []byte) {
	log.Debug().Msg("handleMercadoPagoWebhook iniciado")

	// 1. Parse webhook payload
	var payload MercadoPagoWebhookPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		log.Error().Err(err).Msg("handleMercadoPagoWebhook: erro ao fazer parse")
		util.RespondError(w, http.StatusBadRequest, "invalid_payload", "Payload inválido")
		return
	}

	// 2. Filtrar por tipo de evento (apenas payment)
	if payload.Type != "payment" {
		log.Debug().Str("event_type", payload.Type).Msg("handleMercadoPagoWebhook: evento ignorado")
		w.WriteHeader(http.StatusOK)
		return
	}

	// 3. Validar que status é approved
	if payload.Data.Status != "approved" {
		log.Debug().
			Str("payment_id", payload.Data.ID).
			Str("status", payload.Data.Status).
			Msg("handleMercadoPagoWebhook: pagamento não aprovado")
		w.WriteHeader(http.StatusOK)
		return
	}

	// 4. TODO: Buscar detalhes do pagamento da API do Mercado Pago
	// Aqui precisaríamos fazer uma chamada GET /payments/{payment_id}
	// Para obter userID, planID, etc. dos metadados do pagamento

	log.Info().
		Str("payment_id", payload.Data.ID).
		Msg("handleMercadoPagoWebhook: processando pagamento aprovado")

	// 5. Responder com 200 OK para Mercado Pago reconhecer recebimento
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}

// generatePortalURL gera URL do portal de gerenciamento
func (h *PaymentHandler) generatePortalURL(userID string, provider entity.PaymentProvider) string {
	// TODO: Integração real com portais dos provedores

	switch provider {
	case entity.PaymentProviderStripe:
		// Em produção: Chamar Stripe API para criar sessão de portal
		return fmt.Sprintf("https://billing.stripe.com/portal/session/%s", userID)

	case entity.PaymentProviderMercadoPago:
		// Em produção: Redirecionar para Mercado Pago
		return fmt.Sprintf("https://www.mercadopago.com.br/account/portal/%s", userID)

	default:
		return ""
	}
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// verifyStripeSignatureManual verifica assinatura do Stripe manualmente
// Implementação de referência (não será usada por enquanto)
func verifyStripeSignatureManual(signature string, body []byte, secret string) bool {
	// Parse signature header: t=timestamp,v1=signature
	parts := strings.Split(signature, ",")
	var timestamp, sig string

	for _, part := range parts {
		if strings.HasPrefix(part, "t=") {
			timestamp = strings.TrimPrefix(part, "t=")
		} else if strings.HasPrefix(part, "v1=") {
			sig = strings.TrimPrefix(part, "v1=")
		}
	}

	if timestamp == "" || sig == "" {
		return false
	}

	// Recriar signed content
	signedContent := fmt.Sprintf("%s.%s", timestamp, string(body))

	// Calcular HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedContent))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	// Constant-time comparison
	return hmac.Equal([]byte(sig), []byte(expectedSig))
}
