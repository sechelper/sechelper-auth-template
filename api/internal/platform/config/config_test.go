package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setRequiredEnvironment(t *testing.T) {
	t.Helper()
	values := map[string]string{
		"DATABASE_PASSWORD": "test-password", "IDENTITY_CLIENT_ID": "client", "IDENTITY_CLIENT_SECRET": "secret",
		"IDENTITY_APPLICATION_CODE": "app-test",
		"SESSION_ENCRYPTION_KEY":    "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
	}
	for key, value := range values {
		t.Setenv(key, value)
	}
}

func TestLoadUsesExplicitPathAndEnvironmentOverrides(t *testing.T) {
	setRequiredEnvironment(t)
	t.Setenv("SERVER_PORT", "18080")
	path := filepath.Join("..", "..", "..", "..", "config.example.yaml")
	c, err := Load("--config", path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if c.Server.Port != 18080 {
		t.Fatalf("server.port = %d, want environment override 18080", c.Server.Port)
	}
	if c.Database.Password != "test-password" {
		t.Fatal("database password was not loaded from the environment")
	}
	if c.RateLimit.LoginPerMinute != 10 || c.RateLimit.CallbackPerMinute != 20 || c.RateLimit.RefreshPerMinute != 30 {
		t.Fatalf("rate limits = %+v, want default limits", c.RateLimit)
	}
}

func TestLoadRejectsUnknownKeys(t *testing.T) {
	setRequiredEnvironment(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	contents := "app:\n  name: test\nunknownGroup:\n  value: true\n"
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Load("--config", path)
	if err == nil || !strings.Contains(err.Error(), "unknowngroup") {
		t.Fatalf("Load() error = %v, want unknown-key error", err)
	}
}

func TestProductionRequiresExplicitConfigPath(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("APP_CONFIG_FILE", "")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "required in production") {
		t.Fatalf("Load() error = %v, want explicit production path error", err)
	}
}
