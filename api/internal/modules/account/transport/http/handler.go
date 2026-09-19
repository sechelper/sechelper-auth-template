package http

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	authzhttp "sechelper-auth-template/api/internal/modules/authorization/transport/http"
	"sechelper-auth-template/api/internal/platform/httpkit"
	"sechelper-auth-template/api/internal/platform/session"
)

type Handler struct {
	sessions      session.Store
	cookieName    string
	profileReader func(context.Context, string) (session.Session, error)
}

func NewHandler(sessions session.Store, cookieName string, profileReader ...func(context.Context, string) (session.Session, error)) *Handler {
	handler := &Handler{sessions: sessions, cookieName: cookieName}
	if len(profileReader) > 0 {
		handler.profileReader = profileReader[0]
	}
	return handler
}

// CurrentUser returns the identity data needed by framework clients. It does
// not expose tokens or other session implementation details.
func (h *Handler) CurrentUser(c *gin.Context) {
	if h.sessions == nil {
		httpkit.WriteError(c, http.StatusInternalServerError, httpkit.Error{Code: "ACCOUNT_UNAVAILABLE"})
		return
	}
	sessionID, _ := c.Cookie(h.cookieName)
	reader := func(ctx context.Context, id string) (session.Session, error) { return h.sessions.Get(ctx, id) }
	if h.profileReader != nil {
		reader = h.profileReader
	}
	value, err := reader(c.Request.Context(), sessionID)
	if err != nil {
		httpkit.WriteError(c, http.StatusUnauthorized, httpkit.Error{Code: "UNAUTHORIZED"})
		return
	}
	profile := map[string]any{}
	for _, key := range []string{"sub", "email", "name", "nickname", "picture"} {
		if item, ok := value.ProfileClaims[key].(string); ok && item != "" {
			profile[key] = item
		}
	}
	if _, ok := profile["sub"]; !ok {
		profile["sub"] = value.Subject
	}
	if _, ok := profile["email"]; !ok && value.Email != "" {
		profile["email"] = value.Email
	}
	data := gin.H{
		"subject":         value.Subject,
		"identitySubject": value.Subject,
		"email":           value.Email,
		"applicationCode": value.ApplicationCode,
		"profile":         profile,
		"nickname":        accountProfileNickname(profile, value.Subject),
		"displayName":     accountProfileNickname(profile, value.Subject),
	}
	if picture, ok := profile["picture"].(string); ok && picture != "" {
		data["avatarUrl"] = picture
	}
	if value.PlatformUserUUID != "" {
		data["platformUserUuid"] = value.PlatformUserUUID
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func accountProfileNickname(profile map[string]any, fallback string) string {
	for _, key := range []string{"nickname", "name"} {
		if value, ok := profile[key].(string); ok && value != "" {
			return value
		}
	}
	return fallback
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
