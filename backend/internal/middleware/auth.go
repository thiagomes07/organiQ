// internal/middleware/auth.go
package middleware

import (
	"context"
	"net/http"
	"strings"

	"organiq/internal/util"

	"github.com/rs/zerolog/log"
)

// ContextKey tipo para chaves do context
type ContextKey string

const (
	UserIDContextKey = ContextKey("user_id")
	EmailContextKey  = ContextKey("email")
)

// AuthMiddleware valida JWT token de autenticação
func AuthMiddleware(crypto *util.CryptoService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Extrair token do cookie
			cookie, err := r.Cookie("accessToken")
			if err != nil {
				log.Warn().Msg("accessToken cookie não encontrado")
				util.RespondError(w, http.StatusUnauthorized, "unauthorized", "Token de autenticação não fornecido")
				return
			}

			token := cookie.Value
			if token == "" {
				log.Warn().Msg("accessToken cookie vazio")
				util.RespondError(w, http.StatusUnauthorized, "unauthorized", "Token de autenticação não fornecido")
				return
			}

			// 2. Validar JWT
			claims, err := crypto.ValidateAccessToken(token)
			if err != nil {
				log.Warn().Err(err).Msg("Erro ao validar token")

				// Diferenciar entre token expirado e inválido
				if strings.Contains(err.Error(), "expired") {
					util.RespondError(w, http.StatusUnauthorized, "token_expired", "Token expirado")
				} else {
					util.RespondError(w, http.StatusUnauthorized, "invalid_token", "Token inválido")
				}
				return
			}

			// 3. Injetar claims no context
			ctx := context.WithValue(r.Context(), UserIDContextKey, claims.Sub)
			ctx = context.WithValue(ctx, EmailContextKey, claims.Email)

			// 4. Continuar com próximo handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext extrai user_id do context
func GetUserIDFromContext(r *http.Request) string {
	userID, ok := r.Context().Value(UserIDContextKey).(string)
	if !ok {
		return ""
	}
	return userID
}

// GetEmailFromContext extrai email do context
func GetEmailFromContext(r *http.Request) string {
	email, ok := r.Context().Value(EmailContextKey).(string)
	if !ok {
		return ""
	}
	return email
}

// ============================================
// CORS MIDDLEWARE
// ============================================

// CORSMiddleware já é configurado em main.go com chi/cors
// Este arquivo serve apenas como referência
