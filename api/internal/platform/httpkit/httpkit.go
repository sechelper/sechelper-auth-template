package httpkit

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Error struct {
	Code, Message string
	Details       []Detail
}
type Detail struct {
	Field, Reason string `json:"field,omitempty"`
}

func WriteData(c *gin.Context, status int, data any) { c.JSON(status, gin.H{"data": data}) }
func WriteCollection(c *gin.Context, data any, meta any) {
	c.JSON(http.StatusOK, gin.H{"data": data, "meta": meta})
}
func WriteError(c *gin.Context, status int, err Error) {
	requestID := c.GetHeader("X-Request-ID")
	c.AbortWithStatusJSON(status, gin.H{"error": gin.H{"code": err.Code, "message": err.Message, "details": err.Details, "requestId": requestID}})
}
