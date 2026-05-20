/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package server

import (
	"net/http"

	ginhttp "github.com/gin-gonic/gin"
	ginadapter "origadmin/application/origstudio/internal/pkg/http/gin"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/infra/auth"
	authbiz "origadmin/application/origstudio/internal/features/auth/biz"
)

// GetClaims retrieves claims from the gin.Context.
func GetClaims(c *ginhttp.Context) (*auth.Claims, bool) {
	if val, exists := c.Get("claims"); exists {
		if claims, ok := val.(*auth.Claims); ok {
			return claims, true
		}
	}
	return nil, false
}

// ==================== http.HandlerFunc wrappers (legacy) ====================

// WithJWT wraps an http.HandlerFunc with JWT middleware.
// It retrieves the real gin.Context from the request context,
// runs JWT validation on it, and proceeds only if the token is valid.
func WithJWT(jwtMgr *auth.Manager, h http.HandlerFunc) http.HandlerFunc {
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
		h(w, r)
	}
}

// WithOptionalJWT wraps an http.HandlerFunc with optional JWT middleware.
// If a valid token is present, claims are set in the gin.Context.
// If no token is present or the token is invalid, the handler still proceeds.
func WithOptionalJWT(jwtMgr *auth.Manager, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		if gc == nil {
			h(w, r)
			return
		}
		header := gc.GetHeader("Authorization")
		if len(header) >= 8 && header[:7] == "Bearer " {
			if claims, err := jwtMgr.Parse(header[7:]); err == nil {
				gc.Set("claims", claims)
			}
		}
		if _, exists := gc.Get("claims"); !exists {
			if t := gc.Query("token"); t != "" {
				if claims, err := jwtMgr.Parse(t); err == nil {
					gc.Set("claims", claims)
				}
			}
		}
		h(w, r)
	}
}

// WithAdmin wraps an http.HandlerFunc with JWT + Admin middleware.
func WithAdmin(jwtMgr *auth.Manager, h http.HandlerFunc) http.HandlerFunc {
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
		h(w, r)
	}
}

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

// WithJWTCtx wraps an http2.HandlerFunc with JWT middleware.
func WithJWTCtx(jwtMgr *auth.Manager, h http2.HandlerFunc) http2.HandlerFunc {
	return JWTMiddlewareCtx(jwtMgr)(h)
}

// WithOptionalJWTCtx wraps an http2.HandlerFunc with optional JWT middleware.
// If a valid token is present, claims are set in the context.
// If no token is present or the token is invalid, the handler still proceeds.
func WithOptionalJWTCtx(jwtMgr *auth.Manager, h http2.HandlerFunc) http2.HandlerFunc {
	return OptionalJWTMiddlewareCtx(jwtMgr)(h)
}

// WithAdminCtx wraps an http2.HandlerFunc with JWT + Admin middleware.
func WithAdminCtx(jwtMgr *auth.Manager, h http2.HandlerFunc) http2.HandlerFunc {
	return http2.Chain(JWTMiddlewareCtx(jwtMgr), AdminMiddlewareCtx(jwtMgr))(h)
}

// WithAdminAndPermCtx wraps an http2.HandlerFunc with JWT + Admin + Permission middleware.
func WithAdminAndPermCtx(jwtMgr *auth.Manager, permChecker authbiz.PermissionChecker, permission string, h http2.HandlerFunc) http2.HandlerFunc {
	return http2.Chain(JWTMiddlewareCtx(jwtMgr), AdminMiddlewareCtx(jwtMgr), RequirePermissionCtx(permChecker, permission))(h)
}

// ==================== http2.MiddlewareFunc implementations ====================

// JWTMiddlewareCtx returns a MiddlewareFunc that validates JWT tokens.
// It supports two token sources:
//  1. Authorization header: "Bearer <token>" (standard, for regular API calls)
//  2. Query parameter: "?token=<token>" (fallback, for SSE/EventSource which
//     cannot set custom headers)
func JWTMiddlewareCtx(jwtMgr *auth.Manager) http2.MiddlewareFunc {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx http2.Context) error {
			tokenStr := extractTokenFromCtx(ctx)
			if tokenStr == "" {
				http2.Fail(ctx, http2.AppErrUnauthorized, "missing token")
				return nil
			}
			claims, err := jwtMgr.Parse(tokenStr)
			if err != nil {
				http2.Fail(ctx, http2.AppErrUnauthorized, "invalid token")
				return nil
			}
			ctx.Set("claims", claims)
			return next(ctx)
		}
	}
}

// OptionalJWTMiddlewareCtx returns a MiddlewareFunc that parses JWT token if
// present but does not require it. If a valid token is found, claims are set.
func OptionalJWTMiddlewareCtx(jwtMgr *auth.Manager) http2.MiddlewareFunc {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx http2.Context) error {
			tokenStr := extractTokenFromCtx(ctx)
			if tokenStr != "" {
				if claims, err := jwtMgr.Parse(tokenStr); err == nil {
					ctx.Set("claims", claims)
				}
			}
			return next(ctx)
		}
	}
}

// AdminMiddlewareCtx returns a MiddlewareFunc that requires admin (or staff) role.
// It expects claims to already be set in the context (typically by JWTMiddlewareCtx).
func AdminMiddlewareCtx(jwtMgr *auth.Manager) http2.MiddlewareFunc {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx http2.Context) error {
			claims, ok := GetClaimsCtx(ctx)
			if !ok {
				http2.Fail(ctx, http2.AppErrUnauthorized, "no claims")
				return nil
			}
			if claims.Role != "admin" && !claims.IsStaff {
				http2.Fail(ctx, http2.AppErrForbidden, "admin required")
				return nil
			}
			return next(ctx)
		}
	}
}

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

			if claims.Role == "admin" || claims.IsStaff {
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

// extractTokenFromCtx extracts the JWT token from an http2.Context.
// It checks the Authorization header first, then falls back to the
// "token" query parameter (for SSE/EventSource connections).
func extractTokenFromCtx(ctx http2.Context) string {
	auth := ctx.GetHeader("Authorization")
	if len(auth) >= 8 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	if t := ctx.QueryVar("token"); t != "" {
		return t
	}
	return ""
}

// JWTMiddleware validates Bearer token and injects claims into context.
// It supports two token sources:
//  1. Authorization header: "Bearer <token>" (standard, for regular API calls)
//  2. Query parameter: "?token=<token>" (fallback, for SSE/EventSource which
//     cannot set custom headers)
func JWTMiddleware(jwtMgr *auth.Manager) ginhttp.HandlerFunc {
	return func(c *ginhttp.Context) {
		var tokenStr string

		// Try Authorization header first
		header := c.GetHeader("Authorization")
		if len(header) >= 8 && header[:7] == "Bearer " {
			tokenStr = header[7:]
		}

		// Fallback to query parameter (for SSE/EventSource connections)
		if tokenStr == "" {
			if t := c.Query("token"); t != "" {
				tokenStr = t
			}
		}

		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ginhttp.H{"error": "missing or invalid Authorization header"})
			return
		}
		claims, err := jwtMgr.Parse(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ginhttp.H{"error": "invalid token: " + err.Error()})
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}

// OptionalJWTMiddleware parses Bearer token if present but does not require it.
func OptionalJWTMiddleware(jwtMgr *auth.Manager) ginhttp.HandlerFunc {
	return func(c *ginhttp.Context) {
		header := c.GetHeader("Authorization")
		if len(header) >= 8 && header[:7] == "Bearer " {
			if claims, err := jwtMgr.Parse(header[7:]); err == nil {
				c.Set("claims", claims)
			}
		}
		c.Next()
	}
}

// RequiredRole requires authenticated user to have a specific role.
func RequiredRole(role string) ginhttp.HandlerFunc {
	return func(c *ginhttp.Context) {
		claims, ok := GetClaims(c)
		if !ok || (claims.Role != role && !claims.IsStaff && claims.Role != "admin") {
			c.AbortWithStatusJSON(http.StatusForbidden, ginhttp.H{"error": "permission denied: " + role + " role required"})
			return
		}
		c.Next()
	}
}

// AdminMiddleware requires JWT + admin (or staff) role.
func AdminMiddleware(jwtMgr *auth.Manager) ginhttp.HandlerFunc {
	return func(c *ginhttp.Context) {
		claims, ok := GetClaims(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ginhttp.H{"error": "no claims in context"})
			return
		}

		if claims.Role != "admin" && !claims.IsStaff {
			c.AbortWithStatusJSON(http.StatusForbidden, ginhttp.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}

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

		if claims.Role == "admin" || claims.IsStaff {
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
