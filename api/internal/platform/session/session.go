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
	ErrNotFound = errors.New("session not found")
	ErrConflict = errors.New("session version conflict")
)

type Session struct {
	ID, Subject, Email, ApplicationCode string
	Permissions                         []string
	RefreshTokenCiphertext              string
	ExpiresAt                           time.Time
	Version                             int64
	Revoked                             bool
	CreatedAt, UpdatedAt                time.Time
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
	ListBySubject(context.Context, string, int) ([]Session, error)
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
	_, err = s.db.ExecContext(ctx, `INSERT INTO authentication_sessions (id, subject, email, application_code, permissions, refresh_token_ciphertext, expires_at, version) VALUES ($1,$2,$3,$4,$5,$6,$7,1)`, value.ID, value.Subject, value.Email, value.ApplicationCode, permissions, value.RefreshTokenCiphertext, value.ExpiresAt)
	return err
}
func (s *PostgresStore) ListBySubject(ctx context.Context, subject string, limit int) ([]Session, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, subject, email, application_code, expires_at, created_at, updated_at, revoked_at FROM authentication_sessions WHERE subject=$1 AND expires_at > CURRENT_TIMESTAMP AND revoked_at IS NULL ORDER BY created_at DESC, id DESC LIMIT $2`, subject, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Session, 0)
	for rows.Next() {
		var value Session
		var revokedAt sql.NullTime
		if err := rows.Scan(&value.ID, &value.Subject, &value.Email, &value.ApplicationCode, &value.ExpiresAt, &value.CreatedAt, &value.UpdatedAt, &revokedAt); err != nil {
			return nil, err
		}
		value.Revoked = revokedAt.Valid
		result = append(result, value)
	}
	return result, rows.Err()
}
func (s *PostgresStore) Get(ctx context.Context, id string) (Session, error) {
	var value Session
	var permissions []byte
	var refreshToken sql.NullString
	var revokedAt sql.NullTime
	err := s.db.QueryRowContext(ctx, `SELECT id, subject, email, application_code, permissions, refresh_token_ciphertext, expires_at, version, revoked_at FROM authentication_sessions WHERE id=$1 AND expires_at > CURRENT_TIMESTAMP AND revoked_at IS NULL`, id).Scan(&value.ID, &value.Subject, &value.Email, &value.ApplicationCode, &permissions, &refreshToken, &value.ExpiresAt, &value.Version, &revokedAt)
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

func (s *MemoryStore) ListBySubject(_ context.Context, subject string, limit int) ([]Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Session, 0)
	for _, value := range s.values {
		if value.Subject == subject && !value.Revoked && time.Now().Before(value.ExpiresAt) {
			result = append(result, value)
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
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
