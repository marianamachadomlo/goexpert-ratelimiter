package limiter

import (
	"context"
	"errors"
	"time"

	"github.com/mariana/rate-limiter/internal/limiter/strategy"
)

const BlockedMessage = "you have reached the maximum number of requests or actions allowed within a certain time frame"

var ErrRateLimitExceeded = errors.New("rate limit exceeded")

type Config struct {
	IPLimit       int
	TokenLimit    func(token string) int
	BlockDuration time.Duration
	RateWindow    time.Duration
}

type Service struct {
	storage strategy.StorageStrategy
	config  Config
}

func New(storage strategy.StorageStrategy, config Config) *Service {
	return &Service{
		storage: storage,
		config:  config,
	}
}

type Request struct {
	IP    string
	Token string
}

type Result struct {
	Allowed   bool
	Blocked   bool
	Limit     int
	Exceeded  bool
	Scope     string
	Identifier string
}

func (s *Service) Allow(ctx context.Context, req Request) (Result, error) {
	scope, identifier, limit := s.resolveScope(req)

	result := Result{
		Allowed:    true,
		Limit:      limit,
		Scope:      scope,
		Identifier: identifier,
	}

	blocked, err := s.storage.IsBlocked(ctx, identifier)
	if err != nil {
		return Result{}, err
	}
	if blocked {
		result.Allowed = false
		result.Blocked = true
		return result, ErrRateLimitExceeded
	}

	allowed, _, err := s.storage.IncrementAndCheck(ctx, identifier, limit, s.config.RateWindow)
	if err != nil {
		return Result{}, err
	}
	if !allowed {
		if blockErr := s.storage.SetBlock(ctx, identifier, s.config.BlockDuration); blockErr != nil {
			return Result{}, blockErr
		}
		result.Allowed = false
		result.Exceeded = true
		return result, ErrRateLimitExceeded
	}

	return result, nil
}

func (s *Service) resolveScope(req Request) (scope, identifier string, limit int) {
	if req.Token != "" {
		return "token", "token:" + req.Token, s.config.TokenLimit(req.Token)
	}
	return "ip", "ip:" + req.IP, s.config.IPLimit
}
