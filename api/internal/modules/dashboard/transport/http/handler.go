package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	authzhttp "sechelper-auth-template/api/internal/modules/authorization/transport/http"
	"sechelper-auth-template/api/internal/modules/dashboard/application"
)

type Handler struct{ service *application.Service }

func NewHandler(service *application.Service) *Handler { return &Handler{service: service} }

type dependencyResponse struct {
	Status    string `json:"status"`
	LatencyMs int64  `json:"latencyMs,omitempty"`
	Message   string `json:"message,omitempty"`
}

func (h *Handler) Overview(c *gin.Context) {
	current, ok := authzhttp.Current(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "UNAUTHORIZED", "requestId": c.GetHeader("X-Request-ID")}})
		return
	}
	value := h.service.Overview(c.Request.Context(), application.CurrentUser{
		Subject:         current.Subject,
		ApplicationCode: current.ApplicationCode,
		PermissionCount: len(current.Permissions),
	})
	dependencies := make(map[string]dependencyResponse, len(value.Dependencies))
	for name, dependency := range value.Dependencies {
		dependencies[name] = dependencyResponse{Status: dependency.Status, LatencyMs: dependency.LatencyMs, Message: dependency.Message}
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"application": gin.H{
			"name": value.App.Name, "version": value.App.Version, "environment": value.App.Environment,
			"startedAt": value.App.StartedAt.UTC().Format(time.RFC3339),
		},
		"currentSession": gin.H{
			"subject": value.Current.Subject, "applicationCode": value.Current.ApplicationCode, "permissionCount": value.Current.PermissionCount,
		},
		"dependencies": dependencies,
		"manifest": gin.H{
			"applicationCode": value.Manifest.ApplicationCode, "version": value.Manifest.Version, "status": value.Manifest.Status,
			"contentHash": value.Manifest.ContentHash, "serverRevision": value.Manifest.ServerRevision, "updatedAt": value.Manifest.UpdatedAt.UTC().Format(time.RFC3339),
		},
	}})
}
