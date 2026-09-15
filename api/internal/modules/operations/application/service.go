package application

import (
	"context"
	"database/sql"
	appmetrics "sechelper-auth-template/api/internal/platform/metrics"
	"time"
)

type AppInfo struct {
	Name, Version, Environment, BuildID, SourceRevision string
	StartedAt                                           time.Time
}
type Service struct {
	db      *sql.DB
	app     AppInfo
	metrics *appmetrics.Metrics
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
