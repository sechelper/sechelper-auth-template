package application

import (
	"context"
	"database/sql"
	"time"
)

type AppInfo struct {
	Name           string
	Version        string
	Environment    string
	BuildID        string
	SourceRevision string
	StartedAt      time.Time
}

type ManifestState struct {
	ApplicationCode string
	Version         int64
	Status          string
	ContentHash     string
	ServerRevision  int64
	UpdatedAt       time.Time
}

type DependencyStatus struct {
	Status    string
	LatencyMs int64
	Message   string
}

type CurrentUser struct {
	Subject         string
	ApplicationCode string
	PermissionCount int
}

type Overview struct {
	App          AppInfo
	Current      CurrentUser
	Dependencies map[string]DependencyStatus
	Manifest     ManifestState
	Resources    *ResourceMetrics
}

type ManifestReader func(context.Context) (ManifestState, error)
type CachePinger func(context.Context) error
type ResourceSnapshotReader interface {
	Snapshot() (ResourceMetrics, bool)
}

type Service struct {
	db             *sql.DB
	app            AppInfo
	manifestReader ManifestReader
	cachePinger    CachePinger
	resources      ResourceSnapshotReader
}

func NewService(db *sql.DB, app AppInfo, manifestReader ManifestReader, cachePinger CachePinger, resources ...ResourceSnapshotReader) *Service {
	service := &Service{db: db, app: app, manifestReader: manifestReader, cachePinger: cachePinger}
	if len(resources) > 0 {
		service.resources = resources[0]
	}
	return service
}

func (s *Service) AppInfo() AppInfo { return s.app }

func (s *Service) Overview(ctx context.Context, current CurrentUser) Overview {
	result := Overview{App: s.app, Current: current, Dependencies: map[string]DependencyStatus{
		"api": {Status: "healthy"},
	}}
	if s.resources != nil {
		if resources, ok := s.resources.Snapshot(); ok {
			result.Resources = &resources
		}
	}
	result.Dependencies["postgres"] = s.checkDatabase(ctx)
	result.Dependencies["redis"] = s.checkCache(ctx)
	if s.manifestReader == nil {
		return result
	}
	manifest, err := s.manifestReader(ctx)
	if err != nil {
		result.Dependencies["manifest"] = DependencyStatus{Status: "degraded", Message: "Manifest 状态暂时不可用"}
		return result
	}
	result.Manifest = manifest
	status := manifest.Status
	if status == "" {
		status = "not_synced"
	}
	if status == "consistent" || status == "applied" {
		result.Dependencies["manifest"] = DependencyStatus{Status: "healthy"}
	} else {
		result.Dependencies["manifest"] = DependencyStatus{Status: "degraded", Message: status}
	}
	return result
}

func (s *Service) checkDatabase(ctx context.Context) DependencyStatus {
	started := time.Now()
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := s.db.PingContext(checkCtx); err != nil {
		return DependencyStatus{Status: "unavailable", LatencyMs: time.Since(started).Milliseconds(), Message: "数据库连接失败"}
	}
	return DependencyStatus{Status: "healthy", LatencyMs: time.Since(started).Milliseconds()}
}

func (s *Service) checkCache(ctx context.Context) DependencyStatus {
	if s.cachePinger == nil {
		return DependencyStatus{Status: "not_configured", Message: "未配置 Redis，使用进程内缓存"}
	}
	started := time.Now()
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := s.cachePinger(checkCtx); err != nil {
		return DependencyStatus{Status: "unavailable", LatencyMs: time.Since(started).Milliseconds(), Message: "Redis 连接失败"}
	}
	return DependencyStatus{Status: "healthy", LatencyMs: time.Since(started).Milliseconds()}
}
