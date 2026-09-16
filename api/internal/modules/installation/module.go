package installation

import (
	"github.com/gin-gonic/gin"
	"net/http"
	configurationapplication "sechelper-auth-template/api/internal/modules/configuration/application"
	"sechelper-auth-template/api/internal/platform/httpkit"
	"strings"
)

type Module struct {
	service    *configurationapplication.Service
	installKey string
}

func New(service *configurationapplication.Service, installKey string, protector configurationapplication.Protector) *Module {
	return &Module{service: service, installKey: strings.TrimSpace(installKey)}
}
func (m *Module) RegisterRoutes(router *gin.Engine) {
	router.GET("/v1/install/status", m.status)
	router.PUT("/v1/install/configuration", m.install)
}
func (m *Module) status(c *gin.Context) {
	installed, err := m.service.IsInstalled(c.Request.Context())
	if err != nil {
		httpkit.WriteError(c, http.StatusBadGateway, httpkit.Error{Code: "INSTALL_STATUS_FAILED"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"installed": installed, "available": !installed && m.installKey != ""}})
}
func (m *Module) install(c *gin.Context) {
	if m.installKey == "" || !secureCompare(m.installKey, c.GetHeader("X-Install-Key")) {
		httpkit.WriteError(c, http.StatusForbidden, httpkit.Error{Code: "INSTALL_KEY_REQUIRED"})
		return
	}
	var input struct {
		Entries []configurationapplication.InstallEntry `json:"entries"`
	}
	if err := c.ShouldBindJSON(&input); err != nil || len(input.Entries) == 0 {
		httpkit.WriteError(c, http.StatusBadRequest, httpkit.Error{Code: "INVALID_INSTALL_REQUEST"})
		return
	}
	if err := m.service.Install(c.Request.Context(), input.Entries, "initial-install"); err != nil {
		httpkit.WriteError(c, http.StatusConflict, httpkit.Error{Code: "INSTALL_FAILED"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"installed": true, "saved": len(input.Entries)}})
}
func secureCompare(expected, actual string) bool {
	if len(expected) != len(actual) {
		return false
	}
	var diff byte
	for i := range expected {
		diff |= expected[i] ^ actual[i]
	}
	return diff == 0
}
