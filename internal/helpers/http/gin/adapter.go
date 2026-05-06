package gin

import (
	"context"
	"net/http"

	ginhttp "github.com/gin-gonic/gin"
)

type contextKey string

const ginCtxKey contextKey = "gin_ctx"

// SetGinContext stores the *gin.Context in the request context.
func SetGinContext(r *http.Request, c *ginhttp.Context) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), ginCtxKey, c))
}

// GetGinContext retrieves the *gin.Context from the request context.
func GetGinContext(r *http.Request) *ginhttp.Context {
	if c, ok := r.Context().Value(ginCtxKey).(*ginhttp.Context); ok {
		return c
	}
	return nil
}

// StdRouterAdapter adapts a gin.RouterGroup to accept http.HandlerFunc handlers.
type StdRouterAdapter struct {
	group *ginhttp.RouterGroup
}

func NewStdRouterAdapter(group *ginhttp.RouterGroup) *StdRouterAdapter {
	return &StdRouterAdapter{group: group}
}

func (a *StdRouterAdapter) Group(prefix string) *StdRouterAdapter {
	return &StdRouterAdapter{group: a.group.Group(prefix)}
}

// wrapHandler wraps an http.HandlerFunc to pass the gin.Context through request context.
func wrapHandler(h http.HandlerFunc) ginhttp.HandlerFunc {
	return func(c *ginhttp.Context) {
		r := SetGinContext(c.Request, c)
		h(c.Writer, r)
	}
}

func (a *StdRouterAdapter) GET(path string, handler http.HandlerFunc) {
	a.group.GET(path, wrapHandler(handler))
}

func (a *StdRouterAdapter) POST(path string, handler http.HandlerFunc) {
	a.group.POST(path, wrapHandler(handler))
}

func (a *StdRouterAdapter) PUT(path string, handler http.HandlerFunc) {
	a.group.PUT(path, wrapHandler(handler))
}

func (a *StdRouterAdapter) DELETE(path string, handler http.HandlerFunc) {
	a.group.DELETE(path, wrapHandler(handler))
}

func (a *StdRouterAdapter) PATCH(path string, handler http.HandlerFunc) {
	a.group.PATCH(path, wrapHandler(handler))
}
