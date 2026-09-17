package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
	authzhttp "sechelper-auth-template/api/internal/modules/authorization/transport/http"
	"sechelper-auth-template/api/internal/platform/httpkit"
	"sechelper-auth-template/api/internal/platform/session"
)

type Handler struct {
	sessions   session.Store
	cookieName string
}

func NewHandler(sessions session.Store, cookieName string) *Handler {
	return &Handler{sessions: sessions, cookieName: cookieName}
}

// CurrentUser returns the identity data needed by framework clients. It does
// not expose tokens or other session implementation details.
func (h *Handler) CurrentUser(c *gin.Context) {
	if h.sessions == nil {
		httpkit.WriteError(c, http.StatusInternalServerError, httpkit.Error{Code: "ACCOUNT_UNAVAILABLE"})
		return
	}
	sessionID, _ := c.Cookie(h.cookieName)
	value, err := h.sessions.Get(c.Request.Context(), sessionID)
	if err != nil {
		httpkit.WriteError(c, http.StatusUnauthorized, httpkit.Error{Code: "UNAUTHORIZED"})
		return
	}
	profile := value.ProfileClaims
	if profile == nil {
		profile = map[string]any{}
	}
	data := gin.H{
		"subject":         value.Subject,
		"identitySubject": value.Subject,
		"email":           value.Email,
		"applicationCode": value.ApplicationCode,
		"profile":         profile,
	}
	if value.PlatformUserUUID != "" {
		data["platformUserUuid"] = value.PlatformUserUUID
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (h *Handler) Current(c *gin.Context) {
	value, ok := authzhttp.Current(c)
	if !ok {
		httpkit.WriteError(c, http.StatusUnauthorized, httpkit.Error{Code: "UNAUTHORIZED"})
		return
	}
	permissions := make([]string, 0, len(value.Permissions))
	for permission := range value.Permissions {
		permissions = append(permissions, permission)
	}
	data := gin.H{"subject": value.Subject, "identitySubject": value.Subject, "applicationCode": value.ApplicationCode, "permissions": permissions}
	if value.PlatformUserUUID != "" {
		data["platformUserUuid"] = value.PlatformUserUUID
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}
