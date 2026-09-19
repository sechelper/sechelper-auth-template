package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadUsesExplicitConfigFileOnly(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://app:test@127.0.0.1:5432/auth_template?sslmode=disable")
	t.Setenv("CONFIG_CENTER_ENCRYPTION_KEY", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	t.Setenv("CONFIG_CENTER_INSTALL_KEY", "test-install-key")
	path := filepath.Join("..", "..", "..", "..", "config.example.yaml")
	c, err := Load("--config", path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if c.App.ListenAddr != "127.0.0.1:8080" {
		t.Fatalf("app.listenAddr = %q, want config file value", c.App.ListenAddr)
	}
	if c.Bootstrap.DatabaseURL == "" {
		t.Fatal("bootstrap database URL was not loaded from the config file")
	}
	if c.RateLimit.LoginPerMinute != 10 || c.RateLimit.CallbackPerMinute != 20 || c.RateLimit.RefreshPerMinute != 30 {
		t.Fatalf("rate limits = %+v, want default limits", c.RateLimit)
	}
}

func TestLoadUsesEnvironmentForDatabaseAndRedisConnections(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://app:test@127.0.0.1:5432/from-env?sslmode=disable")
	t.Setenv("REDIS_URL", "redis://:secret@127.0.0.1:6379/0")
	t.Setenv("CONFIG_CENTER_ENCRYPTION_KEY", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	t.Setenv("CONFIG_CENTER_INSTALL_KEY", "install-key")
	c, err := Load("--config", filepath.Join("..", "..", "..", "..", "config.example.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if c.Bootstrap.DatabaseURL != "postgres://app:test@127.0.0.1:5432/from-env?sslmode=disable" || c.App.RedisURL != "redis://:secret@127.0.0.1:6379/0" || c.Bootstrap.EncryptionKey == "" || c.Bootstrap.InstallKey != "install-key" {
		t.Fatalf("secret connection overrides were not loaded")
	}
}

func TestLoadRejectsUnknownKeys(t *testing.T) {
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
	t.Setenv("APP_CONFIG_FILE", "")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "application config is required") {
		t.Fatalf("Load() error = %v, want missing config error", err)
	}
}

func TestConfigurationCenterCannotReplaceBootstrapDatabase(t *testing.T) {
	base := Config{
		App:       AppConfig{Name: "auth-template", Environment: "test", PublicWebOrigin: "https://example.test", APIOrigin: "https://example.test", PrimaryDomain: "example.test"},
		Identity:  IdentityConfig{Issuer: "https://identity.example", AuthorizationEndpoint: "https://identity.example/authorize", TokenEndpoint: "https://identity.example/token", UserinfoEndpoint: "https://identity.example/userinfo", JWKSURL: "https://identity.example/jwks", Audience: "api", ClientID: "client", ClientSecret: "secret", RedirectURI: "https://example.test/v1/auth/callback", ApplicationCode: "app"},
		Session:   SessionConfig{EncryptionKey: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
		CORS:      CORSConfig{AllowedOrigins: []string{"https://example.test"}},
		RateLimit: RateLimitConfig{LoginPerMinute: 10, CallbackPerMinute: 20, RefreshPerMinute: 30},
		Log:       LogConfig{Mode: "stdout", Encoding: "json", Level: "info", TimeKey: "ts", LevelKey: "level", MessageKey: "msg", CallerKey: "caller", StacktraceKey: "stacktrace", TimeEncoding: "iso8601", LevelEncoding: "lowercase", RetentionDays: 1, MaxSizeMB: 1, MaxBackups: 1, MaxAgeDays: 1},
		Bootstrap: BootstrapConfig{DatabaseURL: "postgres://app:config@127.0.0.1:5432/config"},
	}
	resolved, err := ApplyConfigurationValues(base, map[string]string{"DATABASE_URL": "postgres://app:other@127.0.0.1:5432/other", "APP_ENV": "test"})
	if err != nil {
		t.Fatalf("ApplyConfigurationValues() error = %v", err)
	}
	if resolved.Bootstrap.DatabaseURL != base.Bootstrap.DatabaseURL {
		t.Fatalf("bootstrap database URL changed to %q", resolved.Bootstrap.DatabaseURL)
	}
}
