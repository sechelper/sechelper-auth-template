package application

import (
	"context"
	"errors"
	"testing"

	"github.com/gin-gonic/gin"
	"sechelper-auth-template/api/internal/modules/authorization/domain"
	manifestdomain "sechelper-auth-template/api/internal/modules/manifest/domain"
	platformaudit "sechelper-auth-template/api/internal/platform/audit"
	platformlogging "sechelper-auth-template/api/internal/platform/logging"
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

func TestRegistryInjectsLoggerIntoLoggerAwareModules(t *testing.T) {
	module := &loggerAwareModule{registryModule: registryModule{name: "orders", steps: &[]string{}}}
	registry := NewRegistry()
	if err := registry.Add(module); err != nil {
		t.Fatal(err)
	}
	logger := stubLogger{}
	registry.SetLogger(logger)
	if module.logger != Logger(logger) {
		t.Fatal("logger was not injected into the module")
	}
}

type loggerAwareModule struct {
	registryModule
	logger Logger
}

func (m *loggerAwareModule) SetLogger(logger Logger) { m.logger = logger }

type stubLogger struct{}

func (stubLogger) Debug(context.Context, string, ...platformlogging.Field) {}
func (stubLogger) Info(context.Context, string, ...platformlogging.Field)  {}
func (stubLogger) Warn(context.Context, string, ...platformlogging.Field)  {}
func (stubLogger) Error(context.Context, string, ...platformlogging.Field) {}

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
