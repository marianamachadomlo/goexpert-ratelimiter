package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mariana/rate-limiter/internal/limiter"
	"github.com/mariana/rate-limiter/internal/limiter/strategy/memory"
	"github.com/mariana/rate-limiter/internal/middleware"
)

func TestMiddlewareReturns429WithExactMessage(t *testing.T) {
	storage := memory.New()
	service := limiter.New(storage, limiter.Config{
		IPLimit:       1,
		TokenLimit:    func(_ string) int { return 1 },
		BlockDuration: time.Minute,
		RateWindow:    time.Second,
	})

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.New(service).Handler(next)

	first := httptest.NewRequest(http.MethodGet, "/", nil)
	first.RemoteAddr = "127.0.0.1:12345"
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, first)
	if recorder.Code != http.StatusOK {
		t.Fatalf("first request expected 200, got %d", recorder.Code)
	}

	second := httptest.NewRequest(http.MethodGet, "/", nil)
	second.RemoteAddr = "127.0.0.1:12345"
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, second)
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("second request expected 429, got %d", recorder.Code)
	}
	if body := strings.TrimSpace(recorder.Body.String()); body != limiter.BlockedMessage {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestMiddlewareReturnsExactMessageWhenAlreadyBlocked(t *testing.T) {
	storage := memory.New()
	service := limiter.New(storage, limiter.Config{
		IPLimit:       1,
		TokenLimit:    func(_ string) int { return 1 },
		BlockDuration: time.Minute,
		RateWindow:    time.Second,
	})

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.New(service).Handler(next)

	first := httptest.NewRequest(http.MethodGet, "/", nil)
	first.RemoteAddr = "127.0.0.1:12345"
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, first)

	second := httptest.NewRequest(http.MethodGet, "/", nil)
	second.RemoteAddr = "127.0.0.1:12345"
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, second)
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("second request expected 429, got %d", recorder.Code)
	}

	third := httptest.NewRequest(http.MethodGet, "/", nil)
	third.RemoteAddr = "127.0.0.1:12345"
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, third)
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("blocked request expected 429, got %d", recorder.Code)
	}
	if body := strings.TrimSpace(recorder.Body.String()); body != limiter.BlockedMessage {
		t.Fatalf("unexpected body while blocked: %q", body)
	}
}

func TestMiddlewareUsesAPIKeyHeader(t *testing.T) {
	storage := memory.New()
	service := limiter.New(storage, limiter.Config{
		IPLimit: 1,
		TokenLimit: func(token string) int {
			if token == "secret" {
				return 3
			}
			return 1
		},
		BlockDuration: time.Minute,
		RateWindow:    time.Second,
	})

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.New(service).Handler(next)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set(middleware.APIKeyHeader, "secret")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		if recorder.Code != http.StatusOK {
			t.Fatalf("token request %d expected 200, got %d", i+1, recorder.Code)
		}
	}
}
