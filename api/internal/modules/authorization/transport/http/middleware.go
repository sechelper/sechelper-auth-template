package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"sechelper-auth-template/api/internal/modules/authorization/application"
	"sechelper-auth-template/api/internal/modules/authorization/domain"
)

const contextKey = "authorization.context"

type Middleware struct {
	service    *application.Service
	cookieName string
}

func NewMiddleware(service *application.Service, cookieName string) *Middleware {
	return &Middleware{service: service, cookieName: cookieName}
}
func (m *Middleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, _ := c.Cookie(m.cookieName)
		value, err := m.service.Require(c.Request.Context(), sessionID, permission)
		if err != nil {
			writeAuthorizationError(c, err)
			return
		}
		c.Set(contextKey, value)
		c.Next()
	}
}

func (m *Middleware) RequirePermissions(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, _ := c.Cookie(m.cookieName)
		var value domain.Context
		for _, permission := range permissions {
			resolved, err := m.service.Require(c.Request.Context(), sessionID, permission)
			if err != nil {
				writeAuthorizationError(c, err)
				return
			}
			value = resolved
		}
		c.Set(contextKey, value)
		c.Next()
	}
}
func (m *Middleware) RequireAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, _ := c.Cookie(m.cookieName)
		value, err := m.service.Resolve(c.Request.Context(), sessionID)
		if err != nil {
			writeAuthorizationError(c, err)
			return
		}
		c.Set(contextKey, value)
		c.Next()
	}
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
	c.AbortWithStatusJSON(status, gin.H{"error": gin.H{"code": code, "message": code, "requestId": c.GetHeader("X-Request-ID")}})
}
