package account

import (
	"context"
	"github.com/gin-gonic/gin"
	accounthttp "sechelper-auth-template/api/internal/modules/account/transport/http"
	"sechelper-auth-template/api/internal/platform/session"
)

type Module struct{ Handler *accounthttp.Handler }

func New(sessions session.Store, cookieName string, profileReader ...func(context.Context, string) (session.Session, error)) *Module {
	return &Module{Handler: accounthttp.NewHandler(sessions, cookieName, profileReader...)}
}

func (m *Module) RegisterRoutes(v1 *gin.RouterGroup, readAuth gin.HandlerFunc) {
	v1.GET("/account/me", readAuth, m.Handler.CurrentUser)
	group := v1.Group("/admin/account")
	group.GET("", readAuth, m.Handler.Current)
}
