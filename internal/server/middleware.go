/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package server

import (
	"net/http"

	ginhttp "github.com/gin-gonic/gin"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/infra/auth"
	authbiz "origadmin/application/origcms/internal/features/auth/biz"
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
