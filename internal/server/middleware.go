/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/auth"
)

// GinMiddlewareAdapter adapts a gin.HandlerFunc to http.HandlerFunc with shared context
// The middleware and handler share the same gin.Context
func GinMiddlewareAdapter(middleware gin.HandlerFunc, handler gin.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, _ := gin.CreateTestContext(w)
		c.Request = r

		middleware(c)

		if !c.IsAborted() {
			handler(c)
		}
	}
}

// GinHandlerToHTTP converts a gin.HandlerFunc to http.HandlerFunc
func GinHandlerToHTTP(handler gin.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		handler(c)
	}
}

// WithJWT wraps a handler with JWT middleware
// Supports both gin.HandlerFunc and http.HandlerFunc
func WithJWT(jwtMgr *auth.Manager, handler interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, _ := gin.CreateTestContext(w)
		c.Request = r

		JWTMiddleware(jwtMgr)(c)
		if c.IsAborted() {
			return
		}

		switch h := handler.(type) {
		case gin.HandlerFunc:
			h(c)
		case http.HandlerFunc:
			h(w, r)
		}
	}
}

// WithAdmin wraps a gin.HandlerFunc with JWT + Admin middleware
func WithAdmin(jwtMgr *auth.Manager, handler gin.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, _ := gin.CreateTestContext(w)
		c.Request = r

		JWTMiddleware(jwtMgr)(c)
		if c.IsAborted() {
			return
		}

		AdminMiddleware(jwtMgr)(c)
		if c.IsAborted() {
			return
		}

		handler(c)
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

// RequiredRole requires authenticated user to have a specific role.
func RequiredRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := c.MustGet("claims").(*auth.Claims)
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
		claims, ok := c.Get("claims")
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no claims in context"})
			return
		}

		cl := claims.(*auth.Claims)
		if cl.Role != "admin" && !cl.IsStaff {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}