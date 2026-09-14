package application

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"sechelper-auth-template/api/internal/modules/manifest/domain"
)

type fakeRemote struct {
	status  RemoteStatus
	receipt RemoteReceipt
	synced  bool
}

func (f *fakeRemote) Status(context.Context) (RemoteStatus, error) { return f.status, nil }
func (f *fakeRemote) Sync(_ context.Context, _ domain.Snapshot, _ string, _ string) (RemoteReceipt, error) {
	f.synced = true
	return f.receipt, nil
}

type fakeState struct {
	value  State
	exists bool
}

func (f *fakeState) Load(context.Context) (State, error) {
	if !f.exists {
		return State{}, sql.ErrNoRows
	}
	return f.value, nil
}
func (f *fakeState) Save(_ context.Context, value State) error {
	f.value, f.exists = value, true
	return nil
}

func TestSyncAppliesWhenRemoteIsNotConsistent(t *testing.T) {
	remote := &fakeRemote{status: RemoteStatus{Status: "not_synced"}, receipt: RemoteReceipt{Status: "applied", SyncID: "sync-1"}}
	state := &fakeState{}
	service := NewSyncService(NewRegistry("demo"), remote, state, "demo", nil)
	if _, err := service.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !remote.synced || state.value.Status != "applied" {
		t.Fatal("expected manifest to be applied")
	}
}
func TestSyncInvalidatesOnContentChange(t *testing.T) {
	remote := &fakeRemote{status: RemoteStatus{Status: "not_synced"}, receipt: RemoteReceipt{Status: "applied"}}
	state := &fakeState{exists: true, value: State{ContentHash: "sha256:old", ManifestVersion: 1}}
	changed := ""
	service := NewSyncService(NewRegistry("demo"), remote, state, "demo", func(value string) { changed = value })
	if _, err := service.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if changed != "demo" {
		t.Fatalf("expected invalidation callback, got %q", changed)
	}
	if state.value.UpdatedAt.Before(time.Now().Add(-time.Minute)) {
		t.Fatal("state was not updated")
	}
}
