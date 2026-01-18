// internal/handler/wizard_handler.go
package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"organiq/internal/domain/entity"
	"organiq/internal/middleware"
	"organiq/internal/usecase/wizard"
	"organiq/internal/util"
)

// WizardHandler agrupa handlers do wizard (onboarding)
type WizardHandler struct {
	saveBusinessUC     *wizard.SaveBusinessUseCase
	saveCompetitorsUC  *wizard.SaveCompetitorsUseCase
	saveIntegrationsUC *wizard.SaveIntegrationsUseCase
	generateIdeasUC    *wizard.GenerateIdeasUseCase
	getIdeasStatusUC   *wizard.GetIdeasStatusUseCase
	publishArticlesUC  *wizard.PublishArticlesUseCase
	getWizardDataUC    *wizard.GetWizardDataUseCase
}

// NewWizardHandler cria nova instância
func NewWizardHandler(
	saveBusinessUC *wizard.SaveBusinessUseCase,
	saveCompetitorsUC *wizard.SaveCompetitorsUseCase,
	saveIntegrationsUC *wizard.SaveIntegrationsUseCase,
	generateIdeasUC *wizard.GenerateIdeasUseCase,
	getIdeasStatusUC *wizard.GetIdeasStatusUseCase,
	publishArticlesUC *wizard.PublishArticlesUseCase,
	getWizardDataUC *wizard.GetWizardDataUseCase,
) *WizardHandler {
	return &WizardHandler{
		saveBusinessUC:     saveBusinessUC,
		saveCompetitorsUC:  saveCompetitorsUC,
		saveIntegrationsUC: saveIntegrationsUC,
		generateIdeasUC:    generateIdeasUC,
		getIdeasStatusUC:   getIdeasStatusUC,
		publishArticlesUC:  publishArticlesUC,
		getWizardDataUC:    getWizardDataUC,
	}
}

// ============================================
// GET /api/wizard/data
// ============================================

// GetWizardDataResponse response body
type GetWizardDataResponse struct {
	OnboardingStep int                       `json:"onboardingStep"`
	Business       *wizard.BusinessDataOutput `json:"business,omitempty"`
	Competitors    []string                  `json:"competitors,omitempty"`
	HasIntegration bool                      `json:"hasIntegration"`
}

// GetWizardData implementa GET /api/wizard/data
func (h *WizardHandler) GetWizardData(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("WizardHandler GetWizardData iniciado")

	// 1. Extrair user_id do context
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		log.Warn().Msg("GetWizardData: usuário não autenticado")
		util.RespondError(w, http.StatusUnauthorized, "unauthorized", "Usuário não autenticado")
		return
	}

	// 2. Executar use case
	input := wizard.GetWizardDataInput{
		UserID: userID,
	}

	output, err := h.getWizardDataUC.Execute(r.Context(), input)
	if err != nil {
		log.Error().Err(err).Msg("GetWizardData: erro no use case")
		if err.Error() == "user_not_found" {
			util.RespondError(w, http.StatusNotFound, "user_not_found", "Usuário não encontrado")
		} else {
			util.RespondError(w, http.StatusInternalServerError, "server_error", "Erro ao buscar dados")
		}
		return
	}

	// 3. Responder
	response := GetWizardDataResponse{
		OnboardingStep: output.OnboardingStep,
		Business:       output.Business,
		Competitors:    output.Competitors,
		HasIntegration: output.HasIntegration,
	}

	util.RespondJSON(w, http.StatusOK, response)
}

// ============================================
// POST /api/wizard/generate-ideas
// ============================================

// GenerateIdeasRequest request body (vazio, apenas context)
type GenerateIdeasRequest struct{}

// GenerateIdeasResponse response body
type GenerateIdeasResponse struct {
	JobID   string `json:"jobId"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// GenerateIdeas implementa POST /api/wizard/generate-ideas
func (h *WizardHandler) GenerateIdeas(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("WizardHandler GenerateIdeas iniciado")

	// 1. Extrair user_id do context
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		log.Warn().Msg("GenerateIdeas: usuário não autenticado")
		util.RespondError(w, http.StatusUnauthorized, "unauthorized", "Usuário não autenticado")
		return
	}

	// 2. Executar use case
	input := wizard.GenerateIdeasInput{
		UserID: userID,
	}

	output, err := h.generateIdeasUC.Execute(r.Context(), input)
	if err != nil {
		log.Error().Err(err).Msg("GenerateIdeas: erro no use case")

		// Mapear erros de negócio para HTTP
		if err.Error() == "invalid_user_id" {
			util.RespondError(w, http.StatusBadRequest, "invalid_user", "Usuário inválido")
		} else if err.Error() == "user_not_found" {
			util.RespondError(w, http.StatusNotFound, "user_not_found", "Usuário não encontrado")
		} else if err.Error() == "business_profile_not_found" {
			util.RespondError(w, http.StatusBadRequest, "missing_business_profile", "Perfil de negócio não foi preenchido. Complete o step 'business' primeiro.")
		} else if err.Error() == "business_profile_incomplete" {
			util.RespondError(w, http.StatusBadRequest, "incomplete_business_profile", "Perfil de negócio incompleto. Verifique todos os campos obrigatórios.")
		} else if err.Error() == "erro ao buscar usuário" || err.Error() == "erro ao buscar perfil de negócio" {
			util.RespondError(w, http.StatusInternalServerError, "database_error", "Erro ao acessar banco de dados")
		} else if err.Error() == "erro ao processar geração" || err.Error() == "erro ao iniciar processamento" {
			util.RespondError(w, http.StatusInternalServerError, "processing_error", "Erro ao iniciar processamento. Tente novamente.")
		} else {
			util.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		}
		return
	}

	// 3. Responder
	response := GenerateIdeasResponse{
		JobID:   output.JobID,
		Status:  output.Status,
		Message: "Sua solicitação foi enfileirada. Use o jobId para verificar o progresso.",
	}

	util.RespondJSON(w, http.StatusAccepted, response) // 202 Accepted
}

// ============================================
// GET /api/wizard/ideas-status/{jobId}
// ============================================

// GetIdeasStatusResponse response body
type GetIdeasStatusResponse struct {
	JobID    string               `json:"jobId"`
	Status   string               `json:"status"`
	Progress int                  `json:"progress"`
	Message  string               `json:"message"`
	Ideas    []IdeaStatusResponse `json:"ideas,omitempty"`
	Error    *string              `json:"errorMessage,omitempty"`
}

// IdeaStatusResponse resposta de uma ideia
type IdeaStatusResponse struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Summary  string  `json:"summary"`
	Approved bool    `json:"approved"`
	Feedback *string `json:"feedback,omitempty"`
}

// GetIdeasStatus implementa GET /api/wizard/ideas-status/{jobId}
func (h *WizardHandler) GetIdeasStatus(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("WizardHandler GetIdeasStatus iniciado")

	// 1. Extrair user_id do context
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		log.Warn().Msg("GetIdeasStatus: usuário não autenticado")
		util.RespondError(w, http.StatusUnauthorized, "unauthorized", "Usuário não autenticado")
		return
	}

	// 2. Extrair jobId do path parameter
	jobID := chi.URLParam(r, "jobId")
	if jobID == "" {
		log.Warn().Msg("GetIdeasStatus: jobId não fornecido")
		util.RespondError(w, http.StatusBadRequest, "missing_param", "jobId é obrigatório")
		return
	}

	// 3. Executar use case
	input := wizard.GetIdeasStatusInput{
		UserID: userID,
		JobID:  jobID,
	}

	output, err := h.getIdeasStatusUC.Execute(r.Context(), input)
	if err != nil {
		log.Error().Err(err).Msg("GetIdeasStatus: erro no use case")

		// Mapear erros de negócio para HTTP
		if err.Error() == "invalid_user_id" || err.Error() == "invalid_job_id" {
			util.RespondError(w, http.StatusBadRequest, "invalid_param", "Parâmetro inválido")
		} else if err.Error() == "job_not_found" {
			util.RespondError(w, http.StatusNotFound, "job_not_found", "Job não encontrado")
		} else if err.Error() == "access_denied" {
			util.RespondError(w, http.StatusForbidden, "forbidden", "Acesso negado a este job")
		} else if err.Error() == "invalid_job_type" {
			util.RespondError(w, http.StatusBadRequest, "invalid_job_type", "Tipo de job inválido")
		} else {
			util.RespondError(w, http.StatusInternalServerError, "database_error", "Erro ao buscar status")
		}
		return
	}

	// 4. Converter ideias para response
	ideasResponse := make([]IdeaStatusResponse, len(output.Ideas))
	for i, idea := range output.Ideas {
		ideasResponse[i] = IdeaStatusResponse{
			ID:       idea.ID,
			Title:    idea.Title,
			Summary:  idea.Summary,
			Approved: idea.Approved,
			Feedback: idea.Feedback,
		}
	}

	// 5. Responder
	response := GetIdeasStatusResponse{
		JobID:    output.JobID,
		Status:   output.Status,
		Progress: output.Progress,
		Message:  output.Message,
		Ideas:    ideasResponse,
		Error:    output.ErrorMsg,
	}

	util.RespondJSON(w, http.StatusOK, response)
}

// ============================================
// POST /api/wizard/business
// ============================================

// SaveBusinessRequest request body
type SaveBusinessRequest struct {
	Description        string          `json:"description" validate:"required,min=1,max=500"`
	PrimaryObjective   string          `json:"primaryObjective" validate:"required"`
	SecondaryObjective *string         `json:"secondaryObjective"`
	Location           json.RawMessage `json:"location" validate:"required"`
	SiteURL            *string         `json:"siteUrl"`
	HasBlog            bool            `json:"hasBlog"`
	BlogURLs           []string        `json:"blogUrls"`
}

// SaveBusinessResponse response body
type SaveBusinessResponse struct {
	Success      bool    `json:"success"`
	ProfileID    string  `json:"profileId"`
	BrandFileURL *string `json:"brandFileUrl,omitempty"`
}

// SaveBusiness implementa POST /api/wizard/business
func (h *WizardHandler) SaveBusiness(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("WizardHandler SaveBusiness iniciado")

	// 1. Extrair user_id do context (middleware já validou)
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		log.Warn().Msg("SaveBusiness: usuário não autenticado")
		util.RespondError(w, http.StatusUnauthorized, "unauthorized", "Usuário não autenticado")
		return
	}

	// 2. Parse multipart form (max 10MB de dados + arquivo)
	if err := r.ParseMultipartForm(10 * 1024 * 1024); err != nil {
		log.Warn().Err(err).Msg("SaveBusiness: erro ao fazer parse do multipart")
		util.RespondError(w, http.StatusBadRequest, "invalid_form", "Erro ao processar formulário")
		return
	}

	defer r.MultipartForm.RemoveAll()

	// 3. Parse campos de formulário
	description := r.FormValue("description")
	primaryObjective := r.FormValue("primaryObjective")
	secondaryObjective := r.FormValue("secondaryObjective")
	siteURL := r.FormValue("siteUrl")
	hasBlogStr := r.FormValue("hasBlog")
	locationJSON := r.FormValue("location")

	// Validar campos obrigatórios
	if len(description) == 0 || len(primaryObjective) == 0 || len(locationJSON) == 0 {
		log.Warn().Msg("SaveBusiness: campos obrigatórios ausentes")
		util.RespondError(w, http.StatusBadRequest, "validation_error", "Campos obrigatórios ausentes")
		return
	}

	// Parse boolean
	hasBlog := false
	if hasBlogStr == "true" {
		hasBlog = true
	}

	// Parse location JSON
	var location *LocationData
	if err := json.Unmarshal([]byte(locationJSON), &location); err != nil {
		log.Warn().Err(err).Msg("SaveBusiness: location JSON inválido")
		util.RespondError(w, http.StatusBadRequest, "invalid_json", "Location inválida")
		return
	}

	// Parse secondary objective (opcional)
	var secondaryObjPtr *string
	if len(secondaryObjective) > 0 {
		secondaryObjPtr = &secondaryObjective
	}

	// Parse site URL (opcional)
	var siteURLPtr *string
	if len(siteURL) > 0 {
		siteURLPtr = &siteURL
	}

	// Parse blog URLs (pode vir como array de strings ou string JSON)
	blogURLs := r.Form["blogUrls"]
	// Se veio uma única string que parece ser JSON array, fazer parse
	if len(blogURLs) == 1 && strings.HasPrefix(blogURLs[0], "[") {
		var parsedURLs []string
		if err := json.Unmarshal([]byte(blogURLs[0]), &parsedURLs); err == nil {
			blogURLs = parsedURLs
		} else {
			// Se falhou o parse, manter vazio se for "[]"
			if blogURLs[0] == "[]" {
				blogURLs = []string{}
			}
		}
	}

	// 4. Processar upload de arquivo de brand (opcional)
	var brandFile io.Reader
	var brandFileName string
	var brandFileSize int64

	file, fileHeader, err := r.FormFile("brandFile")
	if err == nil {
		defer file.Close()

		brandFile = file
		brandFileName = fileHeader.Filename
		brandFileSize = fileHeader.Size

		log.Debug().
			Str("filename", brandFileName).
			Int64("size", brandFileSize).
			Msg("SaveBusiness: arquivo de brand detectado")
	} else if err != http.ErrMissingFile {
		log.Error().Err(err).Msg("SaveBusiness: erro ao processar arquivo")
		util.RespondError(w, http.StatusBadRequest, "file_error", "Erro ao processar arquivo")
		return
	}
	// Se é MissingFile, continuar (arquivo é opcional)

	// 5. Converter location para entity.Location
	locationReq := &LocationFromRequest{
		Country:          location.Country,
		State:            location.State,
		City:             location.City,
		HasMultipleUnits: location.HasMultipleUnits,
		Units:            location.Units,
	}
	entityLocation := locationReq.ToEntity()

	// 6. Executar use case
	input := wizard.SaveBusinessInput{
		UserID:             userID,
		Description:        description,
		PrimaryObjective:   primaryObjective,
		SecondaryObjective: secondaryObjPtr,
		Location:           entityLocation,
		SiteURL:            siteURLPtr,
		HasBlog:            hasBlog,
		BlogURLs:           blogURLs,
		BrandFile:          brandFile,
		BrandFileName:      brandFileName,
		BrandFileSize:      brandFileSize,
	}

	output, err := h.saveBusinessUC.Execute(r.Context(), input)
	if err != nil {
		log.Error().Err(err).Msg("SaveBusiness: erro no use case")

		// Mapear erros de negócio para HTTP
		if err.Error() == "invalid_user_id" {
			util.RespondError(w, http.StatusBadRequest, "invalid_user", "Usuário inválido")
		} else if err.Error() == "primaryObjective inválido: deve ser 'leads', 'sales' ou 'branding'" {
			util.RespondError(w, http.StatusBadRequest, "invalid_objective", err.Error())
		} else if err.Error() == "location é obrigatório" {
			util.RespondError(w, http.StatusBadRequest, "missing_location", err.Error())
		} else if err.Error() == "arquivo deve ser PDF, JPG ou PNG" {
			util.RespondError(w, http.StatusBadRequest, "invalid_file_type", err.Error())
		} else if err.Error() == "arquivo não pode exceder 5MB" {
			util.RespondError(w, http.StatusBadRequest, "file_too_large", err.Error())
		} else {
			util.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		}
		return
	}

	// 7. Responder
	response := SaveBusinessResponse{
		Success:      output.Success,
		ProfileID:    output.ProfileID,
		BrandFileURL: output.BrandFileURL,
	}

	util.RespondJSON(w, http.StatusCreated, response)
}

// ============================================
// POST /api/wizard/competitors
// ============================================

// SaveCompetitorsRequest request body
type SaveCompetitorsRequest struct {
	CompetitorURLs []string `json:"competitorUrls" validate:"required,max=20"`
}

// SaveCompetitorsResponse response body
type SaveCompetitorsResponse struct {
	Success bool `json:"success"`
	Count   int  `json:"count"`
}

// SaveCompetitors implementa POST /api/wizard/competitors
func (h *WizardHandler) SaveCompetitors(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("WizardHandler SaveCompetitors iniciado")

	// 1. Extrair user_id do context
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		log.Warn().Msg("SaveCompetitors: usuário não autenticado")
		util.RespondError(w, http.StatusUnauthorized, "unauthorized", "Usuário não autenticado")
		return
	}

	// 2. Parse request body
	var req SaveCompetitorsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("SaveCompetitors: erro ao decodificar JSON")
		util.RespondError(w, http.StatusBadRequest, "invalid_json", "JSON inválido")
		return
	}

	// 3. Executar use case
	input := wizard.SaveCompetitorsInput{
		UserID:         userID,
		CompetitorURLs: req.CompetitorURLs,
	}

	output, err := h.saveCompetitorsUC.Execute(r.Context(), input)
	if err != nil {
		log.Error().Err(err).Msg("SaveCompetitors: erro no use case")

		if err.Error() == "invalid_user_id" {
			util.RespondError(w, http.StatusBadRequest, "invalid_user", "Usuário inválido")
		} else if err.Error() == "máximo 10 URLs de concorrentes permitidas" {
			util.RespondError(w, http.StatusBadRequest, "too_many_urls", err.Error())
		} else {
			util.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		}
		return
	}

	// 4. Responder
	response := SaveCompetitorsResponse{
		Success: output.Success,
		Count:   output.Count,
	}

	util.RespondJSON(w, http.StatusCreated, response)
}

// ============================================
// POST /api/wizard/integrations
// ============================================

// SaveIntegrationsRequest request body
type SaveIntegrationsRequest struct {
	WordPress     *WordPressRequest     `json:"wordpress"`
	SearchConsole *SearchConsoleRequest `json:"searchConsole"`
	Analytics     *AnalyticsRequest     `json:"analytics"`
}

// WordPressRequest configuração do WordPress
type WordPressRequest struct {
	SiteURL     string `json:"siteUrl" validate:"required,url"`
	Username    string `json:"username" validate:"required"`
	AppPassword string `json:"appPassword" validate:"required"`
}

// SearchConsoleRequest configuração do Search Console
type SearchConsoleRequest struct {
	PropertyURL string `json:"propertyUrl" validate:"required,url"`
}

// AnalyticsRequest configuração do Google Analytics
type AnalyticsRequest struct {
	MeasurementID string `json:"measurementId" validate:"required"`
}

// SaveIntegrationsResponse response body
type SaveIntegrationsResponse struct {
	Success                bool              `json:"success"`
	WordPressConnected     bool              `json:"wordPressConnected"`
	SearchConsoleConnected bool              `json:"searchConsoleConnected"`
	AnalyticsConnected     bool              `json:"analyticsConnected"`
	Errors                 map[string]string `json:"errors,omitempty"`
}

// SaveIntegrations implementa POST /api/wizard/integrations
func (h *WizardHandler) SaveIntegrations(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("WizardHandler SaveIntegrations iniciado")

	// 1. Extrair user_id do context
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		log.Warn().Msg("SaveIntegrations: usuário não autenticado")
		util.RespondError(w, http.StatusUnauthorized, "unauthorized", "Usuário não autenticado")
		return
	}

	// 2. Parse request body
	var req SaveIntegrationsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("SaveIntegrations: erro ao decodificar JSON")
		util.RespondError(w, http.StatusBadRequest, "invalid_json", "JSON inválido")
		return
	}

	// 3. Converter para input do use case
	var wpInput *wizard.WordPressIntegrationInput
	if req.WordPress != nil {
		wpInput = &wizard.WordPressIntegrationInput{
			SiteURL:     req.WordPress.SiteURL,
			Username:    req.WordPress.Username,
			AppPassword: req.WordPress.AppPassword,
		}
	}

	var scInput *wizard.SearchConsoleIntegrationInput
	if req.SearchConsole != nil {
		scInput = &wizard.SearchConsoleIntegrationInput{
			PropertyURL: req.SearchConsole.PropertyURL,
		}
	}

	var analyticsInput *wizard.AnalyticsIntegrationInput
	if req.Analytics != nil {
		analyticsInput = &wizard.AnalyticsIntegrationInput{
			MeasurementID: req.Analytics.MeasurementID,
		}
	}

	// 4. Executar use case
	input := wizard.SaveIntegrationsInput{
		UserID:        userID,
		WordPress:     wpInput,
		SearchConsole: scInput,
		Analytics:     analyticsInput,
	}

	output, err := h.saveIntegrationsUC.Execute(r.Context(), input)
	if err != nil {
		log.Error().Err(err).Msg("SaveIntegrations: erro no use case")

		if err.Error() == "invalid_user_id" {
			util.RespondError(w, http.StatusBadRequest, "invalid_user", "Usuário inválido")
		} else {
			util.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		}
		return
	}

	// 5. Responder
	response := SaveIntegrationsResponse{
		Success:                output.Success,
		WordPressConnected:     output.WordPressConnected,
		SearchConsoleConnected: output.SearchConsoleConnected,
		AnalyticsConnected:     output.AnalyticsConnected,
	}

	// Incluir erros no response se houver
	if len(output.Errors) > 0 {
		response.Errors = output.Errors
		// Retornar 207 (Multi-Status) se apenas algumas integrações falharam
		util.RespondJSON(w, http.StatusMultiStatus, response)
		return
	}

	// Se tudo ok, retornar 201
	util.RespondJSON(w, http.StatusCreated, response)
}

// ============================================
// HELPER TYPES
// ============================================

// LocationData para parse do JSON
type LocationData struct {
	Country          string     `json:"country"`
	State            string     `json:"state"`
	City             string     `json:"city"`
	HasMultipleUnits bool       `json:"hasMultipleUnits"`
	Units            []UnitData `json:"units"`
}

// UnitData para parse do JSON
type UnitData struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Country string `json:"country"`
	State   string `json:"state"`
	City    string `json:"city"`
}

// LocationFromRequest converter auxiliar
type LocationFromRequest struct {
	Country          string
	State            string
	City             string
	HasMultipleUnits bool
	Units            []UnitData
}

// ToEntity converte para entity.Location
func (l *LocationFromRequest) ToEntity() *entity.Location {
	var units []entity.Unit
	for _, u := range l.Units {
		unitID, _ := uuid.Parse(u.ID)
		if unitID == uuid.Nil {
			unitID = uuid.New()
		}
		units = append(units, entity.Unit{
			ID:      unitID,
			Name:    u.Name,
			Country: u.Country,
			State:   u.State,
			City:    u.City,
		})
	}

	return &entity.Location{
		Country:          l.Country,
		State:            l.State,
		City:             l.City,
		HasMultipleUnits: l.HasMultipleUnits,
		Units:            units,
	}
}

// ============================================
// POST /api/wizard/publish
// ============================================

// PublishArticleItemRequest item de request
type PublishArticleItemRequest struct {
	ID       string  `json:"id" validate:"required"`
	Feedback *string `json:"feedback"`
}

// PublishArticlesRequest request body
type PublishArticlesRequest struct {
	Articles []PublishArticleItemRequest `json:"articles" validate:"required,min=1"`
}

// PublishArticlesResponse response body
type PublishArticlesResponse struct {
	JobID   string `json:"jobId"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Count   int    `json:"count"`
}

// PublishArticles implementa POST /api/wizard/publish
func (h *WizardHandler) PublishArticles(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("WizardHandler PublishArticles iniciado")

	// 1. Extrair user_id do context
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		log.Warn().Msg("PublishArticles: usuário não autenticado")
		util.RespondError(w, http.StatusUnauthorized, "unauthorized", "Usuário não autenticado")
		return
	}

	// 2. Parse request body
	var req PublishArticlesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("PublishArticles: erro ao decodificar JSON")
		util.RespondError(w, http.StatusBadRequest, "invalid_json", "JSON inválido")
		return
	}

	// 3. Validar request
	if err := util.ValidateStruct(req); err != nil {
		log.Warn().Err(err).Msg("PublishArticles: validação falhou")
		util.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// 4. Converter para input do use case
	articles := make([]wizard.PublishArticleItem, len(req.Articles))
	for i, a := range req.Articles {
		articles[i] = wizard.PublishArticleItem{
			ID:       a.ID,
			Feedback: a.Feedback,
		}
	}

	// 5. Executar use case
	input := wizard.PublishArticlesInput{
		UserID:   userID,
		Articles: articles,
	}

	output, err := h.publishArticlesUC.Execute(r.Context(), input)
	if err != nil {
		log.Error().Err(err).Msg("PublishArticles: erro no use case")

		// Mapear erros de negócio para HTTP
		switch err.Error() {
		case "invalid_user_id":
			util.RespondError(w, http.StatusBadRequest, "invalid_user", "Usuário inválido")
		case "no_articles_provided":
			util.RespondError(w, http.StatusBadRequest, "no_articles", "Nenhum artigo fornecido")
		case "articles_not_found":
			util.RespondError(w, http.StatusNotFound, "articles_not_found", "Artigos não encontrados")
		case "integration_not_configured":
			util.RespondError(w, http.StatusBadRequest, "integration_missing", "Integração WordPress não configurada")
		default:
			util.RespondError(w, http.StatusInternalServerError, "processing_error", "Erro ao processar publicação")
		}
		return
	}

	// 6. Responder
	response := PublishArticlesResponse{
		JobID:   output.JobID,
		Status:  output.Status,
		Message: "Publicação enfileirada. Use o jobId para acompanhar o progresso.",
		Count:   output.ArticlesCount,
	}

	util.RespondJSON(w, http.StatusAccepted, response) // 202 Accepted
}
