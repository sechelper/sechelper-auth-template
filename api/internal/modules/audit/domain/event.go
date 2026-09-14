package domain

import "time"

type Event struct {
	ID              string         `json:"id"`
	OccurredAt      time.Time      `json:"occurredAt"`
	EventType       string         `json:"eventType"`
	Category        string         `json:"category"`
	Severity        string         `json:"severity"`
	Outcome         string         `json:"outcome"`
	ReasonCode      string         `json:"reasonCode,omitempty"`
	ActorSubject    string         `json:"actorSubject"`
	ActorEmail      string         `json:"actorEmail,omitempty"`
	ApplicationCode string         `json:"applicationCode"`
	ResourceType    string         `json:"resourceType,omitempty"`
	ResourceID      string         `json:"resourceId,omitempty"`
	Action          string         `json:"action,omitempty"`
	RequestID       string         `json:"requestId,omitempty"`
	CorrelationID   string         `json:"correlationId,omitempty"`
	Source          string         `json:"source"`
	IPHash          string         `json:"-"`
	UserAgent       string         `json:"-"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}
