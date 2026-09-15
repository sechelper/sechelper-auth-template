package application

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"sechelper-auth-template/api/internal/modules/authorization/domain"
	"sechelper-auth-template/api/internal/platform/session"
)

var (
	ErrUnauthorized = errors.New("session is not authenticated")
	ErrForbidden    = errors.New("permission denied")
	ErrDependency   = errors.New("authorization dependency failed")
)

type Cache interface {
	Get(context.Context, string) (domain.Context, bool, error)
	Put(context.Context, string, domain.Context, time.Time) error
	Delete(context.Context, string) error
	DeleteByApplication(context.Context, string) error
}

type MemoryCache struct {
	mu     sync.RWMutex
	values map[string]cachedContext
}

type cachedContext struct {
	value     domain.Context
	expiresAt time.Time
}

func NewMemoryCache() *MemoryCache { return &MemoryCache{values: map[string]cachedContext{}} }
func (c *MemoryCache) Get(_ context.Context, key string) (domain.Context, bool, error) {
	c.mu.RLock()
	value, ok := c.values[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(value.expiresAt) {
		if ok {
			_ = c.Delete(context.Background(), key)
		}
		return domain.Context{}, false, nil
	}
	return value.value, true, nil
}
func (c *MemoryCache) Put(_ context.Context, key string, value domain.Context, expiresAt time.Time) error {
	c.mu.Lock()
	c.values[key] = cachedContext{value: value, expiresAt: expiresAt}
	c.mu.Unlock()
	return nil
}
func (c *MemoryCache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	delete(c.values, key)
	c.mu.Unlock()
	return nil
}
func (c *MemoryCache) DeleteByApplication(_ context.Context, applicationCode string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, value := range c.values {
		if value.value.ApplicationCode == applicationCode {
			delete(c.values, key)
		}
	}
	return nil
}

type Service struct {
	sessions session.Store
	cache    Cache
}

func NewService(sessions session.Store, cache Cache) *Service {
	return &Service{sessions: sessions, cache: cache}
}
func (s *Service) Resolve(ctx context.Context, sessionID string) (domain.Context, error) {
	if sessionID == "" {
		return domain.Context{}, ErrUnauthorized
	}
	current, err := s.sessions.Get(ctx, sessionID)
	if err != nil {
		if errors.Is(err, session.ErrNotFound) {
			return domain.Context{}, ErrUnauthorized
		}
		return domain.Context{}, fmt.Errorf("%w: %v", ErrDependency, err)
	}
	if value, ok, err := s.cache.Get(ctx, sessionID); err != nil {
		return domain.Context{}, fmt.Errorf("%w: %v", ErrDependency, err)
	} else if ok {
		return value, nil
	}
	permissions := make(map[string]struct{}, len(current.Permissions))
	for _, code := range current.Permissions {
		permissions[code] = struct{}{}
	}
	value := domain.Context{Subject: current.Subject, PlatformUserUUID: current.PlatformUserUUID, ApplicationCode: current.ApplicationCode, Permissions: permissions}
	if err := s.cache.Put(ctx, sessionID, value, current.ExpiresAt); err != nil {
		return domain.Context{}, fmt.Errorf("%w: %v", ErrDependency, err)
	}
	return value, nil
}
func (s *Service) Require(ctx context.Context, sessionID, permission string) (domain.Context, error) {
	value, err := s.Resolve(ctx, sessionID)
	if err != nil {
		return domain.Context{}, err
	}
	if !value.HasPermission(permission) {
		return domain.Context{}, ErrForbidden
	}
	return value, nil
}
func (s *Service) InvalidateSession(sessionID string) {
	_ = s.cache.Delete(context.Background(), sessionID)
}
func (s *Service) InvalidateApplication(applicationCode string) {
	_ = s.cache.DeleteByApplication(context.Background(), applicationCode)
}
