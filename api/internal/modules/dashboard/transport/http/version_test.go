package http

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"sechelper-auth-template/api/internal/modules/dashboard/application"
)

func TestVersionReturnsRuntimeReleaseDescriptor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := application.NewService(nil, application.AppInfo{
		Version: "1.4.0", Environment: "test", BuildID: "20260915150445", SourceRevision: "0123456789abcdef0123456789abcdef01234567",
	}, nil, nil)
	engine := gin.New()
	engine.GET("/v1/version", NewHandler(service).Version)
	request := httptest.NewRequest("GET", "/v1/version", nil)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)

	if response.Code != 200 {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", response.Header().Get("Cache-Control"))
	}
	var body struct {
		Data struct {
			ReleaseVersion        string `json:"releaseVersion"`
			DeploymentEnvironment string `json:"deploymentEnvironment"`
			BuildID               string `json:"buildId"`
			SourceRevision        string `json:"sourceRevision"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.ReleaseVersion != "1.4.0" || body.Data.DeploymentEnvironment != "test" || body.Data.BuildID != "20260915150445" || body.Data.SourceRevision != "0123456789abcdef0123456789abcdef01234567" {
		t.Fatalf("unexpected release descriptor: %+v", body.Data)
	}
}
