package middleware

import (
	"net/http"
	"runtime/debug"

	"organiq/internal/util"

	"github.com/rs/zerolog/log"
)

// RecoveryMiddleware captura panics e responde com erro estruturado.
func RecoveryMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error().
						Interface("panic", rec).
						Str("method", r.Method).
						Str("path", r.URL.Path).
						Bytes("stacktrace", debug.Stack()).
						Msg("panic recovered")

					util.RespondInternalServerError(w)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
