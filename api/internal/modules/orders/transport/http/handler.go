//go:build example

package http

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"net/http"
	"sechelper-auth-template/api/internal/modules/orders/application"
	"sechelper-auth-template/api/internal/modules/orders/domain"
	"strconv"
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
	values, err := h.service.List(c.Request.Context(), limit)
	if err != nil {
		writeError(c, http.StatusBadGateway, "ORDER_QUERY_FAILED")
		return
	}
	data := make([]orderResponse, 0, len(values))
	for _, value := range values {
		data = append(data, mapOrder(value))
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "meta": gin.H{"hasMore": len(data) == limit}})
}
func (h *Handler) Get(c *gin.Context) {
	value, err := h.service.Get(c.Request.Context(), c.Param("orderId"))
	if err == application.ErrNotFound || err == sql.ErrNoRows {
		writeError(c, http.StatusNotFound, "ORDER_NOT_FOUND")
		return
	}
	if err != nil {
		writeError(c, http.StatusBadGateway, "ORDER_QUERY_FAILED")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": mapOrder(value)})
}

type orderResponse struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	TotalMinor int64  `json:"totalMinor"`
	Currency   string `json:"currency"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

func mapOrder(value domain.Order) orderResponse {
	return orderResponse{ID: value.ID, Status: value.Status, TotalMinor: value.TotalMinor, Currency: value.Currency, CreatedAt: value.CreatedAt.UTC().Format(timeFormat), UpdatedAt: value.UpdatedAt.UTC().Format(timeFormat)}
}

const timeFormat = "2006-01-02T15:04:05Z07:00"

func writeError(c *gin.Context, status int, code string) {
	c.AbortWithStatusJSON(status, gin.H{"error": gin.H{"code": code, "message": code, "requestId": c.GetHeader("X-Request-ID")}})
}
