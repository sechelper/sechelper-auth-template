package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sechelper-auth-template/api/internal/modules/audit/domain"
	"strings"
	"time"
)

type ListFilter struct {
	EventType, Category, Severity, Outcome, ActorSubject, ResourceType, ResourceID string
	From, To                                                                       time.Time
	Limit                                                                          int
}
type Repository interface {
	Record(context.Context, domain.Event) error
	List(context.Context, ListFilter) ([]domain.Event, error)
}
type Recorder interface {
	Record(context.Context, domain.Event) error
}
type EventDefinition struct{ Category, Severity string }
type Service struct {
	repository  Repository
	definitions map[string]EventDefinition
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository, definitions: map[string]EventDefinition{
		"AUTH_LOGIN_SUCCESS": {"authentication", "info"}, "AUTH_LOGIN_FAILED": {"authentication", "warning"}, "AUTH_LOGOUT": {"authentication", "info"}, "AUTH_SESSION_REFRESH_FAILED": {"authentication", "warning"}, "AUTH_SESSION_REVOKED": {"authentication", "warning"}, "AUTH_SESSION_REVOKE_FAILED": {"authentication", "warning"},
		"RESOURCE_ACCESS_DENIED": {"authorization", "warning"}, "RESOURCE_SCOPE_DENIED": {"authorization", "warning"}, "ACCESS_DECISION_CHECKED": {"authorization", "info"},
		"MANIFEST_SYNC_SUCCEEDED": {"manifest", "info"}, "MANIFEST_SYNC_FAILED": {"manifest", "critical"}, "MANIFEST_CHANGE_DETECTED": {"manifest", "warning"},
		"BACKGROUND_JOB_RETRIED": {"operations", "warning"}, "BACKGROUND_JOB_CANCELLED": {"operations", "warning"}, "SECURITY_CONFIGURATION_CHANGED": {"operations", "critical"},
		"RESOURCE_EXPORT_STARTED": {"resource", "warning"}, "RESOURCE_EXPORT_COMPLETED": {"resource", "warning"}, "RESOURCE_EXPORT_FAILED": {"resource", "critical"},
	}}
}
func (s *Service) Record(ctx context.Context, event domain.Event) error {
	definition, ok := s.definitions[event.EventType]
	if !ok {
		return fmt.Errorf("audit event type %q is not registered", event.EventType)
	}
	if event.EventType == "" || strings.TrimSpace(event.ApplicationCode) == "" {
		return errors.New("audit event type and application code are required")
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now().UTC()
	}
	if event.Category == "" {
		event.Category = definition.Category
	}
	if event.Severity == "" {
		event.Severity = definition.Severity
	}
	if event.Outcome == "" {
		event.Outcome = "success"
	}
	if event.ID == "" {
		return errors.New("audit event id is required")
	}
	return s.repository.Record(ctx, event)
}
func (s *Service) List(ctx context.Context, filter ListFilter) ([]domain.Event, error) {
	if filter.Limit < 1 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	return s.repository.List(ctx, filter)
}
func Metadata(value map[string]any) []byte { raw, _ := json.Marshal(value); return raw }
