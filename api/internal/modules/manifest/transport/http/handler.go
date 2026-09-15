package http

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"sechelper-auth-template/api/internal/modules/manifest/application"
	"sechelper-auth-template/api/internal/platform/httpkit"
)

type Handler struct {
	service       *application.SyncService
	registry      *application.Registry
	auditRecorder func(context.Context, application.State, error, string)
}

func (h *Handler) SetAuditRecorder(recorder func(context.Context, application.State, error, string)) {
	h.auditRecorder = recorder
}

type stateResponse struct {
	ApplicationCode string    `json:"applicationCode"`
	ManifestVersion int64     `json:"manifestVersion"`
	ContentHash     string    `json:"contentHash"`
	CanonicalJSON   string    `json:"canonicalJson"`
	SyncID          string    `json:"syncId"`
	Status          string    `json:"status"`
	ServerRevision  int64     `json:"serverRevision"`
	AcceptedAt      time.Time `json:"acceptedAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
type permissionResponse struct {
	Code        string           `json:"code"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	RiskLevel   string           `json:"riskLevel"`
	APIs        []applicationAPI `json:"apis"`
}
type applicationAPI struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

func toStateResponse(value application.State) stateResponse {
	return stateResponse{ApplicationCode: value.ApplicationCode, ManifestVersion: value.ManifestVersion, ContentHash: value.ContentHash, CanonicalJSON: value.CanonicalJSON, SyncID: value.SyncID, Status: value.Status, ServerRevision: value.ServerRevision, AcceptedAt: value.AcceptedAt, UpdatedAt: value.UpdatedAt}
}

func NewHandler(service *application.SyncService, registry *application.Registry) *Handler {
	return &Handler{service: service, registry: registry}
}
func (h *Handler) Permissions(c *gin.Context) {
	values := h.registry.Permissions()
	data := make([]permissionResponse, 0, len(values))
	for _, value := range values {
		apis := make([]applicationAPI, 0, len(value.APIs))
		for _, api := range value.APIs {
			apis = append(apis, applicationAPI{Method: api.Method, Path: api.Path})
		}
		data = append(data, permissionResponse{Code: value.Code, Name: value.Name, Description: value.Description, RiskLevel: value.RiskLevel, APIs: apis})
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "meta": gin.H{"hasMore": false}})
}
func (h *Handler) Status(c *gin.Context) {
	value, err := h.service.Current(c.Request.Context())
	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "not_synced"}})
		return
	}
	if err != nil {
		writeError(c, http.StatusInternalServerError, "MANIFEST_STATUS_FAILED")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toStateResponse(value)})
}
func (h *Handler) Sync(c *gin.Context) {
	value, err := h.service.Sync(c.Request.Context())
	if h.auditRecorder != nil {
		h.auditRecorder(c.Request.Context(), value, err, httpkit.RequestID(c))
	}
	if err != nil {
		writeError(c, http.StatusBadGateway, "MANIFEST_SYNC_FAILED")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toStateResponse(value)})
}
func writeError(c *gin.Context, status int, code string) {
	httpkit.WriteError(c, status, httpkit.Error{Code: code})
}
