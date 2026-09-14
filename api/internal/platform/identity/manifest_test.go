package identity

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"sechelper-auth-template/api/internal/platform/config"
)

func TestManifestClientUsesClientCredentialsBearerToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/provisioning/token":
			if r.Method != http.MethodPost {
				t.Errorf("token method = %s, want POST", r.Method)
			}
			if id, secret, ok := r.BasicAuth(); !ok || id != "client-1" || secret != "secret-1" {
				t.Errorf("unexpected token client auth: %q %q %v", id, secret, ok)
			}
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			if got := r.Form.Get("grant_type"); got != "client_credentials" {
				t.Errorf("grant_type = %q, want client_credentials", got)
			}
			if got := r.Form.Get("scope"); got != "authorization.manifest" {
				t.Errorf("scope = %q, want authorization.manifest", got)
			}
			if got := r.Form.Get("audience"); got != "provisioning-api" {
				t.Errorf("audience = %q, want provisioning-api", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"service-token","token_type":"Bearer","expires_in":900}`))
		case "/api/v1/provisioning/authorization-manifest/status":
			if got := r.Header.Get("Authorization"); got != "Bearer service-token" {
				t.Errorf("authorization = %q, want Bearer service-token", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "accepted", "manifest_version": 1})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewManifestClient(config.IdentityConfig{
		Issuer:       server.URL,
		ClientID:     "client-1",
		ClientSecret: "secret-1",
	}, server.Client())
	status, err := client.Status(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != "accepted" {
		t.Fatalf("status = %q, want accepted", status.Status)
	}
}
