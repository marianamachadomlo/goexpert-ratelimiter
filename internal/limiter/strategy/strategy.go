package strategy

import (
	"context"
	"time"
)

type StorageStrategy interface {
	IsBlocked(ctx context.Context, key string) (bool, error)
	SetBlock(ctx context.Context, key string, duration time.Duration) error
	IncrementAndCheck(ctx context.Context, key string, limit int, window time.Duration) (allowed bool, count int64, err error)
}
