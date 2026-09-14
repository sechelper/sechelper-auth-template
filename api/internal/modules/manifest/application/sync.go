package application

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"sechelper-auth-template/api/internal/modules/manifest/domain"
)

var ErrSyncRejected = errors.New("manifest synchronization rejected")

type RemoteClient interface {
	Status(context.Context) (RemoteStatus, error)
	Sync(context.Context, domain.Snapshot, string, string) (RemoteReceipt, error)
}
type StateStore interface {
	Load(context.Context) (State, error)
	Save(context.Context, State) error
}
type RemoteStatus struct {
	Status              string
	ManifestVersion     int64
	ContentHash, SyncID string
	ServerRevision      int64
}
type RemoteReceipt struct {
	Status, SyncID string
	ServerRevision int64
}
type State struct {
	ApplicationCode                            string
	ManifestVersion                            int64
	ContentHash, CanonicalJSON, SyncID, Status string
	ServerRevision                             int64
	AcceptedAt, UpdatedAt                      time.Time
}

type SyncService struct {
	registry        *Registry
	remote          RemoteClient
	state           StateStore
	applicationCode string
	mu              sync.Mutex
	onChanged       func(string)
	onError         func(error)
}

func NewSyncService(registry *Registry, remote RemoteClient, state StateStore, applicationCode string, onChanged func(string)) *SyncService {
	return &SyncService{registry: registry, remote: remote, state: state, applicationCode: applicationCode, onChanged: onChanged}
}
func (s *SyncService) Sync(ctx context.Context) (State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	remote, err := s.remote.Status(ctx)
	if err != nil {
		return State{}, fmt.Errorf("manifest status: %w", err)
	}
	if remote.Status != "consistent" && remote.Status != "not_synced" {
		return State{}, fmt.Errorf("unsupported remote manifest status: %s", remote.Status)
	}
	local, localErr := s.state.Load(ctx)
	if localErr != nil && !errors.Is(localErr, sql.ErrNoRows) {
		return State{}, fmt.Errorf("load manifest state: %w", localErr)
	}
	hasLocal := localErr == nil
	version := int64(1)
	if remote.Status == "consistent" && remote.ManifestVersion > 0 {
		version = remote.ManifestVersion
	}
	if hasLocal && local.ManifestVersion > version {
		version = local.ManifestVersion
	}
	snapshot, contentHash, err := s.registry.Snapshot(version)
	if err != nil {
		return State{}, err
	}
	canonicalJSON, err := SnapshotJSON(snapshot)
	if err != nil {
		return State{}, err
	}
	if remote.Status == "consistent" && remote.ContentHash != contentHash {
		version++
		snapshot, contentHash, err = s.registry.Snapshot(version)
		if err != nil {
			return State{}, err
		}
		canonicalJSON, err = SnapshotJSON(snapshot)
		if err != nil {
			return State{}, err
		}
	}
	if remote.Status == "consistent" && remote.ContentHash == contentHash {
		current := State{ApplicationCode: s.applicationCode, ManifestVersion: version, ContentHash: contentHash, CanonicalJSON: string(canonicalJSON), SyncID: remote.SyncID, Status: remote.Status, ServerRevision: remote.ServerRevision, AcceptedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
		if err := s.state.Save(ctx, current); err != nil {
			return State{}, err
		}
		if hasLocal && local.ContentHash != contentHash && s.onChanged != nil {
			s.onChanged(s.applicationCode)
		}
		return current, nil
	}
	key := idempotencyKey(version, contentHash)
	receipt, err := s.remote.Sync(ctx, snapshot, contentHash, key)
	if err != nil {
		return State{}, fmt.Errorf("manifest sync: %w", err)
	}
	if receipt.Status != "applied" && receipt.Status != "consistent" {
		return State{}, fmt.Errorf("%w: %s", ErrSyncRejected, receipt.Status)
	}
	current := State{ApplicationCode: s.applicationCode, ManifestVersion: version, ContentHash: contentHash, CanonicalJSON: string(canonicalJSON), SyncID: receipt.SyncID, Status: receipt.Status, ServerRevision: receipt.ServerRevision, AcceptedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	if err := s.state.Save(ctx, current); err != nil {
		return State{}, err
	}
	if hasLocal && local.ContentHash != contentHash && s.onChanged != nil {
		s.onChanged(s.applicationCode)
	}
	return current, nil
}
func idempotencyKey(version int64, hash string) string {
	raw := sha256.Sum256([]byte(fmt.Sprintf("manifest:%d:%s", version, hash)))
	return "manifest-" + hex.EncodeToString(raw[:])[:32]
}
func SnapshotJSON(snapshot domain.Snapshot) ([]byte, error)       { return json.Marshal(snapshot) }
func (s *SyncService) Current(ctx context.Context) (State, error) { return s.state.Load(ctx) }
func (s *SyncService) SetErrorHandler(handler func(error))        { s.onError = handler }
func (s *SyncService) Start(ctx context.Context, interval time.Duration) context.CancelFunc {
	loopCtx, cancel := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-loopCtx.Done():
				return
			case <-ticker.C:
				if _, err := s.Sync(loopCtx); err != nil && s.onError != nil {
					s.onError(err)
				}
			}
		}
	}()
	return cancel
}
