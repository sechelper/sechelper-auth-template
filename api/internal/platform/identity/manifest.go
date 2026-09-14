package identity

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"sechelper-auth-template/api/internal/modules/manifest/domain"
	"sechelper-auth-template/api/internal/platform/config"
)

type ManifestClient struct {
	issuer string
	client *Client
}
type ManifestStatus struct {
	Status          string `json:"status"`
	ManifestVersion int64  `json:"manifest_version"`
	ContentHash     string `json:"content_hash"`
	SyncID          string `json:"sync_id"`
	ServerRevision  int64  `json:"server_revision"`
}
type ManifestReceipt struct {
	Status         string `json:"status"`
	SyncID         string `json:"sync_id"`
	ServerRevision int64  `json:"server_revision"`
}

func NewManifestClient(identityConfig config.IdentityConfig, httpClient *http.Client) *ManifestClient {
	return &ManifestClient{issuer: strings.TrimRight(identityConfig.Issuer, "/"), client: NewClient(identityConfig, httpClient)}
}
func (c *ManifestClient) Status(ctx context.Context) (ManifestStatus, error) {
	var out ManifestStatus
	err := c.doJSON(ctx, http.MethodGet, c.issuer+"/api/v1/provisioning/authorization-manifest/status", nil, "", &out)
	return out, err
}
func (c *ManifestClient) Sync(ctx context.Context, snapshot domain.Snapshot, contentHash, idempotencyKey string) (ManifestReceipt, error) {
	body := struct {
		ManifestVersion int64               `json:"manifest_version"`
		ContentHash     string              `json:"content_hash"`
		Permissions     []domain.Permission `json:"permissions"`
	}{snapshot.ManifestVersion, contentHash, snapshot.Permissions}
	raw, err := json.Marshal(body)
	if err != nil {
		return ManifestReceipt{}, err
	}
	var out ManifestReceipt
	err = c.doJSON(ctx, http.MethodPut, c.issuer+"/api/v1/provisioning/authorization-manifest", raw, idempotencyKey, &out)
	return out, err
}
func (c *ManifestClient) doJSON(ctx context.Context, method, endpoint string, body []byte, idempotencyKey string, out any) error {
	var reader io.Reader
	if body != nil {
		reader = strings.NewReader(string(body))
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return err
	}
	tokens, err := c.client.ProvisioningToken(ctx, "authorization.manifest")
	if err != nil {
		return fmt.Errorf("manifest access token: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	res, err := c.client.http.Do(req)
	if err != nil {
		return fmt.Errorf("manifest request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("manifest request rejected: http %d", res.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(out); err != nil {
		return fmt.Errorf("manifest response invalid: %w", err)
	}
	return nil
}
