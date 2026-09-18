package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sechelper-auth-template/api/internal/modules/audit/application"
	"sechelper-auth-template/api/internal/modules/audit/domain"
	"strings"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }
func (r *Repository) Record(ctx context.Context, value domain.Event) error {
	metadata, _ := json.Marshal(value.Metadata)
	_, err := r.db.ExecContext(ctx, `INSERT INTO framework.audit_events (id, occurred_at, event_type, category, severity, outcome, reason_code, actor_subject, actor_email, application_code, action, resource_type, resource_id, result, request_id, correlation_id, source, ip_hash, user_agent, metadata) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)`, value.ID, value.OccurredAt, value.EventType, value.Category, value.Severity, value.Outcome, value.ReasonCode, value.ActorSubject, value.ActorEmail, value.ApplicationCode, value.Action, value.ResourceType, value.ResourceID, value.Outcome, value.RequestID, value.CorrelationID, value.Source, value.IPHash, value.UserAgent, metadata)
	return err
}
func (r *Repository) List(ctx context.Context, filter application.ListFilter) ([]domain.Event, error) {
	args := []any{}
	conditions := []string{"1=1"}
	add := func(value string, expression string) {
		if strings.TrimSpace(value) != "" {
			args = append(args, value)
			conditions = append(conditions, fmt.Sprintf(expression, len(args)))
		}
	}
	add(filter.EventType, "event_type=$%d")
	add(filter.Category, "category=$%d")
	add(filter.Severity, "severity=$%d")
	add(filter.Outcome, "COALESCE(NULLIF(outcome,''), result)=$%d")
	add(filter.ActorSubject, "actor_subject=$%d")
	add(filter.ResourceType, "resource_type=$%d")
	add(filter.ResourceID, "resource_id=$%d")
	if !filter.From.IsZero() {
		args = append(args, filter.From)
		conditions = append(conditions, fmt.Sprintf("occurred_at >= $%d", len(args)))
	}
	if !filter.To.IsZero() {
		args = append(args, filter.To)
		conditions = append(conditions, fmt.Sprintf("occurred_at < $%d", len(args)))
	}
	if filter.Limit < 1 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	args = append(args, filter.Limit)
	query := fmt.Sprintf(`SELECT id, occurred_at, event_type, category, severity, COALESCE(NULLIF(outcome,''), result), reason_code, actor_subject, actor_email, application_code, resource_type, resource_id, action, request_id, correlation_id, source, ip_hash, user_agent, metadata FROM framework.audit_events WHERE %s ORDER BY occurred_at DESC, id DESC LIMIT $%d`, strings.Join(conditions, " AND "), len(args))
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	resultItems := make([]domain.Event, 0)
	for rows.Next() {
		var value domain.Event
		var metadata []byte
		if err := rows.Scan(&value.ID, &value.OccurredAt, &value.EventType, &value.Category, &value.Severity, &value.Outcome, &value.ReasonCode, &value.ActorSubject, &value.ActorEmail, &value.ApplicationCode, &value.ResourceType, &value.ResourceID, &value.Action, &value.RequestID, &value.CorrelationID, &value.Source, &value.IPHash, &value.UserAgent, &metadata); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(metadata, &value.Metadata)
		resultItems = append(resultItems, value)
	}
	return resultItems, rows.Err()
}
