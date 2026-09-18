//go:build example

package persistence

import (
	"context"
	"database/sql"
	"sechelper-auth-template/api/internal/business/orders/application"
	"sechelper-auth-template/api/internal/business/orders/domain"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }
func (r *Repository) List(ctx context.Context, limit int) ([]domain.Order, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, status, total_minor, currency, created_at, updated_at FROM business_orders.orders ORDER BY created_at DESC, id DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]domain.Order, 0)
	for rows.Next() {
		var item domain.Order
		if err := rows.Scan(&item.ID, &item.Status, &item.TotalMinor, &item.Currency, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
func (r *Repository) Get(ctx context.Context, id string) (domain.Order, error) {
	var item domain.Order
	err := r.db.QueryRowContext(ctx, `SELECT id, status, total_minor, currency, created_at, updated_at FROM business_orders.orders WHERE id=$1`, id).Scan(&item.ID, &item.Status, &item.TotalMinor, &item.Currency, &item.CreatedAt, &item.UpdatedAt)
	if err == sql.ErrNoRows {
		return domain.Order{}, application.ErrNotFound
	}
	return item, err
}
