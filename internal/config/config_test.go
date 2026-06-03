package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/mariana/rate-limiter/internal/config"
)

func TestLoadReadsEnvironmentVariables(t *testing.T) {
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("RATE_LIMIT_IP", "15")
	t.Setenv("RATE_LIMIT_TOKEN_DEFAULT", "200")
	t.Setenv("BLOCK_DURATION", "2m")
	t.Setenv("REDIS_ADDR", "localhost:6379")
	t.Setenv("TOKEN_LIMITS", `{"vip-token":100}`)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if cfg.ServerPort != "9090" {
		t.Fatalf("unexpected server port: %s", cfg.ServerPort)
	}
	if cfg.IPLimit != 15 {
		t.Fatalf("unexpected ip limit: %d", cfg.IPLimit)
	}
	if cfg.DefaultToken != 200 {
		t.Fatalf("unexpected default token limit: %d", cfg.DefaultToken)
	}
	if cfg.BlockDuration != 2*time.Minute {
		t.Fatalf("unexpected block duration: %s", cfg.BlockDuration)
	}
	if cfg.TokenLimit("vip-token") != 100 {
		t.Fatalf("unexpected token limit: %d", cfg.TokenLimit("vip-token"))
	}
}

func TestLoadRejectsInvalidTokenLimits(t *testing.T) {
	t.Setenv("TOKEN_LIMITS", `{invalid`)
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for invalid TOKEN_LIMITS")
	}
}

func unsetEnv(keys ...string) {
	for _, key := range keys {
		_ = os.Unsetenv(key)
	}
}

func TestLoadUsesDefaultsWhenEnvMissing(t *testing.T) {
	unsetEnv(
		"SERVER_PORT",
		"RATE_LIMIT_IP",
		"RATE_LIMIT_TOKEN_DEFAULT",
		"BLOCK_DURATION",
		"REDIS_ADDR",
		"TOKEN_LIMITS",
	)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if cfg.IPLimit != config.DefaultIPLimit {
		t.Fatalf("expected default ip limit")
	}
}
