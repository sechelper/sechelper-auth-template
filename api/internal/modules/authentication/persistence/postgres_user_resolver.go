package persistence

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"strings"
)

type PostgresUserResolver struct{ db *sql.DB }

func NewPostgresUserResolver(db *sql.DB) *PostgresUserResolver {
	return &PostgresUserResolver{db: db}
}

func (r *PostgresUserResolver) ResolveOrCreate(ctx context.Context, issuer, subject string) (string, error) {
	issuer, subject = strings.TrimSpace(issuer), strings.TrimSpace(subject)
	if issuer == "" || subject == "" {
		return "", fmt.Errorf("issuer and subject are required")
	}
	localID, err := newUUID()
	if err != nil {
		return "", err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `INSERT INTO platform_users (id) VALUES ($1::uuid)`, localID); err != nil {
		return "", err
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO external_identities (issuer, identity_subject, platform_user_uuid) VALUES ($1,$2,$3::uuid) ON CONFLICT (issuer, identity_subject) DO NOTHING`, issuer, subject, localID)
	if err != nil {
		return "", err
	}
	inserted, err := result.RowsAffected()
	if err != nil {
		return "", err
	}
	if inserted == 0 {
		if err := tx.Rollback(); err != nil {
			return "", err
		}
		var existing string
		if err := r.db.QueryRowContext(ctx, `SELECT platform_user_uuid::text FROM external_identities WHERE issuer=$1 AND identity_subject=$2`, issuer, subject).Scan(&existing); err != nil {
			return "", err
		}
		return existing, nil
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return localID, nil
}

func newUUID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", raw[0:4], raw[4:6], raw[6:8], raw[8:10], raw[10:16]), nil
}
