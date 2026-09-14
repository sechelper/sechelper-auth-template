package domain

import "context"

type ResourceRef struct {
	Type string `json:"resourceType"`
	ID   string `json:"resourceId"`
}
type AccessRequest struct {
	Actor    Context
	Resource ResourceRef
	Action   string
}
type Decision struct {
	Allowed      bool   `json:"allowed"`
	ReasonCode   string `json:"reasonCode"`
	Permission   string `json:"permission,omitempty"`
	ResourceType string `json:"resourceType"`
	ResourceID   string `json:"resourceId"`
	Action       string `json:"action"`
}
type ActionDefinition struct {
	Name       string `json:"name"`
	Permission string `json:"permission"`
}
type ResourceDefinition struct {
	Type        string             `json:"type"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Actions     []ActionDefinition `json:"actions"`
}
type ResourcePolicy interface {
	Evaluate(context.Context, AccessRequest) (Decision, error)
}
