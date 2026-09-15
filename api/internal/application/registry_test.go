package application

import (
	"errors"
	"testing"

	"github.com/gin-gonic/gin"
	"sechelper-auth-template/api/internal/modules/authorization/domain"
	manifestdomain "sechelper-auth-template/api/internal/modules/manifest/domain"
	platformaudit "sechelper-auth-template/api/internal/platform/audit"
)

type registryModule struct {
	name   string
	steps  *[]string
	failAt string
}

func (m registryModule) Name() string { return m.name }
func (m registryModule) RegisterPermissions(PermissionRegistrar) error {
	*m.steps = append(*m.steps, m.name+":permissions")
	return m.failure("permissions")
}
func (m registryModule) RegisterResources(ResourceRegistrar) error {
	*m.steps = append(*m.steps, m.name+":resources")
	return m.failure("resources")
}
func (m registryModule) RegisterAuditEvents(AuditRegistrar) error {
	*m.steps = append(*m.steps, m.name+":audit")
	return m.failure("audit")
}
func (m registryModule) RegisterRoutes(*gin.RouterGroup, RouteAuthorizer) error {
	*m.steps = append(*m.steps, m.name+":routes")
	return m.failure("routes")
}
func (m registryModule) failure(step string) error {
	if m.failAt == step {
		return errors.New("registration failed")
	}
	return nil
}

func TestRegistryRegistersCapabilitiesInModuleOrder(t *testing.T) {
	steps := []string{}
	registry := NewRegistry()
	if err := registry.Add(registryModule{name: "first", steps: &steps}); err != nil {
		t.Fatal(err)
	}
	if err := registry.Add(registryModule{name: "second", steps: &steps}); err != nil {
		t.Fatal(err)
	}
	if err := registry.RegisterPermissions(permissionRegistrar{}); err != nil {
		t.Fatal(err)
	}
	if err := registry.RegisterResources(resourceRegistrar{}); err != nil {
		t.Fatal(err)
	}
	if err := registry.RegisterAuditEvents(auditRegistrar{}); err != nil {
		t.Fatal(err)
	}
	if err := registry.RegisterRoutes(nil, routeAuthorizer{}); err != nil {
		t.Fatal(err)
	}
	want := []string{"first:permissions", "second:permissions", "first:resources", "second:resources", "first:audit", "second:audit", "first:routes", "second:routes"}
	if len(steps) != len(want) {
		t.Fatalf("registered steps = %v, want %v", steps, want)
	}
	for i := range want {
		if steps[i] != want[i] {
			t.Fatalf("registered steps = %v, want %v", steps, want)
		}
	}
}

func TestRegistryRejectsDuplicateModuleNamesAndPropagatesRegistrationErrors(t *testing.T) {
	steps := []string{}
	registry := NewRegistry()
	if err := registry.Add(registryModule{name: "orders", steps: &steps}); err != nil {
		t.Fatal(err)
	}
	if err := registry.Add(registryModule{name: "orders", steps: &steps}); err == nil {
		t.Fatal("expected duplicate module error")
	}
	failing := NewRegistry()
	if err := failing.Add(registryModule{name: "broken", steps: &steps, failAt: "resources"}); err != nil {
		t.Fatal(err)
	}
	if err := failing.RegisterResources(resourceRegistrar{}); err == nil {
		t.Fatal("expected resource registration error")
	}
}

type permissionRegistrar struct{}

func (permissionRegistrar) Register(manifestdomain.Permission) error { return nil }

type resourceRegistrar struct{}

func (resourceRegistrar) RegisterResource(domain.ResourceDefinition, domain.ResourcePolicy) error {
	return nil
}

type auditRegistrar struct{}

func (auditRegistrar) RegisterEventType(string, platformaudit.EventDefinition) error { return nil }

type routeAuthorizer struct{}

func (routeAuthorizer) RequirePermission(string) gin.HandlerFunc     { return nil }
func (routeAuthorizer) RequirePermissions(...string) gin.HandlerFunc { return nil }
