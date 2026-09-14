package persistence

import (
	"context"

	"sechelper-auth-template/api/internal/modules/manifest/application"
	"sechelper-auth-template/api/internal/modules/manifest/domain"
	"sechelper-auth-template/api/internal/platform/identity"
)

type IdentityRemote struct{ client *identity.ManifestClient }

func NewIdentityRemote(client *identity.ManifestClient) *IdentityRemote {
	return &IdentityRemote{client: client}
}
func (r *IdentityRemote) Status(ctx context.Context) (application.RemoteStatus, error) {
	value, err := r.client.Status(ctx)
	if err != nil {
		return application.RemoteStatus{}, err
	}
	return application.RemoteStatus{Status: value.Status, ManifestVersion: value.ManifestVersion, ContentHash: value.ContentHash, SyncID: value.SyncID, ServerRevision: value.ServerRevision}, nil
}
func (r *IdentityRemote) Sync(ctx context.Context, snapshot domain.Snapshot, contentHash, key string) (application.RemoteReceipt, error) {
	value, err := r.client.Sync(ctx, snapshot, contentHash, key)
	if err != nil {
		return application.RemoteReceipt{}, err
	}
	return application.RemoteReceipt{Status: value.Status, SyncID: value.SyncID, ServerRevision: value.ServerRevision}, nil
}
