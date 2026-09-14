package security

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"sechelper-auth-template/api/internal/platform/config"
)

func testRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(HostOriginPolicy(config.Config{App: config.AppConfig{APIOrigin: "https://api.example.com", PublicWebOrigin: "https://www.example.com"}, CORS: config.CORSConfig{AllowedOrigins: []string{"https://www.example.com"}}}))
	r.GET("/", func(c *gin.Context) { c.Status(204) })
	return r
}

func TestHostOriginPolicyRejectsUnknownHost(t *testing.T) {
	r := testRouter()
	req := httptest.NewRequest("GET", "https://evil.example/", nil)
	req.Host = "evil.example"
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestHostOriginPolicyAcceptsTrustedForwardedHost(t *testing.T) {
	cfg := config.Config{App: config.AppConfig{APIOrigin: "https://api.example.com", PublicWebOrigin: "https://www.example.com"}, CORS: config.CORSConfig{AllowedOrigins: []string{"https://www.example.com"}}, Security: config.SecurityConfig{TrustedProxyCIDRs: []string{"127.0.0.1/32"}}}
	r := gin.New()
	r.Use(HostOriginPolicy(cfg))
	r.GET("/", func(c *gin.Context) { c.Status(204) })
	req := httptest.NewRequest("GET", "http://127.0.0.1/", nil)
	req.Host = "127.0.0.1"
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("X-Forwarded-Host", "api.example.com")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != 204 {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
}

func TestMetricsAuthRequiresBearerToken(t *testing.T) {
	r := gin.New()
	r.Use(MetricsAuth("secret"))
	r.GET("/metrics", func(c *gin.Context) { c.Status(200) })
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	req = httptest.NewRequest("GET", "/metrics", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("authorized status = %d, want 200", rec.Code)
	}
}
