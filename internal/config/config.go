package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	DefaultServerPort      = "8080"
	DefaultIPLimit         = 10
	DefaultTokenLimit      = 100
	DefaultBlockDuration   = 5 * time.Minute
	DefaultRedisAddr       = "redis:6379"
	DefaultRedisPassword   = ""
	DefaultRedisDB         = 0
	DefaultRateWindow      = time.Second
)

type Config struct {
	ServerPort    string
	IPLimit       int
	DefaultToken  int
	TokenLimits   map[string]int
	BlockDuration time.Duration
	RateWindow    time.Duration
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:    getEnv("SERVER_PORT", DefaultServerPort),
		IPLimit:       getEnvInt("RATE_LIMIT_IP", DefaultIPLimit),
		DefaultToken:  getEnvInt("RATE_LIMIT_TOKEN_DEFAULT", DefaultTokenLimit),
		BlockDuration: getEnvDuration("BLOCK_DURATION", DefaultBlockDuration),
		RateWindow:    getEnvDuration("RATE_WINDOW", DefaultRateWindow),
		RedisAddr:     getEnv("REDIS_ADDR", DefaultRedisAddr),
		RedisPassword: getEnv("REDIS_PASSWORD", DefaultRedisPassword),
		RedisDB:       getEnvInt("REDIS_DB", DefaultRedisDB),
	}

	tokenLimits, err := parseTokenLimits(getEnv("TOKEN_LIMITS", ""))
	if err != nil {
		return nil, fmt.Errorf("invalid TOKEN_LIMITS: %w", err)
	}
	cfg.TokenLimits = tokenLimits

	if cfg.IPLimit <= 0 {
		return nil, fmt.Errorf("RATE_LIMIT_IP must be greater than zero")
	}
	if cfg.DefaultToken <= 0 {
		return nil, fmt.Errorf("RATE_LIMIT_TOKEN_DEFAULT must be greater than zero")
	}
	if cfg.BlockDuration <= 0 {
		return nil, fmt.Errorf("BLOCK_DURATION must be greater than zero")
	}
	if cfg.RateWindow <= 0 {
		return nil, fmt.Errorf("RATE_WINDOW must be greater than zero")
	}

	return cfg, nil
}

func (c *Config) TokenLimit(token string) int {
	if limit, ok := c.TokenLimits[token]; ok {
		return limit
	}
	return c.DefaultToken
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func parseTokenLimits(raw string) (map[string]int, error) {
	if raw == "" {
		return map[string]int{}, nil
	}

	limits := make(map[string]int)
	if err := json.Unmarshal([]byte(raw), &limits); err != nil {
		return nil, err
	}

	for token, limit := range limits {
		if limit <= 0 {
			return nil, fmt.Errorf("limit for token %q must be greater than zero", token)
		}
	}

	return limits, nil
}
