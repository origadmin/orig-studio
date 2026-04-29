/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package handler provides adapters for different HTTP frameworks.
package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GinRouterAdapter adapts a gin.RouterGroup to the Router interface.
type GinRouterAdapter struct {
	group *gin.RouterGroup
}

// NewGinRouterAdapter creates a new GinRouterAdapter.
func NewGinRouterAdapter(group *gin.RouterGroup) *GinRouterAdapter {
	return &GinRouterAdapter{group: group}
}

// Group creates a new route group with the given prefix.
func (a *GinRouterAdapter) Group(prefix string) Router {
	return &GinRouterAdapter{group: a.group.Group(prefix)}
}

type contextKey string

const ginParamsKey contextKey = "gin_params"
const claimsContextKey contextKey = "claims"

// GetGinParams retrieves gin.Params from the request context.
func GetGinParams(r *http.Request) gin.Params {
	if params, ok := r.Context().Value(ginParamsKey).(gin.Params); ok {
		return params
	}
	return nil
}

// wrapWithParams wraps an http.HandlerFunc to store gin.Params in the request context.
func wrapWithParams(h http.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), ginParamsKey, c.Params)
		c.Request = c.Request.WithContext(ctx)
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// GET registers a GET route.
func (a *GinRouterAdapter) GET(path string, handler http.HandlerFunc) {
	a.group.GET(path, wrapWithParams(handler))
}

// POST registers a POST route.
func (a *GinRouterAdapter) POST(path string, handler http.HandlerFunc) {
	a.group.POST(path, wrapWithParams(handler))
}

// PUT registers a PUT route.
func (a *GinRouterAdapter) PUT(path string, handler http.HandlerFunc) {
	a.group.PUT(path, wrapWithParams(handler))
}

// DELETE registers a DELETE route.
func (a *GinRouterAdapter) DELETE(path string, handler http.HandlerFunc) {
	a.group.DELETE(path, wrapWithParams(handler))
}

// PATCH registers a PATCH route.
func (a *GinRouterAdapter) PATCH(path string, handler http.HandlerFunc) {
	a.group.PATCH(path, wrapWithParams(handler))
}

// GinContextAdapter adapts a gin.Context to the Context interface.
type GinContextAdapter struct {
	c *gin.Context
}

// NewGinContextAdapter creates a new GinContextAdapter.
func NewGinContextAdapter(c *gin.Context) *GinContextAdapter {
	return &GinContextAdapter{c: c}
}

// NewGinContextAdapterFromHTTP creates a new GinContextAdapter from http.ResponseWriter and http.Request.
func NewGinContextAdapterFromHTTP(w http.ResponseWriter, r *http.Request) *GinContextAdapter {
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	if params := GetGinParams(r); params != nil {
		c.Params = params
	}
	// Copy claims from request context to gin context so that c.Get("claims") works
	// in handlers that use this adapter. The WithJWT middleware stores claims in
	// the request context using SetClaimsInContext.
	if claims := r.Context().Value(claimsContextKey); claims != nil {
		c.Set("claims", claims)
	}
	return &GinContextAdapter{c: c}
}

// SetClaimsInContext stores claims in the request context so that
// NewGinContextAdapterFromHTTP can copy them to the gin context.
func SetClaimsInContext(r *http.Request, claims interface{}) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), claimsContextKey, claims))
}

// Context returns the underlying context.Context.
func (a *GinContextAdapter) Context() context.Context {
	return a.c.Request.Context()
}

// Request returns the underlying http.Request.
func (a *GinContextAdapter) Request() *http.Request {
	return a.c.Request
}

// ResponseWriter returns the underlying http.ResponseWriter.
func (a *GinContextAdapter) ResponseWriter() http.ResponseWriter {
	return a.c.Writer
}

// Param gets a URL parameter.
func (a *GinContextAdapter) Param(key string) string {
	return a.c.Param(key)
}

// Query gets a query parameter.
func (a *GinContextAdapter) Query(key string) string {
	return a.c.Query(key)
}

// Bind binds the request body to a struct.
func (a *GinContextAdapter) Bind(v interface{}) error {
	return a.c.ShouldBind(v)
}

// JSON sends a JSON response.
func (a *GinContextAdapter) JSON(code int, v interface{}) {
	a.c.JSON(code, v)
}

// Status sends a status code response.
func (a *GinContextAdapter) Status(code int) {
	a.c.Status(code)
}

// Header sets a response header.
func (a *GinContextAdapter) Header(key, value string) {
	a.c.Header(key, value)
}

// Get gets a value from the context.
func (a *GinContextAdapter) Get(key string) interface{} {
	val, _ := a.c.Get(key)
	return val
}

// Set sets a value in the context.
func (a *GinContextAdapter) Set(key string, value interface{}) {
	a.c.Set(key, value)
}

// GinContext returns the underlying *gin.Context.
// This allows server response helpers (server.OK, server.Fail, etc.) to work
// with handlers that use GinContextAdapter.
func (a *GinContextAdapter) GinContext() *gin.Context {
	return a.c
}
