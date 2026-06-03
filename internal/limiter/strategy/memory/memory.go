package memory

import (
	"context"
	"sync"
	"time"

	"github.com/mariana/rate-limiter/internal/limiter/strategy"
)

type counterEntry struct {
	count     int64
	expiresAt time.Time
}

type blockEntry struct {
	expiresAt time.Time
}

type Storage struct {
	mu       sync.Mutex
	counters map[string]counterEntry
	blocks   map[string]blockEntry
}

func New() *Storage {
	return &Storage{
		counters: make(map[string]counterEntry),
		blocks:   make(map[string]blockEntry),
	}
}

func (s *Storage) IsBlocked(_ context.Context, key string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanupExpired(time.Now())

	entry, ok := s.blocks[key]
	if !ok {
		return false, nil
	}
	return time.Now().Before(entry.expiresAt), nil
}

func (s *Storage) SetBlock(_ context.Context, key string, duration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.blocks[key] = blockEntry{expiresAt: time.Now().Add(duration)}
	return nil
}

func (s *Storage) IncrementAndCheck(_ context.Context, key string, limit int, window time.Duration) (bool, int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	s.cleanupExpired(now)

	entry, ok := s.counters[key]
	if !ok || now.After(entry.expiresAt) {
		entry = counterEntry{count: 0, expiresAt: now.Add(window)}
	}

	entry.count++
	s.counters[key] = entry

	return entry.count <= int64(limit), entry.count, nil
}

func (s *Storage) cleanupExpired(now time.Time) {
	for key, entry := range s.counters {
		if now.After(entry.expiresAt) {
			delete(s.counters, key)
		}
	}

	for key, entry := range s.blocks {
		if !now.Before(entry.expiresAt) {
			delete(s.blocks, key)
		}
	}
}

var _ strategy.StorageStrategy = (*Storage)(nil)
