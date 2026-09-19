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

func TestResolveCarriesExternalAndPlatformUserIdentifiers(t *testing.T) {
	store := session.NewMemoryStore()
	localID := "89cf8f29-9954-470d-b3de-8a37d20c6f44"
	if err := store.Create(context.Background(), session.Session{ID: "session-identifiers", Subject: "identity-subject", PlatformUserUUID: localID, ApplicationCode: "demo", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	value, err := NewService(store, NewMemoryCache()).Resolve(context.Background(), "session-identifiers")
	if err != nil {
		t.Fatal(err)
	}
	if value.Subject != "identity-subject" || value.PlatformUserUUID != localID {
		t.Fatalf("resolved identity = (%q, %q), want external subject and platform UUID", value.Subject, value.PlatformUserUUID)
	}
}

func TestRequireForSurfaceRejectsWrongSessionSurface(t *testing.T) {
	store := session.NewMemoryStore()
	if err := store.Create(context.Background(), session.Session{ID: "admin-session", Surface: "admin", Subject: "user-1", ApplicationCode: "admin", Permissions: []string{"admin:access"}, ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	service := NewService(store, NewMemoryCache())
	if _, err := service.RequireForSurface(context.Background(), "admin-session", "public", "admin:access"); err != ErrUnauthorized {
		t.Fatalf("RequireForSurface() error = %v, want ErrUnauthorized", err)
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
