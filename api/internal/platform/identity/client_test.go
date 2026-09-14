package identity

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
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
