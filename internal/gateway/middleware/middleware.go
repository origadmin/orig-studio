/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package middleware provides HTTP middleware for the API gateway.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/origadmin/runtime/log"
	"origadmin/application/origcms/internal/auth"
)

type contextKey string

const ClaimsKey contextKey = "claims"

// Logger is a simple HTTP request logging middleware.
func Logger(logger log.Logger) func(http.Handler) http.Handler {
	helper := log.NewHelper(log.With(logger, "module", "gateway.middleware"))
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			helper.Infof("%s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}

// CORS adds permissive CORS headers for development.
func CORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// JWTAuth validates the Authorization header Bearer token using auth.Manager.
func JWTAuth(mgr *auth.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"error":"unauthorized: missing or invalid header"}`, http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := mgr.Parse(tokenStr)
			if err != nil {
				http.Error(w, `{"error":"unauthorized: `+err.Error()+`"}`, http.StatusUnauthorized)
				return
			}

			// Inject claims into context
			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
