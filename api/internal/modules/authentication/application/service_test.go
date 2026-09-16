package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"sechelper-auth-template/api/internal/platform/identity"
	"sechelper-auth-template/api/internal/platform/session"
)

type loginIdentity struct{ subject string }

func (f loginIdentity) AuthorizationURL(state, nonce, challenge, prompt string) (string, error) {
	return "https://idp.test/authorize", nil
}
func (f loginIdentity) EndSessionURL(idTokenHint, redirectURI, state string) (string, error) {
	return "https://idp.test/logout", nil
}

func TestLoginPromptIsBoundToStateSuffix(t *testing.T) {
	if got := LoginPrompt("random.none"); got != PromptNone {
		t.Fatalf("LoginPrompt(none) = %q", got)
	}
	if got := LoginPrompt("random.login"); got != PromptLogin {
		t.Fatalf("LoginPrompt(login) = %q", got)
	}
}
func (f loginIdentity) ExchangeCode(context.Context, string, string) (identity.TokenSet, error) {
	return identity.TokenSet{AccessToken: "access"}, nil
}
func (f loginIdentity) ValidateIDToken(context.Context, string, string) (identity.IDTokenClaims, error) {
	return identity.IDTokenClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: f.subject}}, nil
}
func (f loginIdentity) GetUserInfo(context.Context, string) (identity.UserInfo, error) {
	return identity.UserInfo{Subject: f.subject, Nickname: "Mina", Picture: "https://cdn.example/avatar.png", Claims: map[string]any{"sub": f.subject, "nickname": "Mina", "picture": "https://cdn.example/avatar.png", "locale": "zh-CN"}}, nil
}
func (f loginIdentity) GetAuthorization(context.Context, string) (identity.Authorization, error) {
	return identity.Authorization{Subject: f.subject, ApplicationCode: "app-1"}, nil
}
func (f loginIdentity) RefreshToken(context.Context, string) (identity.TokenSet, error) {
	return identity.TokenSet{}, errors.New("not used")
}
func (f loginIdentity) RevokeToken(context.Context, string) error { return nil }

type fixedUserResolver struct {
	issuer, subject, platformUserUUID string
}

func (r *fixedUserResolver) ResolveOrCreate(_ context.Context, issuer, subject string) (string, error) {
	r.issuer, r.subject = issuer, subject
	return r.platformUserUUID, nil
}

func TestCompleteLoginStoresResolvedPlatformUserUUID(t *testing.T) {
	ctx := context.Background()
	store := session.NewMemoryStore()
	if err := store.SaveLoginTransaction(ctx, session.LoginTransaction{State: "state", Nonce: "nonce", Verifier: "verifier", ExpiresAt: time.Now().Add(time.Minute)}); err != nil {
		t.Fatal(err)
	}
	resolver := &fixedUserResolver{platformUserUUID: "89cf8f29-9954-470d-b3de-8a37d20c6f44"}
	service := NewService(loginIdentity{subject: "identity-subject"}, store, store, nil, time.Hour, "app-1")
	service.SetUserResolver("https://identity.example", resolver)

	created, err := service.CompleteLogin(ctx, "state", "authorization-code")
	if err != nil {
		t.Fatal(err)
	}
	if resolver.issuer != "https://identity.example" || resolver.subject != "identity-subject" {
		t.Fatalf("resolved identity = (%q, %q), want configured issuer and verified subject", resolver.issuer, resolver.subject)
	}
	if created.Subject != "identity-subject" || created.PlatformUserUUID != resolver.platformUserUUID {
		t.Fatalf("session identity = (%q, %q), want external subject and local UUID", created.Subject, created.PlatformUserUUID)
	}
	if created.ProfileClaims["nickname"] != "Mina" || created.ProfileClaims["locale"] != "zh-CN" {
		t.Fatalf("session profile claims = %#v, want all returned UserInfo claims", created.ProfileClaims)
	}
	stored, err := store.Get(ctx, created.ID)
	if err != nil || stored.PlatformUserUUID != resolver.platformUserUUID || len(stored.ProfileClaims) != 0 {
		t.Fatalf("stored session identity/profile = (%q, %#v), err = %v", stored.PlatformUserUUID, stored.ProfileClaims, err)
	}
	current, err := service.Get(ctx, created.ID)
	if err != nil || current.ProfileClaims["picture"] != "https://cdn.example/avatar.png" {
		t.Fatalf("current session profile cache = %#v, err = %v", current.ProfileClaims, err)
	}
}

type rotationIdentity struct {
	refreshCalls []string
	revoked      []string
	responses    map[string]identity.TokenSet
	invalidGrant bool
}

func (f *rotationIdentity) AuthorizationURL(state, nonce, challenge, prompt string) (string, error) {
	return "https://idp.test/authorize", nil
}
func (f *rotationIdentity) EndSessionURL(idTokenHint, redirectURI, state string) (string, error) {
	return "https://idp.test/logout", nil
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
	if _, err := service.Logout(ctx, "session-1", "https://app.example/logout/callback", "logout-state"); err != nil {
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
