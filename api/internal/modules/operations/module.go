package operations

import (
	"github.com/gin-gonic/gin"
	"sechelper-auth-template/api/internal/modules/operations/application"
	operationshttp "sechelper-auth-template/api/internal/modules/operations/transport/http"
)

type Module struct {
	Service *application.Service
	Handler *operationshttp.Handler
}

func New(service *application.Service) *Module {
	return &Module{Service: service, Handler: operationshttp.NewHandler(service)}
}
func (m *Module) RegisterRoutes(v1 *gin.RouterGroup, auth, deploymentAuth gin.HandlerFunc) {
	v1.GET("/admin/operations/overview", auth, m.Handler.Overview)
	v1.GET("/admin/deployment/guide", deploymentAuth, m.Handler.DeploymentGuide)
}
