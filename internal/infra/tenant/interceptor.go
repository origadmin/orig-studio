package tenant

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
)

type TenantInterceptor struct {
	tenantIDFunc func(context.Context) string
	skipFunc     func(context.Context) bool
}

func NewInterceptor(tenantIDFunc func(context.Context) string) *TenantInterceptor {
	return &TenantInterceptor{
		tenantIDFunc: tenantIDFunc,
		skipFunc:     func(ctx context.Context) bool { return false },
	}
}

func (i *TenantInterceptor) WithSkipFunc(skip func(context.Context) bool) *TenantInterceptor {
	i.skipFunc = skip
	return i
}

func (i *TenantInterceptor) QueryTraverser() ent.InterceptFunc {
	return func(next ent.Querier) ent.Querier {
		return ent.QuerierFunc(func(ctx context.Context, q ent.Query) (ent.Value, error) {
			if i.skipFunc(ctx) {
				return next.Query(ctx, q)
			}

			tenantID := i.tenantIDFunc(ctx)
			if tenantID == "" {
				return next.Query(ctx, q)
			}

			switch q := q.(type) {
			case interface {
				WhereP(...func(*sql.Selector))
			}:
				q.WhereP(
					sql.FieldEQ("tenant_id", tenantID),
				)
			default:
				return nil, fmt.Errorf("tenant interceptor: unsupported query type %T", q)
			}

			return next.Query(ctx, q)
		})
	}
}

type SkipKey struct{}

func SkipContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, SkipKey{}, true)
}

func ShouldSkip(ctx context.Context) bool {
	v, ok := ctx.Value(SkipKey{}).(bool)
	return ok && v
}
