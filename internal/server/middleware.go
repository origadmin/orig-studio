/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/handler"
	authbiz "origadmin/application/origcms/internal/svc-auth/biz"
)

func GetClaims(c *gin.Context) (*auth.Claims, bool) {
	if val, exists := c.Get("claims"); exists {
		if claims, ok := val.(*auth.Claims); ok {
			return claims, true
		}
	}
	if claims, ok := claimsFromContext(c.Request.Context()); ok {
		return claims, true
	}
	return nil, false
}

// GinMiddlewareAdapter adapts a gin.HandlerFunc to http.HandlerFunc with shared context
// The middleware and handler share the same gin.Context
func GinMiddlewareAdapter(middleware gin.HandlerFunc, h gin.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		if params := handler.GetGinParams(r); params != nil {
			c.Params = params
		}

		middleware(c)

		if !c.IsAborted() {
			h(c)
		}
	}
}

// GinHandlerToHTTP converts a gin.HandlerFunc to http.HandlerFunc
func GinHandlerToHTTP(h gin.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		if params := handler.GetGinParams(r); params != nil {
			c.Params = params
		}
		h(c)
	}
}

// WithJWT wraps a handler with JWT middleware
// Supports both gin.HandlerFunc and http.HandlerFunc
func WithJWT(jwtMgr *auth.Manager, h interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		if params := handler.GetGinParams(r); params != nil {
			c.Params = params
		}

		JWTMiddleware(jwtMgr)(c)
		if c.IsAborted() {
			return
		}

		if claimsVal, exists := c.Get("claims"); exists {
			r = r.WithContext(context.WithValue(r.Context(), claimsKey, claimsVal))
			r = handler.SetClaimsInContext(r, claimsVal)
		}

		switch h := h.(type) {
		case gin.HandlerFunc:
			h(c)
		case http.HandlerFunc:
			h(w, r)
		case func(http.ResponseWriter, *http.Request):
			h(w, r)
		}
	}
}

// WithAdmin wraps a gin.HandlerFunc with JWT + Admin middleware
func WithAdmin(jwtMgr *auth.Manager, h gin.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		if params := handler.GetGinParams(r); params != nil {
			c.Params = params
		}

		JWTMiddleware(jwtMgr)(c)
		if c.IsAborted() {
			return
		}

		AdminMiddleware(jwtMgr)(c)
		if c.IsAborted() {
			return
		}

		if claimsVal, exists := c.Get("claims"); exists {
			r = r.WithContext(context.WithValue(r.Context(), claimsKey, claimsVal))
			r = handler.SetClaimsInContext(r, claimsVal)
		}

		h(c)
	}
}

// WithAdminAndPerm wraps a gin.HandlerFunc with JWT + Admin + Permission middleware
func WithAdminAndPerm(jwtMgr *auth.Manager, permChecker authbiz.PermissionChecker, permission string, h gin.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		if params := handler.GetGinParams(r); params != nil {
			c.Params = params
		}

		JWTMiddleware(jwtMgr)(c)
		if c.IsAborted() {
			return
		}

		AdminMiddleware(jwtMgr)(c)
		if c.IsAborted() {
			return
		}

		RequirePermission(permChecker, permission)(c)
		if c.IsAborted() {
			return
		}

		if claimsVal, exists := c.Get("claims"); exists {
			r = r.WithContext(context.WithValue(r.Context(), claimsKey, claimsVal))
			r = handler.SetClaimsInContext(r, claimsVal)
		}

		h(c)
	}
}

// JWTMiddleware validates Bearer token and injects claims into context.
func JWTMiddleware(jwtMgr *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if len(header) < 8 || header[:7] != "Bearer " {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			return
		}
		claims, err := jwtMgr.Parse(header[7:])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token: " + err.Error()})
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}

// OptionalJWTMiddleware parses Bearer token if present but does not require it.
func OptionalJWTMiddleware(jwtMgr *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
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
func RequiredRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := GetClaims(c)
		if !ok || (claims.Role != role && !claims.IsStaff && claims.Role != "admin") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "permission denied: " + role + " role required"})
			return
		}
		c.Next()
	}
}

// AdminMiddleware requires JWT + admin (or staff) role.
func AdminMiddleware(jwtMgr *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := GetClaims(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no claims in context"})
			return
		}

		if claims.Role != "admin" && !claims.IsStaff {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}

type permissionConfig struct {
	ownershipExtractor func(*gin.Context) (string, error)
	categoryExtractor  func(*gin.Context) (string, error)
}

type PermissionOption func(*permissionConfig)

func WithOwnershipCheck(extractor func(*gin.Context) (string, error)) PermissionOption {
	return func(c *permissionConfig) {
		c.ownershipExtractor = extractor
	}
}

func WithResourceCategory(extractor func(*gin.Context) (string, error)) PermissionOption {
	return func(c *permissionConfig) {
		c.categoryExtractor = extractor
	}
}

func RequirePermission(permChecker authbiz.PermissionChecker, permission string, opts ...PermissionOption) gin.HandlerFunc {
	cfg := &permissionConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	return func(c *gin.Context) {
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