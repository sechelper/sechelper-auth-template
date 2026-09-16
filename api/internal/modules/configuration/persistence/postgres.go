package persistence

import (
	"context"
	"database/sql"
	"sechelper-auth-template/api/internal/modules/configuration/domain"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) List(ctx context.Context) ([]domain.Entry, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT key, description, is_secret, version, updated_by, updated_at FROM configuration_entries ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]domain.Entry, 0)
	for rows.Next() {
		var value domain.Entry
		if err := rows.Scan(&value.Key, &value.Description, &value.IsSecret, &value.Version, &value.UpdatedBy, &value.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

func (r *Repository) Values(ctx context.Context) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT key, ciphertext FROM configuration_entries`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, rows.Err()
}

func (r *Repository) Get(ctx context.Context, key string) (domain.Value, error) {
	var value domain.Value
	err := r.db.QueryRowContext(ctx, `SELECT key, description, is_secret, version, updated_by, updated_at, ciphertext FROM configuration_entries WHERE key=$1`, key).
		Scan(&value.Key, &value.Description, &value.IsSecret, &value.Version, &value.UpdatedBy, &value.UpdatedAt, &value.Value)
	return value, err
}

func (r *Repository) Upsert(ctx context.Context, key, description string, secret bool, ciphertext, updatedBy string) (domain.Entry, error) {
	var value domain.Entry
	err := r.db.QueryRowContext(ctx, `INSERT INTO configuration_entries(key, description, is_secret, ciphertext, updated_by) VALUES($1,$2,$3,$4,$5) ON CONFLICT(key) DO UPDATE SET description=EXCLUDED.description, is_secret=EXCLUDED.is_secret, ciphertext=EXCLUDED.ciphertext, version=configuration_entries.version+1, updated_by=EXCLUDED.updated_by, updated_at=CURRENT_TIMESTAMP RETURNING key, description, is_secret, version, updated_by, updated_at`, key, description, secret, ciphertext, updatedBy).
		Scan(&value.Key, &value.Description, &value.IsSecret, &value.Version, &value.UpdatedBy, &value.UpdatedAt)
	return value, err
}

func (r *Repository) Delete(ctx context.Context, key string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM configuration_entries WHERE key=$1`, key)
	return err
}
