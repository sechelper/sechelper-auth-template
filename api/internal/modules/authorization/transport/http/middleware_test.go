package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"sechelper-auth-template/api/internal/modules/authorization/application"
	"sechelper-auth-template/api/internal/modules/authorization/domain"
	"sechelper-auth-template/api/internal/platform/session"
)

func TestMiddlewareMakesPlatformUserUUIDAvailableToBusinessHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := session.NewMemoryStore()
	const platformUserUUID = "89cf8f29-9954-470d-b3de-8a37d20c6f44"
	if err := store.Create(t.Context(), session.Session{ID: "session-1", Subject: "identity-subject", PlatformUserUUID: platformUserUUID, ApplicationCode: "app-1", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.GET("/business", NewMiddleware(application.NewService(store, application.NewMemoryCache()), "session").RequireAuthentication(), func(c *gin.Context) {
		value, ok := domain.FromContext(c.Request.Context())
		if !ok || value.PlatformUserUUID != platformUserUUID {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})
	req := httptest.NewRequest(http.MethodGet, "/business", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "session-1"})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	if response.Code != http.StatusNoContent {
		t.Fatalf("business handler status = %d, want %d", response.Code, http.StatusNoContent)
	}
}
