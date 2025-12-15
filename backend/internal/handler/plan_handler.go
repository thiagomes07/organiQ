package handler

import (
	"net/http"
	"sort"

	"organiq/internal/domain/repository"
	"organiq/internal/util"
)

// PlanHandler expõe endpoints públicos relacionados aos planos.
type PlanHandler struct {
	planRepo repository.PlanRepository
}

// NewPlanHandler cria nova instância do handler.
func NewPlanHandler(planRepo repository.PlanRepository) *PlanHandler {
	return &PlanHandler{planRepo: planRepo}
}

// PlanResponse representa um plano retornado no endpoint público.
type PlanResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	MaxArticles int      `json:"maxArticles"`
	Price       float64  `json:"price"`
	Features    []string `json:"features"`
}

// ListPlans responde com todos os planos ativos.
func (h *PlanHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := util.LoggerFromContext(ctx)

	plans, err := h.planRepo.FindAllActive(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("failed to list plans")
		util.RespondInternalServerError(w)
		return
	}

	// Ordena por preço e nome para garantir consistência entre respostas.
	sort.Slice(plans, func(i, j int) bool {
		if plans[i].Price == plans[j].Price {
			return plans[i].Name < plans[j].Name
		}
		return plans[i].Price < plans[j].Price
	})

	responses := make([]PlanResponse, 0, len(plans))
	for _, plan := range plans {
		features := make([]string, len(plan.Features))
		copy(features, plan.Features)

		responses = append(responses, PlanResponse{
			ID:          plan.ID.String(),
			Name:        plan.Name,
			MaxArticles: plan.MaxArticles,
			Price:       plan.Price,
			Features:    features,
		})
	}

	util.RespondJSON(w, http.StatusOK, responses)
}
