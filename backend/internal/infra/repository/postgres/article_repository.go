// internal/infra/repository/postgres/article_repository.go
package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"organiq/internal/domain/entity"
	"organiq/internal/domain/repository"
)

// ============================================
// ARTICLE JOB REPOSITORY
// ============================================

type ArticleJobRepositoryPostgres struct {
	db *gorm.DB
}

func NewArticleJobRepository(db *gorm.DB) repository.ArticleJobRepository {
	return &ArticleJobRepositoryPostgres{db: db}
}

// Create implementa repository.Create
func (r *ArticleJobRepositoryPostgres) Create(ctx context.Context, job *entity.ArticleJob) error {
	if err := job.Validate(); err != nil {
		log.Error().Err(err).Msg("ArticleJobRepository Create: validação falhou")
		return err
	}

	log.Debug().Str("job_id", job.ID.String()).Str("type", string(job.Type)).Msg("ArticleJobRepository Create")

	if err := r.db.WithContext(ctx).Create(job).Error; err != nil {
		log.Error().Err(err).Msg("ArticleJobRepository Create erro no banco")
		return err
	}

	return nil
}

// FindByID implementa repository.FindByID
func (r *ArticleJobRepositoryPostgres) FindByID(ctx context.Context, id uuid.UUID) (*entity.ArticleJob, error) {
	var job entity.ArticleJob

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&job).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		log.Error().Err(err).Msg("ArticleJobRepository FindByID erro no banco")
		return nil, err
	}

	return &job, nil
}

// FindByUserID implementa repository.FindByUserID
func (r *ArticleJobRepositoryPostgres) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.ArticleJob, error) {
	var jobs []*entity.ArticleJob

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&jobs).
		Error

	if err != nil {
		log.Error().Err(err).Msg("ArticleJobRepository FindByUserID erro no banco")
		return nil, err
	}

	return jobs, nil
}

// FindByUserIDAndType implementa repository.FindByUserIDAndType
func (r *ArticleJobRepositoryPostgres) FindByUserIDAndType(ctx context.Context, userID uuid.UUID, jobType entity.JobType) ([]*entity.ArticleJob, error) {
	var jobs []*entity.ArticleJob

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, jobType).
		Order("created_at DESC").
		Find(&jobs).
		Error

	if err != nil {
		log.Error().Err(err).Msg("ArticleJobRepository FindByUserIDAndType erro no banco")
		return nil, err
	}

	return jobs, nil
}

// FindByStatus implementa repository.FindByStatus
func (r *ArticleJobRepositoryPostgres) FindByStatus(ctx context.Context, status entity.JobStatus) ([]*entity.ArticleJob, error) {
	var jobs []*entity.ArticleJob

	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at ASC").
		Find(&jobs).
		Error

	if err != nil {
		log.Error().Err(err).Msg("ArticleJobRepository FindByStatus erro no banco")
		return nil, err
	}

	return jobs, nil
}

// FindPendingJobs implementa repository.FindPendingJobs
func (r *ArticleJobRepositoryPostgres) FindPendingJobs(ctx context.Context) ([]*entity.ArticleJob, error) {
	var jobs []*entity.ArticleJob

	err := r.db.WithContext(ctx).
		Where("status IN ?", []entity.JobStatus{entity.JobStatusQueued, entity.JobStatusProcessing}).
		Order("created_at ASC").
		Find(&jobs).
		Error

	if err != nil {
		log.Error().Err(err).Msg("ArticleJobRepository FindPendingJobs erro no banco")
		return nil, err
	}

	return jobs, nil
}

// Update implementa repository.Update
func (r *ArticleJobRepositoryPostgres) Update(ctx context.Context, job *entity.ArticleJob) error {
	if err := job.Validate(); err != nil {
		log.Error().Err(err).Msg("ArticleJobRepository Update: validação falhou")
		return err
	}

	log.Debug().Str("job_id", job.ID.String()).Msg("ArticleJobRepository Update")

	if err := r.db.WithContext(ctx).Save(job).Error; err != nil {
		log.Error().Err(err).Msg("ArticleJobRepository Update erro no banco")
		return err
	}

	return nil
}

// Delete implementa repository.Delete
func (r *ArticleJobRepositoryPostgres) Delete(ctx context.Context, id uuid.UUID) error {
	log.Debug().Str("job_id", id.String()).Msg("ArticleJobRepository Delete")

	if err := r.db.WithContext(ctx).Delete(&entity.ArticleJob{}, "id = ?", id).Error; err != nil {
		log.Error().Err(err).Msg("ArticleJobRepository Delete erro no banco")
		return err
	}

	return nil
}

// UpdateStatus implementa repository.UpdateStatus
func (r *ArticleJobRepositoryPostgres) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.JobStatus, progress int) error {
	log.Debug().Str("job_id", id.String()).Str("status", string(status)).Int("progress", progress).Msg("ArticleJobRepository UpdateStatus")

	if err := r.db.WithContext(ctx).
		Model(&entity.ArticleJob{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"status": status, "progress": progress}).
		Error; err != nil {

		log.Error().Err(err).Msg("ArticleJobRepository UpdateStatus erro no banco")
		return err
	}

	return nil
}

// UpdateError implementa repository.UpdateError
func (r *ArticleJobRepositoryPostgres) UpdateError(ctx context.Context, id uuid.UUID, errorMsg string) error {
	log.Debug().Str("job_id", id.String()).Msg("ArticleJobRepository UpdateError")

	if err := r.db.WithContext(ctx).
		Model(&entity.ArticleJob{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"status": entity.JobStatusFailed, "error_message": errorMsg}).
		Error; err != nil {

		log.Error().Err(err).Msg("ArticleJobRepository UpdateError erro no banco")
		return err
	}

	return nil
}

// CountByUserID implementa repository.CountByUserID
func (r *ArticleJobRepositoryPostgres) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Table("article_jobs").
		Where("user_id = ?", userID).
		Count(&count).
		Error; err != nil {

		log.Error().Err(err).Msg("ArticleJobRepository CountByUserID erro no banco")
		return 0, err
	}

	return int(count), nil
}

// CountByUserIDAndType implementa repository.CountByUserIDAndType
func (r *ArticleJobRepositoryPostgres) CountByUserIDAndType(ctx context.Context, userID uuid.UUID, jobType entity.JobType) (int, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Table("article_jobs").
		Where("user_id = ? AND type = ?", userID, jobType).
		Count(&count).
		Error; err != nil {

		log.Error().Err(err).Msg("ArticleJobRepository CountByUserIDAndType erro no banco")
		return 0, err
	}

	return int(count), nil
}

// ============================================
// ARTICLE IDEA REPOSITORY
// ============================================

type ArticleIdeaRepositoryPostgres struct {
	db *gorm.DB
}

func NewArticleIdeaRepository(db *gorm.DB) repository.ArticleIdeaRepository {
	return &ArticleIdeaRepositoryPostgres{db: db}
}

// Create implementa repository.Create
func (r *ArticleIdeaRepositoryPostgres) Create(ctx context.Context, idea *entity.ArticleIdea) error {
	if err := idea.Validate(); err != nil {
		log.Error().Err(err).Msg("ArticleIdeaRepository Create: validação falhou")
		return err
	}

	log.Debug().Str("idea_id", idea.ID.String()).Msg("ArticleIdeaRepository Create")

	if err := r.db.WithContext(ctx).Create(idea).Error; err != nil {
		log.Error().Err(err).Msg("ArticleIdeaRepository Create erro no banco")
		return err
	}

	return nil
}

// CreateBatch implementa repository.CreateBatch
func (r *ArticleIdeaRepositoryPostgres) CreateBatch(ctx context.Context, ideas []*entity.ArticleIdea) error {
	if len(ideas) == 0 {
		return errors.New("ideas não pode estar vazio")
	}

	log.Debug().Int("count", len(ideas)).Msg("ArticleIdeaRepository CreateBatch")

	if err := r.db.WithContext(ctx).CreateInBatches(ideas, 10).Error; err != nil {
		log.Error().Err(err).Msg("ArticleIdeaRepository CreateBatch erro no banco")
		return err
	}

	return nil
}

// FindByID implementa repository.FindByID
func (r *ArticleIdeaRepositoryPostgres) FindByID(ctx context.Context, id uuid.UUID) (*entity.ArticleIdea, error) {
	var idea entity.ArticleIdea

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&idea).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		log.Error().Err(err).Msg("ArticleIdeaRepository FindByID erro no banco")
		return nil, err
	}

	return &idea, nil
}

// FindByJobID implementa repository.FindByJobID
func (r *ArticleIdeaRepositoryPostgres) FindByJobID(ctx context.Context, jobID uuid.UUID) ([]*entity.ArticleIdea, error) {
	var ideas []*entity.ArticleIdea

	err := r.db.WithContext(ctx).
		Where("job_id = ?", jobID).
		Order("created_at ASC").
		Find(&ideas).
		Error

	if err != nil {
		log.Error().Err(err).Msg("ArticleIdeaRepository FindByJobID erro no banco")
		return nil, err
	}

	return ideas, nil
}

// FindByUserID implementa repository.FindByUserID
func (r *ArticleIdeaRepositoryPostgres) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.ArticleIdea, error) {
	var ideas []*entity.ArticleIdea

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&ideas).
		Error

	if err != nil {
		log.Error().Err(err).Msg("ArticleIdeaRepository FindByUserID erro no banco")
		return nil, err
	}

	return ideas, nil
}

// FindApprovedByJobID implementa repository.FindApprovedByJobID
func (r *ArticleIdeaRepositoryPostgres) FindApprovedByJobID(ctx context.Context, jobID uuid.UUID) ([]*entity.ArticleIdea, error) {
	var ideas []*entity.ArticleIdea

	err := r.db.WithContext(ctx).
		Where("job_id = ? AND approved = ?", jobID, true).
		Order("created_at ASC").
		Find(&ideas).
		Error

	if err != nil {
		log.Error().Err(err).Msg("ArticleIdeaRepository FindApprovedByJobID erro no banco")
		return nil, err
	}

	return ideas, nil
}

// FindPendingByUserID implementa repository.FindPendingByUserID
func (r *ArticleIdeaRepositoryPostgres) FindPendingByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.ArticleIdea, error) {
	var ideas []*entity.ArticleIdea

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND approved = ?", userID, false).
		Order("created_at DESC").
		Find(&ideas).
		Error

	if err != nil {
		log.Error().Err(err).Msg("ArticleIdeaRepository FindPendingByUserID erro no banco")
		return nil, err
	}

	return ideas, nil
}

// Update implementa repository.Update
func (r *ArticleIdeaRepositoryPostgres) Update(ctx context.Context, idea *entity.ArticleIdea) error {
	if err := idea.Validate(); err != nil {
		log.Error().Err(err).Msg("ArticleIdeaRepository Update: validação falhou")
		return err
	}

	log.Debug().Str("idea_id", idea.ID.String()).Msg("ArticleIdeaRepository Update")

	if err := r.db.WithContext(ctx).Save(idea).Error; err != nil {
		log.Error().Err(err).Msg("ArticleIdeaRepository Update erro no banco")
		return err
	}

	return nil
}

// Delete implementa repository.Delete
func (r *ArticleIdeaRepositoryPostgres) Delete(ctx context.Context, id uuid.UUID) error {
	log.Debug().Str("idea_id", id.String()).Msg("ArticleIdeaRepository Delete")

	if err := r.db.WithContext(ctx).Delete(&entity.ArticleIdea{}, "id = ?", id).Error; err != nil {
		log.Error().Err(err).Msg("ArticleIdeaRepository Delete erro no banco")
		return err
	}

	return nil
}

// DeleteByJobID implementa repository.DeleteByJobID
func (r *ArticleIdeaRepositoryPostgres) DeleteByJobID(ctx context.Context, jobID uuid.UUID) error {
	log.Debug().Str("job_id", jobID.String()).Msg("ArticleIdeaRepository DeleteByJobID")

	if err := r.db.WithContext(ctx).Delete(&entity.ArticleIdea{}, "job_id = ?", jobID).Error; err != nil {
		log.Error().Err(err).Msg("ArticleIdeaRepository DeleteByJobID erro no banco")
		return err
	}

	return nil
}

// ApproveMultiple implementa repository.ApproveMultiple
func (r *ArticleIdeaRepositoryPostgres) ApproveMultiple(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return errors.New("ids não pode estar vazio")
	}

	log.Debug().Int("count", len(ids)).Msg("ArticleIdeaRepository ApproveMultiple")

	if err := r.db.WithContext(ctx).
		Model(&entity.ArticleIdea{}).
		Where("id IN ?", ids).
		Update("approved", true).
		Error; err != nil {

		log.Error().Err(err).Msg("ArticleIdeaRepository ApproveMultiple erro no banco")
		return err
	}

	return nil
}

// CountApprovedByJobID implementa repository.CountApprovedByJobID
func (r *ArticleIdeaRepositoryPostgres) CountApprovedByJobID(ctx context.Context, jobID uuid.UUID) (int, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Table("article_ideas").
		Where("job_id = ? AND approved = ?", jobID, true).
		Count(&count).
		Error; err != nil {

		log.Error().Err(err).Msg("ArticleIdeaRepository CountApprovedByJobID erro no banco")
		return 0, err
	}

	return int(count), nil
}

// CountByUserID implementa repository.CountByUserID
func (r *ArticleIdeaRepositoryPostgres) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Table("article_ideas").
		Where("user_id = ?", userID).
		Count(&count).
		Error; err != nil {

		log.Error().Err(err).Msg("ArticleIdeaRepository CountByUserID erro no banco")
		return 0, err
	}

	return int(count), nil
}

// DeleteUnapprovedByUserID implementa repository.DeleteUnapprovedByUserID
func (r *ArticleIdeaRepositoryPostgres) DeleteUnapprovedByUserID(ctx context.Context, userID uuid.UUID) error {
	log.Debug().Str("user_id", userID.String()).Msg("ArticleIdeaRepository DeleteUnapprovedByUserID")

	if err := r.db.WithContext(ctx).Delete(&entity.ArticleIdea{}, "user_id = ? AND approved = ?", userID, false).Error; err != nil {
		log.Error().Err(err).Msg("ArticleIdeaRepository DeleteUnapprovedByUserID erro no banco")
		return err
	}

	return nil
}

// CountApprovedByUserID implementa repository.CountApprovedByUserID
func (r *ArticleIdeaRepositoryPostgres) CountApprovedByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Table("article_ideas").
		Where("user_id = ? AND approved = ?", userID, true).
		Count(&count).
		Error; err != nil {

		log.Error().Err(err).Msg("ArticleIdeaRepository CountApprovedByUserID erro no banco")
		return 0, err
	}

	return int(count), nil
}

// CountGenerationsInLastHour implementa repository.CountGenerationsInLastHour
// Conta quantos jobs de geração de ideias foram criados na última hora
func (r *ArticleIdeaRepositoryPostgres) CountGenerationsInLastHour(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int64

	// Conta ideias distintas por generated_at (agrupadas por tempo de geração)
	// Uma "geração" é um grupo de ideias com generated_at muito próximos
	if err := r.db.WithContext(ctx).
		Table("article_ideas").
		Where("user_id = ? AND generated_at > NOW() - INTERVAL '1 hour'", userID).
		Select("COUNT(DISTINCT DATE_TRUNC('minute', generated_at))").
		Scan(&count).
		Error; err != nil {

		log.Error().Err(err).Msg("ArticleIdeaRepository CountGenerationsInLastHour erro no banco")
		return 0, err
	}

	return int(count), nil
}

// ============================================
// ARTICLE REPOSITORY
// ============================================

type ArticleRepositoryPostgres struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) repository.ArticleRepository {
	return &ArticleRepositoryPostgres{db: db}
}

// Create implementa repository.Create
func (r *ArticleRepositoryPostgres) Create(ctx context.Context, article *entity.Article) error {
	if err := article.Validate(); err != nil {
		log.Error().Err(err).Msg("ArticleRepository Create: validação falhou")
		return err
	}

	log.Debug().Str("article_id", article.ID.String()).Msg("ArticleRepository Create")

	if err := r.db.WithContext(ctx).Create(article).Error; err != nil {
		log.Error().Err(err).Msg("ArticleRepository Create erro no banco")
		return err
	}

	return nil
}

// FindByID implementa repository.FindByID
func (r *ArticleRepositoryPostgres) FindByID(ctx context.Context, id uuid.UUID) (*entity.Article, error) {
	var article entity.Article

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&article).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		log.Error().Err(err).Msg("ArticleRepository FindByID erro no banco")
		return nil, err
	}

	return &article, nil
}

// FindByUserID implementa repository.FindByUserID
func (r *ArticleRepositoryPostgres) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Article, error) {
	var articles []*entity.Article

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&articles).
		Error

	if err != nil {
		log.Error().Err(err).Msg("ArticleRepository FindByUserID erro no banco")
		return nil, err
	}

	return articles, nil
}

// FindByUserIDAndStatus implementa repository.FindByUserIDAndStatus
func (r *ArticleRepositoryPostgres) FindByUserIDAndStatus(ctx context.Context, userID uuid.UUID, status entity.ArticleStatus, limit, offset int) ([]*entity.Article, error) {
	var articles []*entity.Article

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&articles).
		Error

	if err != nil {
		log.Error().Err(err).Msg("ArticleRepository FindByUserIDAndStatus erro no banco")
		return nil, err
	}

	return articles, nil
}

// FindByIdeaID implementa repository.FindByIdeaID
func (r *ArticleRepositoryPostgres) FindByIdeaID(ctx context.Context, ideaID uuid.UUID) (*entity.Article, error) {
	var article entity.Article

	err := r.db.WithContext(ctx).Where("idea_id = ?", ideaID).First(&article).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		log.Error().Err(err).Msg("ArticleRepository FindByIdeaID erro no banco")
		return nil, err
	}

	return &article, nil
}

// FindPublishedByUserID implementa repository.FindPublishedByUserID
func (r *ArticleRepositoryPostgres) FindPublishedByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Article, error) {
	var articles []*entity.Article

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, entity.ArticleStatusPublished).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&articles).
		Error

	if err != nil {
		log.Error().Err(err).Msg("ArticleRepository FindPublishedByUserID erro no banco")
		return nil, err
	}

	return articles, nil
}

// FindErrorsByUserID implementa repository.FindErrorsByUserID
func (r *ArticleRepositoryPostgres) FindErrorsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Article, error) {
	var articles []*entity.Article

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, entity.ArticleStatusError).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&articles).
		Error

	if err != nil {
		log.Error().Err(err).Msg("ArticleRepository FindErrorsByUserID erro no banco")
		return nil, err
	}

	return articles, nil
}

// FindByStatus implementa repository.FindByStatus
func (r *ArticleRepositoryPostgres) FindByStatus(ctx context.Context, status entity.ArticleStatus) ([]*entity.Article, error) {
	var articles []*entity.Article

	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at ASC").
		Find(&articles).
		Error

	if err != nil {
		log.Error().Err(err).Msg("ArticleRepository FindByStatus erro no banco")
		return nil, err
	}

	return articles, nil
}

// Update implementa repository.Update
func (r *ArticleRepositoryPostgres) Update(ctx context.Context, article *entity.Article) error {
	if err := article.Validate(); err != nil {
		log.Error().Err(err).Msg("ArticleRepository Update: validação falhou")
		return err
	}

	log.Debug().Str("article_id", article.ID.String()).Msg("ArticleRepository Update")

	if err := r.db.WithContext(ctx).Save(article).Error; err != nil {
		log.Error().Err(err).Msg("ArticleRepository Update erro no banco")
		return err
	}

	return nil
}

// Delete implementa repository.Delete
func (r *ArticleRepositoryPostgres) Delete(ctx context.Context, id uuid.UUID) error {
	log.Debug().Str("article_id", id.String()).Msg("ArticleRepository Delete")

	if err := r.db.WithContext(ctx).Delete(&entity.Article{}, "id = ?", id).Error; err != nil {
		log.Error().Err(err).Msg("ArticleRepository Delete erro no banco")
		return err
	}

	return nil
}

// UpdateStatus implementa repository.UpdateStatus
func (r *ArticleRepositoryPostgres) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ArticleStatus) error {
	log.Debug().Str("article_id", id.String()).Str("status", string(status)).Msg("ArticleRepository UpdateStatus")

	if err := r.db.WithContext(ctx).
		Model(&entity.Article{}).
		Where("id = ?", id).
		Update("status", status).
		Error; err != nil {

		log.Error().Err(err).Msg("ArticleRepository UpdateStatus erro no banco")
		return err
	}

	return nil
}

// UpdateStatusWithError implementa repository.UpdateStatusWithError
func (r *ArticleRepositoryPostgres) UpdateStatusWithError(ctx context.Context, id uuid.UUID, errMsg string) error {
	log.Debug().Str("article_id", id.String()).Msg("ArticleRepository UpdateStatusWithError")

	if err := r.db.WithContext(ctx).
		Model(&entity.Article{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"status": entity.ArticleStatusError, "error_message": errMsg}).
		Error; err != nil {

		log.Error().Err(err).Msg("ArticleRepository UpdateStatusWithError erro no banco")
		return err
	}

	return nil
}

// SetPublished implementa repository.SetPublished
func (r *ArticleRepositoryPostgres) SetPublished(ctx context.Context, id uuid.UUID, postURL string) error {
	log.Debug().Str("article_id", id.String()).Str("post_url", postURL).Msg("ArticleRepository SetPublished")

	if err := r.db.WithContext(ctx).
		Model(&entity.Article{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":    entity.ArticleStatusPublished,
			"post_url":  postURL,
			"error_message": nil,
		}).
		Error; err != nil {

		log.Error().Err(err).Msg("ArticleRepository SetPublished erro no banco")
		return err
	}

	return nil
}

// SetContent implementa repository.SetContent
func (r *ArticleRepositoryPostgres) SetContent(ctx context.Context, id uuid.UUID, content string) error {
	log.Debug().Str("article_id", id.String()).Msg("ArticleRepository SetContent")

	if err := r.db.WithContext(ctx).
		Model(&entity.Article{}).
		Where("id = ?", id).
		Update("content", content).
		Error; err != nil {

		log.Error().Err(err).Msg("ArticleRepository SetContent erro no banco")
		return err
	}

	return nil
}

// CountByUserID implementa repository.CountByUserID
func (r *ArticleRepositoryPostgres) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Table("articles").
		Where("user_id = ?", userID).
		Count(&count).
		Error; err != nil {

		log.Error().Err(err).Msg("ArticleRepository CountByUserID erro no banco")
		return 0, err
	}

	return int(count), nil
}

// CountPublishedByUserID implementa repository.CountPublishedByUserID
func (r *ArticleRepositoryPostgres) CountPublishedByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Table("articles").
		Where("user_id = ? AND status = ?", userID, entity.ArticleStatusPublished).
		Count(&count).
		Error; err != nil {

		log.Error().Err(err).Msg("ArticleRepository CountPublishedByUserID erro no banco")
		return 0, err
	}

	return int(count), nil
}

// CountByUserIDAndStatus implementa repository.CountByUserIDAndStatus
func (r *ArticleRepositoryPostgres) CountByUserIDAndStatus(ctx context.Context, userID uuid.UUID, status entity.ArticleStatus) (int, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Table("articles").
		Where("user_id = ? AND status = ?", userID, status).
		Count(&count).
		Error; err != nil {

		log.Error().Err(err).Msg("ArticleRepository CountByUserIDAndStatus erro no banco")
		return 0, err
	}

	return int(count), nil
}

// CountErrorsByUserID implementa repository.CountErrorsByUserID
func (r *ArticleRepositoryPostgres) CountErrorsByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Table("articles").
		Where("user_id = ? AND status = ?", userID, entity.ArticleStatusError).
		Count(&count).
		Error; err != nil {

		log.Error().Err(err).Msg("ArticleRepository CountErrorsByUserID erro no banco")
		return 0, err
	}

	return int(count), nil
}
