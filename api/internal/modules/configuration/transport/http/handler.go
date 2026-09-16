package http

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	stdhttp "net/http"
	authhttp "sechelper-auth-template/api/internal/modules/authorization/transport/http"
	"sechelper-auth-template/api/internal/modules/configuration/application"
	"sechelper-auth-template/api/internal/platform/httpkit"
)

type Handler struct {
	service *application.Service
	audit   func(string, string, string)
}

func NewHandler(service *application.Service, audit func(string, string, string)) *Handler {
	return &Handler{service: service, audit: audit}
}
func (h *Handler) List(c *gin.Context) {
	values, err := h.service.List(c.Request.Context())
	if err != nil {
		httpkit.WriteError(c, stdhttp.StatusBadGateway, httpkit.Error{Code: "CONFIGURATION_QUERY_FAILED"})
		return
	}
	httpkit.WriteCollection(c, values, gin.H{"hasMore": false})
}
func (h *Handler) Get(c *gin.Context) {
	value, err := h.service.Get(c.Request.Context(), c.Param("key"))
	if err != nil {
		status := stdhttp.StatusBadGateway
		code := "CONFIGURATION_QUERY_FAILED"
		if err == sql.ErrNoRows {
			status, code = stdhttp.StatusNotFound, "CONFIGURATION_NOT_FOUND"
		}
		httpkit.WriteError(c, status, httpkit.Error{Code: code})
		return
	}
	if value.IsSecret {
		value.Value = ""
	}
	httpkit.WriteData(c, stdhttp.StatusOK, value)
}
func (h *Handler) Put(c *gin.Context) {
	var input struct {
		Description string `json:"description"`
		IsSecret    *bool  `json:"isSecret"`
		Value       string `json:"value"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		httpkit.WriteError(c, stdhttp.StatusBadRequest, httpkit.Error{Code: "INVALID_REQUEST"})
		return
	}
	actor := "system"
	if current, ok := authhttp.Current(c); ok {
		actor = current.Subject
	}
	secret := true
	if input.IsSecret != nil {
		secret = *input.IsSecret
	}
	value, err := h.service.Upsert(c.Request.Context(), c.Param("key"), input.Description, secret, input.Value, actor)
	if err != nil {
		httpkit.WriteError(c, stdhttp.StatusBadRequest, httpkit.Error{Code: "INVALID_CONFIGURATION"})
		return
	}
	if h.audit != nil {
		h.audit(actor, "updated", value.Key)
	}
	httpkit.WriteData(c, stdhttp.StatusOK, value)
}
func (h *Handler) Delete(c *gin.Context) {
	actor := "system"
	if current, ok := authhttp.Current(c); ok {
		actor = current.Subject
	}
	if err := h.service.Delete(c.Request.Context(), c.Param("key")); err != nil {
		httpkit.WriteError(c, stdhttp.StatusBadRequest, httpkit.Error{Code: "INVALID_CONFIGURATION"})
		return
	}
	if h.audit != nil {
		h.audit(actor, "deleted", c.Param("key"))
	}
	c.Status(stdhttp.StatusNoContent)
}
