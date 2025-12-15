package middleware

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"organiq/internal/util"

	"github.com/rs/zerolog/log"
)

// IdentifierFunc resolve qual chave utilizar no bucket por requisição.
type IdentifierFunc func(r *http.Request) string

// RateLimiter implementa algoritmo Token Bucket em memória.
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*tokenBucket
	fillRate float64
	capacity float64
	window   time.Duration
}

type tokenBucket struct {
	tokens     float64
	lastRefill time.Time
}

// RateLimitResult contém informações sobre o resultado do rate limit.
type RateLimitResult struct {
	Allowed   bool
	Limit     int
	Remaining int
	ResetAt   time.Time
}

// NewRateLimiter cria um rate limiter baseado em número de requests por janela.
func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	if maxRequests <= 0 {
		maxRequests = 1
	}
	if window <= 0 {
		window = time.Second
	}

	capacity := float64(maxRequests)
	fillRate := capacity / window.Seconds()

	return &RateLimiter{
		buckets:  make(map[string]*tokenBucket),
		fillRate: fillRate,
		capacity: capacity,
		window:   window,
	}
}

// Allow verifica se identificador possui tokens disponíveis.
func (rl *RateLimiter) Allow(identifier string) bool {
	result := rl.Check(identifier)
	return result.Allowed
}

// Check verifica rate limit e retorna informações detalhadas para headers.
func (rl *RateLimiter) Check(identifier string) RateLimitResult {
	if rl == nil {
		return RateLimitResult{Allowed: true, Limit: 0, Remaining: 0}
	}

	if identifier == "" {
		identifier = "anonymous"
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, ok := rl.buckets[identifier]
	if !ok {
		bucket = &tokenBucket{
			tokens:     rl.capacity,
			lastRefill: now,
		}
		rl.buckets[identifier] = bucket
		bucket.tokens -= 1
		return RateLimitResult{
			Allowed:   true,
			Limit:     int(rl.capacity),
			Remaining: int(bucket.tokens),
			ResetAt:   now.Add(rl.window),
		}
	}

	elapsed := now.Sub(bucket.lastRefill).Seconds()
	if elapsed > 0 {
		bucket.tokens = math.Min(rl.capacity, bucket.tokens+elapsed*rl.fillRate)
		bucket.lastRefill = now
	}

	remaining := int(bucket.tokens)
	resetAt := now.Add(time.Duration((rl.capacity-bucket.tokens)/rl.fillRate) * time.Second)

	if bucket.tokens >= 1 {
		bucket.tokens -= 1
		remaining = int(bucket.tokens)
		return RateLimitResult{
			Allowed:   true,
			Limit:     int(rl.capacity),
			Remaining: remaining,
			ResetAt:   resetAt,
		}
	}

	return RateLimitResult{
		Allowed:   false,
		Limit:     int(rl.capacity),
		Remaining: 0,
		ResetAt:   resetAt,
	}
}

// RateLimitMiddleware aplica validação de rate limit para cada request.
func RateLimitMiddleware(limiter *RateLimiter, identifierFn IdentifierFunc) func(http.Handler) http.Handler {
	if limiter == nil {
		return func(next http.Handler) http.Handler { return next }
	}

	if identifierFn == nil {
		identifierFn = DefaultIdentifier
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := identifierFn(r)
			result := limiter.Check(id)

			// Adicionar headers de rate limit conforme spec 3.5
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(result.Limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))

			if !result.Allowed {
				retryAfter := int(time.Until(result.ResetAt).Seconds())
				if retryAfter < 1 {
					retryAfter = 1
				}
				log.Warn().Str("identifier", id).Msg("rate limit exceeded")
				util.RespondTooManyRequests(w, fmt.Sprintf("%d", retryAfter))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// IPIdentifier cria chave baseada apenas no IP.
func IPIdentifier(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return "ip:" + strings.TrimSpace(parts[0])
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "ip:" + r.RemoteAddr
	}

	return "ip:" + host
}

// UserIdentifier cria chave baseada no usuário autenticado.
func UserIdentifier(r *http.Request) string {
	if userID := GetUserIDFromContext(r); userID != "" {
		return "user:" + userID
	}
	return IPIdentifier(r)
}

// DefaultIdentifier cria chave baseada no usuário autenticado ou IP.
func DefaultIdentifier(r *http.Request) string {
	if userID := GetUserIDFromContext(r); userID != "" {
		return "user:" + userID
	}

	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return "ip:" + strings.TrimSpace(parts[0])
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "ip:" + r.RemoteAddr
	}

	return "ip:" + host
}
