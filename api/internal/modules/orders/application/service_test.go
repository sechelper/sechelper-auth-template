//go:build example

package application

import (
	"context"
	"sechelper-auth-template/api/internal/modules/orders/domain"
	"testing"
	"time"
)

type fakeRepository struct{}

func (fakeRepository) List(context.Context, int) ([]domain.Order, error) {
	return []domain.Order{{ID: "order-1"}}, nil
}
func (fakeRepository) Get(context.Context, string) (domain.Order, error) {
	return domain.Order{ID: "order-1", CreatedAt: time.Now()}, nil
}
func TestListBoundsPageSize(t *testing.T) {
	service := NewService(fakeRepository{})
	values, err := service.List(context.Background(), 1000)
	if err != nil {
		t.Fatal(err)
	}
	if len(values) != 1 {
		t.Fatalf("unexpected result length %d", len(values))
	}
}
func TestGetRequiresIdentifier(t *testing.T) {
	service := NewService(fakeRepository{})
	if _, err := service.Get(context.Background(), ""); err != ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}
