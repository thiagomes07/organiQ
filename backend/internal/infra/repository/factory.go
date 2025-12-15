// internal/infra/repository/factory.go
package repository

import (
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	domainrepo "organiq/internal/domain/repository"
	postgresrepo "organiq/internal/infra/repository/postgres"
)

// Re-export todas as interfaces do domain para conveniência
type (
	UserRepository         = domainrepo.UserRepository
	RefreshTokenRepository = domainrepo.RefreshTokenRepository
	PlanRepository         = domainrepo.PlanRepository
	BusinessRepository     = domainrepo.BusinessRepository
	IntegrationRepository  = domainrepo.IntegrationRepository
	ArticleJobRepository   = domainrepo.ArticleJobRepository
	ArticleIdeaRepository  = domainrepo.ArticleIdeaRepository
	ArticleRepository      = domainrepo.ArticleRepository
	PaymentRepository      = domainrepo.PaymentRepository
	PaginatedArticleResult = domainrepo.PaginatedArticleResult
)

// RepositoryContainer agrupa todas as instâncias de repository
// Usa as interfaces do domain/repository
type RepositoryContainer struct {
	User         domainrepo.UserRepository
	RefreshToken domainrepo.RefreshTokenRepository
	Plan         domainrepo.PlanRepository
	Business     domainrepo.BusinessRepository
	Integration  domainrepo.IntegrationRepository
	ArticleJob   domainrepo.ArticleJobRepository
	ArticleIdea  domainrepo.ArticleIdeaRepository
	Article      domainrepo.ArticleRepository
	Payment      domainrepo.PaymentRepository
}

// NewRepositoryContainer cria container com todas as implementações PostgreSQL
func NewRepositoryContainer(db *gorm.DB) *RepositoryContainer {
	log.Info().Msg("Inicializando RepositoryContainer")

	container := &RepositoryContainer{
		User:         postgresrepo.NewUserRepository(db),
		RefreshToken: postgresrepo.NewRefreshTokenRepository(db),
		Plan:         postgresrepo.NewPlanRepository(db),
		Business:     postgresrepo.NewBusinessRepository(db),
		Integration:  postgresrepo.NewIntegrationRepository(db),
		ArticleJob:   postgresrepo.NewArticleJobRepository(db),
		ArticleIdea:  postgresrepo.NewArticleIdeaRepository(db),
		Article:      postgresrepo.NewArticleRepository(db),
		Payment:      postgresrepo.NewPaymentRepository(db),
	}

	log.Info().Msg("RepositoryContainer inicializado com sucesso")
	return container
}
