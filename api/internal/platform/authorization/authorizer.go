package authorization

import "context"

type Subject struct{ ID, Application string }
type Resource struct{ Type, ID string }
type Decision struct {
	Allowed                bool
	ReasonCode, Permission string
}

type Authorizer interface {
	Authorize(context.Context, Subject, Resource, string) (Decision, error)
}
