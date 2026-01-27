// internal/usecase/article/publish_article.go
package article

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"organiq/internal/domain/entity"
	"organiq/internal/domain/repository"
	"organiq/internal/infra/wordpress"
	"organiq/internal/util"
)

// PublishArticleInput dados de entrada
type PublishArticleInput struct {
	UserID    string
	ArticleID string
}

// PublishArticleOutput dados de saída
type PublishArticleOutput struct {
	ArticleID string
	Status    string
	PostURL   string
}

// PublishArticleUseCase implementa o caso de uso de publicação manual
type PublishArticleUseCase struct {
	articleRepo     repository.ArticleRepository
	integrationRepo repository.IntegrationRepository
	cryptoService   *util.CryptoService
}

// NewPublishArticleUseCase cria nova instância
func NewPublishArticleUseCase(
	articleRepo repository.ArticleRepository,
	integrationRepo repository.IntegrationRepository,
	cryptoService *util.CryptoService,
) *PublishArticleUseCase {
	return &PublishArticleUseCase{
		articleRepo:     articleRepo,
		integrationRepo: integrationRepo,
		cryptoService:   cryptoService,
	}
}

// Execute executa a publicação
func (uc *PublishArticleUseCase) Execute(ctx context.Context, input PublishArticleInput) (*PublishArticleOutput, error) {
	log.Debug().
		Str("user_id", input.UserID).
		Str("article_id", input.ArticleID).
		Msg("PublishArticleUseCase Execute iniciado")

	// 1. Validar IDs
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, errors.New("invalid_user_id")
	}

	articleID, err := uuid.Parse(input.ArticleID)
	if err != nil {
		return nil, errors.New("invalid_article_id")
	}

	// 2. Buscar e validar artigo
	article, err := uc.articleRepo.FindByID(ctx, articleID)
	if err != nil {
		log.Error().Err(err).Msg("PublishArticleUseCase: erro ao buscar artigo")
		return nil, errors.New("erro ao buscar artigo")
	}

	if article == nil {
		return nil, errors.New("article_not_found")
	}

	if article.UserID != userID {
		return nil, errors.New("access_denied")
	}

	// 3. Validar status (deve ser Generated ou Error)
	// Permitimos retry de erro, e publication de generated
	if article.Status != entity.ArticleStatusGenerated && article.Status != entity.ArticleStatusError {
		return nil, errors.New("artigo não está pronto para publicação (status incorreto)")
	}

	if article.Content == nil || *article.Content == "" {
		return nil, errors.New("artigo sem conteúdo para publicar")
	}

	// 4. Buscar integração WordPress
	wpIntegration, err := uc.integrationRepo.FindEnabledByUserIDAndType(
		ctx,
		userID,
		entity.IntegrationTypeWordPress,
	)

	if err != nil {
		log.Error().Err(err).Msg("PublishArticleUseCase: erro ao buscar integração")
		return nil, errors.New("erro ao buscar integração")
	}

	if wpIntegration == nil {
		return nil, errors.New("integração WordPress não configurada")
	}

	// 5. Atualizar status para Publishing
	if err := uc.articleRepo.UpdateStatus(ctx, article.ID, entity.ArticleStatusPublishing); err != nil {
		log.Error().Err(err).Msg("PublishArticleUseCase: erro ao atualizar status")
		return nil, errors.New("erro ao preparar publicação")
	}

	// 6. Publicar no WP (Lógica movida do worker)
	wpConfig, err := wpIntegration.GetWordPressConfig()
	if err != nil {
		uc.handleError(ctx, article.ID, "erro na config do wordpress: "+err.Error())
		return nil, errors.New("erro na configuração do WordPress")
	}

	decryptedPassword, err := uc.cryptoService.DecryptAES(wpConfig.AppPassword)
	if err != nil {
		uc.handleError(ctx, article.ID, "erro ao descriptografar senha")
		return nil, errors.New("erro de segurança ao acessar credenciais")
	}

	htmlContent := markdownToHTML(*article.Content)

	wpClient := wordpress.NewClient(wpConfig.SiteURL, wpConfig.Username, decryptedPassword)

	wpPost := &wordpress.Post{
		Title:   article.Title,
		Content: htmlContent,
		Status:  "publish",
	}

	wpResponse, err := wpClient.CreatePost(ctx, wpPost)
	if err != nil {
		log.Error().
			Err(err).
			Str("article_id", article.ID.String()).
			Msg("PublishArticleUseCase: erro ao publicar no WP")
		
		uc.handleError(ctx, article.ID, "erro ao publicar no WordPress: "+err.Error())
		return nil, errors.New("falha ao comunicar com o WordPress: " + err.Error())
	}

	// 7. Sucesso: Atualizar status e URL
	if err := uc.articleRepo.SetPublished(ctx, article.ID, wpResponse.Link); err != nil {
		log.Error().Err(err).Msg("PublishArticleUseCase: erro ao salvar sucesso")
		// Não retornamos erro aqui pois já foi publicado
	}

	return &PublishArticleOutput{
		ArticleID: article.ID.String(),
		Status:    string(entity.ArticleStatusPublished),
		PostURL:   wpResponse.Link,
	}, nil
}

func (uc *PublishArticleUseCase) handleError(ctx context.Context, articleID uuid.UUID, msg string) {
	_ = uc.articleRepo.UpdateStatusWithError(ctx, articleID, msg)
}

// markdownToHTML (duplicado do helper, idealmente estaria em utils/transform)
func markdownToHTML(markdown string) string {
	html := markdown
	html = strings.ReplaceAll(html, "### ", "<h3>")
	html = strings.ReplaceAll(html, "## ", "<h2>")
	html = strings.ReplaceAll(html, "# ", "<h1>")
	html = strings.ReplaceAll(html, "**", "<strong>")
	html = strings.ReplaceAll(html, "*", "<em>")
	html = strings.ReplaceAll(html, "\n\n", "</p><p>")
	html = "<p>" + html + "</p>"
	return html
}
