package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrNotFound     = errors.New("session not found")
	ErrConflict     = errors.New("session version conflict")
	ErrRefreshReuse = errors.New("refresh token reuse detected")
)

type Session struct {
	ID, Subject, PlatformUserUUID, Email, ApplicationCode string
	Permissions                                           []string
	ProfileClaims                                         map[string]any
	RefreshTokenCiphertext                                string
	ExpiresAt                                             time.Time
	Version                                               int64
	Revoked                                               bool
	CreatedAt, UpdatedAt                                  time.Time
}
type LoginTransaction struct {
	State, Nonce, Verifier string
	ExpiresAt              time.Time
}
type LoginTransactionStore interface {
	SaveLoginTransaction(context.Context, LoginTransaction) error
	ConsumeLoginTransaction(context.Context, string) (LoginTransaction, error)
}
type Store interface {
	Create(context.Context, Session) error
	Get(context.Context, string) (Session, error)
	Update(context.Context, Session) error
	Revoke(context.Context, string) error
}

// RefreshRotator is the durable concurrency boundary for refresh-token
// rotation. Implementations must serialize the callback for one session and
// persist its returned session atomically.
type RefreshRotator interface {
	RotateRefreshToken(context.Context, string, func(context.Context, Session) (Session, error)) (Session, error)
}
type MemoryStore struct {
	mu     sync.RWMutex
	values map[string]Session
	logins map[string]LoginTransaction
}

type PostgresStore struct{ db *sql.DB }

func NewPostgresStore(db *sql.DB) *PostgresStore { return &PostgresStore{db: db} }
func (s *PostgresStore) Create(ctx context.Context, value Session) error {
	permissions, err := json.Marshal(value.Permissions)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO authentication_sessions (id, subject, platform_user_uuid, email, application_code, permissions, refresh_token_ciphertext, expires_at, version) VALUES ($1,$2,NULLIF($3,'')::uuid,$4,$5,$6,$7,$8,1)`, value.ID, value.Subject, value.PlatformUserUUID, value.Email, value.ApplicationCode, permissions, value.RefreshTokenCiphertext, value.ExpiresAt)
	return err
}
func (s *PostgresStore) Get(ctx context.Context, id string) (Session, error) {
	var value Session
	var permissions []byte
	var refreshToken sql.NullString
	var revokedAt sql.NullTime
	var platformUserUUID sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT id, subject, platform_user_uuid, email, application_code, permissions, refresh_token_ciphertext, expires_at, version, revoked_at FROM authentication_sessions WHERE id=$1 AND expires_at > CURRENT_TIMESTAMP AND revoked_at IS NULL`, id).Scan(&value.ID, &value.Subject, &platformUserUUID, &value.Email, &value.ApplicationCode, &permissions, &refreshToken, &value.ExpiresAt, &value.Version, &revokedAt)
	if err == sql.ErrNoRows {
		return Session{}, ErrNotFound
	}
	if err != nil {
		return Session{}, err
	}
	if err := json.Unmarshal(permissions, &value.Permissions); err != nil {
		return Session{}, err
	}
	value.RefreshTokenCiphertext = refreshToken.String
	value.PlatformUserUUID = platformUserUUID.String
	value.Revoked = revokedAt.Valid
	return value, nil
}
func (s *PostgresStore) Update(ctx context.Context, value Session) error {
	permissions, err := json.Marshal(value.Permissions)
	if err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `UPDATE authentication_sessions SET email=$2, permissions=$3, refresh_token_ciphertext=$4, expires_at=$5, version=version+1, updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND version=$6 AND revoked_at IS NULL`, value.ID, value.Email, permissions, value.RefreshTokenCiphertext, value.ExpiresAt, value.Version)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil || count == 0 {
		return ErrConflict
	}
	return nil
}
func (s *PostgresStore) RotateRefreshToken(ctx context.Context, id string, rotate func(context.Context, Session) (Session, error)) (Session, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Session{}, err
	}
	defer tx.Rollback()
	current, err := scanSession(tx.QueryRowContext(ctx, `SELECT id, subject, platform_user_uuid, email, application_code, permissions, refresh_token_ciphertext, expires_at, version, revoked_at FROM authentication_sessions WHERE id=$1 FOR UPDATE`, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, ErrNotFound
		}
		return Session{}, err
	}
	if current.Revoked || !time.Now().Before(current.ExpiresAt) {
		return Session{}, ErrNotFound
	}
	value, err := rotate(ctx, current)
	if err != nil {
		if errors.Is(err, ErrRefreshReuse) {
			if _, revokeErr := tx.ExecContext(ctx, `UPDATE authentication_sessions SET revoked_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND revoked_at IS NULL`, id); revokeErr != nil {
				return Session{}, revokeErr
			}
			if commitErr := tx.Commit(); commitErr != nil {
				return Session{}, commitErr
			}
		}
		return Session{}, err
	}
	permissions, err := json.Marshal(value.Permissions)
	if err != nil {
		return Session{}, err
	}
	result, err := tx.ExecContext(ctx, `UPDATE authentication_sessions SET email=$2, permissions=$3, refresh_token_ciphertext=$4, expires_at=$5, version=version+1, updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND version=$6 AND revoked_at IS NULL`, value.ID, value.Email, permissions, value.RefreshTokenCiphertext, value.ExpiresAt, current.Version)
	if err != nil {
		return Session{}, err
	}
	count, err := result.RowsAffected()
	if err != nil || count != 1 {
		return Session{}, ErrConflict
	}
	if err := tx.Commit(); err != nil {
		return Session{}, err
	}
	value.Version = current.Version + 1
	return value, nil
}
func (s *PostgresStore) Revoke(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `UPDATE authentication_sessions SET revoked_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND revoked_at IS NULL`, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil || count == 0 {
		return ErrNotFound
	}
	return nil
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{values: map[string]Session{}, logins: map[string]LoginTransaction{}}
}
func (s *MemoryStore) Create(_ context.Context, value Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if value.Version == 0 {
		value.Version = 1
	}
	s.values[value.ID] = value
	return nil
}
func (s *MemoryStore) Get(_ context.Context, id string) (Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.values[id]
	if !ok || value.Revoked || time.Now().After(value.ExpiresAt) {
		return Session{}, ErrNotFound
	}
	return value, nil
}
func (s *MemoryStore) Update(_ context.Context, value Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.values[value.ID]
	if !ok || current.Revoked || time.Now().After(current.ExpiresAt) {
		return ErrNotFound
	}
	if value.Version == 0 || current.Version != value.Version {
		return ErrConflict
	}
	value.Version++
	s.values[value.ID] = value
	return nil
}
func (s *MemoryStore) RotateRefreshToken(ctx context.Context, id string, rotate func(context.Context, Session) (Session, error)) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.values[id]
	if !ok || current.Revoked || !time.Now().Before(current.ExpiresAt) {
		return Session{}, ErrNotFound
	}
	value, err := rotate(ctx, current)
	if err != nil {
		if errors.Is(err, ErrRefreshReuse) {
			current.Revoked = true
			s.values[id] = current
		}
		return Session{}, err
	}
	value.Version = current.Version + 1
	s.values[id] = value
	return value, nil
}

func (s *MemoryStore) SaveLoginTransaction(_ context.Context, value LoginTransaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logins[value.State] = value
	return nil
}

func (s *MemoryStore) ConsumeLoginTransaction(_ context.Context, state string) (LoginTransaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.logins[state]
	delete(s.logins, state)
	if !ok || time.Now().After(value.ExpiresAt) {
		return LoginTransaction{}, ErrNotFound
	}
	return value, nil
}

func loginStateHash(state string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(state))) }

func (s *PostgresStore) SaveLoginTransaction(ctx context.Context, value LoginTransaction) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO authentication_login_transactions (state_hash, nonce, verifier, expires_at) VALUES ($1,$2,$3,$4)`, loginStateHash(value.State), value.Nonce, value.Verifier, value.ExpiresAt)
	return err
}

func (s *PostgresStore) ConsumeLoginTransaction(ctx context.Context, state string) (LoginTransaction, error) {
	var value LoginTransaction
	err := s.db.QueryRowContext(ctx, `DELETE FROM authentication_login_transactions WHERE state_hash=$1 AND expires_at > CURRENT_TIMESTAMP RETURNING nonce, verifier, expires_at`, loginStateHash(state)).Scan(&value.Nonce, &value.Verifier, &value.ExpiresAt)
	if err == sql.ErrNoRows {
		return LoginTransaction{}, ErrNotFound
	}
	if err != nil {
		return LoginTransaction{}, err
	}
	value.State = state
	return value, nil
}

type sessionScanner interface{ Scan(...any) error }

func scanSession(row sessionScanner) (Session, error) {
	var value Session
	var permissions []byte
	var refreshToken sql.NullString
	var revokedAt sql.NullTime
	var platformUserUUID sql.NullString
	err := row.Scan(&value.ID, &value.Subject, &platformUserUUID, &value.Email, &value.ApplicationCode, &permissions, &refreshToken, &value.ExpiresAt, &value.Version, &revokedAt)
	if err != nil {
		return Session{}, err
	}
	if err := json.Unmarshal(permissions, &value.Permissions); err != nil {
		return Session{}, err
	}
	value.RefreshTokenCiphertext, value.Revoked = refreshToken.String, revokedAt.Valid
	value.PlatformUserUUID = platformUserUUID.String
	return value, nil
}
func (s *MemoryStore) Revoke(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.values[id]
	if !ok {
		return ErrNotFound
	}
	value.Revoked = true
	s.values[id] = value
	return nil
}
func NewID() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
