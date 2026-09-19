package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"sechelper-auth-template/api/internal/modules/authorization/application"
	"sechelper-auth-template/api/internal/modules/authorization/domain"
	"sechelper-auth-template/api/internal/platform/httpkit"
)

const contextKey = "authorization.context"

type Middleware struct {
	service    *application.Service
	cookieName string
	surface    string
}

func NewMiddleware(service *application.Service, cookieName string, surface ...string) *Middleware {
	value := "public"
	if len(surface) > 0 && surface[0] != "" {
		value = surface[0]
	}
	return &Middleware{service: service, cookieName: cookieName, surface: value}
}
func (m *Middleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, _ := c.Cookie(m.cookieName)
		value, err := m.service.RequireForSurface(c.Request.Context(), sessionID, m.surface, permission)
		if err != nil {
			writeAuthorizationError(c, err)
			return
		}
		setCurrent(c, value)
		c.Next()
	}
}

func (m *Middleware) RequirePermissions(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, _ := c.Cookie(m.cookieName)
		var value domain.Context
		for _, permission := range permissions {
			resolved, err := m.service.RequireForSurface(c.Request.Context(), sessionID, m.surface, permission)
			if err != nil {
				writeAuthorizationError(c, err)
				return
			}
			value = resolved
		}
		setCurrent(c, value)
		c.Next()
	}
}
func (m *Middleware) RequireAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, _ := c.Cookie(m.cookieName)
		value, err := m.service.ResolveForSurface(c.Request.Context(), sessionID, m.surface)
		if err != nil {
			writeAuthorizationError(c, err)
			return
		}
		setCurrent(c, value)
		c.Next()
	}
}

func setCurrent(c *gin.Context, value domain.Context) {
	c.Set(contextKey, value)
	c.Request = c.Request.WithContext(domain.WithContext(c.Request.Context(), value))
}

func Current(c *gin.Context) (domain.Context, bool) {
	value, ok := c.Get(contextKey)
	if !ok {
		return domain.Context{}, false
	}
	result, ok := value.(domain.Context)
	return result, ok
}
func writeAuthorizationError(c *gin.Context, err error) {
	status, code := http.StatusUnauthorized, "UNAUTHORIZED"
	if errors.Is(err, application.ErrForbidden) {
		status, code = http.StatusForbidden, "FORBIDDEN"
	} else if errors.Is(err, application.ErrDependency) {
		status, code = http.StatusBadGateway, "AUTHORIZATION_DEPENDENCY_FAILED"
	}
	httpkit.WriteError(c, status, httpkit.Error{Code: code})
}
