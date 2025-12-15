// internal/util/response.go
package util

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

// ============================================
// RESPONSE STRUCTURES
// ============================================

// SuccessResponse resposta bem-sucedida padrão
type SuccessResponse struct {
	Data interface{} `json:"data,omitempty"`
}

// ErrorResponse resposta de erro padrão
type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ============================================
// HELPERS
// ============================================

// RespondJSON envia resposta JSON com status code
func RespondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("Erro ao encodar resposta JSON")
	}
}

// RespondError envia resposta de erro padronizada
func RespondError(w http.ResponseWriter, statusCode int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   errorCode,
		Message: message,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error().Err(err).Msg("Erro ao encodar resposta de erro")
	}
}

// RespondErrorWithDetails envia erro com detalhes adicionais
func RespondErrorWithDetails(w http.ResponseWriter, statusCode int, errorCode, message string, details interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   errorCode,
		Message: message,
		Details: details,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error().Err(err).Msg("Erro ao encodar resposta de erro")
	}
}

// RespondCreated envia resposta 201 Created
func RespondCreated(w http.ResponseWriter, data interface{}) {
	RespondJSON(w, http.StatusCreated, data)
}

// RespondOK envia resposta 200 OK
func RespondOK(w http.ResponseWriter, data interface{}) {
	RespondJSON(w, http.StatusOK, data)
}

// RespondNoContent envia resposta 204 No Content
func RespondNoContent(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusNoContent)
}

// RespondNotFound envia resposta 404 Not Found
func RespondNotFound(w http.ResponseWriter) {
	RespondError(w, http.StatusNotFound, "not_found", "Recurso não encontrado")
}

// RespondBadRequest envia resposta 400 Bad Request
func RespondBadRequest(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusBadRequest, "bad_request", message)
}

// RespondUnauthorized envia resposta 401 Unauthorized
func RespondUnauthorized(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusUnauthorized, "unauthorized", message)
}

// RespondForbidden envia resposta 403 Forbidden
func RespondForbidden(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusForbidden, "forbidden", message)
}

// RespondConflict envia resposta 409 Conflict
func RespondConflict(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusConflict, "conflict", message)
}

// RespondInternalServerError envia resposta 500 Internal Server Error
func RespondInternalServerError(w http.ResponseWriter) {
	RespondError(w, http.StatusInternalServerError, "internal_server_error", "Erro interno do servidor")
}

// RespondTooManyRequests envia resposta 429 Too Many Requests
func RespondTooManyRequests(w http.ResponseWriter, retryAfter string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if retryAfter != "" {
		w.Header().Set("Retry-After", retryAfter)
	}
	w.WriteHeader(http.StatusTooManyRequests)

	response := ErrorResponse{
		Error:   "rate_limit_exceeded",
		Message: "Muitas requisições. Tente novamente em alguns instantes",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error().Err(err).Msg("Erro ao encodar resposta")
	}
}

// ============================================
// COOKIE HELPERS
// ============================================

// SetAccessTokenCookie configura cookie de access token
// O parâmetro secure deve ser true em produção (HTTPS)
func SetAccessTokenCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   15 * 60, // 15 minutos
	})
}

// SetRefreshTokenCookie configura cookie de refresh token
// O parâmetro secure deve ser true em produção (HTTPS)
// Path "/api/auth" permite acesso em refresh e logout
func SetRefreshTokenCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    token,
		Path:     "/api/auth", // Escopo permite refresh e logout
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   7 * 24 * 60 * 60, // 7 dias
	})
}

// ClearAuthCookies limpa cookies de autenticação
// O parâmetro secure deve corresponder ao usado na criação dos cookies
func ClearAuthCookies(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    "",
		Path:     "/api/auth", // Mesmo path de criação
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}
