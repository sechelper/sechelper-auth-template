package account

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSessionAdministrationRoutesAreNotRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	module := New(nil, "session")
	module.RegisterRoutes(router.Group("/v1"), func(c *gin.Context) { c.Next() })

	if got := len(router.Routes()); got != 2 {
		t.Fatalf("registered route count = %d, want current user and admin account routes", got)
	}

	for _, testCase := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/v1/admin/account/sessions"},
		{method: http.MethodDelete, path: "/v1/admin/account/sessions/session-1"},
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(testCase.method, testCase.path, nil)
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusNotFound {
			t.Errorf("%s %s status = %d, want %d", testCase.method, testCase.path, recorder.Code, http.StatusNotFound)
		}
	}
}
