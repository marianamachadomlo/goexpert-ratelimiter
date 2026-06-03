package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/mariana/rate-limiter/internal/limiter/strategy"
)

const (
	blockKeyPrefix   = "block:"
	counterKeyPrefix = "rate:"
)

type Storage struct {
	client *goredis.Client
}

func New(addr, password string, db int) (*Storage, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return &Storage{client: client}, nil
}

func (s *Storage) IsBlocked(ctx context.Context, key string) (bool, error) {
	result, err := s.client.Exists(ctx, blockKeyPrefix+key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (s *Storage) SetBlock(ctx context.Context, key string, duration time.Duration) error {
	return s.client.Set(ctx, blockKeyPrefix+key, "1", duration).Err()
}

func (s *Storage) IncrementAndCheck(ctx context.Context, key string, limit int, window time.Duration) (bool, int64, error) {
	windowKey := fmt.Sprintf("%s%s:%d", counterKeyPrefix, key, time.Now().Unix()/int64(window.Seconds()))
	if window.Seconds() < 1 {
		windowKey = fmt.Sprintf("%s%s:%d", counterKeyPrefix, key, time.Now().UnixNano()/window.Nanoseconds())
	}

	count, err := s.client.Incr(ctx, windowKey).Result()
	if err != nil {
		return false, 0, err
	}

	if count == 1 {
		if err := s.client.Expire(ctx, windowKey, window+time.Second).Err(); err != nil {
			return false, count, err
		}
	}

	return count <= int64(limit), count, nil
}

func (s *Storage) Close() error {
	return s.client.Close()
}

var _ strategy.StorageStrategy = (*Storage)(nil)
