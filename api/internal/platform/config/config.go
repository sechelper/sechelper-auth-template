package config

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Identity  IdentityConfig  `mapstructure:"identity"`
	Manifest  ManifestConfig  `mapstructure:"manifest"`
	Session   SessionConfig   `mapstructure:"session"`
	RateLimit RateLimitConfig `mapstructure:"rateLimit"`
	CORS      CORSConfig      `mapstructure:"cors"`
	Security  SecurityConfig  `mapstructure:"security"`
	Log       LogConfig       `mapstructure:"log"`
	Bootstrap BootstrapConfig `mapstructure:"bootstrap"`
}
type AppConfig struct{ Name, Environment, ListenAddr, PublicWebOrigin, APIOrigin, PrimaryDomain, TestDomainSuffix, RedisURL string }
type ServerConfig struct {
	Host, Address                                           string
	Port                                                    int
	ReadTimeout, WriteTimeout, IdleTimeout, ShutdownTimeout time.Duration
}
type DatabaseConfig struct {
	Driver, Host, Name, User, Password, SSLMode string
	Port                                        int
}

func (c DatabaseConfig) URL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", url.PathEscape(c.User), url.PathEscape(c.Password), c.Host, c.Port, c.Name, c.SSLMode)
}

type IdentityConfig struct {
	Issuer, AuthorizationEndpoint, TokenEndpoint, UserinfoEndpoint, JWKSURL, RevocationEndpoint, EndSessionEndpoint, Audience, ClientID, ClientSecret, RedirectURI, ApplicationCode string
	Scopes                                                                                                                                                                          []string
}
type ManifestConfig struct {
	SyncInterval time.Duration
}
type SessionConfig struct {
	CookieName              string
	TTL                     time.Duration
	Secure                  bool
	SameSite, EncryptionKey string
}
type CORSConfig struct{ AllowedOrigins []string }
type SecurityConfig struct {
	TrustedProxyCIDRs []string `mapstructure:"trustedProxyCidrs"`
	MetricsToken      string   `mapstructure:"metricsToken"`
}
type RateLimitConfig struct {
	LoginPerMinute    int `mapstructure:"loginPerMinute"`
	CallbackPerMinute int `mapstructure:"callbackPerMinute"`
	RefreshPerMinute  int `mapstructure:"refreshPerMinute"`
}
type LogConfig struct {
	Mode, Encoding, Level, Output                                            string
	TimeKey, LevelKey, MessageKey, CallerKey, StacktraceKey                  string
	TimeEncoding, LevelEncoding                                              string
	RetentionDays, MaxSizeMB, MaxBackups, MaxAgeDays                         int
	Development, DisableCaller, DisableStacktrace, Sampling, Compress, Debug bool
}
type BootstrapConfig struct {
	DatabaseURL   string `mapstructure:"databaseUrl"`
	EncryptionKey string `mapstructure:"encryptionKey"`
}

func Load(args ...string) (Config, error) {
	path, explicit, err := resolvePath(args)
	if err != nil {
		return Config{}, err
	}
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("read application config %q: %w", path, err)
	}
	setDefaults(v)
	for key, env := range envBindings() {
		if value, ok := os.LookupEnv(env); ok {
			v.Set(key, value)
		}
	}
	var c Config
	if err := v.UnmarshalExact(&c); err != nil {
		return Config{}, fmt.Errorf("decode application config: %w", err)
	}
	if c.App.Environment == "production" && !explicit {
		return Config{}, errors.New("production requires an explicit --config path or APP_CONFIG_FILE")
	}
	if c.Server.Address == "" {
		c.Server.Address = fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
	}
	if err := Validate(c); err != nil {
		return Config{}, err
	}
	return c, nil
}

// ApplyConfigurationValues overlays the framework-supported environment
// variables supplied by the configuration center. Deployment metadata and
// bootstrap-only values are deliberately excluded from this overlay.
func ApplyConfigurationValues(base Config, values map[string]string) (Config, error) {
	var raw map[string]any
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName: "mapstructure", Result: &raw})
	if err != nil {
		return Config{}, fmt.Errorf("create configuration decoder: %w", err)
	}
	if err := decoder.Decode(base); err != nil {
		return Config{}, fmt.Errorf("encode application config: %w", err)
	}
	v := viper.New()
	setDefaults(v)
	v.MergeConfigMap(raw)
	if value := strings.TrimSpace(values["DATABASE_URL"]); value != "" {
		if err := validateBootstrapDatabaseURL(value); err != nil {
			return Config{}, err
		}
		v.Set("bootstrap.databaseUrl", value)
	}
	bindings := envBindings()
	for key, env := range bindings {
		if env == "APP_ENV" || env == "DATABASE_URL" || env == "CONFIG_CENTER_DATABASE_URL" || env == "CONFIG_CENTER_ENCRYPTION_KEY" {
			continue
		}
		if value, ok := values[env]; ok {
			v.Set(key, value)
		}
	}
	var result Config
	if err := v.UnmarshalExact(&result); err != nil {
		return Config{}, fmt.Errorf("decode configuration center values: %w", err)
	}
	if result.Server.Address == "" {
		result.Server.Address = fmt.Sprintf("%s:%d", result.Server.Host, result.Server.Port)
	}
	if err := Validate(result); err != nil {
		return Config{}, fmt.Errorf("validate configuration center values: %w", err)
	}
	return result, nil
}

func validateBootstrapDatabaseURL(value string) error {
	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Host == "" || u.User == nil {
		return errors.New("DATABASE_URL must be a PostgreSQL URL with host and credentials")
	}
	return nil
}
func resolvePath(args []string) (string, bool, error) {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	requested := fs.String("config", "", "path to the Viper application YAML file")
	if err := fs.Parse(args); err != nil {
		return "", false, err
	}
	if strings.TrimSpace(*requested) != "" {
		return *requested, true, nil
	}
	if p := strings.TrimSpace(os.Getenv("APP_CONFIG_FILE")); p != "" {
		return p, true, nil
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production") {
		return "", false, errors.New("APP_CONFIG_FILE or --config is required in production")
	}
	for _, p := range []string{"config.yaml", "config.yml"} {
		if _, err := os.Stat(p); err == nil {
			return p, false, nil
		}
	}
	return "", false, errors.New("application config is required: pass --config or set APP_CONFIG_FILE")
}
func envBindings() map[string]string {
	return map[string]string{
		"bootstrap.databaseUrl": "DATABASE_URL", "bootstrap.encryptionKey": "CONFIG_CENTER_ENCRYPTION_KEY",
		"app.environment": "APP_ENV", "app.listenAddr": "LISTEN_ADDR", "app.publicWebOrigin": "PUBLIC_WEB_ORIGIN", "app.apiOrigin": "API_ORIGIN", "app.primaryDomain": "PRIMARY_DOMAIN", "app.testDomainSuffix": "TEST_DOMAIN_SUFFIX", "app.redisUrl": "REDIS_URL",
		"server.host": "SERVER_HOST", "server.port": "SERVER_PORT", "server.address": "LISTEN_ADDR", "server.readTimeout": "SERVER_READ_TIMEOUT", "server.writeTimeout": "SERVER_WRITE_TIMEOUT", "server.idleTimeout": "SERVER_IDLE_TIMEOUT", "server.shutdownTimeout": "SERVER_SHUTDOWN_TIMEOUT",
		"database.driver": "DATABASE_DRIVER", "database.host": "DATABASE_HOST", "database.port": "DATABASE_PORT", "database.name": "DATABASE_NAME", "database.user": "DATABASE_USER", "database.password": "DATABASE_PASSWORD", "database.sslMode": "DATABASE_SSL_MODE",
		"identity.issuer": "IDENTITY_ISSUER", "identity.authorizationEndpoint": "IDENTITY_AUTHORIZATION_ENDPOINT", "identity.tokenEndpoint": "IDENTITY_TOKEN_ENDPOINT", "identity.userinfoEndpoint": "IDENTITY_USERINFO_ENDPOINT", "identity.jwksUrl": "IDENTITY_JWKS_URL", "identity.revocationEndpoint": "IDENTITY_REVOCATION_ENDPOINT", "identity.endSessionEndpoint": "IDENTITY_END_SESSION_ENDPOINT", "identity.audience": "IDENTITY_AUDIENCE", "identity.clientId": "IDENTITY_CLIENT_ID", "identity.clientSecret": "IDENTITY_CLIENT_SECRET", "identity.redirectUri": "IDENTITY_REDIRECT_URI", "identity.applicationCode": "IDENTITY_APPLICATION_CODE", "identity.scopes": "IDENTITY_SCOPES",
		"manifest.syncInterval": "MANIFEST_SYNC_INTERVAL", "session.cookieName": "SESSION_COOKIE_NAME", "session.ttl": "SESSION_TTL", "session.secure": "SESSION_SECURE", "session.sameSite": "SESSION_SAME_SITE", "session.encryptionKey": "SESSION_ENCRYPTION_KEY", "cors.allowedOrigins": "ALLOWED_ORIGINS", "security.trustedProxyCidrs": "TRUSTED_PROXY_CIDRS", "security.metricsToken": "METRICS_TOKEN", "log.mode": "LOG_MODE", "log.encoding": "LOG_ENCODING", "log.level": "LOG_LEVEL", "log.output": "LOG_OUTPUT", "log.timeKey": "LOG_TIME_KEY", "log.levelKey": "LOG_LEVEL_KEY", "log.messageKey": "LOG_MESSAGE_KEY", "log.callerKey": "LOG_CALLER_KEY", "log.stacktraceKey": "LOG_STACKTRACE_KEY", "log.timeEncoding": "LOG_TIME_ENCODING", "log.levelEncoding": "LOG_LEVEL_ENCODING", "log.development": "LOG_DEVELOPMENT", "log.disableCaller": "LOG_DISABLE_CALLER", "log.disableStacktrace": "LOG_DISABLE_STACKTRACE", "log.sampling": "LOG_SAMPLING", "log.retentionDays": "LOG_RETENTION_DAYS", "log.maxSizeMB": "LOG_MAX_SIZE_MB", "log.maxBackups": "LOG_MAX_BACKUPS", "log.maxAgeDays": "LOG_MAX_AGE_DAYS", "log.compress": "LOG_COMPRESS", "log.debug": "DEBUG",
		"rateLimit.loginPerMinute": "RATE_LIMIT_LOGIN_PER_MINUTE", "rateLimit.callbackPerMinute": "RATE_LIMIT_CALLBACK_PER_MINUTE", "rateLimit.refreshPerMinute": "RATE_LIMIT_REFRESH_PER_MINUTE",
	}
}
func setDefaults(v *viper.Viper) {
	for key, value := range map[string]any{"app.name": "auth-template", "app.environment": "development", "app.testDomainSuffix": "-test", "server.host": "127.0.0.1", "server.port": 8080, "server.readTimeout": "10s", "server.writeTimeout": "15s", "server.idleTimeout": "60s", "server.shutdownTimeout": "15s", "database.driver": "postgres", "database.port": 5432, "database.sslMode": "disable", "identity.scopes": []string{"openid", "profile", "email"}, "manifest.syncInterval": "10m", "session.cookieName": "auth_template_session", "session.ttl": "8h", "session.secure": false, "session.sameSite": "Lax", "log.mode": "file", "log.encoding": "json", "log.level": "info", "log.output": "logs/app.log", "log.timeKey": "ts", "log.levelKey": "level", "log.messageKey": "msg", "log.callerKey": "caller", "log.stacktraceKey": "stacktrace", "log.timeEncoding": "iso8601", "log.levelEncoding": "lowercase", "log.development": false, "log.disableCaller": false, "log.disableStacktrace": true, "log.sampling": false, "log.retentionDays": 180, "log.maxSizeMB": 100, "log.maxBackups": 10, "log.maxAgeDays": 180, "log.compress": true, "log.debug": false} {
		v.SetDefault(key, value)
	}
	v.SetDefault("rateLimit.loginPerMinute", 10)
	v.SetDefault("rateLimit.callbackPerMinute", 20)
	v.SetDefault("rateLimit.refreshPerMinute", 30)
}
func Validate(c Config) error {
	if c.App.Environment != "development" && c.App.Environment != "test" && c.App.Environment != "production" {
		return fmt.Errorf("unsupported app.environment: %s", c.App.Environment)
	}
	for name, value := range map[string]string{"app.name": c.App.Name, "app.publicWebOrigin": c.App.PublicWebOrigin, "app.apiOrigin": c.App.APIOrigin, "app.primaryDomain": c.App.PrimaryDomain, "database.host": c.Database.Host, "database.name": c.Database.Name, "database.user": c.Database.User, "database.password": c.Database.Password, "identity.issuer": c.Identity.Issuer, "identity.authorizationEndpoint": c.Identity.AuthorizationEndpoint, "identity.tokenEndpoint": c.Identity.TokenEndpoint, "identity.userinfoEndpoint": c.Identity.UserinfoEndpoint, "identity.jwksUrl": c.Identity.JWKSURL, "identity.audience": c.Identity.Audience, "identity.clientId": c.Identity.ClientID, "identity.clientSecret": c.Identity.ClientSecret, "identity.redirectUri": c.Identity.RedirectURI, "identity.applicationCode": c.Identity.ApplicationCode, "session.encryptionKey": c.Session.EncryptionKey} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if c.Server.Port < 1 || c.Server.Port > 65535 || c.Database.Port < 1 || c.Database.Port > 65535 {
		return errors.New("server.port and database.port must be valid TCP ports")
	}
	for name, raw := range map[string]string{"app.publicWebOrigin": c.App.PublicWebOrigin, "app.apiOrigin": c.App.APIOrigin, "identity.issuer": c.Identity.Issuer} {
		u, err := url.Parse(raw)
		if err != nil || u.Scheme == "" || u.Host == "" || (u.Path != "" && u.Path != "/") || u.User != nil {
			return fmt.Errorf("invalid %s", name)
		}
	}
	for name, raw := range map[string]string{"identity.authorizationEndpoint": c.Identity.AuthorizationEndpoint, "identity.tokenEndpoint": c.Identity.TokenEndpoint, "identity.userinfoEndpoint": c.Identity.UserinfoEndpoint} {
		u, err := url.Parse(raw)
		if err != nil || u.Scheme == "" || u.Host == "" || u.User != nil {
			return fmt.Errorf("invalid %s", name)
		}
	}
	for name, raw := range map[string]string{"identity.revocationEndpoint": c.Identity.RevocationEndpoint, "identity.endSessionEndpoint": c.Identity.EndSessionEndpoint} {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		u, err := url.Parse(raw)
		if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
			return fmt.Errorf("invalid %s", name)
		}
	}
	if u, err := url.Parse(c.Identity.JWKSURL); err != nil || (u.Scheme != "https" && !(c.App.Environment != "production" && u.Scheme == "http")) || u.Host == "" || u.User != nil {
		return errors.New("identity.jwksUrl must be an HTTPS URL without credentials in production; HTTP is allowed only for development/test")
	}
	redirect, err := url.Parse(c.Identity.RedirectURI)
	if err != nil || redirect.Scheme == "" || redirect.Host == "" || redirect.Path != "/v1/auth/callback" {
		return errors.New("invalid identity.redirectUri")
	}
	if c.Identity.ClientID == c.Identity.ClientSecret {
		return errors.New("identity.clientId and identity.clientSecret must differ")
	}
	decoded, err := base64.RawStdEncoding.DecodeString(c.Session.EncryptionKey)
	if err != nil || len(decoded) != 32 {
		return errors.New("session.encryptionKey must be base64 RawStdEncoding of 32 bytes")
	}
	if len(c.CORS.AllowedOrigins) == 0 {
		return errors.New("cors.allowedOrigins is required")
	}
	if c.RateLimit.LoginPerMinute <= 0 || c.RateLimit.CallbackPerMinute <= 0 || c.RateLimit.RefreshPerMinute <= 0 {
		return errors.New("rateLimit values must be positive")
	}
	if c.App.Environment == "production" && !c.Session.Secure {
		return errors.New("session.secure must be true in production")
	}
	if c.App.Environment == "production" && strings.TrimSpace(c.App.RedisURL) == "" {
		return errors.New("app.redisUrl is required in production")
	}
	if c.App.Environment == "production" && strings.TrimSpace(c.Security.MetricsToken) == "" {
		return errors.New("security.metricsToken is required in production")
	}
	if c.Log.RetentionDays <= 0 || c.Log.MaxSizeMB <= 0 || c.Log.MaxBackups <= 0 || c.Log.MaxAgeDays <= 0 {
		return errors.New("log rotation limits must be positive")
	}
	if c.Log.Mode != "file" && c.Log.Mode != "local_file" && c.Log.Mode != "stdout" {
		return fmt.Errorf("unsupported log.mode: %s", c.Log.Mode)
	}
	if c.Log.Encoding != "json" || c.Log.TimeEncoding != "iso8601" || c.Log.LevelEncoding != "lowercase" {
		return errors.New("log.encoding must be json, log.timeEncoding must be iso8601, and log.levelEncoding must be lowercase")
	}
	if c.Log.TimeKey == "" || c.Log.LevelKey == "" || c.Log.MessageKey == "" || c.Log.CallerKey == "" || c.Log.StacktraceKey == "" {
		return errors.New("log field keys must not be empty")
	}
	return nil
}
