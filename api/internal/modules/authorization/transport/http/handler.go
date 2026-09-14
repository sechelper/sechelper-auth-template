package http

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler { return &Handler{} }
func (h *Handler) Me(c *gin.Context) {
	value, ok := Current(c)
	if !ok {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	permissions := make([]string, 0, len(value.Permissions))
	for code := range value.Permissions {
		permissions = append(permissions, code)
	}
	sort.Strings(permissions)
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"subject": value.Subject, "applicationCode": value.ApplicationCode, "permissions": permissions}})
}
