package application

import (
	"context"
	"sechelper-auth-template/api/internal/modules/audit/domain"
	"testing"
)

type fakeRepository struct{ events []domain.Event }

func (r *fakeRepository) Record(_ context.Context, event domain.Event) error {
	r.events = append(r.events, event)
	return nil
}
func (r *fakeRepository) List(_ context.Context, _ ListFilter) ([]domain.Event, error) {
	return r.events, nil
}

func TestRecordAppliesRegisteredDefaults(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)
	if err := service.Record(context.Background(), domain.Event{ID: "evt-1", EventType: "RESOURCE_ACCESS_DENIED", ApplicationCode: "app-1"}); err != nil {
		t.Fatal(err)
	}
	if got := repository.events[0]; got.Category != "authorization" || got.Severity != "warning" || got.Outcome != "success" {
		t.Fatalf("unexpected event defaults: %+v", got)
	}
}

func TestRecordRejectsUnregisteredEvent(t *testing.T) {
	service := NewService(&fakeRepository{})
	if err := service.Record(context.Background(), domain.Event{ID: "evt-1", EventType: "HTTP_GET", ApplicationCode: "app-1"}); err == nil {
		t.Fatal("expected unregistered event error")
	}
}
