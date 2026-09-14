package audit

import "context"

type Event struct {
	ActorSubject, Action, ResourceType, ResourceID, Outcome, RequestID string
	Metadata                                                           map[string]any
}

type Recorder interface {
	Record(context.Context, Event) error
}
