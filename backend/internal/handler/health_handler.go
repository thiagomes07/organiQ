package handler

import (
	"context"
	"net/http"
	"time"

	"gorm.io/gorm"

	"organiq/internal/infra/database"
	"organiq/internal/infra/queue"
	"organiq/internal/infra/storage"
	"organiq/internal/util"
)

// HealthHandler expõe endpoints de verificação de saúde dos serviços.
type HealthHandler struct {
	db         *gorm.DB
	storage    storage.StorageService
	queue      queue.QueueService
	queueNames []string
}

// NewHealthHandler cria uma nova instância de HealthHandler.
func NewHealthHandler(
	db *gorm.DB,
	storage storage.StorageService,
	queue queue.QueueService,
	queueNames []string,
) *HealthHandler {
	return &HealthHandler{
		db:         db,
		storage:    storage,
		queue:      queue,
		queueNames: queueNames,
	}
}

// HealthResponse representa a resposta do endpoint.
type HealthResponse struct {
	Status     string            `json:"status"`
	Timestamp  string            `json:"timestamp"`
	Components []ComponentStatus `json:"components"`
}

// ComponentStatus descreve o estado de um componente individual.
type ComponentStatus struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Latency  string `json:"latency"`
	ErrorMsg string `json:"error,omitempty"`
}

// Check executa as verificações e retorna JSON padronizado.
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	components := []ComponentStatus{
		h.checkDatabase(ctx),
		h.checkStorage(ctx),
	}
	components = append(components, h.checkQueues(ctx)...)

	overallStatus := "healthy"
	statusCode := http.StatusOK
	for _, component := range components {
		if component.Status != "up" {
			overallStatus = "degraded"
			statusCode = http.StatusServiceUnavailable
			break
		}
	}

	response := HealthResponse{
		Status:     overallStatus,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Components: components,
	}

	util.RespondJSON(w, statusCode, response)
}

func (h *HealthHandler) checkDatabase(ctx context.Context) ComponentStatus {
	start := time.Now()
	status := ComponentStatus{Name: "database"}

	if err := database.HealthCheck(h.db); err != nil {
		status.Status = "down"
		status.ErrorMsg = err.Error()
		util.LoggerFromContext(ctx).Error().Err(err).Msg("database health check failed")
		status.Latency = time.Since(start).String()
		return status
	}

	status.Status = "up"
	status.Latency = time.Since(start).String()
	return status
}

func (h *HealthHandler) checkStorage(ctx context.Context) ComponentStatus {
	start := time.Now()
	status := ComponentStatus{Name: "storage"}

	if h.storage == nil {
		status.Status = "skipped"
		status.ErrorMsg = "storage not configured"
		status.Latency = "0s"
		return status
	}

	if err := h.storage.HealthCheck(ctx); err != nil {
		status.Status = "down"
		status.ErrorMsg = err.Error()
		util.LoggerFromContext(ctx).Error().Err(err).Msg("storage health check failed")
		status.Latency = time.Since(start).String()
		return status
	}

	status.Status = "up"
	status.Latency = time.Since(start).String()
	return status
}

func (h *HealthHandler) checkQueues(ctx context.Context) []ComponentStatus {
	if h.queue == nil {
		return []ComponentStatus{{
			Name:     "queue",
			Status:   "skipped",
			Latency:  "0s",
			ErrorMsg: "queue service not configured",
		}}
	}

	statuses := make([]ComponentStatus, 0, len(h.queueNames))
	for _, queueName := range h.queueNames {
		if queueName == "" {
			continue
		}

		start := time.Now()
		component := ComponentStatus{Name: "queue:" + queueName}

		if err := h.queue.HealthCheck(ctx, queueName); err != nil {
			component.Status = "down"
			component.ErrorMsg = err.Error()
			util.LoggerFromContext(ctx).Error().Err(err).Str("queue", queueName).Msg("queue health check failed")
			component.Latency = time.Since(start).String()
			statuses = append(statuses, component)
			continue
		}

		component.Status = "up"
		component.Latency = time.Since(start).String()
		statuses = append(statuses, component)
	}

	if len(statuses) == 0 {
		return []ComponentStatus{{
			Name:     "queue",
			Status:   "skipped",
			Latency:  "0s",
			ErrorMsg: "queue names not configured",
		}}
	}

	return statuses
}
