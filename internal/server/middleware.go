/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package server

import (
	stdcontext "context"
	"fmt"
	stdhttp "net/http"
	"time"

	ginhttp "github.com/gin-gonic/gin"
	"origadmin/application/origstudio/internal/infra/auth"
	ginadapter "origadmin/application/origstudio/internal/pkg/http/gin"
	http2 "origadmin/application/origstudio/internal/pkg/http"
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

// ==================== http.HandlerFunc wrappers (legacy, Gin-based) ====================

// WithJWT wraps an stdhttp.HandlerFunc with JWT middleware.
// It retrieves the real gin.Context from the request context,
// runs JWT validation on it, and proceeds only if the token is valid.
func WithJWT(jwtMgr *auth.Manager, h stdhttp.HandlerFunc) stdhttp.HandlerFunc {
	return func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		gc := ginadapter.GetGinContext(r)
		if gc == nil {
			stdhttp.Error(w, "internal error", stdhttp.StatusInternalServerError)
			return
		}
		JWTMiddleware(jwtMgr)(gc)
		if gc.IsAborted() {
			return
		}
		h(w, r)
	}
}

// WithOptionalJWT wraps an stdhttp.HandlerFunc with optional JWT middleware.
func WithOptionalJWT(jwtMgr *auth.Manager, h stdhttp.HandlerFunc) stdhttp.HandlerFunc {
	return func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
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

// WithAdmin wraps an stdhttp.HandlerFunc with JWT + Admin middleware.
func WithAdmin(jwtMgr *auth.Manager, h stdhttp.HandlerFunc) stdhttp.HandlerFunc {
	return func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		gc := ginadapter.GetGinContext(r)
		if gc == nil {
			stdhttp.Error(w, "internal error", stdhttp.StatusInternalServerError)
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

// ==================== http2.HandlerFunc wrappers (adapter-based) ====================

// WithJWTCtx wraps an http2.HandlerFunc with JWT middleware.
func WithJWTCtx(jwtMgr *auth.Manager, h http2.HandlerFunc) http2.HandlerFunc {
	return JWTMiddlewareCtx(jwtMgr)(h)
}

// WithOptionalJWTCtx wraps an http2.HandlerFunc with optional JWT middleware.
func WithOptionalJWTCtx(jwtMgr *auth.Manager, h http2.HandlerFunc) http2.HandlerFunc {
	return OptionalJWTMiddlewareCtx(jwtMgr)(h)
}

// WithAdminCtx wraps an http2.HandlerFunc with JWT + Admin middleware.
func WithAdminCtx(jwtMgr *auth.Manager, h http2.HandlerFunc) http2.HandlerFunc {
	return http2.Chain(JWTMiddlewareCtx(jwtMgr), AdminMiddlewareCtx(jwtMgr))(h)
}

// ==================== Standard net/http Handler wrappers (NO Gin dependency) ====================

type stdClaimsContextKey struct{}

// StdWithJWT wraps a standard stdhttp.Handler with JWT validation middleware.
// It does NOT depend on Gin - extracts token directly from the request.
// Token sources: Authorization header "Bearer <token>" or query "?token=<token>".
// Valid claims are stored in the request context under a private key.
func StdWithJWT(jwtMgr *auth.Manager, h stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		tokenStr := extractStdToken(r)
		if tokenStr == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(stdhttp.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"missing token"}`))
			return
		}
		claims, err := jwtMgr.Parse(tokenStr)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(stdhttp.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid token"}`))
			return
		}
		ctx := stdcontext.WithValue(r.Context(), stdClaimsContextKey{}, claims)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// StdWithAdmin wraps a standard stdhttp.Handler with JWT + admin role validation.
func StdWithAdmin(jwtMgr *auth.Manager, h stdhttp.Handler) stdhttp.Handler {
	return StdWithJWT(jwtMgr, stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		claims, ok := GetClaimsFromStdCtx(r.Context())
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(stdhttp.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"no claims"}`))
			return
		}
		if claims.Role != "admin" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(stdhttp.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"admin required"}`))
			return
		}
		h.ServeHTTP(w, r)
	}))
}

// GetClaimsFromStdCtx retrieves auth.Claims from a standard context.Context.
func GetClaimsFromStdCtx(ctx stdcontext.Context) (*auth.Claims, bool) {
	val := ctx.Value(stdClaimsContextKey{})
	if val == nil {
		return nil, false
	}
	claims, ok := val.(*auth.Claims)
	return claims, ok
}

func extractStdToken(r *stdhttp.Request) string {
	header := r.Header.Get("Authorization")
	if len(header) >= 8 && header[:7] == "Bearer " {
		return header[7:]
	}
	if t := r.URL.Query().Get("token"); t != "" {
		return t
	}
	return ""
}

// ==================== http2.MiddlewareFunc implementations ====================

// JWTMiddlewareCtx returns a MiddlewareFunc that validates JWT tokens.
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

// OptionalJWTMiddlewareCtx returns a MiddlewareFunc that parses JWT token if present.
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

// AdminMiddlewareCtx returns a MiddlewareFunc that requires admin role.
func AdminMiddlewareCtx(jwtMgr *auth.Manager) http2.MiddlewareFunc {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx http2.Context) error {
			claims, ok := GetClaimsCtx(ctx)
			if !ok {
				http2.Fail(ctx, http2.AppErrUnauthorized, "no claims")
				return nil
			}
			if claims.Role != "admin" {
				http2.Fail(ctx, http2.AppErrForbidden, "admin required")
				return nil
			}
			return next(ctx)
		}
	}
}

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

// ==================== Gin middleware implementations ====================

// JWTMiddleware validates Bearer token and injects claims into context.
func JWTMiddleware(jwtMgr *auth.Manager) ginhttp.HandlerFunc {
	return func(c *ginhttp.Context) {
		var tokenStr string

		header := c.GetHeader("Authorization")
		if len(header) >= 8 && header[:7] == "Bearer " {
			tokenStr = header[7:]
		}

		if tokenStr == "" {
			if t := c.Query("token"); t != "" {
				tokenStr = t
			}
		}

		if tokenStr == "" {
			c.AbortWithStatusJSON(stdhttp.StatusUnauthorized, ginhttp.H{"error": "missing or invalid Authorization header"})
			return
		}
		claims, err := jwtMgr.Parse(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(stdhttp.StatusUnauthorized, ginhttp.H{"error": "invalid token: " + err.Error()})
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
		if !ok || (claims.Role != role && claims.Role != "admin") {
			c.AbortWithStatusJSON(stdhttp.StatusForbidden, ginhttp.H{"error": "permission denied: " + role + " role required"})
			return
		}
		c.Next()
	}
}

// AdminMiddleware requires JWT + admin role.
func AdminMiddleware(jwtMgr *auth.Manager) ginhttp.HandlerFunc {
	return func(c *ginhttp.Context) {
		claims, ok := GetClaims(c)
		if !ok {
			c.AbortWithStatusJSON(stdhttp.StatusUnauthorized, ginhttp.H{"error": "no claims in context"})
			return
		}

		if claims.Role != "admin" {
			c.AbortWithStatusJSON(stdhttp.StatusForbidden, ginhttp.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}

// ==================== Base http2.MiddlewareFunc (framework-agnostic) ====================

// RequestIDCtx returns a MiddlewareFunc that generates or propagates an
// X-Request-ID header. If the incoming request already carries one, it is
// preserved; otherwise a new UUID-like ID is generated from timestamp.
func RequestIDCtx() http2.MiddlewareFunc {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx http2.Context) error {
			rid := ctx.GetHeader("X-Request-ID")
			if rid == "" {
				rid = fmt.Sprintf("req-%d", time.Now().UnixNano())
			}
			ctx.Set("request_id", rid)
			ctx.Response().Header().Set("X-Request-ID", rid)
			return next(ctx)
		}
	}
}

// RecoveryCtx returns a MiddlewareFunc that recovers from panics, logs the
// error, and returns a 500 response. This is the framework-agnostic
// counterpart of gin.Recovery().
func RecoveryCtx() http2.MiddlewareFunc {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx http2.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					http2.Fail(ctx, http2.AppErrInternal, "internal server error")
					err = nil
				}
			}()
			return next(ctx)
		}
	}
}

// CORSCtx returns a MiddlewareFunc that adds CORS headers to the response.
// This is the framework-agnostic counterpart of rs/cors handlers.
func CORSCtx(allowedOrigins []string) http2.MiddlewareFunc {
	originSet := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[o] = true
	}
	allowAll := len(allowedOrigins) == 0

	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx http2.Context) error {
			origin := ctx.GetHeader("Origin")
			if origin != "" && (allowAll || originSet[origin]) {
				ctx.Response().Header().Set("Access-Control-Allow-Origin", origin)
				ctx.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
				ctx.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
				ctx.Response().Header().Set("Access-Control-Allow-Credentials", "true")
				ctx.Response().Header().Set("Access-Control-Max-Age", "86400")
			}
			// Handle preflight
			if ctx.Request().Method == stdhttp.MethodOptions {
				ctx.Response().WriteHeader(stdhttp.StatusNoContent)
				return nil
			}
			return next(ctx)
		}
	}
}
