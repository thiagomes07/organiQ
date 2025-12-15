// internal/infra/ai/agent_client.go
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"organiq/config"
)

// AgentClient cliente HTTP para chamar agente próprio ou LLMs externos
type AgentClient struct {
	client      *http.Client
	provider    string
	model       string
	apiKey      string
	baseURL     string
	maxTokens   int
	temperature float64
}

// NewAgentClient cria nova instância do cliente
func NewAgentClient(cfg *config.Config) *AgentClient {
	return &AgentClient{
		client: &http.Client{
			Timeout: 120 * time.Second, // Timeout de 2 minutos para operações de IA
		},
		provider:    cfg.AI.Provider,
		model:       cfg.AI.Model,
		apiKey:      cfg.AI.APIKey,
		maxTokens:   cfg.AI.MaxTokens,
		temperature: cfg.AI.Temperature,
		baseURL:     "https://api.openai.com/v1", // Ajustar para provider
	}
}

// GenerateIdeas gera ideias de artigos usando prompt customizado
func (c *AgentClient) GenerateIdeas(
	ctx context.Context,
	businessInfo string,
	competitors []string,
	articleCount int,
	objectives string,
	location string,
) ([]string, error) {

	log.Debug().
		Int("articleCount", articleCount).
		Msg("AI GenerateIdeas iniciado")

	// Construir prompt
	prompt := c.buildIdeasPrompt(businessInfo, competitors, articleCount, objectives, location)

	response, err := c.callLLM(ctx, prompt)
	if err != nil {
		log.Error().Err(err).Msg("AI GenerateIdeas falhou")
		return nil, err
	}

	// Parse response (assume JSON com array de ideias)
	ideas := parseIdeasResponse(response)

	log.Info().Int("count", len(ideas)).Msg("AI GenerateIdeas bem-sucedido")
	return ideas, nil
}

// GenerateArticle gera conteúdo de artigo usando IA
func (c *AgentClient) GenerateArticle(
	ctx context.Context,
	title string,
	summary string,
	businessInfo string,
	objectives string,
	location string,
	feedback *string,
	brandTone *string,
) (string, error) {

	log.Debug().
		Str("title", title).
		Msg("AI GenerateArticle iniciado")

	// Construir prompt para gerar artigo
	prompt := c.buildArticlePrompt(title, summary, businessInfo, objectives, location, feedback, brandTone)

	response, err := c.callLLM(ctx, prompt)
	if err != nil {
		log.Error().Err(err).Msg("AI GenerateArticle falhou")
		return "", err
	}

	log.Info().
		Str("title", title).
		Int("contentLength", len(response)).
		Msg("AI GenerateArticle bem-sucedido")

	return response, nil
}

// GenerateMetadata gera meta tags e SEO para artigo
func (c *AgentClient) GenerateMetadata(
	ctx context.Context,
	title string,
	content string,
) (map[string]string, error) {

	log.Debug().Str("title", title).Msg("AI GenerateMetadata iniciado")

	prompt := fmt.Sprintf(`
Analise o título e conteúdo do artigo e gere metadados SEO em JSON.

Título: %s

Conteúdo (primeiros 500 chars):
%s

Retorne um JSON com os seguintes campos:
{
  "metaDescription": "Descrição meta (150-160 caracteres)",
  "keywords": "3-5 palavras-chave separadas por vírgula",
  "mainKeyword": "Palavra-chave principal",
  "slug": "url-slug-do-artigo"
}

Responda APENAS com o JSON, sem explicações.
`, title, content[:minInt(500, len(content))])

	response, err := c.callLLM(ctx, prompt)
	if err != nil {
		log.Error().Err(err).Msg("AI GenerateMetadata falhou")
		return nil, err
	}

	metadata := parseMetadataResponse(response)

	log.Info().Msg("AI GenerateMetadata bem-sucedido")
	return metadata, nil
}

// ============================================
// PRIVATE METHODS
// ============================================

// callLLM faz chamada HTTP para LLM
func (c *AgentClient) callLLM(ctx context.Context, prompt string) (string, error) {
	// Construir request body (padrão OpenAI)
	reqBody := map[string]interface{}{
		"model":       c.model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":   c.maxTokens,
		"temperature": c.temperature,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao serializar request para LLM")
		return "", err
	}

	log.Debug().
		Str("provider", c.provider).
		Str("model", c.model).
		Msg("Chamando LLM")

	// Criar HTTP request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/chat/completions", c.baseURL),
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao criar request para LLM")
		return "", err
	}

	// Headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	// Executar request
	resp, err := c.client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao fazer chamada para LLM")
		return "", err
	}
	defer resp.Body.Close()

	// Validar status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Error().
			Int("statusCode", resp.StatusCode).
			Str("body", string(body)).
			Msg("LLM retornou erro")
		return "", fmt.Errorf("LLM error: status %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Error().Err(err).Msg("Erro ao fazer parse do response do LLM")
		return "", err
	}

	// Extrair conteúdo da resposta
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		log.Error().Msg("LLM response inválido: choices não encontrado")
		return "", fmt.Errorf("invalid LLM response structure")
	}

	firstChoice, ok := choices[0].(map[string]interface{})
	if !ok {
		log.Error().Msg("LLM response inválido: primeira choice é inválida")
		return "", fmt.Errorf("invalid LLM response structure")
	}

	message, ok := firstChoice["message"].(map[string]interface{})
	if !ok {
		log.Error().Msg("LLM response inválido: message não encontrado")
		return "", fmt.Errorf("invalid LLM response structure")
	}

	content, ok := message["content"].(string)
	if !ok {
		log.Error().Msg("LLM response inválido: content não é string")
		return "", fmt.Errorf("invalid LLM response structure")
	}

	return content, nil
}

// buildIdeasPrompt constrói prompt para geração de ideias
func (c *AgentClient) buildIdeasPrompt(
	businessInfo string,
	competitors []string,
	articleCount int,
	objectives string,
	location string,
) string {
	competitorStr := ""
	for i, comp := range competitors {
		competitorStr += fmt.Sprintf("%d. %s\n", i+1, comp)
	}

	return fmt.Sprintf(`
Você é um especialista em criação de conteúdo e SEO para empresas locais e pequenos negócios.

SOBRE O NEGÓCIO:
%s

LOCALIZAÇÃO:
%s

OBJETIVOS:
%s

CONCORRENTES:
%s

Gere %d ideias de artigos de blog que:
1. Sejam relevantes para o nicho e localização
2. Tenham alto potencial de SEO local
3. Resolvam problemas comuns dos clientes
4. Sejam diferenciadas das competição

Para cada ideia, forneça:
- Título (80 caracteres máximo)
- Resumo (200 caracteres)

Responda em JSON com formato:
{
  "ideas": [
    {"title": "...", "summary": "..."},
    ...
  ]
}

Responda APENAS com o JSON, sem explicações.
`, businessInfo, location, objectives, competitorStr, articleCount)
}

// buildArticlePrompt constrói prompt para geração de artigo
func (c *AgentClient) buildArticlePrompt(
	title string,
	summary string,
	businessInfo string,
	objectives string,
	location string,
	feedback *string,
	brandTone *string,
) string {
	feedbackStr := ""
	if feedback != nil && len(*feedback) > 0 {
		feedbackStr = fmt.Sprintf("\nFEEDBACK DO CLIENTE:\n%s", *feedback)
	}

	toneStr := "Profissional e amigável"
	if brandTone != nil && len(*brandTone) > 0 {
		toneStr = *brandTone
	}

	return fmt.Sprintf(`
Você é um especialista em copywriting e criação de conteúdo para negócios locais.

TÍTULO DO ARTIGO:
%s

RESUMO:
%s

SOBRE O NEGÓCIO:
%s

LOCALIZAÇÃO:
%s

OBJETIVOS DO NEGÓCIO:
%s

TOM DA MARCA:
%s
%s

Escreva um artigo completo e bem estruturado com:

1. **Introdução** (2-3 parágrafos): Contextualize o problema e por que é importante
2. **Seções de Conteúdo** (3-5 seções com H2): Desenvolvimente os pontos principais
3. **Conclusão** (2-3 parágrafos): Resuma e faça call-to-action

Requisitos:
- Mínimo 1000 palavras
- Otimizado para SEO com uso natural de palavras-chave
- Prático e acionável
- Inclua exemplos reais quando aplicável
- Links internos sugeridos entre [ ]
- Markup em Markdown com # H1, ## H2, ### H3, **negrito**, *itálico*, etc.

Responda apenas com o conteúdo do artigo em Markdown, sem preâmbulos.
`, title, summary, businessInfo, location, objectives, toneStr, feedbackStr)
}

// Helper functions
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func parseIdeasResponse(response string) []string {
	// Parse JSON response
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		log.Error().Err(err).Msg("Erro ao fazer parse de ideias")
		return []string{}
	}

	ideas, ok := result["ideas"].([]interface{})
	if !ok {
		log.Error().Msg("Ideas não é array no response")
		return []string{}
	}

	var titles []string
	for _, idea := range ideas {
		if ideaMap, ok := idea.(map[string]interface{}); ok {
			if title, ok := ideaMap["title"].(string); ok {
				titles = append(titles, title)
			}
		}
	}

	return titles
}

func parseMetadataResponse(response string) map[string]string {
	var metadata map[string]string
	if err := json.Unmarshal([]byte(response), &metadata); err != nil {
		log.Error().Err(err).Msg("Erro ao fazer parse de metadata")
		return make(map[string]string)
	}
	return metadata
}
