package account

import (
	"github.com/gin-gonic/gin"
	accounthttp "sechelper-auth-template/api/internal/modules/account/transport/http"
)

type Module struct{ Handler *accounthttp.Handler }

func New() *Module { return &Module{Handler: accounthttp.NewHandler()} }

func (m *Module) RegisterRoutes(v1 *gin.RouterGroup, readAuth gin.HandlerFunc) {
	group := v1.Group("/admin/account")
	group.GET("", readAuth, m.Handler.Current)
}
