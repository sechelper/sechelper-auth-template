package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"sechelper-auth-template/api/internal/platform/identity"
	"sechelper-auth-template/api/internal/platform/session"
)

type rotationIdentity struct {
	refreshCalls []string
	revoked      []string
	responses    map[string]identity.TokenSet
	invalidGrant bool
}

func (f *rotationIdentity) AuthorizationURL(state, nonce, challenge string) (string, error) {
	return "https://idp.test/authorize", nil
}
func (f *rotationIdentity) ExchangeCode(context.Context, string, string) (identity.TokenSet, error) {
	return identity.TokenSet{}, errors.New("not used in refresh test")
}
func (f *rotationIdentity) ValidateIDToken(context.Context, string, string) (identity.IDTokenClaims, error) {
	return identity.IDTokenClaims{}, errors.New("not used in refresh test")
}
func (f *rotationIdentity) GetUserInfo(context.Context, string) (identity.UserInfo, error) {
	return identity.UserInfo{}, errors.New("not used in refresh test")
}
func (f *rotationIdentity) GetAuthorization(context.Context, string) (identity.Authorization, error) {
	return identity.Authorization{}, errors.New("not used in refresh test")
}
func (f *rotationIdentity) RefreshToken(_ context.Context, value string) (identity.TokenSet, error) {
	f.refreshCalls = append(f.refreshCalls, value)
	if f.invalidGrant {
		return identity.TokenSet{}, &identity.TokenError{StatusCode: 400, ErrorCode: "invalid_grant"}
	}
	response, ok := f.responses[value]
	if !ok {
		return identity.TokenSet{}, errors.New("refresh token reuse rejected")
	}
	return response, nil
}
func (f *rotationIdentity) RevokeToken(_ context.Context, value string) error {
	f.revoked = append(f.revoked, value)
	return nil
}

func TestRefreshRotatesTokenAndExtendsSessionWithoutWaitingForExpiry(t *testing.T) {
	ctx := context.Background()
	store := session.NewMemoryStore()
	protector, err := session.NewTokenProtector("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	if err != nil {
		t.Fatal(err)
	}
	oldToken := "refresh-token-1"
	oldCiphertext, err := protector.Encrypt(oldToken)
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now().Add(time.Minute)
	if err := store.Create(ctx, session.Session{ID: "session-1", Subject: "user-1", ApplicationCode: "app-1", RefreshTokenCiphertext: oldCiphertext, ExpiresAt: started}); err != nil {
		t.Fatal(err)
	}
	fake := &rotationIdentity{responses: map[string]identity.TokenSet{
		oldToken: {AccessToken: "access-2", RefreshToken: "refresh-token-2", ExpiresIn: 900},
	}}
	service := NewService(fake, store, store, protector, 8*time.Hour, "app-1")

	rotated, err := service.Refresh(ctx, "session-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(fake.refreshCalls) != 1 || fake.refreshCalls[0] != oldToken {
		t.Fatalf("RefreshToken calls = %#v, want the original token once", fake.refreshCalls)
	}
	if !rotated.ExpiresAt.After(time.Now()) {
		t.Fatalf("rotated session expiry = %s, want future expiry", rotated.ExpiresAt)
	}
	if rotated.RefreshTokenCiphertext == oldCiphertext {
		t.Fatal("refresh token ciphertext did not rotate")
	}
	newToken, err := protector.Decrypt(rotated.RefreshTokenCiphertext)
	if err != nil || newToken != "refresh-token-2" {
		t.Fatalf("stored refresh token = %q, err = %v", newToken, err)
	}
	delete(fake.responses, oldToken)
	if _, err := fake.RefreshToken(ctx, oldToken); err == nil {
		t.Fatal("mock identity provider accepted reuse of the old refresh token")
	}
	if err := service.Logout(ctx, "session-1"); err != nil {
		t.Fatal(err)
	}
	if len(fake.revoked) != 1 || fake.revoked[0] != "refresh-token-2" {
		t.Fatalf("revoked tokens = %#v, want rotated token", fake.revoked)
	}
}

func TestRefreshInvalidGrantRevokesSessionAsReuseDetection(t *testing.T) {
	ctx := context.Background()
	store := session.NewMemoryStore()
	protector, err := session.NewTokenProtector("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, err := protector.Encrypt("rotated-away-token")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Create(ctx, session.Session{ID: "session-reuse", Subject: "user-1", RefreshTokenCiphertext: ciphertext, ExpiresAt: time.Now().Add(time.Minute)}); err != nil {
		t.Fatal(err)
	}
	fake := &rotationIdentity{responses: map[string]identity.TokenSet{}, invalidGrant: true}
	service := NewService(fake, store, store, protector, time.Hour, "app-1")
	if _, err := service.Refresh(ctx, "session-reuse"); !errors.Is(err, session.ErrRefreshReuse) {
		t.Fatalf("Refresh() error = %v, want ErrRefreshReuse", err)
	}
	if _, err := store.Get(ctx, "session-reuse"); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("session after reuse detection = %v, want revoked/not found", err)
	}
}
