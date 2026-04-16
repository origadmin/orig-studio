/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/auth"
)

// GinMiddlewareAdapter adapts a gin.HandlerFunc to http.HandlerFunc
func GinMiddlewareAdapter(middleware gin.HandlerFunc, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create a gin context
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		
		// Call the middleware first
		middleware(c)
		
		// If the middleware didn't abort, call the next handler
		if !c.IsAborted() {
			next(w, r)
		}
	}
}

// WithJWT wraps a handler with JWT middleware
func WithJWT(jwtMgr *auth.Manager, handler http.HandlerFunc) http.HandlerFunc {
	return GinMiddlewareAdapter(JWTMiddleware(jwtMgr), handler)
}

// GinHandlerToHTTP converts a gin.HandlerFunc to http.HandlerFunc
func GinHandlerToHTTP(handler gin.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		handler(c)
	}
}

// WithAdmin wraps a handler with JWT + Admin middleware
func WithAdmin(jwtMgr *auth.Manager, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		
		// First run JWT middleware
		JWTMiddleware(jwtMgr)(c)
		if c.IsAborted() {
			return
		}
		
		// Then run Admin middleware
		AdminMiddleware(jwtMgr)(c)
		if c.IsAborted() {
			return
		}
		
		// Call the actual handler
		handler(w, r)
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
		// JWTMiddleware should have been run already if using separate groups
		// but for safety we can check here or assume it's in the chain.
		claims, ok := c.Get("claims")
		if !ok {
			// Try to run it inline if not already run
			JWTMiddleware(jwtMgr)(c)
			if c.IsAborted() {
				return
			}
			claims, _ = c.Get("claims")
		}

		cl := claims.(*auth.Claims)
		if cl.Role != "admin" && !cl.IsStaff {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}
