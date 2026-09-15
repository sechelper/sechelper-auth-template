package identity

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"sechelper-auth-template/api/internal/platform/config"
)

type Client struct {
	cfg       config.IdentityConfig
	http      *http.Client
	keysMu    sync.RWMutex
	keys      map[string]any
	keysUntil time.Time
}
type TokenSet struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}
type TokenError struct {
	StatusCode int
	ErrorCode  string
}

func (e *TokenError) Error() string {
	return fmt.Sprintf("identity token rejected: http %d (%s)", e.StatusCode, e.ErrorCode)
}

type IDTokenClaims struct {
	Nonce string `json:"nonce"`
	jwt.RegisteredClaims
}
type UserInfo struct {
	Subject  string         `json:"sub"`
	Username string         `json:"preferred_username"`
	Name     string         `json:"name"`
	Nickname string         `json:"nickname"`
	Picture  string         `json:"picture"`
	Email    string         `json:"email"`
	Claims   map[string]any `json:"-"`
}
type Permission struct {
	Code      string `json:"permission_code"`
	Name      string `json:"name,omitempty"`
	RiskLevel string `json:"risk_level,omitempty"`
}
type Authorization struct {
	Subject         string `json:"subject"`
	ApplicationCode string `json:"application_code"`
	ManifestVersion int64  `json:"manifest_version"`
	Permissions     struct {
		Items []Permission `json:"items"`
	} `json:"permissions"`
}

func NewClient(cfg config.IdentityConfig, httpClient *http.Client) *Client {
	return &Client{cfg: cfg, http: httpClient}
}

func (c *Client) AuthorizationURL(state, nonce, challenge string) (string, error) {
	u, err := url.Parse(c.cfg.AuthorizationEndpoint)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("client_id", c.cfg.ClientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", c.cfg.RedirectURI)
	q.Set("scope", strings.Join(withProfileScope(c.cfg.Scopes), " "))
	q.Set("state", state)
	q.Set("nonce", nonce)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func withProfileScope(scopes []string) []string {
	result := make([]string, 0, len(scopes)+1)
	seen := make(map[string]struct{}, len(scopes)+1)
	for _, value := range scopes {
		for _, scope := range strings.Fields(value) {
			if _, exists := seen[scope]; exists {
				continue
			}
			seen[scope] = struct{}{}
			result = append(result, scope)
		}
	}
	if _, exists := seen["profile"]; !exists {
		result = append(result, "profile")
	}
	return result
}

func (c *Client) ExchangeCode(ctx context.Context, code, verifier string) (TokenSet, error) {
	form := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "redirect_uri": {c.cfg.RedirectURI}, "code_verifier": {verifier}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return TokenSet{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.cfg.ClientID, c.cfg.ClientSecret)
	return c.doToken(req)
}
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (TokenSet, error) {
	form := url.Values{"grant_type": {"refresh_token"}, "refresh_token": {refreshToken}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return TokenSet{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.cfg.ClientID, c.cfg.ClientSecret)
	return c.doToken(req)
}

// ClientCredentialsToken obtains a service access token for calls made by the
// business application to the identity platform. The client secret is sent
// only to the token endpoint; downstream APIs receive a Bearer access token.
func (c *Client) ClientCredentialsToken(ctx context.Context) (TokenSet, error) {
	form := url.Values{"grant_type": {"client_credentials"}}
	if c.cfg.Audience != "" {
		form.Set("audience", c.cfg.Audience)
	}
	if len(c.cfg.Scopes) > 0 {
		form.Set("scope", strings.Join(c.cfg.Scopes, " "))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return TokenSet{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.cfg.ClientID, c.cfg.ClientSecret)
	return c.doToken(req)
}

// ProvisioningToken obtains the machine token used by provisioning APIs. It
// deliberately does not reuse the OIDC audience/scopes configured for browser
// login: provisioning tokens have a fixed audience and an explicit machine
// scope.
func (c *Client) ProvisioningToken(ctx context.Context, scopes ...string) (TokenSet, error) {
	if len(scopes) == 0 {
		return TokenSet{}, errors.New("provisioning token scope is required")
	}
	form := url.Values{
		"grant_type": {"client_credentials"},
		"scope":      {strings.Join(scopes, " ")},
		"audience":   {"provisioning-api"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.cfg.Issuer, "/")+"/api/v1/provisioning/token", strings.NewReader(form.Encode()))
	if err != nil {
		return TokenSet{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.cfg.ClientID, c.cfg.ClientSecret)
	return c.doToken(req)
}

func (c *Client) RevokeToken(ctx context.Context, token string) error {
	if strings.TrimSpace(c.cfg.RevocationEndpoint) == "" || strings.TrimSpace(token) == "" {
		return nil
	}
	form := url.Values{"token": {token}, "token_type_hint": {"refresh_token"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.RevocationEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.cfg.ClientID, c.cfg.ClientSecret)
	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("identity revoke request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("identity revoke rejected: http %d", res.StatusCode)
	}
	return nil
}
func (c *Client) doToken(req *http.Request) (TokenSet, error) {
	res, err := c.http.Do(req)
	if err != nil {
		return TokenSet{}, fmt.Errorf("identity token request: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return TokenSet{}, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var payload struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(body, &payload)
		return TokenSet{}, &TokenError{StatusCode: res.StatusCode, ErrorCode: payload.Error}
	}
	var out TokenSet
	if err := json.Unmarshal(body, &out); err != nil || out.AccessToken == "" {
		return TokenSet{}, errors.New("identity token response is invalid")
	}
	return out, nil
}

// ValidateIDToken verifies the OIDC token returned by the provider. The key set
// is cached briefly so an IdP outage does not turn every request into a network
// dependency, while key rotation is still picked up automatically.
func (c *Client) ValidateIDToken(ctx context.Context, raw, expectedNonce string) (IDTokenClaims, error) {
	if raw == "" {
		return IDTokenClaims{}, errors.New("identity id_token is missing")
	}
	keys, err := c.jwks(ctx, false)
	if err != nil {
		return IDTokenClaims{}, err
	}
	claims := IDTokenClaims{}
	// OIDC ID Tokens are issued to the confidential client, so their `aud`
	// claim is the client_id. The configured business audience applies to the
	// provider's API access token contract and must not be reused here.
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256", "RS384", "RS512"}), jwt.WithAudience(c.cfg.ClientID), jwt.WithIssuer(c.cfg.Issuer))
	token, err := parser.ParseWithClaims(raw, &claims, func(token *jwt.Token) (any, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, errors.New("identity id_token has no key id")
		}
		key, ok := keys[kid]
		if !ok {
			return nil, fmt.Errorf("identity signing key %q is unknown", kid)
		}
		return key, nil
	})
	if err != nil || !token.Valid {
		// Retry once after clearing the cache to support provider key rotation.
		keys, refreshErr := c.jwks(ctx, true)
		if refreshErr != nil {
			return IDTokenClaims{}, fmt.Errorf("validate identity id_token: %w", err)
		}
		token, err = parser.ParseWithClaims(raw, &claims, func(token *jwt.Token) (any, error) {
			kid, _ := token.Header["kid"].(string)
			key, ok := keys[kid]
			if !ok {
				return nil, fmt.Errorf("identity signing key %q is unknown", kid)
			}
			return key, nil
		})
	}
	if err != nil || !token.Valid || claims.Subject == "" || claims.Nonce != expectedNonce {
		return IDTokenClaims{}, errors.New("identity id_token claims are invalid")
	}
	return claims, nil
}

func (c *Client) jwks(ctx context.Context, force bool) (map[string]any, error) {
	c.keysMu.RLock()
	if !force && len(c.keys) > 0 && time.Now().Before(c.keysUntil) {
		keys := c.keys
		c.keysMu.RUnlock()
		return keys, nil
	}
	c.keysMu.RUnlock()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.JWKSURL, nil)
	if err != nil {
		return nil, err
	}
	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("identity jwks request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("identity jwks rejected: http %d", res.StatusCode)
	}
	var document struct {
		Keys []struct {
			Kty string `json:"kty"`
			Kid string `json:"kid"`
			Alg string `json:"alg"`
			Use string `json:"use"`
			N   string `json:"n"`
			E   string `json:"e"`
			Crv string `json:"crv"`
			X   string `json:"x"`
			Y   string `json:"y"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&document); err != nil {
		return nil, err
	}
	keys := map[string]any{}
	for _, item := range document.Keys {
		if item.Kid == "" || item.Use != "sig" && item.Use != "" {
			continue
		}
		switch item.Kty {
		case "RSA":
			n, nErr := base64.RawURLEncoding.DecodeString(item.N)
			e, eErr := base64.RawURLEncoding.DecodeString(item.E)
			if nErr != nil || eErr != nil || len(e) == 0 {
				continue
			}
			exponent := 0
			for _, b := range e {
				exponent = exponent<<8 | int(b)
			}
			if exponent == 0 {
				continue
			}
			keys[item.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: exponent}
		case "EC":
			curve := map[string]elliptic.Curve{"P-256": elliptic.P256(), "P-384": elliptic.P384(), "P-521": elliptic.P521()}[item.Crv]
			x, xErr := base64.RawURLEncoding.DecodeString(item.X)
			y, yErr := base64.RawURLEncoding.DecodeString(item.Y)
			if curve == nil || xErr != nil || yErr != nil {
				continue
			}
			keys[item.Kid] = &ecdsa.PublicKey{Curve: curve, X: new(big.Int).SetBytes(x), Y: new(big.Int).SetBytes(y)}
		}
	}
	if len(keys) == 0 {
		return nil, errors.New("identity jwks contains no usable signing keys")
	}
	c.keysMu.Lock()
	c.keys, c.keysUntil = keys, time.Now().Add(10*time.Minute)
	c.keysMu.Unlock()
	return keys, nil
}
func (c *Client) GetUserInfo(ctx context.Context, accessToken string) (UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.UserinfoEndpoint, nil)
	if err != nil {
		return UserInfo{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	res, err := c.http.Do(req)
	if err != nil {
		return UserInfo{}, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return UserInfo{}, fmt.Errorf("identity userinfo rejected: http %d", res.StatusCode)
	}
	var raw map[string]any
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&raw); err != nil {
		return UserInfo{}, errors.New("identity userinfo response is invalid")
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return UserInfo{}, errors.New("identity userinfo response is invalid")
	}
	var out UserInfo
	if err := json.Unmarshal(encoded, &out); err != nil || out.Subject == "" {
		return UserInfo{}, errors.New("identity userinfo response is invalid")
	}
	out.Claims = raw
	return out, nil
}
func (c *Client) GetAuthorization(ctx context.Context, accessToken string) (Authorization, error) {
	endpoint := strings.TrimRight(c.cfg.Issuer, "/") + "/api/v1/business/authorization/me"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Authorization{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	res, err := c.http.Do(req)
	if err != nil {
		return Authorization{}, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return Authorization{}, fmt.Errorf("identity authorization rejected: http %d", res.StatusCode)
	}
	var out Authorization
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&out); err != nil || out.Subject == "" || out.ApplicationCode == "" {
		return Authorization{}, errors.New("identity authorization response is invalid")
	}
	return out, nil
}
func (c *Client) Timeout() time.Duration { return 15 * time.Second }
