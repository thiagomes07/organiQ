package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"organiq/internal/domain/entity"
	"organiq/internal/middleware"
	"organiq/internal/usecase/account"
	"organiq/internal/util"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// AccountHandler reúne handlers relacionados à conta.
type AccountHandler struct {
	getAccountUC         *account.GetAccountUseCase
	updateProfileUC      *account.UpdateProfileUseCase
	updateIntegrationsUC *account.UpdateIntegrationsUseCase
	getPlanUC            *account.GetPlanUseCase
}

// NewAccountHandler cria nova instância de AccountHandler.
func NewAccountHandler(
	getAccountUC *account.GetAccountUseCase,
	updateProfileUC *account.UpdateProfileUseCase,
	updateIntegrationsUC *account.UpdateIntegrationsUseCase,
	getPlanUC *account.GetPlanUseCase,
) *AccountHandler {
	return &AccountHandler{
		getAccountUC:         getAccountUC,
		updateProfileUC:      updateProfileUC,
		updateIntegrationsUC: updateIntegrationsUC,
		getPlanUC:            getPlanUC,
	}
}

// AccountUserResponse representa o usuário autenticado.
type AccountUserResponse struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	Email                  string `json:"email"`
	PlanID                 string `json:"planId"`
	ArticlesUsed           int    `json:"articlesUsed"`
	HasCompletedOnboarding bool   `json:"hasCompletedOnboarding"`
	CreatedAt              string `json:"createdAt"`
	UpdatedAt              string `json:"updatedAt"`
}

// AccountPlanResponse descreve o plano vigente.
type AccountPlanResponse struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	MaxArticles       int      `json:"maxArticles"`
	ArticlesUsed      int      `json:"articlesUsed"`
	RemainingArticles int      `json:"remainingArticles"`
	LimitReached      bool     `json:"limitReached"`
	Price             float64  `json:"price"`
	Active            bool     `json:"active"`
	Features          []string `json:"features"`
}

// IntegrationStatusResponse resume uma integração configurada.
type IntegrationStatusResponse struct {
	Type        string `json:"type"`
	Enabled     bool   `json:"enabled"`
	Configured  bool   `json:"configured"`
	LastUpdated string `json:"lastUpdated,omitempty"`
}

// AccountResponse agrega todos os dados da conta.
type AccountResponse struct {
	User         AccountUserResponse         `json:"user"`
	Plan         AccountPlanResponse         `json:"plan"`
	Integrations []IntegrationStatusResponse `json:"integrations"`
}

// UpdateProfileRequest payload para atualização de perfil.
type UpdateProfileRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UpdateProfileResponse resposta do update de perfil.
type UpdateProfileResponse struct {
	User AccountUserResponse `json:"user"`
}

// UpdateIntegrationsRequest payload de integrações.
type UpdateIntegrationsRequest struct {
	WordPress *WordPressIntegrationRequest `json:"wordpress"`
	Analytics *AnalyticsIntegrationRequest `json:"analytics"`
}

// WordPressIntegrationRequest request específico do WordPress.
type WordPressIntegrationRequest struct {
	SiteURL     string `json:"siteUrl"`
	Username    string `json:"username"`
	AppPassword string `json:"appPassword"`
	Enabled     bool   `json:"enabled"`
}

// AnalyticsIntegrationRequest request específico do Analytics.
type AnalyticsIntegrationRequest struct {
	MeasurementID string `json:"measurementId"`
	Enabled       bool   `json:"enabled"`
}

// UpdateIntegrationsResponse resposta da atualização de integrações.
type UpdateIntegrationsResponse struct {
	Integrations []IntegrationStatusResponse `json:"integrations"`
}

// GetPlanResponse resposta para detalhes do plano.
type GetPlanResponse struct {
	Plan AccountPlanResponse `json:"plan"`
}

// GetAccount retorna o agregado da conta autenticada.
func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		util.RespondUnauthorized(w, "Usuário não autenticado")
		return
	}

	output, err := h.getAccountUC.Execute(r.Context(), account.GetAccountInput{UserID: userID})
	if err != nil {
		logger := requestLogger(r)
		logger.Warn().Err(err).Msg("get account failed")
		handleAccountError(w, err)
		return
	}

	resp := AccountResponse{
		User:         buildAccountUserResponse(output.User),
		Plan:         buildPlanResponse(output.Plan, output.User.ArticlesUsed),
		Integrations: buildIntegrationStatuses(output.Integrations),
	}

	util.RespondOK(w, resp)
}

// UpdateProfile atualiza nome e email do usuário.
func (h *AccountHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		util.RespondUnauthorized(w, "Usuário não autenticado")
		return
	}

	var req UpdateProfileRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		requestLogger(r).Warn().Err(err).Msg("invalid profile payload")
		util.RespondBadRequest(w, "Payload inválido")
		return
	}

	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(req.Email)
	if name == "" || email == "" {
		util.RespondBadRequest(w, "Nome e email são obrigatórios")
		return
	}

	if !util.IsValidEmail(email) {
		util.RespondBadRequest(w, "Email inválido")
		return
	}

	output, err := h.updateProfileUC.Execute(r.Context(), account.UpdateProfileInput{
		UserID: userID,
		Name:   name,
		Email:  email,
	})
	if err != nil {
		requestLogger(r).Warn().Err(err).Msg("update profile failed")
		handleAccountError(w, err)
		return
	}

	util.RespondOK(w, UpdateProfileResponse{User: buildAccountUserResponse(output.User)})
}

// UpdateIntegrations atualiza configurações de integrações.
func (h *AccountHandler) UpdateIntegrations(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		util.RespondUnauthorized(w, "Usuário não autenticado")
		return
	}

	var req UpdateIntegrationsRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		requestLogger(r).Warn().Err(err).Msg("invalid integrations payload")
		util.RespondBadRequest(w, "Payload inválido")
		return
	}

	if req.WordPress == nil && req.Analytics == nil {
		util.RespondBadRequest(w, "Informe ao menos uma integração para atualizar")
		return
	}

	input := account.UpdateIntegrationsInput{UserID: userID}
	if req.WordPress != nil {
		input.WordPress = &account.WordPressIntegrationInput{
			SiteURL:     strings.TrimSpace(req.WordPress.SiteURL),
			Username:    strings.TrimSpace(req.WordPress.Username),
			AppPassword: strings.TrimSpace(req.WordPress.AppPassword),
			Enabled:     req.WordPress.Enabled,
		}
	}
	if req.Analytics != nil {
		input.Analytics = &account.AnalyticsIntegrationInput{
			MeasurementID: strings.TrimSpace(req.Analytics.MeasurementID),
			Enabled:       req.Analytics.Enabled,
		}
	}

	output, err := h.updateIntegrationsUC.Execute(r.Context(), input)
	if err != nil {
		requestLogger(r).Warn().Err(err).Msg("update integrations failed")
		handleAccountError(w, err)
		return
	}

	resp := UpdateIntegrationsResponse{Integrations: buildIntegrationStatuses(output.Integrations)}
	util.RespondOK(w, resp)
}

// GetPlan retorna os dados do plano do usuário.
func (h *AccountHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		util.RespondUnauthorized(w, "Usuário não autenticado")
		return
	}

	output, err := h.getPlanUC.Execute(r.Context(), account.GetPlanInput{UserID: userID})
	if err != nil {
		requestLogger(r).Warn().Err(err).Msg("get plan failed")
		handleAccountError(w, err)
		return
	}

	resp := GetPlanResponse{Plan: buildPlanResponse(output.Plan, output.ArticlesUsed)}
	util.RespondOK(w, resp)
}

func buildAccountUserResponse(user *entity.User) AccountUserResponse {
	return AccountUserResponse{
		ID:                     user.ID.String(),
		Name:                   user.Name,
		Email:                  user.Email,
		PlanID:                 user.PlanID.String(),
		ArticlesUsed:           user.ArticlesUsed,
		HasCompletedOnboarding: user.HasCompletedOnboarding,
		CreatedAt:              formatTime(user.CreatedAt),
		UpdatedAt:              formatTime(user.UpdatedAt),
	}
}

func buildPlanResponse(plan *entity.Plan, articlesUsed int) AccountPlanResponse {
	remaining := plan.MaxArticles - articlesUsed
	if remaining < 0 {
		remaining = 0
	}
	limitReached := plan.MaxArticles > 0 && articlesUsed >= plan.MaxArticles

	features := make([]string, 0, len(plan.Features))
	for _, feature := range plan.Features {
		features = append(features, feature)
	}

	return AccountPlanResponse{
		ID:                plan.ID.String(),
		Name:              plan.Name,
		MaxArticles:       plan.MaxArticles,
		ArticlesUsed:      articlesUsed,
		RemainingArticles: remaining,
		LimitReached:      limitReached,
		Price:             plan.Price,
		Active:            plan.Active,
		Features:          features,
	}
}

func buildIntegrationStatuses(integrations []*entity.Integration) []IntegrationStatusResponse {
	typeMap := map[entity.IntegrationType]IntegrationStatusResponse{}
	for _, integration := range integrations {
		if integration == nil {
			continue
		}
		typeMap[integration.Type] = IntegrationStatusResponse{
			Type:        string(integration.Type),
			Enabled:     integration.Enabled,
			Configured:  len(integration.Config) > 0,
			LastUpdated: formatTime(integration.UpdatedAt),
		}
	}

	orderedTypes := []entity.IntegrationType{
		entity.IntegrationTypeWordPress,
		entity.IntegrationTypeAnalytics,
		entity.IntegrationTypeSearchConsole,
	}

	statuses := make([]IntegrationStatusResponse, 0, len(orderedTypes)+len(typeMap))
	for _, intType := range orderedTypes {
		if status, ok := typeMap[intType]; ok {
			statuses = append(statuses, status)
			delete(typeMap, intType)
		} else {
			statuses = append(statuses, IntegrationStatusResponse{
				Type:       string(intType),
				Enabled:    false,
				Configured: false,
			})
		}
	}

	if len(typeMap) > 0 {
		var extras []entity.IntegrationType
		for intType := range typeMap {
			extras = append(extras, intType)
		}
		sort.Slice(extras, func(i, j int) bool {
			return extras[i] < extras[j]
		})
		for _, intType := range extras {
			statuses = append(statuses, typeMap[intType])
		}
	}

	return statuses
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func handleAccountError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, account.ErrInvalidUserID):
		util.RespondBadRequest(w, "ID de usuário inválido")
	case errors.Is(err, account.ErrInvalidName):
		util.RespondBadRequest(w, "Nome inválido")
	case errors.Is(err, account.ErrInvalidEmail):
		util.RespondBadRequest(w, "Email inválido")
	case errors.Is(err, account.ErrEmailAlreadyExists):
		util.RespondConflict(w, "Email já cadastrado")
	case errors.Is(err, account.ErrUserNotFound):
		util.RespondNotFound(w)
	case errors.Is(err, account.ErrPlanNotFound):
		util.RespondNotFound(w)
	case errors.Is(err, account.ErrNoIntegrationPayload):
		util.RespondBadRequest(w, "Informe ao menos uma integração")
	case errors.Is(err, account.ErrWordPressConfigIncomplete):
		util.RespondBadRequest(w, "Configuração do WordPress incompleta")
	case errors.Is(err, account.ErrAnalyticsConfigIncomplete):
		util.RespondBadRequest(w, "Configuração do Analytics incompleta")
	default:
		util.RespondInternalServerError(w)
	}
}

func requestLogger(r *http.Request) *zerolog.Logger {
	if logger := log.Ctx(r.Context()); logger != nil {
		return logger
	}
	l := log.Logger
	return &l
}
