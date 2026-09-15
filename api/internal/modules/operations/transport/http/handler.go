package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sechelper-auth-template/api/internal/modules/operations/application"
	"time"
)

type Handler struct{ service *application.Service }

func NewHandler(service *application.Service) *Handler { return &Handler{service: service} }
func (h *Handler) Overview(c *gin.Context) {
	value := h.service.Overview(c.Request.Context())
	dependencies := map[string]any{}
	for name, item := range value.Dependencies {
		dependencies[name] = item
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"application": gin.H{"name": value.App.Name, "version": value.App.Version, "environment": value.App.Environment, "buildId": value.App.BuildID, "sourceRevision": value.App.SourceRevision, "startedAt": value.App.StartedAt.UTC().Format(time.RFC3339)}, "dependencies": dependencies, "metrics": value.Metrics}})
}
