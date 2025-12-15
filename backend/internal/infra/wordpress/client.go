// internal/infra/wordpress/client.go
package wordpress

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// Client cliente para WordPress REST API
type Client struct {
	client   *http.Client
	siteURL  string
	username string
	password string // App password ou senha regular
}

// NewClient cria nova instância do cliente WordPress
func NewClient(siteURL, username, password string) *Client {
	return &Client{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		siteURL:  siteURL,
		username: username,
		password: password,
	}
}

// TestConnection testa conexão e credenciais
func (c *Client) TestConnection(ctx context.Context) error {
	log.Debug().Msg("WordPress TestConnection iniciado")

	_, err := c.getCurrentUser(ctx)
	if err != nil {
		log.Error().Err(err).Msg("WordPress TestConnection falhou")
		return fmt.Errorf("erro ao conectar com WordPress: %w", err)
	}

	log.Info().Msg("WordPress TestConnection bem-sucedido")
	return nil
}

// CreatePost cria novo post no WordPress
func (c *Client) CreatePost(ctx context.Context, post *Post) (*PostResponse, error) {
	if err := post.Validate(); err != nil {
		log.Error().Err(err).Msg("WordPress CreatePost: post inválido")
		return nil, err
	}

	log.Debug().
		Str("title", post.Title).
		Msg("WordPress CreatePost iniciado")

	// Serializar post
	payload, err := json.Marshal(post)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao serializar post")
		return nil, err
	}

	// Fazer requisição
	respBody, statusCode, err := c.doRequest(
		ctx,
		http.MethodPost,
		"/posts",
		payload,
	)

	if err != nil {
		log.Error().Err(err).Msg("WordPress CreatePost falhou")
		return nil, err
	}

	// Validar status code
	if statusCode != http.StatusCreated {
		log.Error().
			Int("statusCode", statusCode).
			Str("body", string(respBody)).
			Msg("WordPress CreatePost retornou erro")
		return nil, fmt.Errorf("wordpress error: status %d", statusCode)
	}

	// Parse response
	var response PostResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		log.Error().Err(err).Msg("Erro ao fazer parse do response de post")
		return nil, err
	}

	log.Info().
		Str("title", post.Title).
		Int("postID", response.ID).
		Str("postURL", response.Link).
		Msg("WordPress CreatePost bem-sucedido")

	return &response, nil
}

// UpdatePost atualiza post existente
func (c *Client) UpdatePost(ctx context.Context, postID int, post *Post) (*PostResponse, error) {
	if postID <= 0 {
		log.Error().Msg("WordPress UpdatePost: postID inválido")
		return nil, fmt.Errorf("postID deve ser > 0")
	}

	log.Debug().
		Str("title", post.Title).
		Int("postID", postID).
		Msg("WordPress UpdatePost iniciado")

	payload, err := json.Marshal(post)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao serializar post")
		return nil, err
	}

	respBody, statusCode, err := c.doRequest(
		ctx,
		http.MethodPost,
		fmt.Sprintf("/posts/%d", postID),
		payload,
	)

	if err != nil {
		log.Error().Err(err).Msg("WordPress UpdatePost falhou")
		return nil, err
	}

	if statusCode != http.StatusOK {
		log.Error().
			Int("statusCode", statusCode).
			Msg("WordPress UpdatePost retornou erro")
		return nil, fmt.Errorf("wordpress error: status %d", statusCode)
	}

	var response PostResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		log.Error().Err(err).Msg("Erro ao fazer parse do response de atualização")
		return nil, err
	}

	log.Info().
		Str("title", post.Title).
		Int("postID", postID).
		Msg("WordPress UpdatePost bem-sucedido")

	return &response, nil
}

// GetPost obtém post existente
func (c *Client) GetPost(ctx context.Context, postID int) (*PostResponse, error) {
	if postID <= 0 {
		log.Error().Msg("WordPress GetPost: postID inválido")
		return nil, fmt.Errorf("postID deve ser > 0")
	}

	log.Debug().Int("postID", postID).Msg("WordPress GetPost iniciado")

	respBody, statusCode, err := c.doRequest(
		ctx,
		http.MethodGet,
		fmt.Sprintf("/posts/%d", postID),
		nil,
	)

	if err != nil {
		log.Error().Err(err).Msg("WordPress GetPost falhou")
		return nil, err
	}

	if statusCode != http.StatusOK {
		log.Error().
			Int("statusCode", statusCode).
			Msg("WordPress GetPost retornou erro")
		return nil, fmt.Errorf("wordpress error: status %d", statusCode)
	}

	var response PostResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		log.Error().Err(err).Msg("Erro ao fazer parse do post")
		return nil, err
	}

	log.Info().Int("postID", postID).Msg("WordPress GetPost bem-sucedido")
	return &response, nil
}

// DeletePost deleta post
func (c *Client) DeletePost(ctx context.Context, postID int, force bool) error {
	if postID <= 0 {
		log.Error().Msg("WordPress DeletePost: postID inválido")
		return fmt.Errorf("postID deve ser > 0")
	}

	log.Debug().Int("postID", postID).Bool("force", force).Msg("WordPress DeletePost iniciado")

	endpoint := fmt.Sprintf("/posts/%d", postID)
	if force {
		endpoint += "?force=true"
	}

	_, statusCode, err := c.doRequest(
		ctx,
		http.MethodDelete,
		endpoint,
		nil,
	)

	if err != nil {
		log.Error().Err(err).Msg("WordPress DeletePost falhou")
		return err
	}

	if statusCode != http.StatusOK {
		log.Error().
			Int("statusCode", statusCode).
			Msg("WordPress DeletePost retornou erro")
		return fmt.Errorf("wordpress error: status %d", statusCode)
	}

	log.Info().Int("postID", postID).Msg("WordPress DeletePost bem-sucedido")
	return nil
}

// ============================================
// PRIVATE METHODS
// ============================================

// doRequest faz requisição HTTP autenticada
func (c *Client) doRequest(
	ctx context.Context,
	method string,
	endpoint string,
	body []byte,
) ([]byte, int, error) {

	// Construir URL
	url := fmt.Sprintf("%s/wp-json/wp/v2%s", c.siteURL, endpoint)

	// Criar request
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}

	if err != nil {
		log.Error().Err(err).Msg("Erro ao criar request para WordPress")
		return nil, 0, err
	}

	// Headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.getAuthHeader())

	log.Debug().
		Str("method", method).
		Str("endpoint", endpoint).
		Msg("WordPress request enviado")

	// Executar request
	resp, err := c.client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao fazer request para WordPress")
		return nil, 0, err
	}
	defer resp.Body.Close()

	// Ler response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao ler response do WordPress")
		return nil, resp.StatusCode, err
	}

	return respBody, resp.StatusCode, nil
}

// getAuthHeader retorna header de autenticação (Basic Auth)
func (c *Client) getAuthHeader() string {
	credentials := fmt.Sprintf("%s:%s", c.username, c.password)
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	return fmt.Sprintf("Basic %s", encoded)
}

// getCurrentUser obtém usuário atual (teste de autenticação)
func (c *Client) getCurrentUser(ctx context.Context) (map[string]interface{}, error) {
	respBody, statusCode, err := c.doRequest(ctx, http.MethodGet, "/users/me", nil)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get current user: status %d", statusCode)
	}

	var user map[string]interface{}
	if err := json.Unmarshal(respBody, &user); err != nil {
		return nil, err
	}

	return user, nil
}

// ============================================
// DATA STRUCTURES
// ============================================

// Post representa um post do WordPress
type Post struct {
	Title       string   `json:"title"`
	Content     string   `json:"content"`
	Status      string   `json:"status"` // draft, publish, pending
	Categories  []int    `json:"categories"`
	Tags        []int    `json:"tags"`
	FeaturedMedia int   `json:"featured_media"`
	Meta        map[string]interface{} `json:"meta"`
	Excerpt     string   `json:"excerpt"`
	// Custom fields (metadados)
	MetaDescription string `json:"yoast_meta,omitempty"` // Para SEO plugins
}

// Validate valida os dados do post
func (p *Post) Validate() error {
	if len(p.Title) == 0 {
		return fmt.Errorf("title é obrigatório")
	}

	if len(p.Title) > 255 {
		return fmt.Errorf("title deve ter no máximo 255 caracteres")
	}

	if len(p.Content) == 0 {
		return fmt.Errorf("content é obrigatório")
	}

	if len(p.Content) < 100 {
		return fmt.Errorf("content deve ter no mínimo 100 caracteres")
	}

	if len(p.Status) == 0 {
		p.Status = "publish"
	}

	validStatuses := map[string]bool{
		"draft":   true,
		"publish": true,
		"pending": true,
		"private": true,
	}

	if !validStatuses[p.Status] {
		return fmt.Errorf("status inválido: %s", p.Status)
	}

	return nil
}

// PostResponse representa resposta do WordPress ao criar/atualizar post
type PostResponse struct {
	ID    int    `json:"id"`
	Title struct {
		Rendered string `json:"rendered"`
	} `json:"title"`
	Content struct {
		Rendered string `json:"rendered"`
	} `json:"content"`
	Link   string `json:"link"`
	Status string `json:"status"`
	Date   string `json:"date"`
	Modified string `json:"modified"`
	Slug   string `json:"slug"`
}
