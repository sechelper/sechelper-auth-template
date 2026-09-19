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
	for _, surface := range []string{application.SurfacePublic, application.SurfaceAdmin} {
		surface := surface
		auth := r.Group("/auth/" + surface)
		auth.GET("/login", func(c *gin.Context) { h.login(c, surface) })
		auth.GET("/session", func(c *gin.Context) { h.session(c, surface) })
		auth.GET("/logout/callback", func(c *gin.Context) { h.logoutCallback(c, surface) })
		auth.POST("/refresh", func(c *gin.Context) { h.refresh(c, surface) })
		auth.POST("/logout", func(c *gin.Context) { h.logout(c, surface) })
	}
	r.GET("/auth/callback", h.callback)
}
func (h *Handler) login(c *gin.Context, surface string) {
	prompt := c.DefaultQuery("prompt", application.PromptLogin)
	if prompt != application.PromptNone && prompt != application.PromptLogin {
		writeError(c, http.StatusBadRequest, "INVALID_LOGIN_PROMPT")
		return
	}
	returnTo := c.Query("return_to")
	if !validReturnTo(surface, returnTo) {
		returnTo = "/"
		if surface == application.SurfaceAdmin {
			returnTo = "/admin/"
		}
	}
	target, err := h.service.BeginLogin(c.Request.Context(), prompt, surface, returnTo)
	if err != nil {
		writeError(c, 502, "IDENTITY_DEPENDENCY_FAILED")
		return
	}
	c.Redirect(http.StatusFound, target)
}
func (h *Handler) callback(c *gin.Context) {
	surface := application.LoginSurface(c.Query("state"))
	if c.Query("error") != "" {
		if application.LoginPrompt(c.Query("state")) == application.PromptNone {
			c.Redirect(http.StatusFound, h.redirectPath(surface)+"?auth=login-required")
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
	surface = value.Surface
	h.setCookie(c, surface, value.ID, value.ExpiresAt)
	redirectTo := h.redirectPath(surface)
	if validReturnTo(surface, value.ReturnTo) {
		redirectTo = value.ReturnTo
	}
	c.Redirect(http.StatusFound, redirectTo)
}
func (h *Handler) session(c *gin.Context, surface string) {
	h.ensureCSRF(c, surface)
	value, err := h.current(c, surface)
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
func (h *Handler) logout(c *gin.Context, surface string) {
	if !h.validCSRF(c, surface) {
		return
	}
	logoutState, err := session.NewID()
	if err != nil {
		writeError(c, http.StatusInternalServerError, "LOGOUT_STATE_FAILED")
		return
	}
	result := application.LogoutResult{}
	if id := h.cookie(c, surface); id != "" {
		result, _ = h.service.Logout(c.Request.Context(), id, h.webOrigin+"/v1/auth/"+surface+"/logout/callback", logoutState)
	}
	h.clearCookie(c, surface)
	if result.EndSessionURL == "" {
		c.JSON(http.StatusOK, gin.H{"loggedOut": true})
		return
	}
	h.setLogoutStateCookie(c, surface, logoutState)
	c.JSON(http.StatusOK, gin.H{"loggedOut": true, "logoutUrl": result.EndSessionURL})
}
func (h *Handler) logoutCallback(c *gin.Context, surface string) {
	state, _ := c.Cookie(h.logoutStateCookieName(surface))
	if state == "" || state != c.Query("state") {
		writeError(c, http.StatusForbidden, "LOGOUT_STATE_FAILED")
		return
	}
	h.clearLogoutStateCookie(c, surface)
	c.Redirect(http.StatusFound, h.redirectPath(surface))
}
func (h *Handler) refresh(c *gin.Context, surface string) {
	if !h.validCSRF(c, surface) {
		return
	}
	value, err := h.service.Refresh(c.Request.Context(), h.cookie(c, surface), surface)
	if err != nil {
		status := http.StatusBadGateway
		if errors.Is(err, session.ErrNotFound) || errors.Is(err, application.ErrRefreshUnavailable) || errors.Is(err, session.ErrRefreshReuse) {
			status = http.StatusUnauthorized
		}
		writeError(c, status, "SESSION_REFRESH_FAILED")
		return
	}
	h.setCookie(c, surface, value.ID, value.ExpiresAt)
	// The CSRF cookie has the same lifetime as the local session. Refreshing
	// the session must refresh both cookies, otherwise a long-lived session
	// will eventually fail its next refresh with CSRF_FAILED.
	h.renewCSRF(c, surface)
	response := identityFields(value)
	response["authenticated"] = true
	response["email"] = value.Email
	response["applicationCode"] = value.ApplicationCode
	response["permissions"] = value.Permissions
	response["expiresAt"] = value.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00")
	c.JSON(http.StatusOK, response)
}

func identityFields(value session.Session) gin.H {
	fields := gin.H{
		"subject":         value.Subject,
		"identitySubject": value.Subject,
		"profile":         publicProfile(value),
		"nickname":        profileNickname(value.ProfileClaims, value.Subject),
		"displayName":     profileNickname(value.ProfileClaims, value.Subject),
	}
	if picture, ok := value.ProfileClaims["picture"].(string); ok && picture != "" {
		fields["avatarUrl"] = picture
	}
	if value.PlatformUserUUID != "" {
		fields["platformUserUuid"] = value.PlatformUserUUID
	}
	return fields
}

func publicProfile(value session.Session) map[string]any {
	profile := map[string]any{}
	for _, key := range []string{"sub", "email", "name", "nickname", "picture"} {
		if item, ok := value.ProfileClaims[key].(string); ok && item != "" {
			profile[key] = item
		}
	}
	if _, ok := profile["sub"]; !ok {
		profile["sub"] = value.Subject
	}
	if _, ok := profile["email"]; !ok && value.Email != "" {
		profile["email"] = value.Email
	}
	return profile
}

func profileNickname(profile map[string]any, fallback string) string {
	for _, key := range []string{"nickname", "name"} {
		if value, ok := profile[key].(string); ok && value != "" {
			return value
		}
	}
	return fallback
}
func (h *Handler) current(c *gin.Context, surface string) (session.Session, error) {
	return h.service.Get(c.Request.Context(), h.cookie(c, surface), surface)
}
func (h *Handler) cookie(c *gin.Context, surface string) string {
	value, _ := c.Cookie(h.cookieName(surface))
	return value
}
func (h *Handler) cookieName(surface string) string {
	if surface == application.SurfaceAdmin {
		return h.cfg.CookieName + "_admin"
	}
	return h.cfg.CookieName
}
func (h *Handler) csrfCookieName(surface string) string { return "auth_template_csrf_" + surface }
func (h *Handler) logoutStateCookieName(surface string) string {
	return "auth_template_logout_state_" + surface
}
func (h *Handler) redirectPath(surface string) string {
	if surface == application.SurfaceAdmin {
		return h.webOrigin + "/admin/"
	}
	return h.webOrigin + "/"
}
func validReturnTo(surface, value string) bool {
	if value == "" || !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") {
		return false
	}
	if surface == application.SurfaceAdmin {
		return strings.HasPrefix(value, "/admin/") || value == "/admin"
	}
	return !strings.HasPrefix(value, "/admin")
}
func (h *Handler) setCookie(c *gin.Context, surface, id string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 1 {
		maxAge = int(h.cfg.TTL.Seconds())
	}
	c.SetSameSite(parseSameSite(h.cfg.SameSite))
	c.SetCookie(h.cookieName(surface), id, maxAge, "/", "", h.cfg.Secure, true)
}
func (h *Handler) clearCookie(c *gin.Context, surface string) {
	c.SetSameSite(parseSameSite(h.cfg.SameSite))
	c.SetCookie(h.cookieName(surface), "", -1, "/", "", h.cfg.Secure, true)
}
func (h *Handler) ensureCSRF(c *gin.Context, surface string) {
	name := h.csrfCookieName(surface)
	value, err := c.Cookie(name)
	if err != nil || value == "" {
		value, _ = session.NewID()
	}
	h.setCSRFCookie(c, name, value)
	c.Header("X-CSRF-Token", value)
}
func (h *Handler) renewCSRF(c *gin.Context, surface string) {
	name := h.csrfCookieName(surface)
	value, err := c.Cookie(name)
	if err != nil || value == "" {
		value, _ = session.NewID()
	}
	h.setCSRFCookie(c, name, value)
	c.Header("X-CSRF-Token", value)
}
func (h *Handler) setCSRFCookie(c *gin.Context, name, value string) {
	c.SetSameSite(parseSameSite(h.cfg.SameSite))
	c.SetCookie(name, value, int(h.cfg.TTL.Seconds()), "/", "", h.cfg.Secure, false)
}
func (h *Handler) setLogoutStateCookie(c *gin.Context, surface, value string) {
	c.SetSameSite(parseSameSite(h.cfg.SameSite))
	c.SetCookie(h.logoutStateCookieName(surface), value, 300, "/", "", h.cfg.Secure, true)
}
func (h *Handler) clearLogoutStateCookie(c *gin.Context, surface string) {
	c.SetSameSite(parseSameSite(h.cfg.SameSite))
	c.SetCookie(h.logoutStateCookieName(surface), "", -1, "/", "", h.cfg.Secure, true)
}
func (h *Handler) validCSRF(c *gin.Context, surface string) bool {
	csrf, _ := c.Cookie(h.csrfCookieName(surface))
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
