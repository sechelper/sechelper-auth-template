package http

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"sechelper-auth-template/api/internal/modules/authentication/application"
	"sechelper-auth-template/api/internal/platform/config"
	"sechelper-auth-template/api/internal/platform/httpkit"
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
	auth.GET("/logout/callback", h.logoutCallback)
	auth.POST("/refresh", h.refresh)
	auth.POST("/logout", h.logout)
}
func (h *Handler) login(c *gin.Context) {
	prompt := c.DefaultQuery("prompt", application.PromptLogin)
	if prompt != application.PromptNone && prompt != application.PromptLogin {
		writeError(c, http.StatusBadRequest, "INVALID_LOGIN_PROMPT")
		return
	}
	target, err := h.service.BeginLogin(c.Request.Context(), prompt)
	if err != nil {
		writeError(c, 502, "IDENTITY_DEPENDENCY_FAILED")
		return
	}
	c.Redirect(http.StatusFound, target)
}
func (h *Handler) callback(c *gin.Context) {
	if c.Query("error") != "" {
		if application.LoginPrompt(c.Query("state")) == application.PromptNone {
			c.Redirect(http.StatusFound, h.webOrigin+"/?auth=login-required")
			return
		}
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
	response := identityFields(value)
	response["authenticated"] = true
	response["email"] = value.Email
	response["applicationCode"] = value.ApplicationCode
	response["permissions"] = value.Permissions
	response["expiresAt"] = value.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00")
	c.JSON(200, response)
}
func (h *Handler) logout(c *gin.Context) {
	if !h.validCSRF(c) {
		return
	}
	logoutState, err := session.NewID()
	if err != nil {
		writeError(c, http.StatusInternalServerError, "LOGOUT_STATE_FAILED")
		return
	}
	result := application.LogoutResult{}
	if id := h.cookie(c); id != "" {
		result, _ = h.service.Logout(c.Request.Context(), id, h.webOrigin+"/v1/auth/logout/callback", logoutState)
	}
	h.clearCookie(c)
	if result.EndSessionURL == "" {
		c.JSON(http.StatusOK, gin.H{"loggedOut": true})
		return
	}
	h.setLogoutStateCookie(c, logoutState)
	c.JSON(http.StatusOK, gin.H{"loggedOut": true, "logoutUrl": result.EndSessionURL})
}
func (h *Handler) logoutCallback(c *gin.Context) {
	state, _ := c.Cookie("auth_template_logout_state")
	if state == "" || state != c.Query("state") {
		writeError(c, http.StatusForbidden, "LOGOUT_STATE_FAILED")
		return
	}
	h.clearLogoutStateCookie(c)
	c.Redirect(http.StatusFound, h.webOrigin+"/")
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
	response := identityFields(value)
	response["authenticated"] = true
	response["email"] = value.Email
	response["applicationCode"] = value.ApplicationCode
	response["permissions"] = value.Permissions
	response["expiresAt"] = value.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00")
	c.JSON(http.StatusOK, response)
}

func identityFields(value session.Session) gin.H {
	fields := gin.H{"subject": value.Subject, "identitySubject": value.Subject}
	if len(value.ProfileClaims) > 0 {
		fields["profile"] = value.ProfileClaims
		fields["nickname"] = profileNickname(value.ProfileClaims, value.Subject)
		if picture, ok := value.ProfileClaims["picture"].(string); ok && picture != "" {
			fields["avatarUrl"] = picture
		}
	}
	if value.PlatformUserUUID != "" {
		fields["platformUserUuid"] = value.PlatformUserUUID
	}
	return fields
}

func profileNickname(profile map[string]any, fallback string) string {
	for _, key := range []string{"nickname", "name", "preferred_username"} {
		if value, ok := profile[key].(string); ok && value != "" {
			return value
		}
	}
	return fallback
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
func (h *Handler) setLogoutStateCookie(c *gin.Context, value string) {
	c.SetSameSite(parseSameSite(h.cfg.SameSite))
	c.SetCookie("auth_template_logout_state", value, 300, "/", "", h.cfg.Secure, true)
}
func (h *Handler) clearLogoutStateCookie(c *gin.Context) {
	c.SetSameSite(parseSameSite(h.cfg.SameSite))
	c.SetCookie("auth_template_logout_state", "", -1, "/", "", h.cfg.Secure, true)
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
	httpkit.WriteError(c, status, httpkit.Error{Code: code})
}
