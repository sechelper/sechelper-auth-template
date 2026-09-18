package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	plataudit "sechelper-auth-template/api/internal/platform/audit"
	"sechelper-auth-template/api/internal/platform/httpkit"
	"sechelper-auth-template/api/internal/platform/identity"
	"sechelper-auth-template/api/internal/platform/session"
)

var (
	ErrInvalidState       = errors.New("invalid authentication state")
	ErrRefreshUnavailable = errors.New("session refresh is unavailable")
)

type LoginState struct {
	State, Nonce, Verifier string
	ExpiresAt              time.Time
}

const (
	PromptNone  = "none"
	PromptLogin = "login"
)

type IdentityProvider interface {
	AuthorizationURL(state, nonce, challenge, prompt string) (string, error)
	EndSessionURL(idTokenHint, redirectURI, state string) (string, error)
	ExchangeCode(context.Context, string, string) (identity.TokenSet, error)
	ValidateIDToken(context.Context, string, string) (identity.IDTokenClaims, error)
	GetUserInfo(context.Context, string) (identity.UserInfo, error)
	GetAuthorization(context.Context, string) (identity.Authorization, error)
	RefreshToken(context.Context, string) (identity.TokenSet, error)
	RevokeToken(context.Context, string) error
}
type UserResolver interface {
	ResolveOrCreate(context.Context, string, string) (string, error)
}
type Service struct {
	identity        IdentityProvider
	sessions        session.Store
	loginStates     session.LoginTransactionStore
	protector       *session.TokenProtector
	refreshMu       sync.Mutex
	onRemoteRevoke  func(error)
	ttl             time.Duration
	applicationCode string
	identityIssuer  string
	userResolver    UserResolver
	auditRecorder   plataudit.Recorder
	profileMu       sync.RWMutex
	profiles        map[string]cachedProfile
}

type cachedProfile struct {
	claims    map[string]any
	expiresAt time.Time
}

type LogoutResult struct {
	EndSessionURL string
}

func (s *Service) SetAuditRecorder(recorder plataudit.Recorder) { s.auditRecorder = recorder }
func (s *Service) SetUserResolver(issuer string, resolver UserResolver) {
	s.identityIssuer, s.userResolver = issuer, resolver
}

func NewService(client IdentityProvider, sessions session.Store, loginStates session.LoginTransactionStore, protector *session.TokenProtector, ttl time.Duration, applicationCode string) *Service {
	return &Service{identity: client, sessions: sessions, loginStates: loginStates, protector: protector, ttl: ttl, applicationCode: applicationCode, profiles: make(map[string]cachedProfile)}
}
func (s *Service) BeginLogin(ctx context.Context, prompt string) (string, error) {
	if prompt != PromptNone && prompt != PromptLogin {
		return "", errors.New("unsupported login prompt")
	}
	randomState, err := randomString(32)
	if err != nil {
		return "", err
	}
	nonce, err := randomString(32)
	if err != nil {
		return "", err
	}
	verifier, err := randomString(48)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	// Keep the prompt bound to the server-validated state without changing the
	// existing login-transaction schema. The random portion remains the value
	// protected by the transaction store's hash.
	state := randomState + "." + prompt
	if err := s.loginStates.SaveLoginTransaction(ctx, session.LoginTransaction{State: state, Nonce: nonce, Verifier: verifier, ExpiresAt: time.Now().Add(5 * time.Minute)}); err != nil {
		return "", err
	}
	return s.identity.AuthorizationURL(state, nonce, challenge, prompt)
}

func LoginPrompt(state string) string {
	if strings.HasSuffix(state, "."+PromptNone) {
		return PromptNone
	}
	return PromptLogin
}
func (s *Service) CompleteLogin(ctx context.Context, state, code string) (result session.Session, resultErr error) {
	defer func() {
		if s.auditRecorder == nil {
			return
		}
		id, err := session.NewID()
		if err != nil {
			return
		}
		eventType := "AUTH_LOGIN_SUCCESS"
		outcome := "success"
		if resultErr != nil {
			eventType = "AUTH_LOGIN_FAILED"
			outcome = "failure"
		}
		_ = s.auditRecorder.Record(ctx, plataudit.Event{ID: "evt-" + id, EventType: eventType, Outcome: outcome, ActorSubject: result.Subject, ApplicationCode: s.applicationCode, ResourceType: "auth_flow", ResourceID: s.applicationCode, Action: "login", RequestID: httpkit.RequestIDFromContext(ctx), Source: "api"})
	}()
	value, err := s.loginStates.ConsumeLoginTransaction(ctx, state)
	if err != nil {
		return session.Session{}, ErrInvalidState
	}
	tokens, err := s.identity.ExchangeCode(ctx, code, value.Verifier)
	if err != nil {
		return session.Session{}, err
	}
	idToken, err := s.identity.ValidateIDToken(ctx, tokens.IDToken, value.Nonce)
	if err != nil {
		return session.Session{}, err
	}
	user, err := s.identity.GetUserInfo(ctx, tokens.AccessToken)
	if err != nil {
		return session.Session{}, err
	}
	auth, err := s.identity.GetAuthorization(ctx, tokens.AccessToken)
	if err != nil {
		return session.Session{}, err
	}
	if auth.ApplicationCode != s.applicationCode || auth.Subject != user.Subject || idToken.Subject != user.Subject {
		return session.Session{}, errors.New("identity application mismatch")
	}
	if s.userResolver == nil || s.identityIssuer == "" {
		return session.Session{}, errors.New("local user identity resolver is not configured")
	}
	platformUserUUID, err := s.userResolver.ResolveOrCreate(ctx, s.identityIssuer, user.Subject)
	if err != nil {
		return session.Session{}, fmt.Errorf("resolve local user identity: %w", err)
	}
	id, err := session.NewID()
	if err != nil {
		return session.Session{}, err
	}
	ttl := s.ttl
	if tokens.ExpiresIn > 0 && time.Duration(tokens.ExpiresIn)*time.Second < ttl {
		ttl = time.Duration(tokens.ExpiresIn) * time.Second
	}
	permissions := make([]string, 0, len(auth.Permissions.Items))
	for _, permission := range auth.Permissions.Items {
		permissions = append(permissions, permission.Code)
	}
	refreshToken := ""
	if tokens.RefreshToken != "" {
		refreshToken, err = s.protector.Encrypt(tokens.RefreshToken)
		if err != nil {
			return session.Session{}, err
		}
	}
	encryptedIDToken := ""
	if tokens.IDToken != "" {
		if s.protector == nil {
			return session.Session{}, errors.New("session token protector is not configured")
		}
		encryptedIDToken, err = s.protector.Encrypt(tokens.IDToken)
		if err != nil {
			return session.Session{}, err
		}
	}
	result = session.Session{ID: id, Subject: user.Subject, PlatformUserUUID: platformUserUUID, Email: user.Email, ApplicationCode: auth.ApplicationCode, Permissions: permissions, ProfileClaims: profileClaims(user), RefreshTokenCiphertext: refreshToken, IDTokenCiphertext: encryptedIDToken, ExpiresAt: time.Now().Add(ttl)}
	persisted := result
	persisted.ProfileClaims = nil
	if err := s.sessions.Create(ctx, persisted); err != nil {
		return session.Session{}, err
	}
	s.storeProfile(result.ID, result.ProfileClaims, result.ExpiresAt)
	return result, nil
}

func profileClaims(user identity.UserInfo) map[string]any {
	if user.Claims != nil {
		return user.Claims
	}
	claims := map[string]any{}
	for key, value := range map[string]string{
		"sub": user.Subject, "preferred_username": user.Username, "name": user.Name,
		"nickname": user.Nickname, "picture": user.Picture, "email": user.Email,
	} {
		if value != "" {
			claims[key] = value
		}
	}
	return claims
}
func (s *Service) Refresh(ctx context.Context, id string) (result session.Session, resultErr error) {
	defer func() {
		if resultErr == nil || s.auditRecorder == nil {
			return
		}
		eventID, err := session.NewID()
		if err != nil {
			return
		}
		_ = s.auditRecorder.Record(ctx, plataudit.Event{ID: "evt-" + eventID, EventType: "AUTH_SESSION_REFRESH_FAILED", Outcome: "failure", ApplicationCode: s.applicationCode, ResourceType: "session", ResourceID: id, Action: "refresh", RequestID: httpkit.RequestIDFromContext(ctx), Source: "api"})
	}()
	if rotator, ok := s.sessions.(session.RefreshRotator); ok {
		value, err := rotator.RotateRefreshToken(ctx, id, func(ctx context.Context, current session.Session) (session.Session, error) {
			current.ProfileClaims = s.loadProfile(id, current.ExpiresAt)
			return s.rotateSession(ctx, current)
		})
		if err == nil {
			s.storeProfile(id, value.ProfileClaims, value.ExpiresAt)
		}
		return value, err
	}
	// Compatibility path for custom stores. Production uses the durable
	// RotateRefreshToken implementation above.
	s.refreshMu.Lock()
	defer s.refreshMu.Unlock()
	current, err := s.sessions.Get(ctx, id)
	if err != nil {
		return session.Session{}, err
	}
	current.ProfileClaims = s.loadProfile(id, current.ExpiresAt)
	if current.RefreshTokenCiphertext == "" {
		return session.Session{}, ErrRefreshUnavailable
	}
	current, err = s.rotateSession(ctx, current)
	if err != nil {
		return session.Session{}, err
	}
	if err := s.sessions.Update(ctx, current); err != nil {
		return session.Session{}, err
	}
	s.storeProfile(id, current.ProfileClaims, current.ExpiresAt)
	return current, nil
}
func (s *Service) rotateSession(ctx context.Context, current session.Session) (session.Session, error) {
	if current.RefreshTokenCiphertext == "" {
		return session.Session{}, ErrRefreshUnavailable
	}
	raw, err := s.protector.Decrypt(current.RefreshTokenCiphertext)
	if err != nil {
		return session.Session{}, err
	}
	tokens, err := s.identity.RefreshToken(ctx, raw)
	if err != nil {
		var tokenErr *identity.TokenError
		if errors.As(err, &tokenErr) && tokenErr.ErrorCode == "invalid_grant" {
			return session.Session{}, session.ErrRefreshReuse
		}
		return session.Session{}, err
	}
	if tokens.RefreshToken != "" {
		current.RefreshTokenCiphertext, err = s.protector.Encrypt(tokens.RefreshToken)
		if err != nil {
			return session.Session{}, err
		}
	}
	ttl := s.ttl
	if tokens.ExpiresIn > 0 && time.Duration(tokens.ExpiresIn)*time.Second < ttl {
		ttl = time.Duration(tokens.ExpiresIn) * time.Second
	}
	current.ExpiresAt = time.Now().Add(ttl)
	return current, nil
}
func (s *Service) Get(ctx context.Context, id string) (session.Session, error) {
	value, err := s.sessions.Get(ctx, id)
	if err != nil {
		return session.Session{}, err
	}
	value.ProfileClaims = s.loadProfile(id, value.ExpiresAt)
	return value, nil
}
func (s *Service) Logout(ctx context.Context, id, redirectURI, state string) (LogoutResult, error) {
	current, err := s.sessions.Get(ctx, id)
	if err != nil && !errors.Is(err, session.ErrNotFound) {
		return LogoutResult{}, err
	}
	if err == nil && current.RefreshTokenCiphertext != "" {
		if raw, decryptErr := s.protector.Decrypt(current.RefreshTokenCiphertext); decryptErr == nil {
			// Local revocation remains authoritative if the provider is unavailable.
			revokeErr := s.identity.RevokeToken(ctx, raw)
			if s.onRemoteRevoke != nil {
				s.onRemoteRevoke(revokeErr)
			}
		}
	}
	logout := LogoutResult{}
	if err == nil {
		idToken := ""
		if current.IDTokenCiphertext != "" && s.protector != nil {
			idToken, _ = s.protector.Decrypt(current.IDTokenCiphertext)
		}
		logout.EndSessionURL, _ = s.identity.EndSessionURL(idToken, redirectURI, state)
	}
	err = s.sessions.Revoke(ctx, id)
	s.profileMu.Lock()
	delete(s.profiles, id)
	s.profileMu.Unlock()
	if s.auditRecorder != nil {
		eventID, idErr := session.NewID()
		if idErr == nil {
			eventType, outcome, subject := "AUTH_LOGOUT", "success", ""
			if err != nil {
				eventType, outcome = "AUTH_LOGOUT", "failure"
			}
			if current.Subject != "" {
				subject = current.Subject
			}
			_ = s.auditRecorder.Record(ctx, plataudit.Event{ID: "evt-" + eventID, EventType: eventType, Outcome: outcome, ActorSubject: subject, ApplicationCode: s.applicationCode, ResourceType: "session", ResourceID: id, Action: "logout", RequestID: httpkit.RequestIDFromContext(ctx), Source: "api"})
		}
	}
	return logout, err
}

func (s *Service) storeProfile(id string, claims map[string]any, expiresAt time.Time) {
	if len(claims) == 0 {
		return
	}
	now := time.Now()
	s.profileMu.Lock()
	for key, profile := range s.profiles {
		if !now.Before(profile.expiresAt) {
			delete(s.profiles, key)
		}
	}
	s.profiles[id] = cachedProfile{claims: claims, expiresAt: expiresAt}
	s.profileMu.Unlock()
}

func (s *Service) loadProfile(id string, sessionExpiresAt time.Time) map[string]any {
	s.profileMu.Lock()
	defer s.profileMu.Unlock()
	profile, ok := s.profiles[id]
	if !ok {
		return nil
	}
	if !time.Now().Before(profile.expiresAt) || !time.Now().Before(sessionExpiresAt) {
		delete(s.profiles, id)
		return nil
	}
	return profile.claims
}

func (s *Service) SetRemoteRevokeObserver(observer func(error)) { s.onRemoteRevoke = observer }
func randomString(size int) (string, error) {
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
