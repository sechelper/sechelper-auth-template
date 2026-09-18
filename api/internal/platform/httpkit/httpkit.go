package httpkit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

const requestIDKey = "httpkit.requestId"

type requestIDContextKey struct{}

type Error struct {
	Code      string   `json:"code"`
	Message   string   `json:"message"`
	Details   []Detail `json:"details,omitempty"`
	RequestID string   `json:"requestId"`
}

type Detail struct {
	Field  string `json:"field,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type errorEnvelope struct {
	Error Error `json:"error"`
}

// RequestIDMiddleware creates a server-owned correlation identifier for each
// request. A caller-supplied X-Request-ID is not trusted as the canonical ID.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var raw [16]byte
		id := "req-unavailable"
		if _, err := rand.Read(raw[:]); err == nil {
			id = "req-" + hex.EncodeToString(raw[:])
		}
		c.Set(requestIDKey, id)
		c.Request = c.Request.WithContext(WithRequestID(c.Request.Context(), id))
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// WithRequestID stores the server-owned request identifier in a standard
// context so service-layer code can attach it to events without depending on
// Gin.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDContextKey{}, id)
}

// RequestIDFromContext returns the request identifier attached by the HTTP
// middleware, or an empty string for non-HTTP/background work.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDContextKey{}).(string); ok {
		return id
	}
	return ""
}

func RequestID(c *gin.Context) string {
	if value, ok := c.Get(requestIDKey); ok {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}

func WriteData(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

func WriteCollection(c *gin.Context, data any, meta any) {
	c.JSON(http.StatusOK, gin.H{"data": data, "meta": meta})
}

func WriteError(c *gin.Context, status int, err Error) {
	if err.Code == "" {
		err.Code = "INTERNAL_ERROR"
	}
	if err.Message == "" {
		err.Message = err.Code
	}
	err.RequestID = RequestID(c)
	c.AbortWithStatusJSON(status, errorEnvelope{Error: err})
}
