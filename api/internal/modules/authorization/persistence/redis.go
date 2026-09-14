package persistence

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"sechelper-auth-template/api/internal/modules/authorization/domain"
)

type RedisCache struct {
	client *redis.Client
	prefix string
}
type cacheValue struct {
	Subject         string   `json:"subject"`
	ApplicationCode string   `json:"applicationCode"`
	Permissions     []string `json:"permissions"`
}

func NewRedisCache(redisURL string) (*RedisCache, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(options)
	return &RedisCache{client: client, prefix: "authz:"}, nil
}
func (c *RedisCache) Ping(ctx context.Context) error { return c.client.Ping(ctx).Err() }
func (c *RedisCache) Close() error                   { return c.client.Close() }
func (c *RedisCache) key(sessionID string) string    { return c.prefix + sessionID }
func (c *RedisCache) appKey(applicationCode string) string {
	return c.prefix + "application:" + applicationCode
}
func (c *RedisCache) Get(ctx context.Context, sessionID string) (domain.Context, bool, error) {
	raw, err := c.client.Get(ctx, c.key(sessionID)).Bytes()
	if err == redis.Nil {
		return domain.Context{}, false, nil
	}
	if err != nil {
		return domain.Context{}, false, err
	}
	var value cacheValue
	if err := json.Unmarshal(raw, &value); err != nil {
		return domain.Context{}, false, err
	}
	permissions := make(map[string]struct{}, len(value.Permissions))
	for _, code := range value.Permissions {
		permissions[code] = struct{}{}
	}
	return domain.Context{Subject: value.Subject, ApplicationCode: value.ApplicationCode, Permissions: permissions}, true, nil
}
func (c *RedisCache) Put(ctx context.Context, sessionID string, value domain.Context, expiresAt time.Time) error {
	permissions := make([]string, 0, len(value.Permissions))
	for code := range value.Permissions {
		permissions = append(permissions, code)
	}
	raw, err := json.Marshal(cacheValue{Subject: value.Subject, ApplicationCode: value.ApplicationCode, Permissions: permissions})
	if err != nil {
		return err
	}
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}
	pipe := c.client.TxPipeline()
	pipe.Set(ctx, c.key(sessionID), raw, ttl)
	pipe.SAdd(ctx, c.appKey(value.ApplicationCode), sessionID)
	pipe.Expire(ctx, c.appKey(value.ApplicationCode), ttl)
	_, err = pipe.Exec(ctx)
	return err
}
func (c *RedisCache) Delete(ctx context.Context, sessionID string) error {
	raw, err := c.client.Get(ctx, c.key(sessionID)).Bytes()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return err
	}
	var value cacheValue
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	pipe := c.client.TxPipeline()
	pipe.Del(ctx, c.key(sessionID))
	pipe.SRem(ctx, c.appKey(value.ApplicationCode), sessionID)
	_, err = pipe.Exec(ctx)
	return err
}
func (c *RedisCache) DeleteByApplication(ctx context.Context, applicationCode string) error {
	key := c.appKey(applicationCode)
	sessions, err := c.client.SMembers(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	pipe := c.client.TxPipeline()
	for _, sessionID := range sessions {
		pipe.Del(ctx, c.key(sessionID))
	}
	pipe.Del(ctx, key)
	_, err = pipe.Exec(ctx)
	return err
}
