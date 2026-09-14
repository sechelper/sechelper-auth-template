package application

import (
	"context"
	"sechelper-auth-template/api/internal/modules/authorization/domain"
	"testing"
)

type testPolicy struct{}

func (testPolicy) Evaluate(_ context.Context, request domain.AccessRequest) (domain.Decision, error) {
	return domain.Decision{Allowed: request.Resource.ID == "owned", ReasonCode: "RESOURCE_POLICY", ResourceType: request.Resource.Type, ResourceID: request.Resource.ID, Action: request.Action}, nil
}

func TestResourceServiceChecksPermissionBeforePolicy(t *testing.T) {
	registry := NewResourceRegistry()
	if err := registry.Register(domain.ResourceDefinition{Type: "document", Actions: []domain.ActionDefinition{{Name: "read", Permission: "document:read"}}}, testPolicy{}); err != nil {
		t.Fatal(err)
	}
	service := NewResourceService(registry)
	decision, err := service.Check(context.Background(), domain.AccessRequest{Actor: domain.Context{Subject: "user-1"}, Resource: domain.ResourceRef{Type: "document", ID: "owned"}, Action: "read"})
	if err != nil {
		t.Fatal(err)
	}
	if decision.Allowed || decision.ReasonCode != "MISSING_PERMISSION" {
		t.Fatalf("unexpected decision: %+v", decision)
	}
	decision, err = service.Check(context.Background(), domain.AccessRequest{Actor: domain.Context{Subject: "user-1", Permissions: map[string]struct{}{"document:read": {}}}, Resource: domain.ResourceRef{Type: "document", ID: "other"}, Action: "read"})
	if err != nil {
		t.Fatal(err)
	}
	if decision.Allowed || decision.ReasonCode != "RESOURCE_POLICY" {
		t.Fatalf("unexpected decision: %+v", decision)
	}
}

func TestResourceRegistryRejectsDuplicateActions(t *testing.T) {
	registry := NewResourceRegistry()
	err := registry.Register(domain.ResourceDefinition{Type: "document", Actions: []domain.ActionDefinition{{Name: "read", Permission: "document:read"}, {Name: "read", Permission: "document:read"}}}, nil)
	if err == nil {
		t.Fatal("expected duplicate action error")
	}
}
