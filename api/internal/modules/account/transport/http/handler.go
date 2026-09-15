package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sechelper-auth-template/api/internal/modules/account/application"
	authzhttp "sechelper-auth-template/api/internal/modules/authorization/transport/http"
	"sechelper-auth-template/api/internal/platform/httpkit"
	"sechelper-auth-template/api/internal/platform/session"
	"time"
)

type Handler struct{ service *application.Service }

func NewHandler(service *application.Service) *Handler { return &Handler{service: service} }

type sessionResponse struct {
	ID               string    `json:"id"`
	Subject          string    `json:"subject"`
	IdentitySubject  string    `json:"identitySubject"`
	PlatformUserUUID string    `json:"platformUserUuid,omitempty"`
	Email            string    `json:"email"`
	ApplicationCode  string    `json:"applicationCode"`
	ExpiresAt        time.Time `json:"expiresAt"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	Revoked          bool      `json:"revoked"`
}

func mapSession(value session.Session) sessionResponse {
	return sessionResponse{ID: value.ID, Subject: value.Subject, IdentitySubject: value.Subject, PlatformUserUUID: value.PlatformUserUUID, Email: value.Email, ApplicationCode: value.ApplicationCode, ExpiresAt: value.ExpiresAt, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt, Revoked: value.Revoked}
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
func (h *Handler) ListSessions(c *gin.Context) {
	value, ok := authzhttp.Current(c)
	if !ok {
		httpkit.WriteError(c, http.StatusUnauthorized, httpkit.Error{Code: "UNAUTHORIZED"})
		return
	}
	items, err := h.service.ListSessions(c.Request.Context(), value.Subject, 100)
	if err != nil {
		writeError(c, http.StatusBadGateway, "SESSION_QUERY_FAILED")
		return
	}
	data := make([]sessionResponse, 0, len(items))
	for _, item := range items {
		data = append(data, mapSession(item))
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "meta": gin.H{"hasMore": false}})
}
func (h *Handler) RevokeSession(c *gin.Context) {
	current, ok := authzhttp.Current(c)
	if !ok {
		httpkit.WriteError(c, http.StatusUnauthorized, httpkit.Error{Code: "UNAUTHORIZED"})
		return
	}
	if err := h.service.RevokeSession(c.Request.Context(), current.Subject, c.Param("sessionId")); err != nil {
		writeError(c, http.StatusNotFound, "SESSION_NOT_FOUND")
		return
	}
	c.Status(http.StatusNoContent)
}
func writeError(c *gin.Context, status int, code string) {
	httpkit.WriteError(c, status, httpkit.Error{Code: code})
}
