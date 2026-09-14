package application

import (
	"context"
	"errors"
	"sechelper-auth-template/api/internal/modules/orders/domain"
)

var ErrNotFound = errors.New("order not found")

type Repository interface {
	List(ctx context.Context, limit int) ([]domain.Order, error)
	Get(ctx context.Context, id string) (domain.Order, error)
}
type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }
func (s *Service) List(ctx context.Context, limit int) ([]domain.Order, error) {
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repository.List(ctx, limit)
}
func (s *Service) Get(ctx context.Context, id string) (domain.Order, error) {
	if id == "" {
		return domain.Order{}, ErrNotFound
	}
	return s.repository.Get(ctx, id)
}
