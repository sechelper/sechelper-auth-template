package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type Limiter interface {
	Allow(context.Context, string, int, time.Duration) (bool, time.Duration, error)
}

type bucket struct {
	count   int
	resetAt time.Time
}
type Memory struct {
	mu     sync.Mutex
	values map[string]bucket
}

func NewMemory() *Memory { return &Memory{values: map[string]bucket{}} }
func (l *Memory) Allow(_ context.Context, key string, limit int, window time.Duration) (bool, time.Duration, error) {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	b := l.values[key]
	if b.resetAt.IsZero() || !now.Before(b.resetAt) {
		b = bucket{resetAt: now.Add(window)}
	}
	b.count++
	l.values[key] = b
	remaining := time.Until(b.resetAt)
	if remaining < 0 {
		remaining = 0
	}
	return b.count <= limit, remaining, nil
}

type Redis struct{ client *redis.Client }

func NewRedis(rawURL string) (*Redis, error) {
	options, err := redis.ParseURL(rawURL)
	if err != nil {
		return nil, err
	}
	return &Redis{client: redis.NewClient(options)}, nil
}
func (l *Redis) Ping(ctx context.Context) error { return l.client.Ping(ctx).Err() }
func (l *Redis) Close() error                   { return l.client.Close() }
func (l *Redis) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, time.Duration, error) {
	windowSeconds := int64(window / time.Second)
	if windowSeconds < 1 {
		windowSeconds = 1
	}
	windowID := time.Now().Unix() / windowSeconds
	redisKey := fmt.Sprintf("ratelimit:%s:%d", key, windowID)
	count, err := l.client.Incr(ctx, redisKey).Result()
	if err != nil {
		return false, 0, err
	}
	if count == 1 {
		_ = l.client.Expire(ctx, redisKey, time.Duration(windowSeconds)*time.Second).Err()
	}
	remaining := time.Duration(windowSeconds-(time.Now().Unix()%windowSeconds)) * time.Second
	return count <= int64(limit), remaining, nil
}
