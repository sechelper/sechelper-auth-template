package audit

import (
	"context"
	"time"
)

type Event struct {
	ID, EventType, Category, Severity, Outcome, ReasonCode             string
	ActorSubject, ActorEmail, ApplicationCode                          string
	ResourceType, ResourceID, Action, RequestID, CorrelationID, Source string
	OccurredAt                                                         time.Time
	Metadata                                                           map[string]any
}

type EventDefinition struct {
	Category string
	Severity string
}

type DefinitionRegistrar interface {
	RegisterEventType(string, EventDefinition) error
}

type Recorder interface {
	Record(context.Context, Event) error
}
