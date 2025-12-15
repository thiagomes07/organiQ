// internal/handler/auth_handler.go
package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"organiq/internal/domain/repository"
	"organiq/internal/middleware"
	"organiq/internal/usecase/auth"
	"organiq/internal/util"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// AuthHandler agrupa handlers de autenticação
type AuthHandler struct {
	registerUC   *auth.RegisterUserUseCase
	loginUC      *auth.LoginUserUseCase
	refreshUC    *auth.RefreshAccessTokenUseCase
	logoutUC     *auth.LogoutUserUseCase
	getMeUC      *auth.GetMeUseCase
	planRepo     repository.PlanRepository
	isProduction bool
}

// NewAuthHandler cria nova instância
func NewAuthHandler(
	registerUC *auth.RegisterUserUseCase,
	loginUC *auth.LoginUserUseCase,
	refreshUC *auth.RefreshAccessTokenUseCase,
	logoutUC *auth.LogoutUserUseCase,
	getMeUC *auth.GetMeUseCase,
	planRepo repository.PlanRepository,
) *AuthHandler {
	// Detectar ambiente de produção
	env := os.Getenv("ENV")
	isProduction := env == "production"

	return &AuthHandler{
		registerUC:   registerUC,
		loginUC:      loginUC,
		refreshUC:    refreshUC,
		logoutUC:     logoutUC,
		getMeUC:      getMeUC,
		planRepo:     planRepo,
		isProduction: isProduction,
	}
}

// ============================================
// POST /api/auth/register
// ============================================

// RegisterRequest request body para registro
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// RegisterResponse response do registro
type RegisterResponse struct {
	User *UserResponse `json:"user"`
}

// UserResponse resposta contendo dados do usuário
type UserResponse struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	Email                  string `json:"email"`
	PlanID                 string `json:"planId"`
	PlanName               string `json:"planName"`
	MaxArticles            int    `json:"maxArticles"`
	ArticlesUsed           int    `json:"articlesUsed"`
	HasCompletedOnboarding bool   `json:"hasCompletedOnboarding"`
	CreatedAt              string `json:"createdAt"`
}

// Register implementa POST /api/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// 1. Parse request body
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("Erro ao decodificar register request")
		util.RespondError(w, http.StatusBadRequest, "invalid_json", "JSON inválido")
		return
	}

	// 2. Validar entrada com validator/v10 (spec 3.7)
	if err := util.ValidateStruct(req); err != nil {
		log.Warn().Err(err).Msg("Erro de validação no register")
		util.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// 3. Executar use case
	input := auth.RegisterUserInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	output, err := h.registerUC.Execute(r.Context(), input)
	if err != nil {
		log.Warn().Err(err).Str("email", req.Email).Msg("Erro no registro")

		// Mapear erro de negócio para HTTP
		if err.Error() == "email_already_exists" {
			util.RespondError(w, http.StatusConflict, "email_exists", "Email já cadastrado")
			return
		}

		util.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// 4. Configurar cookies (HttpOnly, Secure em produção, SameSite=Strict)
	h.setAuthCookies(w, output.AccessToken, output.RefreshToken)

	// 5. Responder
	planName := "Free"
	maxArticles := 0

	// Se temos plan ID, poderia buscar name do plano
	// Por simplicidade, usamos "Free" para novo usuário

	userResp := UserResponse{
		ID:                     output.User.ID.String(),
		Name:                   output.User.Name,
		Email:                  output.User.Email,
		PlanID:                 output.User.PlanID.String(),
		PlanName:               planName,
		MaxArticles:            maxArticles,
		ArticlesUsed:           output.User.ArticlesUsed,
		HasCompletedOnboarding: output.User.HasCompletedOnboarding,
		CreatedAt:              output.User.CreatedAt.Format(time.RFC3339),
	}

	util.RespondJSON(w, http.StatusCreated, RegisterResponse{User: &userResp})
}

// ============================================
// POST /api/auth/login
// ============================================

// LoginRequest request body para login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse response do login (idêntica a RegisterResponse)
type LoginResponse struct {
	User *UserResponse `json:"user"`
}

// Login implementa POST /api/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// 1. Parse request body
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("Erro ao decodificar login request")
		util.RespondError(w, http.StatusBadRequest, "invalid_json", "JSON inválido")
		return
	}

	// 2. Validar entrada com validator/v10 (spec 3.7)
	if err := util.ValidateStruct(req); err != nil {
		log.Warn().Err(err).Msg("Erro de validação no login")
		util.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// 3. Executar use case
	input := auth.LoginUserInput{
		Email:    req.Email,
		Password: req.Password,
	}

	output, err := h.loginUC.Execute(r.Context(), input)
	if err != nil {
		log.Warn().Err(err).Str("email", req.Email).Msg("Erro no login")

		if err.Error() == "invalid_credentials" {
			util.RespondError(w, http.StatusUnauthorized, "invalid_credentials", "Email ou senha inválidos")
			return
		}

		util.RespondError(w, http.StatusInternalServerError, "server_error", "Erro interno do servidor")
		return
	}

	// 3. Configurar cookies
	h.setAuthCookies(w, output.AccessToken, output.RefreshToken)

	// 4. Responder com dados do usuário
	// Buscar nome do plano do banco
	planName := h.getPlanName(r.Context(), output.User.PlanID.String())
	maxArticles := h.getMaxArticles(r.Context(), output.User.PlanID.String())

	userResp := UserResponse{
		ID:                     output.User.ID.String(),
		Name:                   output.User.Name,
		Email:                  output.User.Email,
		PlanID:                 output.User.PlanID.String(),
		PlanName:               planName,
		MaxArticles:            maxArticles,
		ArticlesUsed:           output.User.ArticlesUsed,
		HasCompletedOnboarding: output.User.HasCompletedOnboarding,
		CreatedAt:              output.User.CreatedAt.Format(time.RFC3339),
	}

	util.RespondJSON(w, http.StatusOK, LoginResponse{User: &userResp})
}

// ============================================
// POST /api/auth/refresh
// ============================================

// Refresh implementa POST /api/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	// 1. Extrair refresh token do cookie
	cookie, err := r.Cookie("refreshToken")
	if err != nil {
		log.Warn().Msg("refreshToken cookie não encontrado no refresh")
		util.RespondError(w, http.StatusUnauthorized, "unauthorized", "Refresh token não fornecido")
		return
	}

	refreshToken := cookie.Value
	if refreshToken == "" {
		util.RespondError(w, http.StatusUnauthorized, "unauthorized", "Refresh token inválido")
		return
	}

	// 2. Executar use case
	input := auth.RefreshAccessTokenInput{
		RefreshToken: refreshToken,
	}

	output, err := h.refreshUC.Execute(r.Context(), input)
	if err != nil {
		log.Warn().Err(err).Msg("Erro ao fazer refresh do token")

		if err.Error() == "refresh_token_expired" {
			util.RespondError(w, http.StatusUnauthorized, "token_expired", "Refresh token expirado")
			return
		}

		util.RespondError(w, http.StatusUnauthorized, "invalid_token", "Refresh token inválido")
		return
	}

	// 3. Setar novo access token no cookie
	h.setAccessTokenCookie(w, output.AccessToken)

	// 4. Retornar 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// ============================================
// POST /api/auth/logout
// ============================================

// Logout implementa POST /api/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// 1. Extrair dados do context (middleware já validou)
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		util.RespondError(w, http.StatusUnauthorized, "unauthorized", "Usuário não autenticado")
		return
	}

	// 2. Extrair refresh token
	cookie, err := r.Cookie("refreshToken")
	if err != nil {
		// Mesmo sem refresh token, logout é bem sucedido
		// Apenas limpar cookies
		h.clearAuthCookies(w)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// 3. Executar use case para deletar refresh token
	input := auth.LogoutUserInput{
		UserID:       userID,
		RefreshToken: cookie.Value,
	}

	if err := h.logoutUC.Execute(r.Context(), input); err != nil {
		log.Warn().Err(err).Str("user_id", userID).Msg("Erro ao fazer logout")
		// Continuar mesmo com erro
	}

	// 4. Limpar cookies
	h.clearAuthCookies(w)

	// 5. Retornar 204
	w.WriteHeader(http.StatusNoContent)
}

// ============================================
// GET /api/auth/me
// ============================================

// GetMeResponse response de /auth/me
type GetMeResponse struct {
	User *UserResponse `json:"user"`
}

// GetMe implementa GET /api/auth/me
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	// 1. Extrair user_id do context (middleware já validou)
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		util.RespondError(w, http.StatusUnauthorized, "unauthorized", "Usuário não autenticado")
		return
	}

	// 2. Executar use case
	input := auth.GetMeInput{
		UserID: userID,
	}

	output, err := h.getMeUC.Execute(r.Context(), input)
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID).Msg("Erro ao buscar usuário")

		if err.Error() == "user_not_found" {
			util.RespondError(w, http.StatusNotFound, "not_found", "Usuário não encontrado")
			return
		}

		util.RespondError(w, http.StatusInternalServerError, "server_error", "Erro interno")
		return
	}

	// 3. Preparar response
	planName := h.getPlanName(r.Context(), output.User.PlanID.String())
	maxArticles := h.getMaxArticles(r.Context(), output.User.PlanID.String())

	userResp := UserResponse{
		ID:                     output.User.ID.String(),
		Name:                   output.User.Name,
		Email:                  output.User.Email,
		PlanID:                 output.User.PlanID.String(),
		PlanName:               planName,
		MaxArticles:            maxArticles,
		ArticlesUsed:           output.User.ArticlesUsed,
		HasCompletedOnboarding: output.User.HasCompletedOnboarding,
		CreatedAt:              output.User.CreatedAt.Format(time.RFC3339),
	}

	util.RespondJSON(w, http.StatusOK, GetMeResponse{User: &userResp})
}

// ============================================
// HELPERS
// ============================================

// setAuthCookies configura access e refresh token cookies
func (h *AuthHandler) setAuthCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	// Access Token (15 minutos)
	// HttpOnly=true, Secure=true em produção, SameSite=Strict (spec 3.3)
	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.isProduction, // true em produção (HTTPS)
		SameSite: http.SameSiteStrictMode,
		MaxAge:   15 * 60, // 15 minutos
	})

	// Refresh Token (7 dias)
	// Path inclui /api/auth para acessar em refresh E logout
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		Path:     "/api/auth", // Escopo permite refresh e logout
		HttpOnly: true,
		Secure:   h.isProduction,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   7 * 24 * 60 * 60, // 7 dias
	})
}

// setAccessTokenCookie configura apenas o cookie de access token
func (h *AuthHandler) setAccessTokenCookie(w http.ResponseWriter, accessToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.isProduction,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   15 * 60,
	})
}

// clearAuthCookies limpa os cookies de autenticação
func (h *AuthHandler) clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.isProduction,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    "",
		Path:     "/api/auth", // Mesmo path de criação
		HttpOnly: true,
		Secure:   h.isProduction,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

// getPlanName retorna o nome do plano buscando do banco
func (h *AuthHandler) getPlanName(ctx context.Context, planID string) string {
	if h.planRepo == nil {
		return "Free"
	}

	planUUID, err := uuid.Parse(planID)
	if err != nil {
		return "Free"
	}

	plan, err := h.planRepo.FindByID(ctx, planUUID)
	if err != nil || plan == nil {
		return "Free"
	}

	return plan.Name
}

// getMaxArticles retorna max articles do plano buscando do banco
func (h *AuthHandler) getMaxArticles(ctx context.Context, planID string) int {
	if h.planRepo == nil {
		return 0
	}

	planUUID, err := uuid.Parse(planID)
	if err != nil {
		return 0
	}

	plan, err := h.planRepo.FindByID(ctx, planUUID)
	if err != nil || plan == nil {
		return 0
	}

	return plan.MaxArticles
}
