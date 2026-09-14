package metrics

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type Metrics struct {
	AuthLoginTotal, AuthCallbackTotal, AuthRefreshTotal     atomic.Int64
	AuthLogoutTotal, RateLimitedTotal, RateLimitErrorsTotal atomic.Int64
	RemoteRevokeTotal, RemoteRevokeErrorsTotal              atomic.Int64
}

func (m *Metrics) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		values := []struct {
			name  string
			value int64
		}{
			{"auth_login_total", m.AuthLoginTotal.Load()}, {"auth_callback_total", m.AuthCallbackTotal.Load()}, {"auth_refresh_total", m.AuthRefreshTotal.Load()},
			{"auth_logout_total", m.AuthLogoutTotal.Load()}, {"auth_rate_limited_total", m.RateLimitedTotal.Load()}, {"auth_rate_limit_errors_total", m.RateLimitErrorsTotal.Load()},
			{"auth_remote_revoke_total", m.RemoteRevokeTotal.Load()}, {"auth_remote_revoke_errors_total", m.RemoteRevokeErrorsTotal.Load()},
		}
		for _, item := range values {
			_, _ = fmt.Fprintf(w, "# TYPE %s counter\n%s %d\n", item.name, item.name, item.value)
		}
	}
}
