package http

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"sechelper-auth-template/api/internal/modules/authorization/application"
	"sechelper-auth-template/api/internal/modules/authorization/domain"
	"sechelper-auth-template/api/internal/platform/httpkit"
	"strings"
)

type ResourceHandler struct {
	service  *application.ResourceService
	recorder func(context.Context, domain.Context, domain.Decision, string)
}

func NewResourceHandler(service *application.ResourceService) *ResourceHandler {
	return &ResourceHandler{service: service}
}
func (h *ResourceHandler) SetDecisionRecorder(recorder func(context.Context, domain.Context, domain.Decision, string)) {
	h.recorder = recorder
}

type accessCheckRequest struct {
	ResourceType string `json:"resourceType"`
	ResourceID   string `json:"resourceId"`
	Action       string `json:"action"`
}

func (h *ResourceHandler) Check(c *gin.Context) {
	var input accessCheckRequest
	if err := c.ShouldBindJSON(&input); err != nil || strings.TrimSpace(input.ResourceType) == "" || strings.TrimSpace(input.ResourceID) == "" || strings.TrimSpace(input.Action) == "" {
		writeResourceError(c, http.StatusUnprocessableEntity, "INVALID_ACCESS_CHECK")
		return
	}
	actor, ok := Current(c)
	if !ok {
		writeResourceError(c, http.StatusUnauthorized, "UNAUTHORIZED")
		return
	}
	decision, err := h.service.Check(c.Request.Context(), domain.AccessRequest{Actor: actor, Resource: domain.ResourceRef{Type: input.ResourceType, ID: input.ResourceID}, Action: input.Action})
	if errors.Is(err, application.ErrUnknownResource) || errors.Is(err, application.ErrUnknownAction) {
		writeResourceError(c, http.StatusUnprocessableEntity, "UNKNOWN_RESOURCE_ACTION")
		return
	}
	if err != nil {
		writeResourceError(c, http.StatusBadGateway, "ACCESS_CHECK_FAILED")
		return
	}
	if h.recorder != nil {
		h.recorder(c.Request.Context(), actor, decision, httpkit.RequestID(c))
	}
	c.JSON(http.StatusOK, gin.H{"data": decision})
}
func (h *ResourceHandler) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": h.service.Definitions(), "meta": gin.H{"hasMore": false}})
}
func writeResourceError(c *gin.Context, status int, code string) {
	httpkit.WriteError(c, status, httpkit.Error{Code: code})
}
