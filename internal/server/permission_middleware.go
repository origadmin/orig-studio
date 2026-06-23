package server

import (
	"net/http"

	ginhttp "github.com/gin-gonic/gin"
	ginadapter "origadmin/application/origstudio/internal/pkg/http/gin"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/infra/auth"
	authbiz "origadmin/application/origstudio/internal/features/auth/biz"
)

// ==================== http.HandlerFunc wrappers (legacy) ====================

// WithAdminAndPerm wraps an http.HandlerFunc with JWT + Admin + Permission middleware.
func WithAdminAndPerm(jwtMgr *auth.Manager, permChecker authbiz.PermissionChecker, permission string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		if gc == nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		JWTMiddleware(jwtMgr)(gc)
		if gc.IsAborted() {
			return
		}
		AdminMiddleware(jwtMgr)(gc)
		if gc.IsAborted() {
			return
		}
		RequirePermission(permChecker, permission)(gc)
		if gc.IsAborted() {
			return
		}
		h(w, r)
	}
}

// ==================== http2.HandlerFunc wrappers (new) ====================

// WithAdminAndPermCtx wraps an http2.HandlerFunc with JWT + Admin + Permission middleware.
func WithAdminAndPermCtx(jwtMgr *auth.Manager, permChecker authbiz.PermissionChecker, permission string, h http2.HandlerFunc) http2.HandlerFunc {
	return http2.Chain(JWTMiddlewareCtx(jwtMgr), AdminMiddlewareCtx(jwtMgr), RequirePermissionCtx(permChecker, permission))(h)
}

// ==================== http2.MiddlewareFunc implementations ====================

// RequirePermissionCtx returns a MiddlewareFunc that checks if the authenticated
// user has the specified permission.
func RequirePermissionCtx(permChecker authbiz.PermissionChecker, permission string) http2.MiddlewareFunc {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx http2.Context) error {
			claims, ok := GetClaimsCtx(ctx)
			if !ok {
				http2.Fail(ctx, http2.AppErrUnauthorized, "authentication required")
				return nil
			}

			if claims.Role == "admin" {
				return next(ctx)
			}

			userID := claims.GetUserID()
			allowed, err := permChecker.CheckPermission(ctx.Request().Context(), userID, permission, "")
			if err != nil || !allowed {
				http2.Fail(ctx, http2.AppErrForbidden, "permission denied")
				return nil
			}
			return next(ctx)
		}
	}
}

// ==================== gin.HandlerFunc implementations ====================

type permissionConfig struct {
	ownershipExtractor func(*ginhttp.Context) (string, error)
	categoryExtractor  func(*ginhttp.Context) (string, error)
}

type PermissionOption func(*permissionConfig)

func WithOwnershipCheck(extractor func(*ginhttp.Context) (string, error)) PermissionOption {
	return func(c *permissionConfig) {
		c.ownershipExtractor = extractor
	}
}

func WithResourceCategory(extractor func(*ginhttp.Context) (string, error)) PermissionOption {
	return func(c *permissionConfig) {
		c.categoryExtractor = extractor
	}
}

func RequirePermission(permChecker authbiz.PermissionChecker, permission string, opts ...PermissionOption) ginhttp.HandlerFunc {
	cfg := &permissionConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	return func(c *ginhttp.Context) {
		claims, ok := GetClaims(c)
		if !ok {
			Fail(c, ErrUnauthorized, "authentication required")
			c.Abort()
			return
		}

		if claims.Role == "admin" {
			c.Next()
			return
		}

		userID := claims.GetUserID()

		categoryID := ""
		if cfg.categoryExtractor != nil {
			if catID, err := cfg.categoryExtractor(c); err == nil && catID != "" {
				categoryID = catID
			}
		}

		allowed, err := permChecker.CheckPermission(c.Request.Context(), userID, permission, categoryID)
		if err == nil && allowed {
			c.Next()
			return
		}

		if cfg.ownershipExtractor != nil {
			if ownerID, err := cfg.ownershipExtractor(c); err == nil && ownerID == userID {
				c.Next()
				return
			}
		}

		Fail(c, ErrForbidden, "insufficient permissions")
		c.Abort()
	}
}