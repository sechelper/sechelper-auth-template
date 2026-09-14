package authorization

import (
	"github.com/gin-gonic/gin"
	"sechelper-auth-template/api/internal/modules/authorization/application"
	"sechelper-auth-template/api/internal/modules/authorization/domain"
	authhttp "sechelper-auth-template/api/internal/modules/authorization/transport/http"
	"sechelper-auth-template/api/internal/platform/session"
)

type Module struct {
	Service          *application.Service
	Middleware       *authhttp.Middleware
	Handler          *authhttp.Handler
	ResourceRegistry *application.ResourceRegistry
	ResourceService  *application.ResourceService
	ResourceHandler  *authhttp.ResourceHandler
}

func New(sessions session.Store, cookieName string, cache application.Cache) *Module {
	service := application.NewService(sessions, cache)
	resourceRegistry := application.NewResourceRegistry()
	resourceService := application.NewResourceService(resourceRegistry)
	return &Module{Service: service, Middleware: authhttp.NewMiddleware(service, cookieName), Handler: authhttp.NewHandler(), ResourceRegistry: resourceRegistry, ResourceService: resourceService, ResourceHandler: authhttp.NewResourceHandler(resourceService)}
}
func (m *Module) RegisterRoutes(v1 *gin.RouterGroup) {
	v1.GET("/authorization/me", m.Middleware.RequireAuthentication(), m.Handler.Me)
}

func (m *Module) RegisterResource(definition domain.ResourceDefinition, policy domain.ResourcePolicy) error {
	return m.ResourceRegistry.Register(definition, policy)
}

func (m *Module) RegisterResourceRoutes(v1 *gin.RouterGroup, auth gin.HandlerFunc) {
	v1.GET("/admin/resources", auth, m.ResourceHandler.List)
	v1.POST("/admin/access-decisions/check", auth, m.ResourceHandler.Check)
}
