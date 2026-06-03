package limiter_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mariana/rate-limiter/internal/limiter"
	"github.com/mariana/rate-limiter/internal/limiter/strategy/memory"
)

func TestServiceBlocksAfterIPLimitExceeded(t *testing.T) {
	storage := memory.New()
	service := limiter.New(storage, limiter.Config{
		IPLimit:       3,
		TokenLimit:    func(_ string) int { return 100 },
		BlockDuration: time.Minute,
		RateWindow:    time.Second,
	})

	ctx := context.Background()
	req := limiter.Request{IP: "127.0.0.1"}

	for i := 0; i < 3; i++ {
		result, err := service.Allow(ctx, req)
		if err != nil {
			t.Fatalf("request %d failed: %v", i+1, err)
		}
		if !result.Allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	result, err := service.Allow(ctx, req)
	if !errors.Is(err, limiter.ErrRateLimitExceeded) {
		t.Fatalf("expected ErrRateLimitExceeded, got %v", err)
	}
	if result.Allowed {
		t.Fatal("fourth request should be rejected")
	}

	result, err = service.Allow(ctx, req)
	if !errors.Is(err, limiter.ErrRateLimitExceeded) {
		t.Fatalf("expected ErrRateLimitExceeded, got %v", err)
	}
	if result.Allowed || !result.Blocked {
		t.Fatal("blocked IP should keep receiving rejections")
	}
}

func TestTokenPrecedenceOverIPLimit(t *testing.T) {
	storage := memory.New()
	service := limiter.New(storage, limiter.Config{
		IPLimit: 2,
		TokenLimit: func(token string) int {
			if token == "premium-token" {
				return 5
			}
			return 2
		},
		BlockDuration: time.Minute,
		RateWindow:    time.Second,
	})

	ctx := context.Background()
	ipReq := limiter.Request{IP: "10.0.0.1"}

	for i := 0; i < 2; i++ {
		result, err := service.Allow(ctx, ipReq)
		if err != nil || !result.Allowed {
			t.Fatalf("ip request %d should pass: %+v err=%v", i+1, result, err)
		}
	}

	_, err := service.Allow(ctx, ipReq)
	if !errors.Is(err, limiter.ErrRateLimitExceeded) {
		t.Fatalf("expected IP limit exceeded, got %v", err)
	}

	tokenReq := limiter.Request{IP: "10.0.0.1", Token: "premium-token"}
	for i := 0; i < 5; i++ {
		result, err := service.Allow(ctx, tokenReq)
		if err != nil || !result.Allowed {
			t.Fatalf("token request %d should pass with higher limit: %+v err=%v", i+1, result, err)
		}
		if result.Scope != "token" {
			t.Fatalf("expected token scope, got %s", result.Scope)
		}
	}

	_, err = service.Allow(ctx, tokenReq)
	if !errors.Is(err, limiter.ErrRateLimitExceeded) {
		t.Fatalf("expected token limit exceeded, got %v", err)
	}
}

func TestTokenUsesConfiguredLimitInsteadOfDefault(t *testing.T) {
	storage := memory.New()
	service := limiter.New(storage, limiter.Config{
		IPLimit: 1,
		TokenLimit: func(token string) int {
			limits := map[string]int{
				"vip-token": 10,
			}
			if limit, ok := limits[token]; ok {
				return limit
			}
			return 1
		},
		BlockDuration: time.Minute,
		RateWindow:    time.Second,
	})

	ctx := context.Background()
	req := limiter.Request{IP: "192.168.0.10", Token: "vip-token"}

	for i := 0; i < 10; i++ {
		result, err := service.Allow(ctx, req)
		if err != nil || !result.Allowed || result.Limit != 10 {
			t.Fatalf("request %d failed: limit=%d err=%v", i+1, result.Limit, err)
		}
	}
}
