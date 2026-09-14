package http

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"sechelper-auth-template/api/internal/modules/authentication/application"
	"sechelper-auth-template/api/internal/platform/config"
	"sechelper-auth-template/api/internal/platform/session"
)

type Handler struct {
	service   *application.Service
	cfg       config.SessionConfig
	webOrigin string
}

func NewHandler(service *application.Service, cfg config.SessionConfig, webOrigin string) *Handler {
	return &Handler{service: service, cfg: cfg, webOrigin: webOrigin}
}
func (h *Handler) Register(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	auth.GET("/login", h.login)
	auth.GET("/callback", h.callback)
	auth.GET("/session", h.session)
	auth.POST("/refresh", h.refresh)
	auth.POST("/logout", h.logout)
}
func (h *Handler) login(c *gin.Context) {
	target, err := h.service.BeginLogin(c.Request.Context())
	if err != nil {
		writeError(c, 502, "IDENTITY_DEPENDENCY_FAILED")
		return
	}
	c.Redirect(http.StatusFound, target)
}
func (h *Handler) callback(c *gin.Context) {
	if c.Query("error") != "" {
		writeError(c, 401, "IDENTITY_LOGIN_FAILED")
		return
	}
	value, err := h.service.CompleteLogin(c.Request.Context(), c.Query("state"), c.Query("code"))
	if err != nil {
		writeError(c, 401, "AUTHENTICATION_FAILED")
		return
	}
	h.setCookie(c, value.ID, value.ExpiresAt)
	c.Redirect(http.StatusFound, h.webOrigin+"/")
}
func (h *Handler) session(c *gin.Context) {
	h.ensureCSRF(c)
	value, err := h.current(c)
	if err != nil {
		c.JSON(200, gin.H{"authenticated": false})
		return
	}
	c.JSON(200, gin.H{"authenticated": true, "subject": value.Subject, "email": value.Email, "applicationCode": value.ApplicationCode, "permissions": value.Permissions, "expiresAt": value.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00")})
}
func (h *Handler) logout(c *gin.Context) {
	if !h.validCSRF(c) {
		return
	}
	if id := h.cookie(c); id != "" {
		_ = h.service.Logout(c.Request.Context(), id)
	}
	h.clearCookie(c)
	c.Status(http.StatusNoContent)
}
func (h *Handler) refresh(c *gin.Context) {
	if !h.validCSRF(c) {
		return
	}
	value, err := h.service.Refresh(c.Request.Context(), h.cookie(c))
	if err != nil {
		status := http.StatusBadGateway
		if errors.Is(err, session.ErrNotFound) || errors.Is(err, application.ErrRefreshUnavailable) || errors.Is(err, session.ErrRefreshReuse) {
			status = http.StatusUnauthorized
		}
		writeError(c, status, "SESSION_REFRESH_FAILED")
		return
	}
	h.setCookie(c, value.ID, value.ExpiresAt)
	c.JSON(http.StatusOK, gin.H{"authenticated": true, "subject": value.Subject, "email": value.Email, "applicationCode": value.ApplicationCode, "permissions": value.Permissions, "expiresAt": value.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00")})
}
func (h *Handler) current(c *gin.Context) (session.Session, error) {
	return h.service.Get(c.Request.Context(), h.cookie(c))
}
func (h *Handler) cookie(c *gin.Context) string { value, _ := c.Cookie(h.cfg.CookieName); return value }
func (h *Handler) setCookie(c *gin.Context, id string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 1 {
		maxAge = int(h.cfg.TTL.Seconds())
	}
	c.SetSameSite(parseSameSite(h.cfg.SameSite))
	c.SetCookie(h.cfg.CookieName, id, maxAge, "/", "", h.cfg.Secure, true)
}
func (h *Handler) clearCookie(c *gin.Context) {
	c.SetSameSite(parseSameSite(h.cfg.SameSite))
	c.SetCookie(h.cfg.CookieName, "", -1, "/", "", h.cfg.Secure, true)
}
func (h *Handler) ensureCSRF(c *gin.Context) {
	value, err := c.Cookie("auth_template_csrf")
	if err != nil || value == "" {
		value, _ = session.NewID()
		c.SetSameSite(parseSameSite(h.cfg.SameSite))
		c.SetCookie("auth_template_csrf", value, int(h.cfg.TTL.Seconds()), "/", "", h.cfg.Secure, false)
	}
	c.Header("X-CSRF-Token", value)
}
func (h *Handler) validCSRF(c *gin.Context) bool {
	csrf, _ := c.Cookie("auth_template_csrf")
	if csrf == "" || csrf != c.GetHeader("X-CSRF-Token") {
		writeError(c, http.StatusForbidden, "CSRF_FAILED")
		return false
	}
	return true
}
func parseSameSite(value string) http.SameSite {
	switch strings.ToLower(value) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
func writeError(c *gin.Context, status int, code string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": code, "requestId": c.GetHeader("X-Request-ID")}})
}
