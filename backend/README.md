# 📘 Especificação Técnica do Backend - organiQ

> **Versão:** 1.0  
> **Data:** Dezembro 2024  
> **Autor:** Arquitetura Backend  
> **Stack:** Go (Golang) + Clean Architecture

---

## 📋 Índice

1. [Visão Arquitetural](#1-visão-arquitetural)
2. [Design de Infraestrutura Plugável](#2-design-de-infraestrutura-plugável)
3. [Segurança e Autenticação](#3-segurança-e-autenticação)
4. [Especificação da API](#4-especificação-da-api)
5. [Modelagem de Dados](#5-modelagem-de-dados)
6. [Estrutura de Arquivos](#6-estrutura-de-arquivos)
7. [Considerações de Deploy](#7-considerações-de-deploy)

---

## 1. Visão Arquitetural

### 1.1 Princípios de Design

**Stateless API**
- Toda autenticação via JWT em cookies HttpOnly
- Nenhum estado compartilhado entre instâncias
- Escalabilidade horizontal sem sessions

**Clean Architecture**
- Separação estrita entre camadas
- Dependências apontam para o centro (domínio)
- Interfaces para inversão de dependência
- Testabilidade máxima

**Security First**
- Todas as rotas protegidas por padrão
- Criptografia robusta (Argon2id + Pepper)
- Cookies seguros (HttpOnly, Secure, SameSite Strict)
- Rate limiting agressivo

**Processamento Assíncrono**
- Operações pesadas (IA, publicação) em filas
- Resposta imediata ao cliente (202 Accepted)
- Workers independentes para processamento

### 1.2 Diagrama Conceitual

```
┌──────────────────────────────────────────────────────┐
│                    PRESENTATION                       │
│  ┌────────────────────────────────────────────────┐  │
│  │  HTTP Server (Chi Router)                      │  │
│  │  - CORS Middleware                             │  │
│  │  - Auth Middleware (JWT Validation)            │  │
│  │  - Rate Limit Middleware                       │  │
│  │  - Logger Middleware                           │  │
│  │  - Recovery Middleware                         │  │
│  └────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────┘
                         ↓
┌──────────────────────────────────────────────────────┐
│                    APPLICATION                        │
│  ┌────────────────────────────────────────────────┐  │
│  │  Use Cases (Business Logic)                    │  │
│  │  - RegisterUser                                │  │
│  │  - GenerateArticleIdeas                        │  │
│  │  - PublishArticles                             │  │
│  │  - ManageIntegrations                          │  │
│  └────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────┘
                         ↓
┌──────────────────────────────────────────────────────┐
│                      DOMAIN                           │
│  ┌────────────────────────────────────────────────┐  │
│  │  Entities & Interfaces (Contratos)             │  │
│  │  - User, Article, Plan, Integration            │  │
│  │  - Repository Interfaces                       │  │
│  │  - Service Interfaces                          │  │
│  └────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────┘
                         ↓
┌──────────────────────────────────────────────────────┐
│                  INFRASTRUCTURE                       │
│  ┌────────────┬──────────────┬────────────────────┐  │
│  │ Repository │  Queue       │  External Services │  │
│  │ (DB)       │  (SQS)       │  (OpenAI, WP, S3)  │  │
│  │ - Postgres │  - Workers   │  - HTTP Clients    │  │
│  │ - RDS      │  - Retry     │  - OAuth Flows     │  │
│  └────────────┴──────────────┴────────────────────┘  │
└──────────────────────────────────────────────────────┘
```

### 1.3 Stack Tecnológico

**Core Framework**
- **Go 1.22+**: Linguagem base
- **Chi Router v5**: HTTP router leve e idiomático
- **GORM v2**: ORM com suporte a migrations
- **Validator v10**: Validação de structs

**Banco de Dados**
- **PostgreSQL 16**: SGBD principal
- **pgx/v5**: Driver nativo de alta performance
- **GORM**: Abstração ORM

**Autenticação & Crypto**
- **golang-jwt/jwt v5**: Geração/validação de JWT
- **crypto/argon2**: Hashing de senhas
- **crypto/aes**: Criptografia de campos sensíveis

**Processamento Assíncrono**
- **AWS SQS**: Fila de mensagens
- **Worker Pool**: Goroutines com controle de concorrência

**Observabilidade**
- **zerolog**: Logs estruturados em JSON
- **chi/middleware**: Métricas HTTP
- **AWS X-Ray**: Distributed tracing

**Clients Externos**
- **net/http**: HTTP client nativo
- **OpenAI Go SDK**: Cliente oficial OpenAI
- **AWS SDK Go v2**: Serviços AWS (S3, SQS, Secrets Manager)

---

## 2. Design de Infraestrutura Plugável

### 2.1 Princípio de Substituição via Interfaces

**Objetivo**: Alternar entre drivers locais e de produção apenas mudando variáveis de ambiente, sem modificar código.

### 2.2 Abstração de Banco de Dados

**Interface de Repository Pattern**

Todas as operações de banco devem implementar interfaces no domínio:

```
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id uuid.UUID) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
    Update(ctx context.Context, user *User) error
}
```

**Implementações Concretas**

**Local (Desenvolvimento)**
- Arquivo: `internal/infra/repository/postgres/user_repository.go`
- Conexão: Container PostgreSQL via Docker Compose
- DSN: `postgres://organiq:dev_password@localhost:5432/organiq_dev?sslmode=disable`

**Produção**
- Arquivo: Mesma implementação (`user_repository.go`)
- Conexão: AWS RDS PostgreSQL
- DSN: `postgres://user:pass@organiq-prod.us-east-1.rds.amazonaws.com:5432/organiq?sslmode=require`

**Factory Pattern para Inicialização**

Arquivo: `internal/infra/repository/factory.go`

Lógica:
1. Ler variável `DB_HOST` do ambiente
2. Se `localhost` → SSL disabled
3. Se domínio AWS → SSL required
4. Retornar implementação Postgres via interface

### 2.3 Abstração de Storage (Blob)

**Interface de Storage**

```
type StorageService interface {
    Upload(ctx context.Context, key string, data io.Reader, contentType string) (url string, error)
    Download(ctx context.Context, key string) (io.ReadCloser, error)
    Delete(ctx context.Context, key string) error
}
```

**Implementações Concretas**

**Local (Desenvolvimento)**
- Arquivo: `internal/infra/storage/minio_storage.go`
- Container: MinIO via Docker Compose
- Endpoint: `http://localhost:9000`
- Bucket: `organiq-dev-brand-files`

**Produção**
- Arquivo: `internal/infra/storage/s3_storage.go`
- Serviço: AWS S3
- Endpoint: `https://s3.us-east-1.amazonaws.com`
- Bucket: `organiq-prod-brand-files`

**Factory Pattern para Inicialização**

Arquivo: `internal/infra/storage/factory.go`

Lógica:
1. Ler variável `STORAGE_TYPE` do ambiente
2. Se `minio` → Retornar MinIO client
3. Se `s3` → Retornar S3 client
4. Ambos implementam mesma interface

### 2.4 Abstração de Fila

**Interface de Queue**

```
type QueueService interface {
    SendMessage(ctx context.Context, queueName string, message []byte) error
    ReceiveMessages(ctx context.Context, queueName string, maxMessages int) ([]Message, error)
    DeleteMessage(ctx context.Context, queueName string, receiptHandle string) error
}
```

**Implementações Concretas**

**Local (Desenvolvimento)**
- Arquivo: `internal/infra/queue/localstack_sqs.go`
- Container: LocalStack via Docker Compose
- Endpoint: `http://localhost:4566`

**Produção**
- Arquivo: `internal/infra/queue/aws_sqs.go`
- Serviço: AWS SQS
- Endpoint: AWS padrão com credenciais IAM

### 2.5 Variáveis de Ambiente

**Desenvolvimento (`.env.local`)**
```env
ENV=development
DB_HOST=localhost
DB_PORT=5432
DB_USER=organiq
DB_PASSWORD=dev_password
DB_NAME=organiq_dev
DB_SSL_MODE=disable

STORAGE_TYPE=minio
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=organiq-dev-brand-files

QUEUE_ENDPOINT=http://localhost:4566
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test
```

**Produção (AWS Secrets Manager + Environment Variables)**
```env
ENV=production
DB_HOST=organiq-prod.abc123.us-east-1.rds.amazonaws.com
DB_PORT=5432
DB_USER=organiq_prod
DB_PASSWORD=<secret-from-secrets-manager>
DB_NAME=organiq_prod
DB_SSL_MODE=require

STORAGE_TYPE=s3
S3_BUCKET=organiq-prod-brand-files
S3_REGION=us-east-1

# AWS SDK usa credenciais IAM do ECS automaticamente
```

### 2.6 Injeção de Dependências

**Arquivo Central: `cmd/api/main.go`**

Fluxo de inicialização:
1. Carregar configuração do ambiente
2. Inicializar logger estruturado
3. Conectar ao banco via factory (retorna interface)
4. Inicializar storage via factory (retorna interface)
5. Inicializar queue via factory (retorna interface)
6. Instanciar repositories com dependências
7. Instanciar use cases com repositories
8. Instanciar handlers com use cases
9. Configurar router com handlers
10. Iniciar HTTP server

**Vantagens**:
- Mudança de ambiente = mudança de variáveis
- Código dos use cases não conhece infraestrutura
- Testes usam mocks das interfaces

---

## 3. Segurança e Autenticação

### 3.1 Criptografia de Senhas

**Algoritmo: Argon2id**

Parâmetros recomendados:
- Time cost (iterations): 3
- Memory cost: 64 MB (64 * 1024 KB)
- Parallelism: 4 threads
- Salt: 16 bytes (gerado por senha)
- Pepper: 32 bytes (global, em variável de ambiente)
- Key length: 32 bytes (256 bits)

**Fluxo de Hashing**:
1. Usuário envia senha plaintext
2. Servidor gera salt aleatório de 16 bytes
3. Concatena `senha + pepper`
4. Aplica Argon2id com salt
5. Armazena no banco: `base64(salt) + "$" + base64(hash)`

**Fluxo de Verificação**:
1. Usuário envia senha plaintext
2. Servidor busca hash armazenado
3. Extrai salt do hash armazenado
4. Refaz Argon2id com `senha enviada + pepper + salt extraído`
5. Compara hash gerado com hash armazenado (constant-time comparison)

**Variável de Ambiente**:
```env
PASSWORD_PEPPER=a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6  # 32 bytes hex
```

**Rotação de Pepper**:
- Armazenar versão do pepper no hash (prefixo)
- Manter peppers antigos para verificação
- Forçar rehash no próximo login

### 3.2 Tokens JWT

**Access Token**

Características:
- Algoritmo: HS256 (HMAC + SHA256)
- Duração: 15 minutos
- Entrega: Cookie `accessToken` (HttpOnly, Secure, SameSite=Strict)
- Secret: 256 bits em variável de ambiente

Claims:
```json
{
  "sub": "user_uuid",
  "email": "user@example.com",
  "exp": 1234567890,
  "iat": 1234567890,
  "jti": "unique_token_id"
}
```

**Refresh Token**

Características:
- Formato: UUID v4
- Duração: 7 dias
- Entrega: Cookie `refreshToken` (HttpOnly, Secure, SameSite=Strict)
- Storage: Tabela `refresh_tokens` com hash SHA-256

Campos da tabela:
- `id`: UUID (PK)
- `user_id`: UUID (FK para users)
- `token_hash`: String (SHA-256 do token, indexed unique)
- `expires_at`: Timestamp
- `last_used_at`: Timestamp (nullable)
- `created_at`: Timestamp

**Fluxo de Refresh**:
1. Access token expira (15min)
2. Cliente recebe 401 em requisição protegida
3. Frontend chama `POST /api/auth/refresh` (envia refresh token no cookie)
4. Backend valida refresh token (hash + expiry + existe no banco)
5. Gera novo access token (15min)
6. Atualiza `last_used_at` do refresh token
7. Retorna 204 com novo access token no cookie
8. Frontend refaz requisição original

### 3.3 Configuração de Cookies

**Desenvolvimento**:
```
Set-Cookie: accessToken={jwt}; HttpOnly; SameSite=Strict; Path=/; Max-Age=900
Set-Cookie: refreshToken={uuid}; HttpOnly; SameSite=Strict; Path=/api/auth/refresh; Max-Age=604800
```

**Produção**:
```
Set-Cookie: accessToken={jwt}; HttpOnly; Secure; SameSite=Strict; Path=/; Max-Age=900; Domain=.organiq.com.br
Set-Cookie: refreshToken={uuid}; HttpOnly; Secure; SameSite=Strict; Path=/api/auth/refresh; Max-Age=604800; Domain=.organiq.com.br
```

**Atributos Explicados**:
- **HttpOnly**: JavaScript não pode acessar (previne XSS)
- **Secure**: Apenas HTTPS (produção)
- **SameSite=Strict**: Previne CSRF (não envia em requests cross-origin)
- **Path**: Limita escopo do cookie
- **Domain**: Permite subdomínios (`.organiq.com.br`)
- **Max-Age**: Tempo de vida em segundos

### 3.4 Middleware de Autenticação

**Arquivo**: `internal/middleware/auth.go`

Fluxo lógico:
1. Extrair cookie `accessToken` da requisição
2. Se não existir → 401 Unauthorized
3. Validar JWT:
   - Verificar assinatura com secret
   - Verificar expiry (`exp` claim)
   - Verificar formato dos claims
4. Extrair `user_id` do claim `sub`
5. Injetar `user_id` no contexto da requisição
6. Chamar próximo handler

**Casos de Erro**:
- Token ausente: `{"error": "unauthorized", "message": "Token de autenticação não fornecido"}`
- Token inválido: `{"error": "unauthorized", "message": "Token inválido"}`
- Token expirado: `{"error": "unauthorized", "message": "Token expirado"}`

### 3.5 Rate Limiting

**Estratégia**: Token Bucket Algorithm

**Implementação**: In-memory com mapa sincronizado + limpeza periódica

**Limites por Endpoint**:

| Endpoint | Limite | Janela | Scope |
|----------|--------|--------|-------|
| `POST /auth/login` | 5 req | 15min | IP |
| `POST /auth/register` | 5 req | 15min | IP |
| `POST /wizard/generate-ideas` | 10 req | 1h | User |
| `POST /wizard/publish` | 10 req | 1h | User |
| `GET /articles` | 100 req | 1min | User |
| Global (todas as rotas) | 1000 req | 1min | IP |

**Headers de Resposta** (para clientes):
```
X-RateLimit-Limit: 5
X-RateLimit-Remaining: 2
X-RateLimit-Reset: 1234567890
```

**Resposta ao Exceder Limite**:
```json
HTTP 429 Too Many Requests
{
  "error": "rate_limit_exceeded",
  "message": "Muitas requisições. Tente novamente em X segundos",
  "retryAfter": 120
}
```

### 3.6 Headers de Segurança

**Middleware CORS**:
```
Access-Control-Allow-Origin: https://organiq.com.br
Access-Control-Allow-Methods: GET, POST, PATCH, DELETE
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Allow-Credentials: true
Access-Control-Max-Age: 3600
```

**Headers de Segurança Adicionais**:
```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
Content-Security-Policy: default-src 'self'
Referrer-Policy: origin-when-cross-origin
```

### 3.7 Validação de Entrada

**Biblioteca**: go-playground/validator v10

Regras aplicadas:
- **Email**: formato RFC 5322
- **UUID**: formato UUID v4 válido
- **URL**: esquema http/https válido
- **String**: min/max length, regex patterns
- **Enum**: valores permitidos (whitelist)
- **Nested structs**: validação recursiva

**Sanitização**:
- Remover whitespace de emails/URLs
- HTML escape em strings de texto livre
- Rejeitar caracteres perigosos (null bytes, scripts)

**Limite de Payload**:
- JSON: 10 MB
- Multipart/form-data (upload): 5 MB por arquivo

---

## 4. Especificação da API

### 4.1 Convenções Gerais

**Base URL**: `/api`

**Content-Type**: `application/json` (exceto uploads)

**Formato de Timestamps**: ISO 8601 UTC
```
Exemplo: "2024-12-09T10:30:00Z"
```

**Formato de IDs**: UUID v4
```
Exemplo: "550e8400-e29b-41d4-a716-446655440000"
```

**Paginação**: Query params `page` (default 1) e `limit` (default 10, max 100)

**Respostas de Erro Padronizadas**:
```json
{
  "error": "error_code",
  "message": "Mensagem legível para o usuário",
  "details": {
    "field": "campo específico com erro"
  }
}
```

---

### 4.2 Domínio: Autenticação

#### **POST /api/auth/register**

Cadastro de novo usuário.

**Request Body**:
```json
{
  "name": "João Silva",
  "email": "joao@example.com",
  "password": "SenhaForte123"
}
```

**Validações**:
- `name`: min 2, max 100 caracteres
- `email`: formato válido, único no banco
- `password`: min 8 caracteres, 1 maiúscula, 1 minúscula, 1 número

**Response 201 Created**:
```json
{
  "user": {
    "id": "uuid",
    "name": "João Silva",
    "email": "joao@example.com",
    "planId": "uuid",
    "planName": "Free",
    "maxArticles": 0,
    "articlesUsed": 0,
    "hasCompletedOnboarding": false,
    "createdAt": "2024-12-09T10:30:00Z"
  }
}
```

**Set-Cookie**:
```
accessToken={jwt}; HttpOnly; Secure; SameSite=Strict; Max-Age=900
refreshToken={uuid}; HttpOnly; Secure; SameSite=Strict; Max-Age=604800
```

**Erros**:
- 400: Validação falhou
- 409: Email já cadastrado
- 429: Rate limit excedido

---

#### **POST /api/auth/login**

Autenticação de usuário existente.

**Request Body**:
```json
{
  "email": "joao@example.com",
  "password": "SenhaForte123"
}
```

**Response 200 OK**: Idêntico a `/register`

**Erros**:
- 401: Credenciais inválidas
- 429: Rate limit (5 tentativas em 15min)

---

#### **POST /api/auth/refresh**

Renova access token usando refresh token.

**Headers Required**:
```
Cookie: refreshToken={uuid}
```

**Response 204 No Content**

**Set-Cookie**:
```
accessToken={novo_jwt}; HttpOnly; Secure; SameSite=Strict; Max-Age=900
```

**Erros**:
- 401: Refresh token inválido/expirado

---

#### **POST /api/auth/logout**

Logout do usuário (invalida refresh token).

**Headers Required**:
```
Cookie: refreshToken={uuid}
```

**Response 204 No Content**

**Set-Cookie** (limpa cookies):
```
accessToken=; Max-Age=0
refreshToken=; Max-Age=0
```

---

#### **GET /api/auth/me**

Retorna dados do usuário autenticado.

**Headers Required**:
```
Cookie: accessToken={jwt}
```

**Response 200 OK**:
```json
{
  "user": {
    "id": "uuid",
    "name": "João Silva",
    "email": "joao@example.com",
    "planId": "uuid",
    "planName": "Pro",
    "maxArticles": 15,
    "articlesUsed": 7,
    "hasCompletedOnboarding": true,
    "createdAt": "2024-12-09T10:30:00Z"
  }
}
```

---

### 4.3 Domínio: Planos

#### **GET /api/plans**

Lista planos disponíveis.

**Autenticação**: Não requer

**Response 200 OK**:
```json
[
  {
    "id": "uuid",
    "name": "Starter",
    "maxArticles": 5,
    "price": 49.90,
    "features": [
      "5 matérias/mês",
      "SEO básico",
      "Suporte email"
    ]
  },
  {
    "id": "uuid",
    "name": "Pro",
    "maxArticles": 15,
    "price": 99.90,
    "features": [
      "15 matérias/mês",
      "SEO avançado",
      "Suporte prioritário"
    ]
  }
]
```

---

### 4.4 Domínio: Pagamentos

#### **POST /api/payments/create-checkout**

Cria sessão de checkout para assinatura de plano.

**Autenticação**: Requer access token

**Request Body**:
```json
{
  "planId": "uuid"
}
```

**Response 201 Created**:
```json
{
  "checkoutUrl": "https://stripe.com/checkout/session/...",
  "sessionId": "uuid"
}
```

**Lógica Backend**:
1. Validar se usuário já tem pagamento ativo
2. Criar sessão no provedor (Stripe/Mercado Pago)
3. Salvar registro `Payment` com status `pending`
4. Retornar URL de redirect

---

#### **GET /api/payments/status/:sessionId**

Verifica status de pagamento (polling).

**Autenticação**: Requer access token

**Response 200 OK**:
```json
{
  "id": "uuid",
  "status": "pending",
  "planId": "uuid"
}
```

Valores de `status`: `pending`, `paid`, `failed`

---

#### **POST /api/payments/webhook**

Webhook de confirmação de pagamento.

**Autenticação**: Validação de assinatura do provedor (Stripe Signature, Mercado Pago X-Signature)

**Request Body**: Payload do provedor (variável)

**Response 200 OK**: (Vazio)

**Lógica Backend**:
1. Validar assinatura do webhook
2. Atualizar status do `Payment` no banco
3. Se `status = paid`:
   - Atualizar `User.planId`
   - Resetar `User.hasCompletedOnboarding = false`
   - Enviar email de confirmação

---

#### **POST /api/payments/create-portal-session**

Cria sessão do portal de gerenciamento de assinatura.

**Autenticação**: Requer access token

**Response 200 OK**:
```json
{
  "url": "https://billing.stripe.com/portal/session/..."
}
```

---

### 4.5 Domínio: Wizard (Onboarding)

#### **POST /api/wizard/business**

Salva informações do negócio do usuário.

**Autenticação**: Requer access token

**Content-Type**: `multipart/form-data`

**Form Fields**:
```
description: string (textarea, max 500 chars)
primaryObjective: enum("leads", "sales", "branding")
secondaryObjective: enum("leads", "sales", "branding") | null
location: JSON string (BusinessLocation)
siteUrl: string (URL) | null
hasBlog: boolean
blogUrls: JSON array de strings (URLs) | []
articleCount: integer (1-50)
brandFile: File (PDF/JPG/PNG, max 5MB) | null
```

**Estrutura de `location` (JSON)**:
```json
{
  "country": "Brasil",
  "state": "São Paulo",
  "city": "São Paulo",
  "hasMultipleUnits": true,
  "units": [
    {
      "id": "uuid",
      "name": "Matriz",
      "country": "Brasil",
      "state": "SP",
      "city": "São Paulo"
    }
  ]
}
```

**Response 201 Created**:
```json
{
  "success": true
}
```

**Lógica Backend**:
1. Validar se usuário tem `planId` válido (pagamento ativo)
2. Se `brandFile` fornecido:
   - Upload para S3/MinIO
   - Salvar URL no banco
3. Criar/atualizar `BusinessProfile` com JSONB para location

---

#### **POST /api/wizard/competitors**

Salva URLs de concorrentes para análise.

**Autenticação**: Requer access token

**Request Body**:
```json
{
  "competitorUrls": [
    "https://concorrente1.com",
    "https://concorrente2.com"
  ]
}
```

**Validações**:
- Max 10 URLs
- Formato URL válido

**Response 201 Created**:
```json
{
  "success": true
}
```

**Lógica Backend**:
1. Deletar concorrentes anteriores do usuário
2. Inserir novos concorrentes em batch

---

#### **POST /api/wizard/integrations**

Salva configurações de integrações (WordPress, Google).

**Autenticação**: Requer access token

**Request Body**:
```json
{
  "wordpress": {
    "siteUrl": "https://meusite.com",
    "username": "admin",
    "appPassword": "xxxx xxxx xxxx xxxx xxxx xxxx"
  },
  "searchConsole": {
    "enabled": true,
    "propertyUrl": "https://meusite.com"
  },
  "analytics": {
    "enabled": true,
    "measurementId": "G-XXXXXXXXXX"
  }
}
```

**Response 201 Created**:
```json
{
  "success": true
}
```

**Lógica Backend**:
1. Validar credenciais do WordPress:
   - Testar autenticação com `GET /wp-json/wp/v2/users/me`
   - Retornar erro se falhar
2. Criptografar `appPassword` antes de salvar (AES-256-GCM)
3. Upsert integrações por tipo (WordPress, Search Console, Analytics)

**Erros**:
- 422: Credenciais do WordPress inválidas

---

#### **POST /api/wizard/generate-ideas** 

**Response 202 Accepted**:
```json
{
  "jobId": "uuid",
  "status": "processing",
  "message": "Geração de ideias iniciada. Use o jobId para verificar o progresso."
}
```

**Lógica Backend**:
1. Validar que usuário completou passos anteriores (business + integrations)
2. Buscar dados do usuário:
   - `BusinessProfile` (objetivos, localização, URLs)
   - `Competitors` (URLs dos concorrentes)
   - `Integration` (Search Console se disponível)
3. Criar registro `ArticleJob` com status `queued`
4. Enviar mensagem para SQS `article-generation-queue`:
```json
{
  "jobId": "uuid",
  "userId": "uuid",
  "type": "generate_ideas",
  "payload": {
    "businessInfo": {},
    "competitors": [],
    "articleCount": 5
  }
}
```
5. Retornar 202 com `jobId` imediatamente

**Worker Assíncrono**:
- Pool de goroutines escuta a fila
- Processa mensagem:
  1. Atualizar `ArticleJob.status = "processing"`
  2. **Análise de Concorrentes**:
     - Scraping dos URLs (extração de títulos, meta descriptions)
     - Análise de tópicos com OpenAI (GPT-4)
  3. **Análise do Search Console** (se conectado):
     - Buscar top 20 queries com impressões > 100
     - Filtrar por taxa de clique < 5% (oportunidades)
  4. **Geração de Ideias com OpenAI**:
     - Prompt personalizado com:
       - Objetivos do negócio
       - Localização (para SEO local)
       - Análise da concorrência
       - Keywords do Search Console
     - Gerar N ideias (baseado em `articleCount`)
  5. Salvar ideias na tabela `ArticleIdeas`
  6. Atualizar `ArticleJob.status = "completed"`
- Em caso de erro:
  - Atualizar `ArticleJob.status = "failed"`
  - Salvar erro em `ArticleJob.error_message`
  - Implementar retry com exponential backoff (3 tentativas)

---

#### **GET /api/wizard/ideas-status/:jobId**

Verifica status da geração de ideias (polling).

**Autenticação**: Requer access token

**Path Params**:
- `jobId`: UUID do job

**Response 200 OK** (Processando):
```json
{
  "jobId": "uuid",
  "status": "processing",
  "progress": 65,
  "message": "Analisando concorrentes..."
}
```

**Response 200 OK** (Completado):
```json
{
  "jobId": "uuid",
  "status": "completed",
  "ideas": [
    {
      "id": "uuid",
      "title": "5 Estratégias de Marketing Digital para Clínicas em 2024",
      "summary": "Descubra as principais tendências de marketing digital que estão transformando a forma como clínicas atraem e retêm pacientes.",
      "approved": false,
      "feedback": null
    },
    {
      "id": "uuid",
      "title": "Como Aumentar a Taxa de Conversão do seu Site Médico",
      "summary": "Aprenda técnicas práticas de otimização que podem dobrar o número de agendamentos online.",
      "approved": false,
      "feedback": null
    }
  ]
}
```

**Response 200 OK** (Erro):
```json
{
  "jobId": "uuid",
  "status": "failed",
  "errorMessage": "Erro ao conectar com OpenAI API. Tente novamente."
}
```

**Valores de `status`**: `queued`, `processing`, `completed`, `failed`

**Lógica Backend**:
1. Buscar `ArticleJob` por ID e validar ownership
2. Se `status = completed`:
   - Buscar ideias relacionadas
   - Retornar com array completo
3. Se `status = processing`:
   - Retornar apenas status + progress
4. Se `status = failed`:
   - Retornar erro para usuário
5. Implementar cache de 30s (Redis) para reduzir carga

**Erros**:
- 404: Job não encontrado
- 403: Job pertence a outro usuário

---

#### **POST /api/wizard/publish**

Inicia publicação assíncrona de matérias aprovadas.

**Autenticação**: Requer access token

**Request Body**:
```json
{
  "articles": [
    {
      "id": "uuid",
      "feedback": "Focar em pequenas empresas e mencionar nosso produto X"
    },
    {
      "id": "uuid",
      "feedback": null
    }
  ]
}
```

**Validações**:
- Mínimo 1 artigo
- Todos os IDs devem existir e pertencer ao usuário
- Validar que usuário não excedeu `maxArticles` do plano

**Response 202 Accepted**:
```json
{
  "jobId": "uuid",
  "status": "processing",
  "articlesCount": 2
}
```

**Lógica Backend**:
1. Validar limite de artigos:
   - `articlesUsed + articlesCount <= maxArticles`
   - Se exceder: retornar erro 422
2. Marcar ideias como `approved = true`
3. Criar `ArticleJob` com tipo `publish`
4. Para cada artigo:
   - Criar registro `Article` com status `generating`
   - Enviar mensagem para SQS:
```json
{
  "articleId": "uuid",
  "userId": "uuid",
  "ideaId": "uuid",
  "feedback": "string ou null"
}
```
5. Incrementar `User.articlesUsed`
6. Retornar 202

**Worker Assíncrono (por artigo)**:
1. Atualizar `Article.status = "generating"`
2. **Escrever Conteúdo com OpenAI**:
   - Prompt incluindo:
     - Título e resumo da ideia
     - Feedback do usuário (se houver)
     - Objetivos do negócio
     - Localização (para SEO local)
     - Tom da marca (análise do brandFile se existir)
   - Modelo: GPT-4 (ou GPT-4 Turbo)
   - Formato: Markdown estruturado (H2, H3, listas, parágrafos)
3. **Otimização SEO**:
   - Gerar meta description (150-160 chars)
   - Extrair palavra-chave principal
   - Sugerir slug URL
4. **Publicar no WordPress**:
   - Atualizar `Article.status = "publishing"`
   - Converter Markdown para HTML
   - Fazer POST `/wp-json/wp/v2/posts`:
```json
{
  "title": "Título do Artigo",
  "content": "<html>...</html>",
  "status": "publish",
  "meta": {
    "description": "Meta description..."
  },
  "categories": [1],
  "tags": ["SEO", "Marketing"]
}
```
5. Salvar `Article.postUrl` e `Article.status = "published"`
6. Se erro em qualquer etapa:
   - `Article.status = "error"`
   - Salvar conteúdo gerado em `Article.content` (para republicar)
   - Salvar mensagem de erro em `Article.errorMessage`
   - Implementar retry (3 tentativas com backoff)

**Erros**:
- 422: Limite de artigos excedido
- 422: IDs inválidos ou não aprovados

---

### 4.6 Domínio: Artigos

#### **GET /api/articles**

Lista matérias do usuário com paginação e filtros.

**Autenticação**: Requer access token

**Query Params**:
- `page`: integer (default 1, min 1)
- `limit`: integer (default 10, min 1, max 100)
- `status`: enum (`all`, `generating`, `publishing`, `published`, `error`) (default `all`)

**Response 200 OK**:
```json
{
  "articles": [
    {
      "id": "uuid",
      "title": "5 Estratégias de Marketing Digital para Clínicas",
      "createdAt": "2024-12-09T10:30:00Z",
      "status": "published",
      "postUrl": "https://meusite.com/5-estrategias-marketing-digital-clinicas"
    },
    {
      "id": "uuid",
      "title": "Como Aumentar a Taxa de Conversão",
      "createdAt": "2024-12-09T11:15:00Z",
      "status": "error",
      "errorMessage": "Falha ao conectar com WordPress: Invalid credentials"
    }
  ],
  "total": 47,
  "page": 1,
  "limit": 10
}
```

**Lógica Backend**:
1. Buscar artigos do usuário com filtros
2. Ordernar por `createdAt DESC`
3. Aplicar paginação (LIMIT/OFFSET)
4. Contar total para metadata

---

#### **GET /api/articles/:id**

Retorna detalhes de um artigo específico.

**Autenticação**: Requer access token

**Response 200 OK**:
```json
{
  "id": "uuid",
  "title": "5 Estratégias de Marketing Digital",
  "createdAt": "2024-12-09T10:30:00Z",
  "status": "published",
  "postUrl": "https://meusite.com/artigo",
  "content": "<html>Conteúdo completo...</html>",
  "errorMessage": null
}
```

**Erros**:
- 404: Artigo não encontrado
- 403: Artigo pertence a outro usuário

---

#### **POST /api/articles/:id/republish**

Retenta publicação de artigo com erro.

**Autenticação**: Requer access token

**Response 202 Accepted**:
```json
{
  "message": "Republicação iniciada",
  "articleId": "uuid"
}
```

**Lógica Backend**:
1. Validar que artigo tem `status = "error"`
2. Validar que `content` existe (foi gerado anteriormente)
3. Atualizar `status = "publishing"`
4. Enviar para fila (mesmo worker de publicação)

---

### 4.7 Domínio: Conta

#### **GET /api/account**

Retorna informações completas da conta do usuário.

**Autenticação**: Requer access token

**Response 200 OK**:
```json
{
  "profile": {
    "id": "uuid",
    "name": "João Silva",
    "email": "joao@example.com",
    "createdAt": "2024-12-09T10:30:00Z"
  },
  "plan": {
    "id": "uuid",
    "name": "Pro",
    "maxArticles": 15,
    "articlesUsed": 7,
    "price": 99.90,
    "nextBillingDate": "2025-01-09T00:00:00Z"
  },
  "integrations": {
    "wordpress": {
      "connected": true,
      "siteUrl": "https://meusite.com"
    },
    "searchConsole": {
      "connected": true,
      "propertyUrl": "https://meusite.com"
    },
    "analytics": {
      "connected": true,
      "measurementId": "G-XXXXXXXXXX"
    }
  }
}
```

---

#### **PATCH /api/account/profile**

Atualiza informações do perfil do usuário.

**Autenticação**: Requer access token

**Request Body**:
```json
{
  "name": "João Pedro Silva"
}
```

**Response 200 OK**:
```json
{
  "user": {
    "id": "uuid",
    "name": "João Pedro Silva",
    "email": "joao@example.com"
  }
}
```

---

#### **PATCH /api/account/integrations**

Atualiza configurações de integrações.

**Autenticação**: Requer access token

**Request Body**:
```json
{
  "wordpress": {
    "siteUrl": "https://novosite.com",
    "username": "admin",
    "appPassword": "xxxx xxxx xxxx xxxx"
  },
  "searchConsole": {
    "enabled": false
  },
  "analytics": {
    "enabled": true,
    "measurementId": "G-YYYYYYYYYY"
  }
}
```

**Response 200 OK**:
```json
{
  "success": true
}
```

**Lógica Backend**: Idêntica ao endpoint do wizard, com validação de credenciais do WordPress.

---

#### **GET /api/account/plan**

Retorna detalhes do plano atual.

**Autenticação**: Requer access token

**Response 200 OK**:
```json
{
  "id": "uuid",
  "name": "Pro",
  "maxArticles": 15,
  "articlesUsed": 7,
  "price": 99.90,
  "nextBillingDate": "2025-01-09T00:00:00Z",
  "canUpgrade": true
}
```

---

### 4.8 Domínio: Health Check

#### **GET /api/health**

Verifica saúde da API e dependências.

**Autenticação**: Não requer

**Response 200 OK**:
```json
{
  "status": "healthy",
  "timestamp": "2024-12-09T10:30:00Z",
  "version": "1.0.0",
  "dependencies": {
    "database": "healthy",
    "storage": "healthy",
    "queue": "healthy",
    "openai": "healthy"
  }
}
```

**Response 503 Service Unavailable** (se alguma dependência falhar):
```json
{
  "status": "unhealthy",
  "timestamp": "2024-12-09T10:30:00Z",
  "version": "1.0.0",
  "dependencies": {
    "database": "healthy",
    "storage": "unhealthy",
    "queue": "healthy",
    "openai": "degraded"
  }
}
```

**Lógica Backend**:
1. Testar conexão com banco (query simples)
2. Testar storage (HEAD bucket)
3. Testar fila (listar queues)
4. Testar OpenAI (ping endpoint)
5. Retornar agregado

---

## 5. Modelagem de Dados

### 5.1 Schema Lógico (PostgreSQL)

#### **Tabela: users**

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    plan_id UUID REFERENCES plans(id),
    articles_used INTEGER NOT NULL DEFAULT 0,
    has_completed_onboarding BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_plan_id ON users(plan_id);
```

#### **Tabela: refresh_tokens**

```sql
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
```

#### **Tabela: plans**

```sql
CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    max_articles INTEGER NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    features JSONB NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Seed Data**:
```sql
INSERT INTO plans (name, max_articles, price, features) VALUES
('Free', 0, 0.00, '["Teste grátis", "Suporte limitado"]'),
('Starter', 5, 49.90, '["5 matérias/mês", "SEO básico", "Suporte email"]'),
('Pro', 15, 99.90, '["15 matérias/mês", "SEO avançado", "Suporte prioritário"]'),
('Enterprise', 50, 249.90, '["50 matérias/mês", "SEO premium", "Suporte dedicado"]');
```

#### **Tabela: payments**

```sql
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES plans(id),
    provider VARCHAR(50) NOT NULL, -- 'stripe' | 'mercadopago'
    provider_session_id VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL, -- 'pending' | 'paid' | 'failed'
    amount DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_payments_provider_session_id ON payments(provider_session_id);
CREATE INDEX idx_payments_status ON payments(status);
```

#### **Tabela: business_profiles**

```sql
CREATE TABLE business_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    primary_objective VARCHAR(20) NOT NULL, -- 'leads' | 'sales' | 'branding'
    secondary_objective VARCHAR(20),
    location JSONB NOT NULL, -- Estrutura BusinessLocation
    site_url TEXT,
    has_blog BOOLEAN NOT NULL DEFAULT FALSE,
    blog_urls JSONB, -- Array de URLs
    brand_file_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_business_profiles_user_id ON business_profiles(user_id);
```

**Exemplo de `location` JSONB**:
```json
{
  "country": "Brasil",
  "state": "São Paulo",
  "city": "São Paulo",
  "hasMultipleUnits": true,
  "units": [
    {
      "id": "uuid",
      "name": "Matriz",
      "country": "Brasil",
      "state": "SP",
      "city": "São Paulo"
    }
  ]
}
```

#### **Tabela: competitors**

```sql
CREATE TABLE competitors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_competitors_user_id ON competitors(user_id);
```

#### **Tabela: integrations**

```sql
CREATE TABLE integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- 'wordpress' | 'search_console' | 'analytics'
    config JSONB NOT NULL, -- Configuração específica do tipo (encrypted)
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, type)
);

CREATE INDEX idx_integrations_user_id ON integrations(user_id);
```

**Exemplo de `config` JSONB (WordPress)**:
```json
{
  "siteUrl": "https://meusite.com",
  "username": "admin",
  "appPassword": "encrypted_base64_string"
}
```

#### **Tabela: article_jobs**

```sql
CREATE TABLE article_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL, -- 'generate_ideas' | 'publish'
    status VARCHAR(20) NOT NULL, -- 'queued' | 'processing' | 'completed' | 'failed'
    progress INTEGER NOT NULL DEFAULT 0, -- 0-100
    payload JSONB NOT NULL, -- Dados de entrada
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_article_jobs_user_id ON article_jobs(user_id);
CREATE INDEX idx_article_jobs_status ON article_jobs(status);
```

#### **Tabela: article_ideas**

```sql
CREATE TABLE article_ideas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_id UUID NOT NULL REFERENCES article_jobs(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    summary TEXT NOT NULL,
    approved BOOLEAN NOT NULL DEFAULT FALSE,
    feedback TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_article_ideas_user_id ON article_ideas(user_id);
CREATE INDEX idx_article_ideas_job_id ON article_ideas(job_id);
```

#### **Tabela: articles**

```sql
CREATE TABLE articles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    idea_id UUID REFERENCES article_ideas(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    content TEXT, -- HTML gerado
    status VARCHAR(20) NOT NULL, -- 'generating' | 'publishing' | 'published' | 'error'
    post_url TEXT,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_articles_user_id ON articles(user_id);
CREATE INDEX idx_articles_status ON articles(status);
CREATE INDEX idx_articles_created_at ON articles(created_at DESC);
```

### 5.2 Relacionamentos

```
users (1) ─────────── (*) refresh_tokens
users (1) ─────────── (1) business_profiles
users (1) ─────────── (*) competitors
users (1) ─────────── (*) integrations
users (1) ─────────── (*) article_jobs
users (1) ─────────── (*) article_ideas
users (1) ─────────── (*) articles
users (*) ─────────── (1) plans
article_jobs (1) ──── (*) article_ideas
article_ideas (1) ─── (1) articles (opcional)
```

---

## 6. Estrutura de Arquivos

### 6.1 Árvore de Diretórios (Clean Architecture)

```
backend/
├── cmd/
│   └── api/
│       └── main.go                    # Entrypoint da aplicação
│
├── internal/
│   ├── domain/                        # Camada de Domínio (entidades + interfaces)
│   │   ├── entity/
│   │   │   ├── user.go
│   │   │   ├── plan.go
│   │   │   ├── article.go
│   │   │   ├── business_profile.go
│   │   │   ├── integration.go
│   │   │   └── payment.go
│   │   │
│   │   └── repository/               # Interfaces de Repository
│   │       ├── user_repository.go
│   │       ├── plan_repository.go
│   │       ├── article_repository.go
│   │       ├── business_repository.go
│   │       └── integration_repository.go
│   │
│   ├── usecase/                      # Camada de Aplicação (lógica de negócio)
│   │   ├── auth/
│   │   │   ├── register.go
│   │   │   ├── login.go
│   │   │   ├── refresh.go
│   │   │   └── logout.go
│   │   │
│   │   ├── wizard/
│   │   │   ├── save_business.go
│   │   │   ├── save_competitors.go
│   │   │   ├── save_integrations.go
│   │   │   ├── generate_ideas.go
│   │   │   └── publish_articles.go
│   │   │
│   │   ├── article/
│   │   │   ├── list_articles.go
│   │   │   ├── get_article.go
│   │   │   └── republish_article.go
│   │   │
│   │   └── account/
│   │       ├── get_account.go
│   │       ├── update_profile.go
│   │       ├── update_integrations.go
│   │       └── get_plan.go
│   │
│   ├── infra/                        # Camada de Infraestrutura
│   │   ├── repository/               # Implementações de Repository
│   │   │   ├── postgres/
│   │   │   │   ├── user_repository.go
│   │   │   │   ├── plan_repository.go
│   │   │   │   ├── article_repository.go
│   │   │   │   ├── business_repository.go
│   │   │   │   └── integration_repository.go
│   │   │   │
│   │   │   └── factory.go            # Factory de Repositories
│   │   │
│   │   ├── storage/                  # Abstração de Blob Storage
│   │   │   ├── interface.go
│   │   │   ├── minio_storage.go      # Implementação MinIO (dev)
│   │   │   ├── s3_storage.go         # Implementação S3 (prod)
│   │   │   └── factory.go
│   │   │
│   │   ├── queue/                    # Abstração de Filas
│   │   │   ├── interface.go
│   │   │   ├── localstack_sqs.go     # Implementação LocalStack (dev)
│   │   │   ├── aws_sqs.go            # Implementação AWS SQS (prod)
│   │   │   └── factory.go
│   │   │
│   │   ├── ai/                       # Serviços de IA
│   │   │   ├── openai_client.go
│   │   │   └── prompts.go
│   │   │
│   │   ├── wordpress/                # Cliente WordPress
│   │   │   └── client.go
│   │   │
│   │   └── database/                 # Configuração do banco
│   │       ├── connection.go
│   │       └── migrations/
│   │           ├── 001_create_users.sql
│   │           ├── 002_create_plans.sql
│   │           └── ...
│   │
│   ├── handler/                      # Camada de Apresentação (HTTP handlers)
│   │   ├── auth_handler.go
│   │   ├── plan_handler.go
│   │   ├── payment_handler.go
│   │   ├── wizard_handler.go
│   │   ├── article_handler.go
│   │   ├── account_handler.go
│   │   └── health_handler.go
│   │
│   ├── middleware/                   # Middlewares HTTP
│   │   ├── auth.go
│   │   ├── cors.go
│   │   ├── rate_limit.go
│   │   ├── logger.go
│   │   └── recovery.go
│   │
│   ├── worker/                       # Workers assíncronos
│   │   ├── article_generator.go
│   │   ├── article_publisher.go
│   │   └── pool.go
│   │
│   └── util/                         # Utilitários
│       ├── crypto.go                 # Argon2, AES, JWT
│       ├── validator.go              # Validações personalizadas
│       ├── response.go               # Helpers de resposta HTTP
│       └── logger.go                 # Configuração do logger
│
├── config/                           # Configurações
│   ├── config.go                     # Struct de configuração
│   └── env.go                        # Carregamento de variáveis
│
├── docker-compose.yml                # Stack de desenvolvimento
├── Dockerfile                        # Build da aplicação
├── go.mod
├── go.sum
└── README.md
```

### 6.2 Responsabilidade das Camadas

#### **Domain (`internal/domain/`)**

**Responsabilidade**: Definir as regras de negócio puras e contratos.

**Contém**:
- **Entities**: Structs de domínio (User, Article, etc) com validações de negócio
- **Interfaces de Repository**: Contratos que a camada de infra deve implementar
- **Value Objects**: Tipos personalizados (Email, Password, etc)

**Regras**:
- Não depende de NENHUMA outra camada
- Não importa pacotes externos (exceto Go stdlib)
- Toda a aplicação depende do domínio

**Exemplo** (`domain/entity/user.go`):
```go
type User struct {
    ID                      uuid.UUID
    Name                    string
    Email                   string
    PasswordHash            string
    PlanID                  uuid.UUID
    ArticlesUsed            int
    HasCompletedOnboarding  bool
    CreatedAt               time.
    UpdatedAt               time.Time
}

// Validação de negócio
func (u *User) Validate() error {
    if len(u.Name) < 2 || len(u.Name) > 100 {
        return errors.New("nome deve ter entre 2 e 100 caracteres")
    }
    
    if !isValidEmail(u.Email) {
        return errors.New("email inválido")
    }
    
    if u.ArticlesUsed < 0 {
        return errors.New("artigos usados não pode ser negativo")
    }
    
    return nil
}

func (u *User) CanGenerateArticles(count int, maxArticles int) bool {
    return u.ArticlesUsed + count <= maxArticles
}

func (u *User) IncrementArticlesUsed(count int) error {
    u.ArticlesUsed += count
    u.UpdatedAt = time.Now()
    return nil
}

func isValidEmail(email string) bool {
    // Implementação de validação de email
    return true // Simplificado
}
```

---

#### **Application (`internal/usecase/`)**

**Responsabilidade**: Orquestrar fluxos de negócio usando entities e repositories.

**Contém**:
- Use cases específicos (RegisterUser, GenerateArticleIdeas, etc)
- Coordenação de múltiplos repositories
- Aplicação de regras de negócio complexas
- Gestão de transações

**Regras**:
- Depende APENAS do Domain (entities + interfaces)
- Não conhece detalhes de HTTP, banco, ou infraestrutura
- Recebe interfaces como dependências (injeção)
- Retorna erros de domínio, não códigos HTTP

**Exemplo** (`usecase/auth/register.go`):
```go
package auth

import (
    "context"
    "errors"
    
    "github.com/google/uuid"
    "organiq/internal/domain/entity"
    "organiq/internal/domain/repository"
    "organiq/internal/util"
)

type RegisterUserInput struct {
    Name     string
    Email    string
    Password string
}

type RegisterUserOutput struct {
    User         *entity.User
    AccessToken  string
    RefreshToken string
}

type RegisterUserUseCase struct {
    userRepo    repository.UserRepository
    planRepo    repository.PlanRepository
    crypto      *util.CryptoService
    jwtService  *util.JWTService
}

func NewRegisterUserUseCase(
    userRepo repository.UserRepository,
    planRepo repository.PlanRepository,
    crypto *util.CryptoService,
    jwt *util.JWTService,
) *RegisterUserUseCase {
    return &RegisterUserUseCase{
        userRepo:   userRepo,
        planRepo:   planRepo,
        crypto:     crypto,
        jwtService: jwt,
    }
}

func (uc *RegisterUserUseCase) Execute(ctx context.Context, input RegisterUserInput) (*RegisterUserOutput, error) {
    // 1. Validar entrada
    if len(input.Password) < 8 {
        return nil, errors.New("senha deve ter no mínimo 8 caracteres")
    }
    
    // 2. Verificar se email já existe
    existing, _ := uc.userRepo.FindByEmail(ctx, input.Email)
    if existing != nil {
        return nil, errors.New("email já cadastrado")
    }
    
    // 3. Buscar plano Free (padrão)
    freePlan, err := uc.planRepo.FindByName(ctx, "Free")
    if err != nil {
        return nil, errors.New("erro ao buscar plano padrão")
    }
    
    // 4. Hash da senha com Argon2id
    passwordHash, err := uc.crypto.HashPassword(input.Password)
    if err != nil {
        return nil, err
    }
    
    // 5. Criar entidade User
    user := &entity.User{
        ID:                     uuid.New(),
        Name:                   input.Name,
        Email:                  input.Email,
        PasswordHash:           passwordHash,
        PlanID:                 freePlan.ID,
        ArticlesUsed:           0,
        HasCompletedOnboarding: false,
    }
    
    // 6. Validar regras de negócio
    if err := user.Validate(); err != nil {
        return nil, err
    }
    
    // 7. Persistir no banco
    if err := uc.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }
    
    // 8. Gerar tokens
    accessToken, err := uc.jwtService.GenerateAccessToken(user.ID, user.Email)
    if err != nil {
        return nil, err
    }
    
    refreshToken, err := uc.jwtService.GenerateRefreshToken(user.ID)
    if err != nil {
        return nil, err
    }
    
    return &RegisterUserOutput{
        User:         user,
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
    }, nil
}
```

---

#### **Infrastructure (`internal/infra/`)**

**Responsabilidade**: Implementar os contratos do Domain com tecnologias concretas.

**Contém**:
- Implementações de Repository usando GORM
- Clients externos (OpenAI, WordPress, AWS)
- Conexões com banco, storage, filas
- Migrations SQL

**Regras**:
- Implementa interfaces do Domain
- Pode importar bibliotecas externas (GORM, AWS SDK, etc)
- Isolado em pacotes por tipo (repository, storage, queue)
- Usa Factory Pattern para alternar entre dev/prod

**Exemplo** (`infra/repository/postgres/user_repository.go`):
```go
package postgres

import (
    "context"
    "errors"
    
    "github.com/google/uuid"
    "gorm.io/gorm"
    "organiq/internal/domain/entity"
    "organiq/internal/domain/repository"
)

type UserRepositoryPostgres struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
    return &UserRepositoryPostgres{db: db}
}

func (r *UserRepositoryPostgres) Create(ctx context.Context, user *entity.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepositoryPostgres) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    var user entity.User
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
    
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    
    return &user, err
}

func (r *UserRepositoryPostgres) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
    var user entity.User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    
    return &user, err
}

func (r *UserRepositoryPostgres) Update(ctx context.Context, user *entity.User) error {
    return r.db.WithContext(ctx).Save(user).Error
}
```

---

#### **Presentation (`internal/handler/`)**

**Responsabilidade**: Adaptar HTTP para use cases e vice-versa.

**Contém**:
- Handlers HTTP (request → use case → response)
- Parsing de request body/params
- Validação de entrada (formato HTTP)
- Mapeamento de erros para códigos HTTP
- Serialização de responses

**Regras**:
- Depende de Use Cases (application layer)
- Conhece detalhes HTTP (status codes, headers, cookies)
- Não contém lógica de negócio
- Delega tudo para use cases

**Exemplo** (`handler/auth_handler.go`):
```go
package handler

import (
    "encoding/json"
    "net/http"
    
    "organiq/internal/usecase/auth"
    "organiq/internal/util"
)

type AuthHandler struct {
    registerUC *auth.RegisterUserUseCase
    loginUC    *auth.LoginUserUseCase
}

func NewAuthHandler(
    registerUC *auth.RegisterUserUseCase,
    loginUC *auth.LoginUserUseCase,
) *AuthHandler {
    return &AuthHandler{
        registerUC: registerUC,
        loginUC:    loginUC,
    }
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request body
    var req struct {
        Name     string `json:"name" validate:"required,min=2,max=100"`
        Email    string `json:"email" validate:"required,email"`
        Password string `json:"password" validate:"required,min=8"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        util.RespondError(w, http.StatusBadRequest, "invalid_json", "JSON inválido")
        return
    }
    
    // 2. Validar formato
    if err := util.ValidateStruct(req); err != nil {
        util.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
        return
    }
    
    // 3. Executar use case
    input := auth.RegisterUserInput{
        Name:     req.Name,
        Email:    req.Email,
        Password: req.Password,
    }
    
    output, err := h.registerUC.Execute(r.Context(), input)
    if err != nil {
        // Mapear erro de negócio para HTTP
        if err.Error() == "email já cadastrado" {
            util.RespondError(w, http.StatusConflict, "email_exists", err.Error())
            return
        }
        
        util.RespondError(w, http.StatusInternalServerError, "server_error", "Erro interno")
        return
    }
    
    // 4. Configurar cookies
    util.SetAccessTokenCookie(w, output.AccessToken)
    util.SetRefreshTokenCookie(w, output.RefreshToken)
    
    // 5. Responder
    util.RespondJSON(w, http.StatusCreated, map[string]interface{}{
        "user": map[string]interface{}{
            "id":                     output.User.ID,
            "name":                   output.User.Name,
            "email":                  output.User.Email,
            "planId":                 output.User.PlanID,
            "planName":               "Free",
            "maxArticles":            0,
            "articlesUsed":           output.User.ArticlesUsed,
            "hasCompletedOnboarding": output.User.HasCompletedOnboarding,
            "createdAt":              output.User.CreatedAt.Format(time.RFC3339),
        },
    })
}
```

---

## 7. Considerações de Deploy

### 7.1 Arquitetura de Compute (AWS ECS Fargate)

**Por que Fargate?**
- **Serverless para containers**: Não gerencia instâncias EC2
- **Escalabilidade automática**: Ajusta tasks baseado em métricas
- **Pay-per-use**: Paga apenas pelo tempo de execução
- **Isolamento**: Cada task roda em ambiente isolado
- **Compatibilidade**: Mesma imagem Docker dev/prod

**Configuração Recomendada**:
- **CPU**: 0.5 vCPU (API) / 1 vCPU (Workers)
- **Memória**: 1 GB (API) / 2 GB (Workers)
- **Tasks mínimas**: 2 (alta disponibilidade)
- **Tasks máximas**: 10 (escala sob demanda)

**Task Definition (Simplificado)**:
```json
{
  "family": "organiq-api",
  "cpu": "512",
  "memory": "1024",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "containerDefinitions": [
    {
      "name": "api",
      "image": "123456789.dkr.ecr.us-east-1.amazonaws.com/organiq-api:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {"name": "ENV", "value": "production"},
        {"name": "DB_HOST", "value": "organiq-prod.rds.amazonaws.com"}
      ],
      "secrets": [
        {"name": "DB_PASSWORD", "valueFrom": "arn:aws:secretsmanager:..."}
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/organiq-api",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "api"
        }
      }
    }
  ]
}
```

---

### 7.2 Networking e Load Balancing

**Application Load Balancer (ALB)**

**Função**: Distribuir tráfego entre tasks do Fargate + SSL Termination

**Configuração**:
- **Scheme**: Internet-facing
- **Listeners**:
  - HTTP (80) → Redirect para HTTPS
  - HTTPS (443) → Target Group (ECS tasks)
- **SSL/TLS**: Certificado ACM (AWS Certificate Manager)
- **Health Check**:
  - Path: `/api/health`
  - Interval: 30 segundos
  - Timeout: 5 segundos
  - Healthy threshold: 2
  - Unhealthy threshold: 3

**Target Group**:
- **Protocol**: HTTP
- **Port**: 8080 (porta do container)
- **Target type**: IP (Fargate usa awsvpc mode)
- **Deregistration delay**: 30 segundos

**Security Groups**:

*ALB Security Group*:
- Inbound: 80 (HTTP) de 0.0.0.0/0
- Inbound: 443 (HTTPS) de 0.0.0.0/0
- Outbound: 8080 para ECS Security Group

*ECS Security Group*:
- Inbound: 8080 de ALB Security Group
- Outbound: 443 para 0.0.0.0/0 (chamadas externas)
- Outbound: 5432 para RDS Security Group

*RDS Security Group*:
- Inbound: 5432 de ECS Security Group

---

### 7.3 Autoscaling

**Estratégia**: Target Tracking Scaling (automático baseado em métricas)

**Métricas de Escala**:

**1. CPU Utilization**
- Target: 70%
- Scale-out: Quando média de 3 minutos > 70%
- Scale-in: Quando média de 15 minutos < 50%
- Cooldown: 180 segundos

**2. Request Count per Target**
- Target: 1000 requests/minuto por task
- Permite prever sobrecarga antes do CPU saturar

**Configuração (Terraform-style)**:
```hcl
resource "aws_appautoscaling_target" "ecs_target" {
  max_capacity       = 10
  min_capacity       = 2
  resource_id        = "service/organiq-cluster/organiq-api"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "cpu_scaling" {
  name               = "cpu-target-tracking"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs_target.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs_target.scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs_target.service_namespace

  target_tracking_scaling_policy_configuration {
    target_value       = 70.0
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    scale_in_cooldown  = 300
    scale_out_cooldown = 60
  }
}
```

**Exemplo de Escala**:
```
Cenário: 500 req/s (pico de tráfego)
- Tasks atuais: 2
- CPU médio por task: 85%
- Autoscaling detecta > 70%
- Provisiona +2 tasks (total: 4)
- CPU médio cai para 42%
- Tráfego normaliza (100 req/s)
- Após 15min < 50%, remove 2 tasks
```

---

### 7.4 Build Multi-Stage (Dockerfile)

**Objetivo**: Gerar imagem Docker otimizada (< 50 MB) separando build e runtime.

**Conceito Multi-Stage**:
1. **Stage 1 (Builder)**: Compila o binário Go com todas as dependências
2. **Stage 2 (Runtime)**: Copia apenas o binário compilado para imagem mínima

**Vantagens**:
- Imagem final sem compilador Go, módulos, cache
- Redução de 500 MB → 20-30 MB
- Menos superfície de ataque (segurança)
- Deploy mais rápido (pull/push de imagem menor)

**Estrutura Lógica** (não é código executável):
```dockerfile
# ==========================================
# STAGE 1: BUILD
# ==========================================
# Imagem base com Go completo (golang:1.22-alpine)
# - Instalar dependências de build (git, ca-certificates)
# - Copiar go.mod e go.sum
# - Download de módulos (go mod download)
# - Copiar código fonte completo
# - Compilar binário estático:
#   * CGO_ENABLED=0 (sem dependências C)
#   * GOOS=linux (target Linux)
#   * Flags de otimização (-ldflags="-s -w")
# - Resultado: binário standalone em /app/api

# ==========================================
# STAGE 2: RUNTIME
# ==========================================
# Imagem base minimalista (alpine:latest ou scratch)
# - Copiar APENAS o binário compilado do Stage 1
# - Copiar certificados CA (para HTTPS externo)
# - Definir user não-root (segurança)
# - Expor porta 8080
# - Comando de inicialização: ./api

# Resultado final: 
# - Imagem < 50 MB
# - Sem código fonte
# - Sem ferramentas de build
```

**Fluxo de Build**:
```bash
# 1. Build da imagem
docker build -t organiq-api:latest .

# 2. Tag para ECR
docker tag organiq-api:latest 123456.dkr.ecr.us-east-1.amazonaws.com/organiq-api:latest

# 3. Push para ECR
docker push 123456.dkr.ecr.us-east-1.amazonaws.com/organiq-api:latest

# 4. ECS faz pull da imagem ao criar tasks
```

**Otimizações Adicionais**:
- Layer caching: `go mod download` em layer separado
- `.dockerignore`: Excluir `.git`, testes, documentação
- Build args: Passar versão/commit SHA como label

---

### 7.5 CI/CD Pipeline (Resumo)

**Ferramentas**: GitHub Actions ou AWS CodePipeline

**Stages**:
1. **Lint & Test**: `go vet`, `golangci-lint`, `go test -race`
2. **Build**: Compilar binário + rodar testes de integração
3. **Docker Build**: Multi-stage build da imagem
4. **Push to ECR**: Upload da imagem para registry
5. **Deploy**: Atualizar ECS Service com nova imagem
6. **Health Check**: Validar `/api/health` após deploy
7. **Rollback**: Reverter se health check falhar

**Zero Downtime Deploy**:
- ECS cria novas tasks com imagem nova
- Aguarda health check passar
- Registra novas tasks no ALB Target Group
- Remove tasks antigas apenas após novas estarem healthy

---

### 7.6 Monitoramento e Observabilidade

**CloudWatch Logs**:
- Logs estruturados em JSON (zerolog)
- Retention: 7 dias (dev), 30 dias (prod)
- Filtros para erros (level: error, warn)

**CloudWatch Metrics**:
- CPU/Memory utilization (ECS)
- Request count (ALB)
- Response time (ALB target response time)
- Unhealthy targets (ALB)
- Queue depth (SQS)

**Alarmes Críticos**:
- CPU > 85% por 5 minutos → SNS notification
- Unhealthy targets > 0 por 2 minutos → PagerDuty
- Error rate > 5% por 1 minuto → Slack webhook

**AWS X-Ray**:
- Distributed tracing entre API → RDS → SQS → OpenAI
- Identificar bottlenecks em requisições lentas
- Visualizar latência por componente

---

## 🎯 Conclusão

Esta especificação define um backend Go com **Clean Architecture**, preparado para escalar de desenvolvimento local (Docker Compose + MinIO + LocalStack) até produção AWS (ECS Fargate + RDS + S3 + SQS) **sem modificar código**, apenas variáveis de ambiente.

**Destaques Técnicos**:
- ✅ Stateless API (JWT em cookies)
- ✅ Segurança robusta (Argon2id + Pepper + AES-256)
- ✅ Processamento assíncrono (SQS + Workers)
- ✅ Infraestrutura plugável (interfaces + factories)
- ✅ Deploy serverless (Fargate + Autoscaling)
- ✅ Build otimizado (Multi-stage < 50 MB)