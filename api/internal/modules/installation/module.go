package installation

import (
	"context"
	"database/sql"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	configurationapplication "sechelper-auth-template/api/internal/modules/configuration/application"
	configurationpersistence "sechelper-auth-template/api/internal/modules/configuration/persistence"
	"sechelper-auth-template/api/internal/platform/httpkit"
	"strings"
	"time"
)

type Module struct {
	service    *configurationapplication.Service
	installKey string
	protector  configurationapplication.Protector
}

func New(service *configurationapplication.Service, installKey string, protector configurationapplication.Protector) *Module {
	return &Module{service: service, installKey: strings.TrimSpace(installKey), protector: protector}
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
	databaseURL := ""
	for _, entry := range input.Entries {
		if entry.Key == "DATABASE_URL" {
			databaseURL = strings.TrimSpace(entry.Value)
		}
	}
	if databaseURL == "" {
		httpkit.WriteError(c, http.StatusBadRequest, httpkit.Error{Code: "DATABASE_URL_REQUIRED"})
		return
	}
	parsed, err := url.Parse(databaseURL)
	if err != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") || parsed.Host == "" || parsed.User == nil {
		httpkit.WriteError(c, http.StatusBadRequest, httpkit.Error{Code: "INVALID_DATABASE_URL"})
		return
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		httpkit.WriteError(c, http.StatusBadRequest, httpkit.Error{Code: "DATABASE_CONNECT_FAILED"})
		return
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		httpkit.WriteError(c, http.StatusBadRequest, httpkit.Error{Code: "DATABASE_CONNECT_FAILED"})
		return
	}
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS configuration_entries (key TEXT PRIMARY KEY, description TEXT NOT NULL DEFAULT '', is_secret BOOLEAN NOT NULL DEFAULT TRUE, ciphertext TEXT NOT NULL, version BIGINT NOT NULL DEFAULT 1, updated_by TEXT NOT NULL DEFAULT '', updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP)`); err != nil {
		httpkit.WriteError(c, http.StatusBadRequest, httpkit.Error{Code: "CONFIGURATION_TABLE_FAILED"})
		return
	}
	targetService := configurationapplication.NewService(configurationpersistence.NewRepository(db), m.protector)
	if err := targetService.Install(ctx, input.Entries, "initial-install"); err != nil {
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
