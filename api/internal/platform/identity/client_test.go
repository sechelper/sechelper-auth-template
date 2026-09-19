package identity

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"sechelper-auth-template/api/internal/platform/config"
)

func TestValidateIDTokenChecksNonceAndClaims(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/jwks" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{
			"kty": "RSA", "kid": "test-key", "use": "sig", "alg": "RS256",
			"n": base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
			"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.PublicKey.E)).Bytes()),
		}}})
	}))
	defer server.Close()

	c := NewClient(config.IdentityConfig{Issuer: "https://issuer.example", JWKSURL: server.URL + "/jwks", ClientID: "client-1", Audience: "business-api"}, server.Client())
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": "https://issuer.example", "aud": "client-1", "sub": "user-1", "nonce": "nonce-1",
		"exp": time.Now().Add(time.Minute).Unix(), "iat": time.Now().Unix(),
	})
	token.Header["kid"] = "test-key"
	raw, err := token.SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := c.ValidateIDToken(t.Context(), raw, "nonce-1")
	if err != nil || claims.Subject != "user-1" {
		t.Fatalf("ValidateIDToken() = (%+v, %v)", claims, err)
	}
	if _, err := c.ValidateIDToken(t.Context(), raw, "wrong-nonce"); err == nil {
		t.Fatal("ValidateIDToken accepted an invalid nonce")
	}
}

func TestAuthorizationURLRequestsProfileScope(t *testing.T) {
	c := NewClient(config.IdentityConfig{
		AuthorizationEndpoint: "https://issuer.example/authorize",
		ClientID:              "client-1",
		RedirectURI:           "https://app.example/callback",
		Scopes:                []string{"openid", "email"},
	}, http.DefaultClient)
	raw, err := c.AuthorizationURL("state", "nonce", "challenge", "none")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Query().Get("scope") != "openid email profile offline_access" {
		t.Fatalf("authorization scope = %q, want profile included", parsed.Query().Get("scope"))
	}
	if parsed.Query().Get("prompt") != "none" {
		t.Fatalf("authorization prompt = %q, want none", parsed.Query().Get("prompt"))
	}
}

func TestAuthorizationURLDeduplicatesConfiguredScopes(t *testing.T) {
	c := NewClient(config.IdentityConfig{
		AuthorizationEndpoint: "https://issuer.example/authorize",
		ClientID:              "client-1",
		RedirectURI:           "https://app.example/callback",
		Scopes:                []string{"openid", "profile", "email", "profile"},
	}, http.DefaultClient)
	raw, err := c.AuthorizationURL("state", "nonce", "challenge", "login")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Query().Get("scope") != "openid profile email offline_access" {
		t.Fatalf("authorization scope = %q, want each scope once", parsed.Query().Get("scope"))
	}
}

func TestEndSessionURLIncludesOIDCLogoutParameters(t *testing.T) {
	c := NewClient(config.IdentityConfig{EndSessionEndpoint: "https://issuer.example/oauth2/logout"}, http.DefaultClient)
	raw, err := c.EndSessionURL("id-token", "https://app.example/v1/auth/logout/callback", "logout-state")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	if query.Get("id_token_hint") != "id-token" {
		t.Fatalf("id_token_hint = %q, want id-token", query.Get("id_token_hint"))
	}
	if query.Get("post_logout_redirect_uri") != "https://app.example/v1/auth/logout/callback" {
		t.Fatalf("post_logout_redirect_uri = %q, want logout callback", query.Get("post_logout_redirect_uri"))
	}
	if query.Get("state") != "logout-state" {
		t.Fatalf("state = %q, want logout-state", query.Get("state"))
	}
}

func TestGetUserInfoPreservesAllReturnedClaims(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer user-token" {
			t.Fatalf("unexpected userinfo authorization: %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sub":"user-1","nickname":"Mina","picture":"https://cdn.example/avatar.png","locale":"zh-CN","updated_at":1700000000}`))
	}))
	defer server.Close()
	c := NewClient(config.IdentityConfig{UserinfoEndpoint: server.URL}, server.Client())
	value, err := c.GetUserInfo(t.Context(), "user-token")
	if err != nil {
		t.Fatal(err)
	}
	if value.Nickname != "Mina" || value.Picture != "https://cdn.example/avatar.png" || value.Claims["locale"] != "zh-CN" {
		t.Fatalf("userinfo claims not preserved: %+v", value)
	}
}

func TestClientCredentialsTokenUsesConfiguredClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		clientID, clientSecret, ok := r.BasicAuth()
		if !ok || clientID != "client-1" || clientSecret != "secret-1" {
			t.Fatalf("unexpected client authentication: %q %q %v", clientID, clientSecret, ok)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.Form.Get("grant_type") != "client_credentials" || r.Form.Get("audience") != "business-api" {
			t.Fatalf("unexpected token form: %v", r.Form)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"service-token","token_type":"Bearer","expires_in":900}`))
	}))
	defer server.Close()

	c := NewClient(config.IdentityConfig{TokenEndpoint: server.URL + "/token", ClientID: "client-1", ClientSecret: "secret-1", Audience: "business-api"}, server.Client())
	tokens, err := c.ClientCredentialsToken(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if tokens.AccessToken != "service-token" {
		t.Fatalf("access token = %q, want service-token", tokens.AccessToken)
	}
}

func TestRefreshTokenUsesRefreshTokenGrantAndRotatesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		clientID, clientSecret, ok := r.BasicAuth()
		if !ok || clientID != "client-1" || clientSecret != "secret-1" {
			t.Fatalf("unexpected client authentication: %q %q %v", clientID, clientSecret, ok)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if got := r.Form.Get("grant_type"); got != "refresh_token" {
			t.Fatalf("grant_type = %q, want refresh_token", got)
		}
		if got := r.Form.Get("refresh_token"); got != "refresh-token-1" {
			t.Fatalf("refresh_token = %q, want refresh-token-1", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"access-token-2","refresh_token":"refresh-token-2","token_type":"Bearer","expires_in":900}`))
	}))
	defer server.Close()

	c := NewClient(config.IdentityConfig{TokenEndpoint: server.URL + "/token", ClientID: "client-1", ClientSecret: "secret-1"}, server.Client())
	tokens, err := c.RefreshToken(t.Context(), "refresh-token-1")
	if err != nil {
		t.Fatal(err)
	}
	if tokens.AccessToken != "access-token-2" || tokens.RefreshToken != "refresh-token-2" || tokens.ExpiresIn != 900 {
		t.Fatalf("refresh response = %+v, want rotated token set", tokens)
	}
}
