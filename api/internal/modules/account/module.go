package account

import (
	"context"
	"github.com/gin-gonic/gin"
	accounthttp "sechelper-auth-template/api/internal/modules/account/transport/http"
	"sechelper-auth-template/api/internal/platform/session"
)

type Module struct {
	Handler      *accounthttp.Handler
	AdminHandler *accounthttp.Handler
}

func New(sessions session.Store, cookieName string, profileReaders ...func(context.Context, string) (session.Session, error)) *Module {
	adminCookieName := cookieName + "_admin"
	var profileReader []func(context.Context, string) (session.Session, error)
	if len(profileReaders) > 0 {
		profileReader = profileReaders[:1]
	}
	return &Module{Handler: accounthttp.NewHandler(sessions, cookieName, profileReader...), AdminHandler: accounthttp.NewHandler(sessions, adminCookieName, profileReader...)}
}

func (m *Module) RegisterRoutes(v1 *gin.RouterGroup, publicAuth gin.HandlerFunc, adminAuth ...gin.HandlerFunc) {
	adminGuard := publicAuth
	if len(adminAuth) > 0 {
		adminGuard = adminAuth[0]
	}
	v1.GET("/account/me", publicAuth, m.Handler.CurrentUser)
	group := v1.Group("/admin/account")
	group.GET("", adminGuard, m.AdminHandler.Current)
}
