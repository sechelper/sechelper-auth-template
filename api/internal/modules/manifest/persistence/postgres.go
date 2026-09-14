package persistence

import (
	"context"
	"database/sql"

	"sechelper-auth-template/api/internal/modules/manifest/application"
)

type PostgresStateStore struct{ db *sql.DB }

func NewPostgresStateStore(db *sql.DB) *PostgresStateStore { return &PostgresStateStore{db: db} }
func (s *PostgresStateStore) Load(ctx context.Context) (application.State, error) {
	var value application.State
	var canonicalJSON string
	var syncID sql.NullString
	var acceptedAt sql.NullTime
	err := s.db.QueryRowContext(ctx, `SELECT application_code, manifest_version, content_hash, canonical_json, sync_id, status, server_revision, accepted_at, updated_at FROM authorization_manifest_state WHERE id=1`).Scan(&value.ApplicationCode, &value.ManifestVersion, &value.ContentHash, &canonicalJSON, &syncID, &value.Status, &value.ServerRevision, &acceptedAt, &value.UpdatedAt)
	if err != nil {
		return value, err
	}
	value.CanonicalJSON = canonicalJSON
	value.SyncID = syncID.String
	if acceptedAt.Valid {
		value.AcceptedAt = acceptedAt.Time
	}
	return value, nil
}
func (s *PostgresStateStore) Save(ctx context.Context, value application.State) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO authorization_manifest_state (id, application_code, manifest_version, content_hash, canonical_json, sync_id, status, server_revision, accepted_at, updated_at) VALUES (1,$1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT (id) DO UPDATE SET application_code=EXCLUDED.application_code, manifest_version=EXCLUDED.manifest_version, content_hash=EXCLUDED.content_hash, canonical_json=EXCLUDED.canonical_json, sync_id=EXCLUDED.sync_id, status=EXCLUDED.status, server_revision=EXCLUDED.server_revision, accepted_at=EXCLUDED.accepted_at, updated_at=EXCLUDED.updated_at`, value.ApplicationCode, value.ManifestVersion, value.ContentHash, value.CanonicalJSON, value.SyncID, value.Status, value.ServerRevision, value.AcceptedAt, value.UpdatedAt)
	return err
}
