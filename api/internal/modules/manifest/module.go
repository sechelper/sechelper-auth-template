package manifest

import (
	"context"
	"github.com/gin-gonic/gin"
	"sechelper-auth-template/api/internal/modules/manifest/application"
	"sechelper-auth-template/api/internal/modules/manifest/domain"
	manifesthttp "sechelper-auth-template/api/internal/modules/manifest/transport/http"
)

type Module struct {
	Registry *application.Registry
	Sync     *application.SyncService
	Handler  *manifesthttp.Handler
}

func New(applicationCode string) *Module {
	return &Module{Registry: application.NewRegistry(applicationCode)}
}
func (m *Module) Register(p domain.Permission) error { return m.Registry.Register(p) }
func (m *Module) AttachSync(service *application.SyncService) {
	m.Sync = service
	m.Handler = manifesthttp.NewHandler(service, m.Registry)
}
func (m *Module) SetAuditRecorder(recorder func(context.Context, application.State, error, string)) {
	if m.Handler != nil {
		m.Handler.SetAuditRecorder(recorder)
	}
}
func (m *Module) RegisterRoutes(v1 *gin.RouterGroup, statusAuth, syncAuth gin.HandlerFunc) {
	if m.Handler == nil {
		return
	}
	group := v1.Group("/internal/authorization-manifest")
	group.GET("", statusAuth, m.Handler.Status)
	v1.GET("/admin/permissions", statusAuth, m.Handler.Permissions)
	group.POST("/sync", syncAuth, m.Handler.Sync)
}
