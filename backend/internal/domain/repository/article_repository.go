// internal/domain/repository/article_repository.go
package repository

import (
	"context"

	"organiq/internal/domain/entity"

	"github.com/google/uuid"
)

// ============================================
// ARTICLE JOB REPOSITORY
// ============================================

// ArticleJobRepository define contrato para operações com article jobs
type ArticleJobRepository interface {
	// Create insere novo job
	Create(ctx context.Context, job *entity.ArticleJob) error

	// FindByID busca job por ID
	FindByID(ctx context.Context, id uuid.UUID) (*entity.ArticleJob, error)

	// FindByUserID retorna todos os jobs de um usuário
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.ArticleJob, error)

	// FindByUserIDAndType retorna jobs de um tipo específico de um usuário
	FindByUserIDAndType(ctx context.Context, userID uuid.UUID, jobType entity.JobType) ([]*entity.ArticleJob, error)

	// FindByStatus retorna jobs com status específico
	FindByStatus(ctx context.Context, status entity.JobStatus) ([]*entity.ArticleJob, error)

	// FindPendingJobs retorna jobs enfileirados/processando
	FindPendingJobs(ctx context.Context) ([]*entity.ArticleJob, error)

	// Update atualiza job existente
	Update(ctx context.Context, job *entity.ArticleJob) error

	// Delete deleta job
	Delete(ctx context.Context, id uuid.UUID) error

	// UpdateStatus atualiza apenas o status e progresso
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.JobStatus, progress int) error

	// UpdateError atualiza status para falhado com mensagem
	UpdateError(ctx context.Context, id uuid.UUID, errorMsg string) error

	// CountByUserID retorna total de jobs de um usuário
	CountByUserID(ctx context.Context, userID uuid.UUID) (int, error)

	// CountByUserIDAndType retorna total de jobs de um tipo para um usuário
	CountByUserIDAndType(ctx context.Context, userID uuid.UUID, jobType entity.JobType) (int, error)
}

// ============================================
// ARTICLE IDEA REPOSITORY
// ============================================

// ArticleIdeaRepository define contrato para operações com ideias de artigos
type ArticleIdeaRepository interface {
	// Create insere nova ideia
	Create(ctx context.Context, idea *entity.ArticleIdea) error

	// CreateBatch insere múltiplas ideias em lote
	CreateBatch(ctx context.Context, ideas []*entity.ArticleIdea) error

	// FindByID busca ideia por ID
	FindByID(ctx context.Context, id uuid.UUID) (*entity.ArticleIdea, error)

	// FindByJobID retorna todas as ideias de um job
	FindByJobID(ctx context.Context, jobID uuid.UUID) ([]*entity.ArticleIdea, error)

	// FindByUserID retorna todas as ideias de um usuário
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.ArticleIdea, error)

	// FindApprovedByJobID retorna ideias aprovadas de um job
	FindApprovedByJobID(ctx context.Context, jobID uuid.UUID) ([]*entity.ArticleIdea, error)

	// FindPendingByUserID retorna ideias pendentes (não aprovadas) de um usuário
	FindPendingByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.ArticleIdea, error)

	// Update atualiza ideia existente
	Update(ctx context.Context, idea *entity.ArticleIdea) error

	// Delete deleta ideia
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByJobID deleta todas as ideias de um job
	DeleteByJobID(ctx context.Context, jobID uuid.UUID) error

	// DeleteUnapprovedByUserID deleta todas as ideias não aprovadas de um usuário
	DeleteUnapprovedByUserID(ctx context.Context, userID uuid.UUID) error

	// ApproveMultiple marca múltiplas ideias como aprovadas
	ApproveMultiple(ctx context.Context, ids []uuid.UUID) error

	// CountApprovedByJobID retorna quantidade de ideias aprovadas de um job
	CountApprovedByJobID(ctx context.Context, jobID uuid.UUID) (int, error)

	// CountByUserID retorna total de ideias de um usuário
	CountByUserID(ctx context.Context, userID uuid.UUID) (int, error)

	// CountApprovedByUserID retorna total de ideias aprovadas de um usuário
	CountApprovedByUserID(ctx context.Context, userID uuid.UUID) (int, error)

	// CountGenerationsInLastHour conta quantos jobs de geração o usuário executou na última hora
	CountGenerationsInLastHour(ctx context.Context, userID uuid.UUID) (int, error)
}

// ============================================
// ARTICLE REPOSITORY
// ============================================

// ArticleRepository define contrato para operações com artigos
type ArticleRepository interface {
	// Create insere novo artigo
	Create(ctx context.Context, article *entity.Article) error

	// FindByID busca artigo por ID
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Article, error)

	// FindByUserID retorna artigos de um usuário com paginação
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int, sortBy, order string) ([]*entity.Article, error)

	// FindByUserIDAndStatus retorna artigos de um usuário com status específico
	FindByUserIDAndStatus(ctx context.Context, userID uuid.UUID, status entity.ArticleStatus, limit, offset int, sortBy, order string) ([]*entity.Article, error)

	// FindByIdeaID busca artigo associado a uma ideia
	FindByIdeaID(ctx context.Context, ideaID uuid.UUID) (*entity.Article, error)

	// FindPublishedByUserID retorna artigos publicados de um usuário
	FindPublishedByUserID(ctx context.Context, userID uuid.UUID, limit, offset int, sortBy, order string) ([]*entity.Article, error)

	// FindErrorsByUserID retorna artigos com erro de um usuário
	FindErrorsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int, sortBy, order string) ([]*entity.Article, error)

	// FindByStatus retorna artigos com status específico
	FindByStatus(ctx context.Context, status entity.ArticleStatus) ([]*entity.Article, error)

	// Update atualiza artigo existente
	Update(ctx context.Context, article *entity.Article) error

	// Delete deleta artigo
	Delete(ctx context.Context, id uuid.UUID) error

	// UpdateStatus atualiza apenas o status do artigo
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ArticleStatus) error

	// UpdateStatusWithError atualiza status para erro com mensagem
	UpdateStatusWithError(ctx context.Context, id uuid.UUID, errMsg string) error

	// SetPublished marca artigo como publicado com URL
	SetPublished(ctx context.Context, id uuid.UUID, postURL string) error

	// SetContent atualiza conteúdo do artigo
	SetContent(ctx context.Context, id uuid.UUID, content string) error

	// CountByUserID retorna total de artigos de um usuário
	CountByUserID(ctx context.Context, userID uuid.UUID) (int, error)

	// CountPublishedByUserID retorna total de artigos publicados de um usuário
	CountPublishedByUserID(ctx context.Context, userID uuid.UUID) (int, error)

	// CountByUserIDAndStatus retorna total de artigos com status específico
	CountByUserIDAndStatus(ctx context.Context, userID uuid.UUID, status entity.ArticleStatus) (int, error)

	// CountErrorsByUserID retorna total de artigos com erro de um usuário
	CountErrorsByUserID(ctx context.Context, userID uuid.UUID) (int, error)
}

// PaginatedArticleResult representa resultado paginado de artigos
type PaginatedArticleResult struct {
	Articles  []*entity.Article
	Total     int
	Page      int
	PageSize  int
	TotalPages int
}
