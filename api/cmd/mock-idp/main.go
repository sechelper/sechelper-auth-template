package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type mockServer struct {
	key             *rsa.PrivateKey
	applicationCode string
	mu              sync.Mutex
	codes           map[string]string
	refreshTokens   map[string]struct{}
	manifest        map[string]any
}

func main() {
	healthcheck := flag.Bool("healthcheck", false, "check the local mock provider")
	flag.Parse()
	if *healthcheck {
		os.Exit(0)
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}
	appCode := os.Getenv("MOCK_IDP_APPLICATION_CODE")
	if appCode == "" {
		appCode = "local-app"
	}
	s := &mockServer{key: key, applicationCode: appCode, codes: map[string]string{}, refreshTokens: map[string]struct{}{}, manifest: map[string]any{"status": "not_synced", "manifest_version": int64(0), "content_hash": "", "sync_id": "", "server_revision": int64(0)}}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.health)
	mux.HandleFunc("/.well-known/jwks.json", s.jwks)
	mux.HandleFunc("/oauth/authorize", s.authorize)
	mux.HandleFunc("/oauth/token", s.token)
	mux.HandleFunc("/oauth/userinfo", s.userinfo)
	mux.HandleFunc("/oauth/revoke", s.revoke)
	mux.HandleFunc("/api/v1/business/authorization/me", s.authorization)
	mux.HandleFunc("/api/v1/provisioning/token", s.provisioningToken)
	mux.HandleFunc("/api/v1/provisioning/authorization-manifest/status", s.manifestStatus)
	mux.HandleFunc("/api/v1/provisioning/authorization-manifest", s.manifestSync)
	log.Printf("mock identity provider listening on :9000")
	log.Fatal(http.ListenAndServe(":9000", mux))
}

func (s *mockServer) health(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }

func (s *mockServer) authorize(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	redirect, err := url.Parse(q.Get("redirect_uri"))
	if err != nil || redirect.Scheme == "" || redirect.Host == "" || q.Get("state") == "" || q.Get("nonce") == "" {
		http.Error(w, "invalid authorization request", http.StatusBadRequest)
		return
	}
	code := randomToken()
	s.mu.Lock()
	s.codes[code] = q.Get("nonce")
	s.mu.Unlock()
	params := redirect.Query()
	params.Set("code", code)
	params.Set("state", q.Get("state"))
	redirect.RawQuery = params.Encode()
	http.Redirect(w, r, redirect.String(), http.StatusFound)
}

func (s *mockServer) token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	nonce := ""
	s.mu.Lock()
	if r.FormValue("grant_type") == "authorization_code" {
		nonce = s.codes[r.FormValue("code")]
		delete(s.codes, r.FormValue("code"))
	} else if r.FormValue("grant_type") == "refresh_token" {
		if _, ok := s.refreshTokens[r.FormValue("refresh_token")]; !ok {
			s.mu.Unlock()
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant"})
			return
		}
		nonce = "refreshed"
	} else if r.FormValue("grant_type") == "client_credentials" {
		nonce = "machine"
	} else {
		s.mu.Unlock()
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported_grant_type"})
		return
	}
	refresh := randomToken()
	s.refreshTokens[refresh] = struct{}{}
	s.mu.Unlock()
	access := "access-" + randomToken()
	idToken, err := s.idToken(nonce)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server_error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"access_token": access, "refresh_token": refresh, "id_token": idToken, "token_type": "Bearer", "expires_in": 3600})
}

func (s *mockServer) userinfo(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"sub": "local-user", "preferred_username": "local-user", "name": "Local User", "email": "local@example.test"})
}

func (s *mockServer) authorization(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"subject": "local-user", "application_code": s.applicationCode, "manifest_version": 1, "permissions": map[string]any{"items": []map[string]any{{"permission_code": "admin:access", "name": "Admin", "risk_level": "normal"}, {"permission_code": "auth:session", "name": "Current session", "risk_level": "normal"}, {"permission_code": "auth:manifest:read", "name": "Read manifest", "risk_level": "normal"}, {"permission_code": "auth:manifest:sync", "name": "Sync manifest", "risk_level": "normal"}, {"permission_code": "audit:read", "name": "Audit", "risk_level": "normal"}}}})
}

func (s *mockServer) provisioningToken(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"access_token": "provisioning-" + randomToken(), "token_type": "Bearer", "expires_in": 3600})
}

func (s *mockServer) manifestStatus(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	writeJSON(w, http.StatusOK, s.manifest)
}

func (s *mockServer) manifestSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var input struct {
		ManifestVersion int64  `json:"manifest_version"`
		ContentHash     string `json:"content_hash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.ManifestVersion < 1 || input.ContentHash == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_manifest"})
		return
	}
	s.mu.Lock()
	s.manifest = map[string]any{"status": "applied", "manifest_version": input.ManifestVersion, "content_hash": input.ContentHash, "sync_id": "local-sync", "server_revision": input.ManifestVersion}
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, s.manifest)
}

func (s *mockServer) revoke(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *mockServer) jwks(w http.ResponseWriter, _ *http.Request) {
	modulus := base64.RawURLEncoding.EncodeToString(s.key.PublicKey.N.Bytes())
	exponent := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(s.key.PublicKey.E)).Bytes())
	writeJSON(w, http.StatusOK, map[string]any{"keys": []map[string]string{{"kty": "RSA", "kid": "local-key", "alg": "RS256", "use": "sig", "n": modulus, "e": exponent}}})
}

func (s *mockServer) idToken(nonce string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{"iss": "http://mock-idp:9000", "aud": "local-client", "sub": "local-user", "nonce": nonce, "iat": now.Unix(), "exp": now.Add(time.Hour).Unix()}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "local-key"
	return token.SignedString(s.key)
}

func randomToken() string {
	value := make([]byte, 18)
	if _, err := rand.Read(value); err != nil {
		return "fallback-token"
	}
	return hex.EncodeToString(value)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
