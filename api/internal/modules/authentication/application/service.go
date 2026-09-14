package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	auditapplication "sechelper-auth-template/api/internal/modules/audit/application"
	auditdomain "sechelper-auth-template/api/internal/modules/audit/domain"
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

type IdentityProvider interface {
	AuthorizationURL(state, nonce, challenge string) (string, error)
	ExchangeCode(context.Context, string, string) (identity.TokenSet, error)
	ValidateIDToken(context.Context, string, string) (identity.IDTokenClaims, error)
	GetUserInfo(context.Context, string) (identity.UserInfo, error)
	GetAuthorization(context.Context, string) (identity.Authorization, error)
	RefreshToken(context.Context, string) (identity.TokenSet, error)
	RevokeToken(context.Context, string) error
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
	auditRecorder   auditapplication.Recorder
}

func (s *Service) SetAuditRecorder(recorder auditapplication.Recorder) { s.auditRecorder = recorder }

func NewService(client IdentityProvider, sessions session.Store, loginStates session.LoginTransactionStore, protector *session.TokenProtector, ttl time.Duration, applicationCode string) *Service {
	return &Service{identity: client, sessions: sessions, loginStates: loginStates, protector: protector, ttl: ttl, applicationCode: applicationCode}
}
func (s *Service) BeginLogin(ctx context.Context) (string, error) {
	state, err := randomString(32)
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
	if err := s.loginStates.SaveLoginTransaction(ctx, session.LoginTransaction{State: state, Nonce: nonce, Verifier: verifier, ExpiresAt: time.Now().Add(5 * time.Minute)}); err != nil {
		return "", err
	}
	return s.identity.AuthorizationURL(state, nonce, challenge)
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
		_ = s.auditRecorder.Record(ctx, auditdomain.Event{ID: "evt-" + id, EventType: eventType, Outcome: outcome, ActorSubject: result.Subject, ApplicationCode: s.applicationCode, ResourceType: "auth_flow", ResourceID: s.applicationCode, Action: "login", Source: "api"})
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
	result = session.Session{ID: id, Subject: user.Subject, Email: user.Email, ApplicationCode: auth.ApplicationCode, Permissions: permissions, RefreshTokenCiphertext: refreshToken, ExpiresAt: time.Now().Add(ttl)}
	if err := s.sessions.Create(ctx, result); err != nil {
		return session.Session{}, err
	}
	return result, nil
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
		_ = s.auditRecorder.Record(ctx, auditdomain.Event{ID: "evt-" + eventID, EventType: "AUTH_SESSION_REFRESH_FAILED", Outcome: "failure", ApplicationCode: s.applicationCode, ResourceType: "session", ResourceID: id, Action: "refresh", Source: "api"})
	}()
	// The local lock prevents refresh-token rotation races inside one process.
	// The persistent store must still provide an optimistic/concurrent update
	// guard for multi-instance deployments.
	s.refreshMu.Lock()
	defer s.refreshMu.Unlock()
	current, err := s.sessions.Get(ctx, id)
	if err != nil {
		return session.Session{}, err
	}
	if current.RefreshTokenCiphertext == "" {
		return session.Session{}, ErrRefreshUnavailable
	}
	raw, err := s.protector.Decrypt(current.RefreshTokenCiphertext)
	if err != nil {
		return session.Session{}, err
	}
	tokens, err := s.identity.RefreshToken(ctx, raw)
	if err != nil {
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
	if err := s.sessions.Update(ctx, current); err != nil {
		return session.Session{}, err
	}
	return current, nil
}
func (s *Service) Get(ctx context.Context, id string) (session.Session, error) {
	return s.sessions.Get(ctx, id)
}
func (s *Service) Logout(ctx context.Context, id string) error {
	current, err := s.sessions.Get(ctx, id)
	if err != nil && !errors.Is(err, session.ErrNotFound) {
		return err
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
	err = s.sessions.Revoke(ctx, id)
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
			_ = s.auditRecorder.Record(ctx, auditdomain.Event{ID: "evt-" + eventID, EventType: eventType, Outcome: outcome, ActorSubject: subject, ApplicationCode: s.applicationCode, ResourceType: "session", ResourceID: id, Action: "logout", Source: "api"})
		}
	}
	return err
}
func (s *Service) SetRemoteRevokeObserver(observer func(error)) { s.onRemoteRevoke = observer }
func randomString(size int) (string, error) {
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
