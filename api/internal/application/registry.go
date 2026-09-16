// Package application contains the only business-module composition boundary.
// Framework packages must never import this package's business implementations.
package application

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"sechelper-auth-template/api/internal/modules/authorization/domain"
	manifestdomain "sechelper-auth-template/api/internal/modules/manifest/domain"
	platformaudit "sechelper-auth-template/api/internal/platform/audit"
)

type PermissionRegistrar interface {
	Register(manifestdomain.Permission) error
}

type ResourceRegistrar interface {
	RegisterResource(domain.ResourceDefinition, domain.ResourcePolicy) error
}

type AuditRegistrar = platformaudit.DefinitionRegistrar

type RouteAuthorizer interface {
	RequirePermission(string) gin.HandlerFunc
	RequirePermissions(...string) gin.HandlerFunc
}

// ConfigurationProvider is the only supported business-module configuration
// boundary. Values are resolved by the host; modules never read env or storage.
type ConfigurationProvider interface {
	Get(context.Context, string) (string, bool)
}

// BusinessModule is the server-side extension protocol. It keeps module code
// behind host-provided contracts while letting the host assemble permissions,
// resource definitions, audit event types, and protected routes consistently.
type BusinessModule interface {
	Name() string
	RegisterPermissions(PermissionRegistrar) error
	RegisterResources(ResourceRegistrar) error
	RegisterAuditEvents(AuditRegistrar) error
	RegisterRoutes(*gin.RouterGroup, RouteAuthorizer) error
}
type ConfigurationAware interface{ SetConfiguration(ConfigurationProvider) }

type Registry struct{ modules []BusinessModule }

func NewRegistry() *Registry { return &Registry{} }

func (r *Registry) Add(module BusinessModule) error {
	if module == nil || module.Name() == "" {
		return fmt.Errorf("business module name is required")
	}
	for _, existing := range r.modules {
		if existing.Name() == module.Name() {
			return fmt.Errorf("duplicate business module: %s", module.Name())
		}
	}
	r.modules = append(r.modules, module)
	return nil
}

func (r *Registry) Modules() []BusinessModule { return append([]BusinessModule(nil), r.modules...) }

func (r *Registry) SetConfiguration(provider ConfigurationProvider) {
	for _, module := range r.modules {
		if aware, ok := module.(ConfigurationAware); ok {
			aware.SetConfiguration(provider)
		}
	}
}

func (r *Registry) RegisterPermissions(registrar PermissionRegistrar) error {
	for _, module := range r.modules {
		if err := module.RegisterPermissions(registrar); err != nil {
			return fmt.Errorf("register permissions for business module %q: %w", module.Name(), err)
		}
	}
	return nil
}

func (r *Registry) RegisterResources(registrar ResourceRegistrar) error {
	for _, module := range r.modules {
		if err := module.RegisterResources(registrar); err != nil {
			return fmt.Errorf("register resources for business module %q: %w", module.Name(), err)
		}
	}
	return nil
}

func (r *Registry) RegisterAuditEvents(registrar AuditRegistrar) error {
	for _, module := range r.modules {
		if err := module.RegisterAuditEvents(registrar); err != nil {
			return fmt.Errorf("register audit events for business module %q: %w", module.Name(), err)
		}
	}
	return nil
}

func (r *Registry) RegisterRoutes(router *gin.RouterGroup, authorizer RouteAuthorizer) error {
	for _, module := range r.modules {
		if err := module.RegisterRoutes(router, authorizer); err != nil {
			return fmt.Errorf("register routes for business module %q: %w", module.Name(), err)
		}
	}
	return nil
}
