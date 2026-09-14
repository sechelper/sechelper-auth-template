package application

import (
	"context"
	"errors"
	auditapplication "sechelper-auth-template/api/internal/modules/audit/application"
	auditdomain "sechelper-auth-template/api/internal/modules/audit/domain"
	"sechelper-auth-template/api/internal/platform/session"
)

type Service struct {
	sessions      session.Store
	auditRecorder auditapplication.Recorder
}

func NewService(sessions session.Store) *Service                       { return &Service{sessions: sessions} }
func (s *Service) SetAuditRecorder(recorder auditapplication.Recorder) { s.auditRecorder = recorder }
func (s *Service) ListSessions(ctx context.Context, subject string, limit int) ([]session.Session, error) {
	return s.sessions.ListBySubject(ctx, subject, limit)
}
func (s *Service) RevokeSession(ctx context.Context, subject, id string) error {
	value, err := s.sessions.Get(ctx, id)
	if err != nil || value.Subject != subject {
		return errors.New("session not found")
	}
	err = s.sessions.Revoke(ctx, id)
	if s.auditRecorder != nil {
		eventID, idErr := session.NewID()
		if idErr == nil {
			eventType, outcome := "AUTH_SESSION_REVOKED", "success"
			if err != nil {
				eventType, outcome = "AUTH_SESSION_REVOKE_FAILED", "failure"
			}
			_ = s.auditRecorder.Record(ctx, auditdomain.Event{ID: "evt-" + eventID, EventType: eventType, Outcome: outcome, ActorSubject: subject, ApplicationCode: value.ApplicationCode, ResourceType: "session", ResourceID: id, Action: "revoke", Source: "admin_ui"})
		}
	}
	return err
}
