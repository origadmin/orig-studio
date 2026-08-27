package tenant

import "context"

type contextKey struct{}

var tenantKey = contextKey{}

type Context struct {
	ID       string
	Slug     string
	Domain   string
	Plan     string
	IsSystem bool
}

func FromContext(ctx context.Context) *Context {
	v, ok := ctx.Value(tenantKey).(*Context)
	if !ok {
		return nil
	}
	return v
}

func WithContext(ctx context.Context, tc *Context) context.Context {
	return context.WithValue(ctx, tenantKey, tc)
}

func IsSystemTenant(ctx context.Context) bool {
	tc := FromContext(ctx)
	return tc == nil || tc.IsSystem
}

func GetID(ctx context.Context) string {
	tc := FromContext(ctx)
	if tc == nil {
		return ""
	}
	return tc.ID
}
