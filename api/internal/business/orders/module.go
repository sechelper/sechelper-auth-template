//go:build example

package orders

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	frameworkapp "sechelper-auth-template/api/internal/application"
	ordersapp "sechelper-auth-template/api/internal/business/orders/application"
	"sechelper-auth-template/api/internal/business/orders/persistence"
	ordershttp "sechelper-auth-template/api/internal/business/orders/transport/http"
	authzdomain "sechelper-auth-template/api/internal/modules/authorization/domain"
	"sechelper-auth-template/api/internal/modules/manifest/domain"
)

type Module struct {
	Service *ordersapp.Service
	Handler *ordershttp.Handler
	logger  frameworkapp.Logger
}

func New(db *sql.DB) *Module {
	service := ordersapp.NewService(persistence.NewRepository(db))
	return &Module{Service: service, Handler: ordershttp.NewHandler(service)}
}
func (m *Module) Name() string                         { return "orders" }
func (m *Module) SetLogger(logger frameworkapp.Logger) { m.logger = logger }

func (m *Module) RegisterPermissions(registry frameworkapp.PermissionRegistrar) error {
	return registry.Register(domain.Permission{Code: "order:read", Name: "查看订单", Description: "查看订单列表和详情", RiskLevel: "normal", APIs: []domain.API{{Method: "GET", Path: "/v1/orders"}, {Method: "GET", Path: "/v1/orders/{orderId}"}}})
}
func (m *Module) RegisterResources(registry frameworkapp.ResourceRegistrar) error {
	return registry.RegisterResource(authzdomain.ResourceDefinition{Type: "example-order", Name: "示例订单", Description: "用于验证业务模块资源访问接入的案例资源", Actions: []authzdomain.ActionDefinition{{Name: "read", Permission: "order:read"}}}, nil)
}
func (m *Module) RegisterAuditEvents(_ frameworkapp.AuditRegistrar) error { return nil }

func (m *Module) RegisterRoutes(v1 *gin.RouterGroup, authorizer frameworkapp.RouteAuthorizer) error {
	group := v1.Group("/orders")
	group.GET("", authorizer.RequirePermissions("admin:access", "order:read"), m.Handler.List)
	group.GET("/:orderId", authorizer.RequirePermissions("admin:access", "order:read"), m.Handler.Get)
	return nil
}
