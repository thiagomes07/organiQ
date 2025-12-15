package middleware

import (
	"net/http"
	"strings"
	"time"

	"organiq/internal/util"

	chimw "github.com/go-chi/chi/v5/middleware"
)

// LoggerMiddleware adiciona logs estruturados por request e enriquece o contexto.
func LoggerMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID := chimw.GetReqID(r.Context())
			remoteIP := clientIPFromRequest(r)
			ctx := util.WithContextFields(r.Context(), util.ContextFields{
				RequestID: reqID,
				UserID:    GetUserIDFromContext(r),
				Email:     GetEmailFromContext(r),
				IP:        remoteIP,
			})

			logger := util.LoggerFromContext(ctx)
			recorder := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

			logger.Debug().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Msg("request started")

			next.ServeHTTP(recorder, r.WithContext(ctx))

			duration := time.Since(start)
			event := logger.Info()
			switch {
			case recorder.statusCode >= http.StatusInternalServerError:
				event = logger.Error()
			case recorder.statusCode >= http.StatusBadRequest:
				event = logger.Warn()
			}

			event.
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("request_id", reqID).
				Int("status", recorder.statusCode).
				Dur("duration", duration).
				Int("bytes", recorder.bytesWritten).
				Msg("request completed")
		})
	}
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.bytesWritten += n
	return n, err
}

func clientIPFromRequest(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	return r.RemoteAddr
}
