package security

import (
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"sechelper-auth-template/api/internal/platform/config"
)

// HostOriginPolicy rejects requests whose externally visible host or origin is
// outside the configured deployment boundary. Forwarded headers are accepted
// only from explicitly trusted proxy networks.
func HostOriginPolicy(cfg config.Config) gin.HandlerFunc {
	allowedHosts := map[string]struct{}{originHost(cfg.App.APIOrigin): {}, originHost(cfg.App.PublicWebOrigin): {}}
	trusted := parseCIDRs(cfg.Security.TrustedProxyCIDRs)
	return func(c *gin.Context) {
		if !validHost(c.Request, allowedHosts, trusted) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "HOST_POLICY_REJECTED", "message": "HOST_POLICY_REJECTED", "requestId": c.GetHeader("X-Request-ID")}})
			return
		}
		if origin := c.GetHeader("Origin"); origin != "" && !allowedOrigin(origin, cfg.CORS.AllowedOrigins) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": gin.H{"code": "ORIGIN_POLICY_REJECTED", "message": "ORIGIN_POLICY_REJECTED", "requestId": c.GetHeader("X-Request-ID")}})
			return
		}
		c.Next()
	}
}

func MetricsAuth(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if token == "" || c.GetHeader("Authorization") != "Bearer "+token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "METRICS_UNAUTHORIZED", "message": "METRICS_UNAUTHORIZED", "requestId": c.GetHeader("X-Request-ID")}})
			return
		}
		c.Next()
	}
}

func originHost(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Hostname())
}
func allowedOrigin(value string, allowed []string) bool {
	for _, item := range allowed {
		if value == item {
			return true
		}
	}
	return false
}
func validHost(r *http.Request, allowed map[string]struct{}, trusted []*net.IPNet) bool {
	host := r.Host
	remote := remoteIP(r.RemoteAddr)
	if isTrusted(remote, trusted) {
		if forwarded := r.Header.Get("X-Forwarded-Host"); forwarded != "" {
			host = strings.TrimSpace(strings.Split(forwarded, ",")[0])
		}
	}
	name := strings.ToLower(hostname(host))
	_, ok := allowed[name]
	return ok
}
func hostname(value string) string {
	if h, _, err := net.SplitHostPort(value); err == nil {
		return h
	}
	return value
}
func remoteIP(value string) net.IP {
	if h, _, err := net.SplitHostPort(value); err == nil {
		return net.ParseIP(h)
	}
	return net.ParseIP(value)
}
func parseCIDRs(values []string) []*net.IPNet {
	result := make([]*net.IPNet, 0, len(values))
	for _, value := range values {
		if _, network, err := net.ParseCIDR(strings.TrimSpace(value)); err == nil {
			result = append(result, network)
		}
	}
	return result
}
func isTrusted(ip net.IP, networks []*net.IPNet) bool {
	for _, network := range networks {
		if ip != nil && network.Contains(ip) {
			return true
		}
	}
	return false
}
