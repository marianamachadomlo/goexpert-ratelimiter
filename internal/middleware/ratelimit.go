package middleware

import (
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/mariana/rate-limiter/internal/limiter"
)

const APIKeyHeader = "API_KEY"

type RateLimiter struct {
	service *limiter.Service
}

func New(service *limiter.Service) *RateLimiter {
	return &RateLimiter{service: service}
}

func (m *RateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result, err := m.service.Allow(r.Context(), limiter.Request{
			IP:    clientIP(r),
			Token: strings.TrimSpace(r.Header.Get(APIKeyHeader)),
		})
		if err != nil && !errors.Is(err, limiter.ErrRateLimitExceeded) {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if !result.Allowed {
			writeTooManyRequests(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeTooManyRequests(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusTooManyRequests)
	_, _ = w.Write([]byte(limiter.BlockedMessage))
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}

	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
