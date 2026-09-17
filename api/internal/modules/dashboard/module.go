package dashboard

import (
	"github.com/gin-gonic/gin"
	"sechelper-auth-template/api/internal/modules/dashboard/application"
	dashboardhttp "sechelper-auth-template/api/internal/modules/dashboard/transport/http"
)

type Module struct {
	Service *application.Service
	Handler *dashboardhttp.Handler
}

func New(service *application.Service) *Module {
	return &Module{Service: service, Handler: dashboardhttp.NewHandler(service)}
}

func (m *Module) RegisterRoutes(v1 *gin.RouterGroup, auth gin.HandlerFunc) {
	v1.GET("/version", m.Handler.Version)
	v1.GET("/runtime-config", m.Handler.RuntimeConfig)
	v1.GET("/admin/dashboard/overview", auth, m.Handler.Overview)
	v1.GET("/admin/dashboard/resources", auth, m.Handler.Resources)
}
