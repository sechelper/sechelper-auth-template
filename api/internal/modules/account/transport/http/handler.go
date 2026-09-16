package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
	authzhttp "sechelper-auth-template/api/internal/modules/authorization/transport/http"
	"sechelper-auth-template/api/internal/platform/httpkit"
)

type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

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
