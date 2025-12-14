# ==========================================
# STAGE 1: BUILD
# ==========================================
FROM golang:1.23-alpine AS builder

# Instalar dependências de build
RUN apk add --no-cache git ca-certificates tzdata

# Definir diretório de trabalho
WORKDIR /app

# Copiar arquivos de dependências
COPY go.mod go.sum ./

# Download de módulos
RUN go mod download

# Copiar código fonte completo
COPY . .

# Compilar binário estático otimizado
# -ldflags="-s -w": remove símbolos de debug para reduzir tamanho
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o api \
    ./cmd/api/main.go

# ==========================================
# STAGE 2: RUNTIME
# ==========================================
FROM alpine:latest

# Instalar certificados CA, tzdata e curl/wget para healthcheck
RUN apk --no-cache add ca-certificates tzdata curl wget

# Criar usuário não-root para segurança
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Definir diretório de trabalho
WORKDIR /app

# Copiar binário compilado do stage anterior
COPY --from=builder /app/api ./api

# Criar diretório para migrations
RUN mkdir -p /app/migrations

# CORREÇÃO PRINCIPAL AQUI:
# Copia os arquivos SQL diretamente. O Docker falhará se não existirem (o que é bom),
# ao invés de usar hacks de shell.
COPY --from=builder /app/internal/infra/database/migrations/*.sql /app/migrations/

# Alterar ownership para usuário não-root
RUN chown -R appuser:appuser /app && \
    chmod +x /app/api

# Trocar para usuário não-root
USER appuser

# Expor porta da aplicação
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Comando de inicialização
CMD ["./api"]