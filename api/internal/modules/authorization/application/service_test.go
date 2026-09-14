package application

import (
	"context"
	"testing"
	"time"

	"sechelper-auth-template/api/internal/platform/session"
)

func TestServiceRequiresPermission(t *testing.T) {
	store := session.NewMemoryStore()
	if err := store.Create(context.Background(), session.Session{ID: "session-1", Subject: "user-1", ApplicationCode: "demo", Permissions: []string{"order:read"}, ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	service := NewService(store, NewMemoryCache())
	if _, err := service.Require(context.Background(), "session-1", "order:read"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Require(context.Background(), "session-1", "order:approve"); err != ErrForbidden {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestServiceRejectsMissingSession(t *testing.T) {
	service := NewService(session.NewMemoryStore(), NewMemoryCache())
	if _, err := service.Resolve(context.Background(), "missing"); err != ErrUnauthorized {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestServiceRejectsRevokedSessionEvenWhenCached(t *testing.T) {
	store := session.NewMemoryStore()
	if err := store.Create(context.Background(), session.Session{ID: "session-2", Subject: "user-2", ApplicationCode: "demo", Permissions: []string{"order:read"}, ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	service := NewService(store, NewMemoryCache())
	if _, err := service.Resolve(context.Background(), "session-2"); err != nil {
		t.Fatal(err)
	}
	if err := store.Revoke(context.Background(), "session-2"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Resolve(context.Background(), "session-2"); err != ErrUnauthorized {
		t.Fatalf("expected revoked session to be unauthorized, got %v", err)
	}
}
