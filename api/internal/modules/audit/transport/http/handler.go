package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sechelper-auth-template/api/internal/modules/audit/application"
	"sechelper-auth-template/api/internal/modules/audit/domain"
	"sechelper-auth-template/api/internal/platform/httpkit"
	"strconv"
	"time"
)

type Handler struct{ service *application.Service }

func NewHandler(service *application.Service) *Handler { return &Handler{service: service} }
func (h *Handler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("page[size]", "20"))
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	filter := application.ListFilter{EventType: c.Query("filter[eventType]"), Category: c.Query("filter[category]"), Severity: c.Query("filter[severity]"), Outcome: c.Query("filter[outcome]"), ActorSubject: c.Query("filter[actorSubject]"), ResourceType: c.Query("filter[resourceType]"), ResourceID: c.Query("filter[resourceId]"), Limit: limit}
	if raw := c.Query("filter[from]"); raw != "" {
		filter.From, _ = time.Parse(time.RFC3339, raw)
	}
	if raw := c.Query("filter[to]"); raw != "" {
		filter.To, _ = time.Parse(time.RFC3339, raw)
	}
	values, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		writeError(c, http.StatusBadGateway, "AUDIT_QUERY_FAILED")
		return
	}
	data := make([]domain.Event, 0, len(values))
	for _, value := range values {
		value.IPHash = ""
		value.UserAgent = ""
		data = append(data, value)
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "meta": gin.H{"hasMore": len(data) == limit}})
}
func writeError(c *gin.Context, status int, code string) {
	httpkit.WriteError(c, status, httpkit.Error{Code: code})
}
