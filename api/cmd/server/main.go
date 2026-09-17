package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"sechelper-auth-template/api/internal/modules/account"
	"sechelper-auth-template/api/internal/modules/audit"
	auditapplication "sechelper-auth-template/api/internal/modules/audit/application"
	auditpersistence "sechelper-auth-template/api/internal/modules/audit/persistence"
	"sechelper-auth-template/api/internal/modules/authentication/application"
	authpersistence "sechelper-auth-template/api/internal/modules/authentication/persistence"
	authhttp "sechelper-auth-template/api/internal/modules/authentication/transport/http"
	"sechelper-auth-template/api/internal/modules/authorization"
	authorizationapplication "sechelper-auth-template/api/internal/modules/authorization/application"
	authorizationdomain "sechelper-auth-template/api/internal/modules/authorization/domain"
	authorizationpersistence "sechelper-auth-template/api/internal/modules/authorization/persistence"
	"sechelper-auth-template/api/internal/modules/configuration"
	configurationapplication "sechelper-auth-template/api/internal/modules/configuration/application"
	configurationpersistence "sechelper-auth-template/api/internal/modules/configuration/persistence"
	"sechelper-auth-template/api/internal/modules/dashboard"
	dashboardapplication "sechelper-auth-template/api/internal/modules/dashboard/application"
	"sechelper-auth-template/api/internal/modules/installation"
	"sechelper-auth-template/api/internal/modules/manifest"
	manifestapplication "sechelper-auth-template/api/internal/modules/manifest/application"
	"sechelper-auth-template/api/internal/modules/manifest/domain"
	manifestpersistence "sechelper-auth-template/api/internal/modules/manifest/persistence"
	"sechelper-auth-template/api/internal/modules/operations"
	operationsapplication "sechelper-auth-template/api/internal/modules/operations/application"
	plataudit "sechelper-auth-template/api/internal/platform/audit"
	"sechelper-auth-template/api/internal/platform/config"
	"sechelper-auth-template/api/internal/platform/httpkit"
	"sechelper-auth-template/api/internal/platform/identity"
	"sechelper-auth-template/api/internal/platform/logging"
	appmetrics "sechelper-auth-template/api/internal/platform/metrics"
	"sechelper-auth-template/api/internal/platform/ratelimit"
	platformsecurity "sechelper-auth-template/api/internal/platform/security"
	"sechelper-auth-template/api/internal/platform/session"
)

var releaseVersion = "dev"
var buildID = "dev-local"
var buildRevision = "working-tree"
var buildEnvironment = "production"

func main() {
	startedAt := time.Now().UTC()
	cfg, err := config.Load(os.Args[1:]...)
	if err != nil {
		panic(err)
	}
	db, err := openDatabase(cfg)
	if err != nil {
		panic(err)
	}
	centerKey := cfg.Bootstrap.EncryptionKey
	if centerKey == "" {
		centerKey = cfg.Session.EncryptionKey
	}
	centerProtector, err := session.NewTokenProtector(centerKey)
	if err != nil {
		_ = db.Close()
		panic(fmt.Errorf("configuration center protector: %w", err))
	}
	centerReader := configurationapplication.NewService(configurationpersistence.NewRepository(db), centerProtector)
	installed, err := centerReader.IsInstalled(context.Background())
	if err != nil {
		_ = db.Close()
		panic(fmt.Errorf("read installation state: %w", err))
	}
	cfg.Bootstrap.InstallMode = !installed && cfg.Bootstrap.InstallKey != ""
	centerCtx, cancelCenter := context.WithTimeout(context.Background(), 5*time.Second)
	centerValues, err := centerReader.Resolve(centerCtx)
	cancelCenter()
	if err != nil {
		_ = db.Close()
		panic(fmt.Errorf("load configuration center: %w", err))
	}
	resolvedCfg := cfg
	if installed {
		resolvedCfg, err = config.ApplyConfigurationValues(cfg, centerValues)
	}
	if err != nil {
		_ = db.Close()
		panic(err)
	}
	if resolvedCfg.Bootstrap.DatabaseURL != cfg.Bootstrap.DatabaseURL {
		_ = db.Close()
		db, err = openDatabase(resolvedCfg)
		if err != nil {
			panic(fmt.Errorf("open configuration center database: %w", err))
		}
	}
	cfg = resolvedCfg
	if !cfg.Bootstrap.InstallMode && cfg.App.Environment != buildEnvironment {
		panic(fmt.Errorf("binary build environment %q does not match config environment %q", buildEnvironment, cfg.App.Environment))
	}
	logger, closeLogger, err := logging.New(cfg.Log, "auth-template", cfg.App.Environment, releaseVersion)
	if err != nil {
		panic(err)
	}
	logger = logger.With(zap.String("buildId", buildID), zap.String("sourceRevision", buildRevision))
	defer closeLogger()
	defer db.Close()
	identityClient := identity.NewClient(cfg.Identity, &http.Client{Timeout: 15 * time.Second})
	limiter, closeLimiter, err := newRateLimiter(cfg)
	if err != nil {
		logger.Fatal("rate_limiter.create.failed", zap.Error(err))
	}
	defer closeLimiter()
	serviceMetrics := &appmetrics.Metrics{}
	sessions := session.NewPostgresStore(db)
	protectorKey := cfg.Session.EncryptionKey
	if protectorKey == "" && cfg.Bootstrap.InstallMode {
		protectorKey = centerKey
	}
	protector, err := session.NewTokenProtector(protectorKey)
	if err != nil {
		logger.Fatal("session.protector.failed", zap.Error(err))
	}
	authService := application.NewService(identityClient, sessions, sessions, protector, cfg.Session.TTL, cfg.Identity.ApplicationCode)
	authService.SetUserResolver(cfg.Identity.Issuer, authpersistence.NewPostgresUserResolver(db))
	authService.SetRemoteRevokeObserver(func(err error) {
		serviceMetrics.RemoteRevokeTotal.Add(1)
		if err != nil {
			serviceMetrics.RemoteRevokeErrorsTotal.Add(1)
			logger.Warn("identity.revoke.failed", zap.Error(err))
		}
	})
	var authorizationCache authorizationapplication.Cache = authorizationapplication.NewMemoryCache()
	var closeAuthorizationCache func()
	if cfg.App.RedisURL != "" {
		redisCache, cacheErr := authorizationpersistence.NewRedisCache(cfg.App.RedisURL)
		if cacheErr != nil {
			logger.Fatal("authorization_cache.create.failed", zap.Error(cacheErr))
		}
		redisPingCtx, cancelRedisPing := context.WithTimeout(context.Background(), 5*time.Second)
		if cacheErr := redisCache.Ping(redisPingCtx); cacheErr != nil {
			cancelRedisPing()
			logger.Fatal("authorization_cache.ping.failed", zap.Error(cacheErr))
		}
		cancelRedisPing()
		authorizationCache = redisCache
		closeAuthorizationCache = func() { _ = redisCache.Close() }
	} else {
		closeAuthorizationCache = func() {}
	}
	defer closeAuthorizationCache()
	authorizationModule := authorization.New(sessions, cfg.Session.CookieName, authorizationCache)
	accountModule := account.New(sessions, cfg.Session.CookieName, authService.Get)
	auditModule := audit.New(auditapplication.NewService(auditpersistence.NewRepository(db)))
	configurationService := configurationapplication.NewService(configurationpersistence.NewRepository(db), centerProtector)
	configurationModule := configuration.New(configurationService, func(ctx context.Context, actor, action, key string) {
		eventID, err := session.NewID()
		if err != nil {
			return
		}
		_ = auditModule.Service.Record(ctx, plataudit.Event{ID: "evt-" + eventID, EventType: "SECURITY_CONFIGURATION_CHANGED", Outcome: "success", ActorSubject: actor, ApplicationCode: cfg.Identity.ApplicationCode, ResourceType: "configuration", ResourceID: key, Action: action, Source: "admin_ui"})
	})
	installationModule := installation.New(configurationService, cfg.Bootstrap.InstallKey, cfg.Bootstrap.InstallLockFile, centerProtector)
	authService.SetAuditRecorder(auditModule.Service)
	authorizationModule.ResourceHandler.SetDecisionRecorder(func(ctx context.Context, actor authorizationdomain.Context, decision authorizationdomain.Decision, requestID string) {
		result := "denied"
		if decision.Allowed {
			result = "success"
		}
		eventID, err := session.NewID()
		if err != nil {
			return
		}
		_ = auditModule.Service.Record(ctx, plataudit.Event{ID: "evt-" + eventID, ActorSubject: actor.Subject, ApplicationCode: actor.ApplicationCode, EventType: "ACCESS_DECISION_CHECKED", Outcome: result, ResourceType: decision.ResourceType, ResourceID: decision.ResourceID, Action: decision.Action, RequestID: requestID, Source: "admin_ui", Metadata: map[string]any{"reasonCode": decision.ReasonCode, "permission": decision.Permission}})
	})
	manifestModule := manifest.New(cfg.Identity.ApplicationCode)
	manifestRemote := manifestpersistence.NewIdentityRemote(identity.NewManifestClient(cfg.Identity, &http.Client{Timeout: 15 * time.Second}))
	manifestState := manifestpersistence.NewPostgresStateStore(db)
	manifestSync := manifestapplication.NewSyncService(manifestModule.Registry, manifestRemote, manifestState, cfg.Identity.ApplicationCode, authorizationModule.Service.InvalidateApplication)
	manifestModule.AttachSync(manifestSync)
	manifestModule.SetAuditRecorder(func(ctx context.Context, state manifestapplication.State, err error, requestID string) {
		eventID, idErr := session.NewID()
		if idErr != nil {
			return
		}
		eventType, outcome := "MANIFEST_SYNC_SUCCEEDED", "success"
		if err != nil {
			eventType, outcome = "MANIFEST_SYNC_FAILED", "failure"
		}
		_ = auditModule.Service.Record(ctx, plataudit.Event{ID: "evt-" + eventID, EventType: eventType, Outcome: outcome, ActorSubject: "", ApplicationCode: cfg.Identity.ApplicationCode, ResourceType: "authorization_manifest", ResourceID: cfg.Identity.ApplicationCode, Action: "sync", RequestID: requestID, Source: "admin_ui", Metadata: map[string]any{"manifestVersion": state.ManifestVersion, "serverRevision": state.ServerRevision}})
	})
	manifestSync.SetErrorHandler(func(err error) {
		logger.Error("manifest.scheduled_sync.failed", zap.Error(err), zap.String("component", "manifest"), zap.String("operation", "scheduled_sync"))
	})
	dashboardManifestReader := func(ctx context.Context) (dashboardapplication.ManifestState, error) {
		value, err := manifestSync.Current(ctx)
		if err != nil {
			return dashboardapplication.ManifestState{}, err
		}
		return dashboardapplication.ManifestState{ApplicationCode: value.ApplicationCode, Version: value.ManifestVersion, Status: value.Status, ContentHash: value.ContentHash, ServerRevision: value.ServerRevision, UpdatedAt: value.UpdatedAt}, nil
	}
	var cachePinger dashboardapplication.CachePinger
	if checker, ok := authorizationCache.(interface{ Ping(context.Context) error }); ok {
		cachePinger = checker.Ping
	}
	resourceCollector := dashboardapplication.NewSystemResourcesCollector(".", dashboardapplication.DefaultResourceSampleInterval)
	defer resourceCollector.Close()
	dashboardModule := dashboard.New(dashboardapplication.NewService(db, dashboardapplication.AppInfo{Name: cfg.App.Name, Version: releaseVersion, Environment: cfg.App.Environment, BuildID: buildID, SourceRevision: buildRevision, StartedAt: startedAt}, dashboardManifestReader, cachePinger, resourceCollector))
	appInfo := operationsapplication.AppInfo{Name: cfg.App.Name, Version: releaseVersion, Environment: cfg.App.Environment, BuildID: buildID, SourceRevision: buildRevision, StartedAt: startedAt}
	operationsService := operationsapplication.NewService(db, appInfo, serviceMetrics)
	operationsService.SetDeploymentInfo(operationsapplication.DeploymentInfo{App: appInfo, BootstrapDatabaseConfigured: cfg.Bootstrap.DatabaseURL != "", BootstrapEncryptionConfigured: cfg.Bootstrap.EncryptionKey != "", ConfigurationLoaded: true})
	operationsModule := operations.New(operationsService)
	if err := registerFrameworkPermissions(manifestModule); err != nil {
		logger.Fatal("manifest registration failed", zap.Error(err))
	}
	businessRuntime, err := newBusinessRuntime(db)
	if err != nil {
		logger.Fatal("business registration failed", zap.Error(err))
	}
	businessRuntime.SetConfiguration(configurationModule.Provider())
	if err := businessRuntime.RegisterPermissions(manifestModule); err != nil {
		logger.Fatal("business permission registration failed", zap.Error(err))
	}
	if err := businessRuntime.RegisterResources(authorizationModule); err != nil {
		logger.Fatal("business resource registration failed", zap.Error(err))
	}
	if err := businessRuntime.RegisterAuditEvents(auditModule.Service); err != nil {
		logger.Fatal("business audit event registration failed", zap.Error(err))
	}
	startupCtx, cancelStartup := context.WithTimeout(context.Background(), 15*time.Second)
	manifestReady := false
	if _, err := manifestSync.Sync(startupCtx); err != nil {
		cancelStartup()
		logger.Error("manifest.startup_sync.failed", zap.Error(err))
	} else {
		manifestReady = true
		cancelStartup()
	}
	stopManifestSync := manifestSync.Start(context.Background(), cfg.Manifest.SyncInterval)
	defer stopManifestSync()
	router, err := buildRouter(cfg, logger, db, func() bool { return manifestReady }, limiter, serviceMetrics, authhttp.NewHandler(authService, cfg.Session, cfg.App.PublicWebOrigin), authorizationModule, manifestModule, dashboardModule, accountModule, auditModule, operationsModule, configurationModule, installationModule, businessRuntime)
	if err != nil {
		logger.Fatal("http routes registration failed", zap.Error(err))
	}
	server := &http.Server{Addr: cfg.App.ListenAddr, Handler: router, ReadTimeout: cfg.Server.ReadTimeout, WriteTimeout: cfg.Server.WriteTimeout, IdleTimeout: cfg.Server.IdleTimeout}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		logger.Info("server.started", zap.String("address", cfg.App.ListenAddr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server.failed", zap.Error(err))
		}
	}()
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()
	_ = server.Shutdown(shutdown)
	logger.Info("server.stopped")
}

func buildRouter(cfg config.Config, logger *zap.Logger, db *sql.DB, manifestReady func() bool, limiter ratelimit.Limiter, serviceMetrics *appmetrics.Metrics, authHandler *authhttp.Handler, authorizationModule *authorization.Module, manifestModule *manifest.Module, dashboardModule *dashboard.Module, accountModule *account.Module, auditModule *audit.Module, operationsModule *operations.Module, configurationModule *configuration.Module, installationModule *installation.Module, businessRuntime *businessRuntime) (*gin.Engine, error) {
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery(), httpkit.RequestIDMiddleware(), platformsecurity.HostOriginPolicy(cfg), requestLogger(logger), cors(cfg.CORS.AllowedOrigins), rateLimitAuth(limiter, cfg.RateLimit, serviceMetrics))
	installationModule.RegisterRoutes(r)
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	metrics := r.Group("/metrics")
	if cfg.App.Environment == "production" {
		metrics.Use(platformsecurity.MetricsAuth(cfg.Security.MetricsToken))
	}
	metrics.GET("", gin.WrapH(serviceMetrics.Handler()))
	r.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil || !manifestReady() {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	v1 := r.Group("/v1")
	authHandler.Register(v1)
	authorizationModule.RegisterRoutes(v1)
	dashboardModule.RegisterRoutes(v1, authorizationModule.Middleware.RequirePermission("admin:access"))
	accountModule.RegisterRoutes(v1, authorizationModule.Middleware.RequirePermission("auth:session"))
	authorizationModule.RegisterResourceRoutes(v1, authorizationModule.Middleware.RequirePermission("admin:access"))
	auditModule.RegisterRoutes(v1, authorizationModule.Middleware.RequirePermission("audit:read"))
	operationsModule.RegisterRoutes(v1, authorizationModule.Middleware.RequirePermission("admin:access"), authorizationModule.Middleware.RequirePermission("admin:access"))
	configurationModule.RegisterRoutes(v1, authorizationModule.Middleware.RequirePermissions("admin:access", "configuration:read"), authorizationModule.Middleware.RequirePermissions("admin:access", "configuration:write"))
	manifestModule.RegisterRoutes(v1, authorizationModule.Middleware.RequirePermissions("admin:access", "auth:manifest:read"), authorizationModule.Middleware.RequirePermissions("admin:access", "auth:manifest:sync"))
	if err := businessRuntime.RegisterRoutes(v1, authorizationModule.Middleware); err != nil {
		return nil, err
	}
	return r, nil
}

func openDatabase(cfg config.Config) (*sql.DB, error) {
	dsn := cfg.Bootstrap.DatabaseURL
	if dsn == "" {
		return nil, fmt.Errorf("database.open.failed: bootstrap.databaseUrl is required")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("database.open.failed: %w", err)
	}
	pingCtx, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPing()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database.ping.failed: %w", err)
	}
	return db, nil
}

func newRateLimiter(cfg config.Config) (ratelimit.Limiter, func(), error) {
	if cfg.App.RedisURL == "" {
		return ratelimit.NewMemory(), func() {}, nil
	}
	value, err := ratelimit.NewRedis(cfg.App.RedisURL)
	if err != nil {
		return nil, nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = value.Ping(ctx)
	cancel()
	if err != nil {
		_ = value.Close()
		return nil, nil, err
	}
	return value, func() { _ = value.Close() }, nil
}

func rateLimitAuth(limiter ratelimit.Limiter, cfg config.RateLimitConfig, serviceMetrics *appmetrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := 0
		switch c.Request.URL.Path {
		case "/v1/auth/login":
			limit = cfg.LoginPerMinute
			serviceMetrics.AuthLoginTotal.Add(1)
		case "/v1/auth/callback":
			limit = cfg.CallbackPerMinute
			serviceMetrics.AuthCallbackTotal.Add(1)
		case "/v1/auth/refresh":
			limit = cfg.RefreshPerMinute
			serviceMetrics.AuthRefreshTotal.Add(1)
		case "/v1/auth/logout":
			serviceMetrics.AuthLogoutTotal.Add(1)
		}
		if limit == 0 {
			c.Next()
			return
		}
		clientIP := c.Request.RemoteAddr
		if host, _, splitErr := net.SplitHostPort(c.Request.RemoteAddr); splitErr == nil {
			clientIP = host
		}
		key := clientIP + ":" + c.Request.URL.Path
		allowed, retryAfter, err := limiter.Allow(c.Request.Context(), key, limit, time.Minute)
		if err != nil {
			serviceMetrics.RateLimitErrorsTotal.Add(1)
			httpkit.WriteError(c, http.StatusServiceUnavailable, httpkit.Error{Code: "RATE_LIMIT_DEPENDENCY_FAILED"})
			return
		}
		if !allowed {
			serviceMetrics.RateLimitedTotal.Add(1)
			seconds := int(retryAfter.Seconds())
			if seconds < 1 {
				seconds = 1
			}
			c.Header("Retry-After", fmt.Sprintf("%d", seconds))
			httpkit.WriteError(c, http.StatusTooManyRequests, httpkit.Error{Code: "RATE_LIMITED"})
			return
		}
		c.Next()
	}
}

func registerFrameworkPermissions(m *manifest.Module) error {
	if err := m.Register(domain.Permission{Code: "admin:access", Name: "进入管理端", Description: "访问本应用后台管理能力和平台概览接口", RiskLevel: "privileged", APIs: []domain.API{{Method: "GET", Path: "/v1/authorization/me"}, {Method: "GET", Path: "/v1/admin/dashboard/overview"}, {Method: "GET", Path: "/v1/admin/dashboard/resources"}, {Method: "GET", Path: "/v1/admin/deployment/guide"}, {Method: "GET", Path: "/v1/admin/resources"}, {Method: "POST", Path: "/v1/admin/access-decisions/check"}, {Method: "GET", Path: "/v1/admin/operations/overview"}}}); err != nil {
		return err
	}
	if err := m.Register(domain.Permission{Code: "auth:session", Name: "查看当前会话", Description: "读取当前登录会话和账号上下文", RiskLevel: "normal", APIs: []domain.API{{Method: "GET", Path: "/v1/auth/session"}, {Method: "GET", Path: "/v1/admin/account"}}}); err != nil {
		return err
	}
	if err := m.Register(domain.Permission{Code: "auth:manifest:read", Name: "查看认证 Manifest", Description: "查看本 Application 的 Manifest 同步状态", RiskLevel: "privileged", APIs: []domain.API{{Method: "GET", Path: "/v1/internal/authorization-manifest"}}}); err != nil {
		return err
	}
	if err := m.Register(domain.Permission{Code: "audit:read", Name: "查看操作审计", Description: "查看管理端操作审计记录", RiskLevel: "privileged", APIs: []domain.API{{Method: "GET", Path: "/v1/admin/audit-events"}}}); err != nil {
		return err
	}
	if err := m.Register(domain.Permission{Code: "configuration:read", Name: "查看配置中心", Description: "查看配置中心中的环境变量元数据", RiskLevel: "privileged", APIs: []domain.API{{Method: "GET", Path: "/v1/admin/configuration"}, {Method: "GET", Path: "/v1/admin/configuration/{key}"}}}); err != nil {
		return err
	}
	if err := m.Register(domain.Permission{Code: "configuration:write", Name: "修改配置中心", Description: "新增、替换或删除配置中心值", RiskLevel: "critical", APIs: []domain.API{{Method: "PUT", Path: "/v1/admin/configuration/{key}"}, {Method: "DELETE", Path: "/v1/admin/configuration/{key}"}}}); err != nil {
		return err
	}
	if err := m.Register(domain.Permission{Code: "deployment:read", Name: "查看框架检查", Description: "查看框架运行检查和配置中心接入状态", RiskLevel: "privileged", APIs: []domain.API{{Method: "GET", Path: "/v1/admin/deployment/guide"}}}); err != nil {
		return err
	}
	return m.Register(domain.Permission{Code: "auth:manifest:sync", Name: "同步认证 Manifest", Description: "触发本 Application 的 Manifest 同步", RiskLevel: "critical", APIs: []domain.API{{Method: "POST", Path: "/v1/internal/authorization-manifest/sync"}}})
}
func requestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		logger.Info("http.request", zap.String("method", c.Request.Method), zap.String("path", c.FullPath()), zap.Int("status", c.Writer.Status()), zap.Int64("durationMs", time.Since(started).Milliseconds()), zap.String("requestId", httpkit.RequestID(c)))
	}
}
func cors(origins []string) gin.HandlerFunc {
	allowed := map[string]struct{}{}
	for _, origin := range origins {
		allowed[origin] = struct{}{}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allowed[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Expose-Headers", "X-CSRF-Token, X-Request-ID")
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token, X-Request-ID")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
