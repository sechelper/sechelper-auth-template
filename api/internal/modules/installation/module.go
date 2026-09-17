package installation

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	configurationapplication "sechelper-auth-template/api/internal/modules/configuration/application"
	"sechelper-auth-template/api/internal/platform/httpkit"
	"strings"
	"time"
)

type Module struct {
	service    *configurationapplication.Service
	installKey string
	lockFile   string
}

func New(service *configurationapplication.Service, installKey, lockFile string, protector configurationapplication.Protector) *Module {
	lockFile = strings.TrimSpace(lockFile)
	if lockFile == "" {
		lockFile = "logs/install.lock"
	}
	return &Module{service: service, installKey: strings.TrimSpace(installKey), lockFile: lockFile}
}
func (m *Module) RegisterRoutes(router *gin.Engine) {
	router.GET("/internal/install-guard", m.guard)
	router.PUT("/v1/install/configuration", m.install)
}
func (m *Module) guard(c *gin.Context) {
	installed, err := m.service.IsInstalled(c.Request.Context())
	if err != nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}
	if installed || m.installKey == "" {
		c.Status(http.StatusForbidden)
		return
	}
	if m.lockExists() {
		c.Status(http.StatusForbidden)
		return
	}
	c.Status(http.StatusNoContent)
}
func (m *Module) install(c *gin.Context) {
	if m.lockExists() {
		httpkit.WriteError(c, http.StatusInternalServerError, httpkit.Error{Code: "INSTALL_LOCKED"})
		return
	}
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
	if missing := configurationapplication.MissingInstallKeys(input.Entries); len(missing) > 0 {
		details := make([]httpkit.Detail, 0, len(missing))
		for _, key := range missing {
			details = append(details, httpkit.Detail{Field: key, Reason: "required"})
		}
		httpkit.WriteError(c, http.StatusBadRequest, httpkit.Error{Code: "REQUIRED_INSTALL_FIELDS_MISSING", Message: "关键安装配置不能为空", Details: details})
		return
	}
	if err := m.createLock(); err != nil {
		if errors.Is(err, errInstallLocked) {
			httpkit.WriteError(c, http.StatusInternalServerError, httpkit.Error{Code: "INSTALL_LOCKED"})
		} else {
			httpkit.WriteError(c, http.StatusInternalServerError, httpkit.Error{Code: "INSTALL_LOCK_WRITE_FAILED"})
		}
		return
	}
	if err := m.service.Install(c.Request.Context(), input.Entries, "initial-install"); err != nil {
		_ = os.Remove(m.lockFile)
		httpkit.WriteError(c, http.StatusConflict, httpkit.Error{Code: "INSTALL_FAILED"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"installed": true, "saved": len(input.Entries), "restartRequired": true}})
}

var errInstallLocked = errors.New("installation lock already exists")

func (m *Module) lockExists() bool {
	_, err := os.Stat(m.lockFile)
	return err == nil
}

func (m *Module) createLock() error {
	file, err := os.OpenFile(m.lockFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return errInstallLocked
		}
		return err
	}
	defer file.Close()
	_, err = file.WriteString(time.Now().UTC().Format(time.RFC3339) + "\n")
	return err
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
