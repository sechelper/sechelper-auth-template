package audit

import (
	"github.com/gin-gonic/gin"
	"sechelper-auth-template/api/internal/modules/audit/application"
	audithttp "sechelper-auth-template/api/internal/modules/audit/transport/http"
)

type Module struct {
	Service *application.Service
	Handler *audithttp.Handler
}

func New(service *application.Service) *Module {
	return &Module{Service: service, Handler: audithttp.NewHandler(service)}
}
func (m *Module) RegisterRoutes(v1 *gin.RouterGroup, auth gin.HandlerFunc) {
	v1.GET("/admin/audit-events", auth, m.Handler.List)
}
