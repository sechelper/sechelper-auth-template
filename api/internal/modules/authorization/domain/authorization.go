package domain

import "context"

type Context struct {
	Subject string
	// PlatformUserUUID is the stable local user identifier for business data ownership.
	PlatformUserUUID string
	ApplicationCode  string
	ManifestVersion  int64
	Permissions      map[string]struct{}
}

type contextKey struct{}

// WithContext attaches the authenticated principal to a request context.
func WithContext(ctx context.Context, value Context) context.Context {
	return context.WithValue(ctx, contextKey{}, value)
}

// FromContext returns the authenticated principal installed by authorization middleware.
// Business handlers can use PlatformUserUUID for local data ownership and auditing.
func FromContext(ctx context.Context) (Context, bool) {
	value, ok := ctx.Value(contextKey{}).(Context)
	return value, ok
}

func (c Context) HasPermission(code string) bool {
	_, ok := c.Permissions[code]
	return ok
}
