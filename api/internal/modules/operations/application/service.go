package application

import (
	"context"
	"database/sql"
	appmetrics "sechelper-auth-template/api/internal/platform/metrics"
	"time"
)

type AppInfo struct {
	Name           string    `json:"name"`
	Version        string    `json:"version"`
	Environment    string    `json:"environment"`
	BuildID        string    `json:"buildId"`
	SourceRevision string    `json:"sourceRevision"`
	StartedAt      time.Time `json:"startedAt"`
}
type Service struct {
	db         *sql.DB
	app        AppInfo
	metrics    *appmetrics.Metrics
	deployment DeploymentInfo
}

func NewService(db *sql.DB, app AppInfo, metrics *appmetrics.Metrics) *Service {
	return &Service{db: db, app: app, metrics: metrics}
}

type Status struct {
	Status    string `json:"status"`
	LatencyMs int64  `json:"latencyMs,omitempty"`
	Message   string `json:"message,omitempty"`
}
type Overview struct {
	App          AppInfo
	Dependencies map[string]Status
	Metrics      map[string]int64
}

type DeploymentInfo struct {
	App                           AppInfo
	BootstrapDatabaseConfigured   bool
	BootstrapEncryptionConfigured bool
	ConfigurationLoaded           bool
}
type DeploymentCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}
type DeploymentGuide struct {
	Application   AppInfo           `json:"application"`
	Bootstrap     []DeploymentCheck `json:"bootstrap"`
	Configuration []DeploymentCheck `json:"configuration"`
	Checks        []DeploymentCheck `json:"checks"`
	Ready         bool              `json:"ready"`
}

func (s *Service) SetDeploymentInfo(value DeploymentInfo) { s.deployment = value }

func (s *Service) DeploymentGuide(ctx context.Context) DeploymentGuide {
	guide := DeploymentGuide{Application: s.deployment.App, Bootstrap: []DeploymentCheck{
		{Name: "database", Status: checkStatus(s.deployment.BootstrapDatabaseConfigured), Message: "数据库 bootstrap 连接"},
		{Name: "encryptionKey", Status: checkStatus(s.deployment.BootstrapEncryptionConfigured), Message: "配置中心加密密钥"},
	}, Configuration: []DeploymentCheck{{Name: "runtimeLoaded", Status: checkStatus(s.deployment.ConfigurationLoaded), Message: "启动时加载配置中心"}}}
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := s.db.PingContext(checkCtx); err != nil {
		guide.Checks = append(guide.Checks, DeploymentCheck{Name: "databaseConnectivity", Status: "unavailable", Message: "数据库连接失败"})
	} else {
		guide.Checks = append(guide.Checks, DeploymentCheck{Name: "databaseConnectivity", Status: "healthy", Message: "数据库连接正常"})
	}
	var tableExists, migrationExists bool
	if err := s.db.QueryRowContext(checkCtx, `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema='configuration' AND table_name='configuration_entries')`).Scan(&tableExists); err != nil {
		guide.Checks = append(guide.Checks, DeploymentCheck{Name: "configurationTable", Status: "unavailable", Message: "无法检查配置中心表"})
	} else {
		guide.Checks = append(guide.Checks, DeploymentCheck{Name: "configurationTable", Status: checkStatus(tableExists), Message: "配置中心数据表"})
	}
	if err := s.db.QueryRowContext(checkCtx, `SELECT EXISTS (SELECT 1 FROM framework.schema_migrations WHERE scope='configuration' AND version='008_configuration_center.sql')`).Scan(&migrationExists); err != nil {
		guide.Checks = append(guide.Checks, DeploymentCheck{Name: "configurationMigration", Status: "unavailable", Message: "无法检查配置迁移"})
	} else {
		guide.Checks = append(guide.Checks, DeploymentCheck{Name: "configurationMigration", Status: checkStatus(migrationExists), Message: "配置中心迁移"})
	}
	guide.Ready = true
	for _, group := range [][]DeploymentCheck{guide.Bootstrap, guide.Configuration, guide.Checks} {
		for _, item := range group {
			if item.Status != "healthy" {
				guide.Ready = false
			}
		}
	}
	return guide
}

func checkStatus(ok bool) string {
	if ok {
		return "healthy"
	}
	return "attention"
}

func (s *Service) Overview(ctx context.Context) Overview {
	started := time.Now()
	status := Status{Status: "healthy", LatencyMs: 0}
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := s.db.PingContext(checkCtx); err != nil {
		status = Status{Status: "unavailable", Message: "数据库连接失败"}
	} else {
		status.LatencyMs = time.Since(started).Milliseconds()
	}
	result := Overview{App: s.app, Dependencies: map[string]Status{"api": {Status: "healthy"}, "postgres": status}, Metrics: map[string]int64{}}
	if s.metrics != nil {
		result.Metrics = map[string]int64{"authLoginTotal": s.metrics.AuthLoginTotal.Load(), "authCallbackTotal": s.metrics.AuthCallbackTotal.Load(), "authRefreshTotal": s.metrics.AuthRefreshTotal.Load(), "authLogoutTotal": s.metrics.AuthLogoutTotal.Load(), "rateLimitedTotal": s.metrics.RateLimitedTotal.Load(), "remoteRevokeErrorsTotal": s.metrics.RemoteRevokeErrorsTotal.Load()}
	}
	return result
}
