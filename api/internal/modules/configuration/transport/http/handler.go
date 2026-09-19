package http

import (
	"context"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	stdhttp "net/http"
	"os"
	authhttp "sechelper-auth-template/api/internal/modules/authorization/transport/http"
	"sechelper-auth-template/api/internal/modules/configuration/application"
	"sechelper-auth-template/api/internal/platform/httpkit"
	"syscall"
	"time"
)

func restartProcess() {
	timer := time.NewTimer(250 * time.Millisecond)
	defer timer.Stop()
	<-timer.C
	_ = syscall.Kill(os.Getpid(), syscall.SIGTERM)
}

type Handler struct {
	service *application.Service
	audit   func(context.Context, string, string, string)
	restart func()
}

func NewHandler(service *application.Service, audit func(context.Context, string, string, string), restart ...func()) *Handler {
	restartProcess := restartProcess
	if len(restart) > 0 && restart[0] != nil {
		restartProcess = restart[0]
	}
	return &Handler{service: service, audit: audit, restart: restartProcess}
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
	h.put(c, true)
}

func (h *Handler) Restart(c *gin.Context) {
	c.JSON(stdhttp.StatusOK, gin.H{"data": gin.H{"restarting": true}, "meta": gin.H{"restartRequired": true}})
	go h.restart()
}
func (h *Handler) PutBusiness(c *gin.Context) {
	h.put(c, false)
}
func (h *Handler) put(c *gin.Context, allowFramework bool) {
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
	upsert := h.service.UpsertBusiness
	if allowFramework {
		upsert = h.service.Upsert
	}
	value, err := upsert(c.Request.Context(), c.Param("key"), input.Description, secret, input.Value, actor)
	if err != nil {
		if errors.Is(err, application.ErrFrameworkVariable) {
			httpkit.WriteError(c, stdhttp.StatusForbidden, httpkit.Error{Code: "FRAMEWORK_CONFIGURATION_READ_ONLY"})
			return
		}
		httpkit.WriteError(c, stdhttp.StatusBadRequest, httpkit.Error{Code: "INVALID_CONFIGURATION"})
		return
	}
	if h.audit != nil {
		h.audit(c.Request.Context(), actor, "updated", value.Key)
	}
	c.JSON(stdhttp.StatusOK, gin.H{"data": value, "meta": gin.H{"restartRequired": true}})
}
func (h *Handler) Delete(c *gin.Context) {
	h.delete(c, true)
}
func (h *Handler) DeleteBusiness(c *gin.Context) {
	h.delete(c, false)
}
func (h *Handler) delete(c *gin.Context, allowFramework bool) {
	actor := "system"
	if current, ok := authhttp.Current(c); ok {
		actor = current.Subject
	}
	deleteConfiguration := h.service.DeleteBusiness
	if allowFramework {
		deleteConfiguration = h.service.Delete
	}
	if err := deleteConfiguration(c.Request.Context(), c.Param("key")); err != nil {
		if errors.Is(err, application.ErrFrameworkVariable) {
			httpkit.WriteError(c, stdhttp.StatusForbidden, httpkit.Error{Code: "FRAMEWORK_CONFIGURATION_READ_ONLY"})
			return
		}
		httpkit.WriteError(c, stdhttp.StatusBadRequest, httpkit.Error{Code: "INVALID_CONFIGURATION"})
		return
	}
	if h.audit != nil {
		h.audit(c.Request.Context(), actor, "deleted", c.Param("key"))
	}
	c.Status(stdhttp.StatusNoContent)
}
