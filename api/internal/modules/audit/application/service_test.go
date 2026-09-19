package application

import (
	"context"
	"sechelper-auth-template/api/internal/modules/audit/domain"
	platformaudit "sechelper-auth-template/api/internal/platform/audit"
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
	if err := service.Record(context.Background(), platformaudit.Event{ID: "evt-1", EventType: "RESOURCE_ACCESS_DENIED", ApplicationCode: "app-1"}); err != nil {
		t.Fatal(err)
	}
	if got := repository.events[0]; got.Category != "authorization" || got.Severity != "warning" || got.Outcome != "success" {
		t.Fatalf("unexpected event defaults: %+v", got)
	}
}

func TestRecordRejectsUnregisteredEvent(t *testing.T) {
	service := NewService(&fakeRepository{})
	if err := service.Record(context.Background(), platformaudit.Event{ID: "evt-1", EventType: "HTTP_GET", ApplicationCode: "app-1"}); err == nil {
		t.Fatal("expected unregistered event error")
	}
}

func TestAuthenticationLifecycleEventsAreNotAuditEvents(t *testing.T) {
	service := NewService(&fakeRepository{})
	for _, eventType := range []string{"AUTH_LOGIN_SUCCESS", "AUTH_LOGIN_FAILED", "AUTH_LOGOUT", "AUTH_SESSION_REVOKED", "AUTH_SESSION_REVOKE_FAILED", "AUTH_SESSION_REFRESH_FAILED"} {
		if err := service.Record(context.Background(), platformaudit.Event{ID: "evt-" + eventType, EventType: eventType, ApplicationCode: "app-1"}); err == nil {
			t.Fatalf("event %q should not be registered", eventType)
		}
	}
	if err := service.RegisterEventType("AUTH_LOGIN_SUCCESS", platformaudit.EventDefinition{Category: "authentication", Severity: "info"}); err == nil {
		t.Fatal("authentication category should not be accepted")
	}
}

func TestBusinessModuleCanRegisterAuditEventType(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)
	if err := service.RegisterEventType("RESOURCE_CREATED", platformaudit.EventDefinition{Category: "business", Severity: "info"}); err != nil {
		t.Fatal(err)
	}
	if err := service.Record(context.Background(), platformaudit.Event{ID: "evt-resource-1", EventType: "RESOURCE_CREATED", ApplicationCode: "catalog"}); err != nil {
		t.Fatal(err)
	}
	if got := repository.events[0]; got.Category != "business" || got.Severity != "info" {
		t.Fatalf("registered event defaults were not applied: %+v", got)
	}
}
