package installation

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGuardReturnsInternalServerErrorWhenInstallLockExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	lockFile := filepath.Join(t.TempDir(), "install.lock")
	if err := os.WriteFile(lockFile, []byte("locked\n"), 0o600); err != nil {
		t.Fatalf("write install lock: %v", err)
	}

	module := &Module{lockFile: lockFile}
	router := gin.New()
	router.GET("/internal/install-guard", module.guard)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/internal/install-guard", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("guard status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
}
