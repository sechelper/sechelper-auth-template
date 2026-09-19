package configuration

import (
	"context"
	"github.com/gin-gonic/gin"
	"sechelper-auth-template/api/internal/modules/configuration/application"
	configurationhttp "sechelper-auth-template/api/internal/modules/configuration/transport/http"
)

type Module struct {
	Service *application.Service
	Handler *configurationhttp.Handler
}

func New(service *application.Service, audit func(context.Context, string, string, string)) *Module {
	return &Module{Service: service, Handler: configurationhttp.NewHandler(service, func(ctx context.Context, actor, action, key string) {
		if audit != nil {
			audit(ctx, actor, action, key)
		}
	})}
}
func (m *Module) RegisterRoutes(v1 *gin.RouterGroup, readAuth, writeAuth gin.HandlerFunc) {
	group := v1.Group("/admin/configuration")
	group.GET("", readAuth, m.Handler.List)
	group.GET("/:key", readAuth, m.Handler.Get)
	group.PUT("/:key", writeAuth, m.Handler.Put)
	group.DELETE("/:key", writeAuth, m.Handler.Delete)
	group.POST("/restart", writeAuth, m.Handler.Restart)
	business := v1.Group("/admin/business-configuration")
	business.PUT("/:key", writeAuth, m.Handler.PutBusiness)
	business.DELETE("/:key", writeAuth, m.Handler.DeleteBusiness)
}

type Provider struct{ service *application.Service }

func (p Provider) Get(ctx context.Context, key string) (string, bool) {
	return p.service.GetValue(ctx, key)
}
func (m *Module) Provider() Provider { return Provider{service: m.Service} }
