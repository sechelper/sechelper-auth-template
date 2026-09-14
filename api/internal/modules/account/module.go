package account

import (
	"github.com/gin-gonic/gin"
	"sechelper-auth-template/api/internal/modules/account/application"
	accounthttp "sechelper-auth-template/api/internal/modules/account/transport/http"
	auditapplication "sechelper-auth-template/api/internal/modules/audit/application"
	"sechelper-auth-template/api/internal/platform/session"
)

type Module struct {
	Service *application.Service
	Handler *accounthttp.Handler
}

func New(sessions session.Store) *Module {
	service := application.NewService(sessions)
	return &Module{Service: service, Handler: accounthttp.NewHandler(service)}
}
func (m *Module) SetAuditRecorder(recorder auditapplication.Recorder) {
	m.Service.SetAuditRecorder(recorder)
}
func (m *Module) RegisterRoutes(v1 *gin.RouterGroup, readAuth, revokeAuth gin.HandlerFunc) {
	group := v1.Group("/admin/account")
	group.GET("", readAuth, m.Handler.Current)
	group.GET("/sessions", readAuth, m.Handler.ListSessions)
	group.DELETE("/sessions/:sessionId", revokeAuth, m.Handler.RevokeSession)
}
