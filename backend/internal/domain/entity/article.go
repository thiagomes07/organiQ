// internal/domain/entity/article.go
package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ============================================
// ARTICLE JOB
// ============================================

// ArticleJob representa um job de geração ou publicação de artigos
type ArticleJob struct {
	ID           uuid.UUID       `gorm:"primaryKey" json:"id"`
	UserID       uuid.UUID       `gorm:"index;column:user_id" json:"userId"`
	Type         JobType         `gorm:"index;column:type" json:"type"`
	Status       JobStatus       `gorm:"index;column:status" json:"status"`
	Progress     int             `gorm:"column:progress" json:"progress"`
	Payload      JobPayload      `gorm:"type:jsonb;column:payload" json:"payload"`
	ErrorMessage *string         `gorm:"column:error_message" json:"errorMessage"`
	CreatedAt    time.Time       `gorm:"index;column:created_at" json:"createdAt"`
	UpdatedAt    time.Time       `gorm:"column:updated_at" json:"updatedAt"`
}

// JobType enum para tipos de job
type JobType string

const (
	JobTypeGenerateIdeas JobType = "generate_ideas"
	JobTypePublish       JobType = "publish"
)

// IsValid verifica se o tipo é válido
func (jt JobType) IsValid() bool {
	return jt == JobTypeGenerateIdeas || jt == JobTypePublish
}

// JobStatus enum para status do job
type JobStatus string

const (
	JobStatusQueued      JobStatus = "queued"
	JobStatusProcessing  JobStatus = "processing"
	JobStatusCompleted   JobStatus = "completed"
	JobStatusFailed      JobStatus = "failed"
)

// IsValid verifica se o status é válido
func (js JobStatus) IsValid() bool {
	return js == JobStatusQueued ||
		js == JobStatusProcessing ||
		js == JobStatusCompleted ||
		js == JobStatusFailed
}

// JobPayload payload genérico do job
type JobPayload map[string]interface{}

// Scan implementa sql.Scanner para JobPayload
func (jp *JobPayload) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion failed for JobPayload")
	}
	return json.Unmarshal(bytes, &jp)
}

// Value implementa driver.Valuer para JobPayload
func (jp JobPayload) Value() (driver.Value, error) {
	return json.Marshal(jp)
}

// TableName especifica o nome da tabela
func (ArticleJob) TableName() string {
	return "article_jobs"
}

// Validate valida as regras de negócio
func (aj *ArticleJob) Validate() error {
	if aj.ID == uuid.Nil {
		return errors.New("id é obrigatório")
	}

	if aj.UserID == uuid.Nil {
		return errors.New("user_id é obrigatório")
	}

	if !aj.Type.IsValid() {
		return errors.New("tipo de job inválido")
	}

	if !aj.Status.IsValid() {
		return errors.New("status de job inválido")
	}

	if aj.Progress < 0 || aj.Progress > 100 {
		return errors.New("progress deve estar entre 0 e 100")
	}

	if len(aj.Payload) == 0 {
		return errors.New("payload não pode estar vazio")
	}

	return nil
}

// SetQueued marca job como enfileirado
func (aj *ArticleJob) SetQueued() {
	aj.Status = JobStatusQueued
	aj.Progress = 0
	aj.ErrorMessage = nil
	aj.UpdatedAt = time.Now()
}

// SetProcessing marca job como em processamento
func (aj *ArticleJob) SetProcessing(progress int) error {
	if progress < 0 || progress > 100 {
		return errors.New("progress deve estar entre 0 e 100")
	}

	aj.Status = JobStatusProcessing
	aj.Progress = progress
	aj.UpdatedAt = time.Now()
	return nil
}

// SetCompleted marca job como completo
func (aj *ArticleJob) SetCompleted() {
	aj.Status = JobStatusCompleted
	aj.Progress = 100
	aj.ErrorMessage = nil
	aj.UpdatedAt = time.Now()
}

// SetFailed marca job como falhado com mensagem de erro
func (aj *ArticleJob) SetFailed(errMsg string) {
	aj.Status = JobStatusFailed
	msg := errMsg
	aj.ErrorMessage = &msg
	aj.UpdatedAt = time.Now()
}

// IsComplete verifica se job foi completado
func (aj *ArticleJob) IsComplete() bool {
	return aj.Status == JobStatusCompleted || aj.Status == JobStatusFailed
}

// ============================================
// ARTICLE IDEA
// ============================================

// ArticleIdea representa uma ideia de artigo gerada
type ArticleIdea struct {
	ID          uuid.UUID `gorm:"primaryKey" json:"id"`
	UserID      uuid.UUID `gorm:"index;column:user_id" json:"userId"`
	JobID       uuid.UUID `gorm:"index;column:job_id" json:"jobId"`
	Title       string    `gorm:"type:text;column:title" json:"title"`
	Summary     string    `gorm:"type:text;column:summary" json:"summary"`
	Approved    bool      `gorm:"index;column:approved" json:"approved"`
	Feedback    *string   `gorm:"column:feedback" json:"feedback"`
	GeneratedAt time.Time `gorm:"index;column:generated_at" json:"generatedAt"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"createdAt"`
}

// TableName especifica o nome da tabela
func (ArticleIdea) TableName() string {
	return "article_ideas"
}

// Validate valida as regras de negócio
func (ai *ArticleIdea) Validate() error {
	if ai.ID == uuid.Nil {
		return errors.New("id é obrigatório")
	}

	if ai.UserID == uuid.Nil {
		return errors.New("user_id é obrigatório")
	}

	if ai.JobID == uuid.Nil {
		return errors.New("job_id é obrigatório")
	}

	if len(ai.Title) == 0 || len(ai.Title) > 500 {
		return errors.New("title deve ter entre 1 e 500 caracteres")
	}

	if len(ai.Summary) == 0 || len(ai.Summary) > 2000 {
		return errors.New("summary deve ter entre 1 e 2000 caracteres")
	}

	if ai.Feedback != nil && len(*ai.Feedback) > 1000 {
		return errors.New("feedback deve ter no máximo 1000 caracteres")
	}

	return nil
}

// Approve marca a ideia como aprovada
func (ai *ArticleIdea) Approve() {
	ai.Approved = true
}

// Reject marca a ideia como rejeitada com feedback
func (ai *ArticleIdea) Reject(feedback string) error {
	if len(feedback) == 0 {
		return errors.New("feedback é obrigatório para rejeição")
	}

	if len(feedback) > 1000 {
		return errors.New("feedback deve ter no máximo 1000 caracteres")
	}

	ai.Approved = false
	ai.Feedback = &feedback
	return nil
}

// SetFeedback define feedback para a ideia
func (ai *ArticleIdea) SetFeedback(feedback string) error {
	if len(feedback) > 1000 {
		return errors.New("feedback deve ter no máximo 1000 caracteres")
	}

	if len(feedback) == 0 {
		ai.Feedback = nil
	} else {
		ai.Feedback = &feedback
	}
	return nil
}

// ============================================
// ARTICLE
// ============================================

// Article representa um artigo publicado
type Article struct {
	ID           uuid.UUID      `gorm:"primaryKey" json:"id"`
	UserID       uuid.UUID      `gorm:"index;column:user_id" json:"userId"`
	IdeaID       *uuid.UUID     `gorm:"column:idea_id" json:"ideaId"`
	Title        string         `gorm:"type:text;column:title" json:"title"`
	Content      *string        `gorm:"type:text;column:content" json:"content"`
	Status       ArticleStatus  `gorm:"index;column:status" json:"status"`
	PostURL      *string        `gorm:"column:post_url" json:"postUrl"`
	ErrorMessage *string        `gorm:"column:error_message" json:"errorMessage"`
	CreatedAt    time.Time      `gorm:"index;column:created_at" json:"createdAt"`
	UpdatedAt    time.Time      `gorm:"column:updated_at" json:"updatedAt"`
}

// ArticleStatus enum para status do artigo
type ArticleStatus string

const (
	ArticleStatusGenerating  ArticleStatus = "generating"
	ArticleStatusPublishing  ArticleStatus = "publishing"
	ArticleStatusPublished   ArticleStatus = "published"
	ArticleStatusError       ArticleStatus = "error"
)

// IsValid verifica se o status é válido
func (as ArticleStatus) IsValid() bool {
	return as == ArticleStatusGenerating ||
		as == ArticleStatusPublishing ||
		as == ArticleStatusPublished ||
		as == ArticleStatusError
}

// TableName especifica o nome da tabela
func (Article) TableName() string {
	return "articles"
}

// Validate valida as regras de negócio
func (a *Article) Validate() error {
	if a.ID == uuid.Nil {
		return errors.New("id é obrigatório")
	}

	if a.UserID == uuid.Nil {
		return errors.New("user_id é obrigatório")
	}

	if len(a.Title) == 0 || len(a.Title) > 500 {
		return errors.New("title deve ter entre 1 e 500 caracteres")
	}

	if !a.Status.IsValid() {
		return errors.New("status de artigo inválido")
	}

	if a.PostURL != nil && len(*a.PostURL) == 0 {
		return errors.New("postUrl deve ser não-vazio se fornecido")
	}

	if a.ErrorMessage != nil && len(*a.ErrorMessage) == 0 {
		return errors.New("errorMessage deve ser não-vazio se fornecido")
	}

	return nil
}

// SetGenerating marca artigo como em geração
func (a *Article) SetGenerating() {
	a.Status = ArticleStatusGenerating
	a.ErrorMessage = nil
	a.UpdatedAt = time.Now()
}

// SetPublishing marca artigo como em publicação
func (a *Article) SetPublishing() {
	a.Status = ArticleStatusPublishing
	a.ErrorMessage = nil
	a.UpdatedAt = time.Now()
}

// SetPublished marca artigo como publicado com URL
func (a *Article) SetPublished(postURL string) error {
	if len(postURL) == 0 {
		return errors.New("postURL não pode estar vazio")
	}

	a.Status = ArticleStatusPublished
	a.PostURL = &postURL
	a.ErrorMessage = nil
	a.UpdatedAt = time.Now()
	return nil
}

// SetError marca artigo como erro com mensagem
func (a *Article) SetError(errMsg string) {
	a.Status = ArticleStatusError
	msg := errMsg
	a.ErrorMessage = &msg
	a.UpdatedAt = time.Now()
}

// SetContent define conteúdo do artigo
func (a *Article) SetContent(content string) error {
	if len(content) == 0 {
		return errors.New("content não pode estar vazio")
	}

	a.Content = &content
	a.UpdatedAt = time.Now()
	return nil
}

// IsPublished verifica se artigo foi publicado
func (a *Article) IsPublished() bool {
	return a.Status == ArticleStatusPublished
}

// HasError verifica se artigo tem erro
func (a *Article) HasError() bool {
	return a.Status == ArticleStatusError
}

// CanRetry verifica se artigo pode ser republicado
func (a *Article) CanRetry() bool {
	return a.Status == ArticleStatusError && a.Content != nil
}
