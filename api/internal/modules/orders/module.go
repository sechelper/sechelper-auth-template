package orders

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	authzdomain "sechelper-auth-template/api/internal/modules/authorization/domain"
	"sechelper-auth-template/api/internal/modules/manifest/domain"
	"sechelper-auth-template/api/internal/modules/orders/application"
	"sechelper-auth-template/api/internal/modules/orders/persistence"
	ordershttp "sechelper-auth-template/api/internal/modules/orders/transport/http"
)

type Module struct {
	Service *application.Service
	Handler *ordershttp.Handler
}

func New(db *sql.DB) *Module {
	service := application.NewService(persistence.NewRepository(db))
	return &Module{Service: service, Handler: ordershttp.NewHandler(service)}
}
func (m *Module) RegisterPermissions(registry interface{ Register(domain.Permission) error }) error {
	return registry.Register(domain.Permission{Code: "order:read", Name: "查看订单", Description: "查看订单列表和详情", RiskLevel: "normal", APIs: []domain.API{{Method: "GET", Path: "/v1/orders"}, {Method: "GET", Path: "/v1/orders/{orderId}"}}})
}
func (m *Module) RegisterResources(registry interface {
	RegisterResource(authzdomain.ResourceDefinition, authzdomain.ResourcePolicy) error
}) error {
	return registry.RegisterResource(authzdomain.ResourceDefinition{Type: "example-order", Name: "示例订单", Description: "用于验证业务模块资源访问接入的案例资源", Actions: []authzdomain.ActionDefinition{{Name: "read", Permission: "order:read"}}}, nil)
}
func (m *Module) RegisterRoutes(v1 *gin.RouterGroup, auth gin.HandlerFunc) {
	group := v1.Group("/orders")
	group.GET("", auth, m.Handler.List)
	group.GET("/:orderId", auth, m.Handler.Get)
}
