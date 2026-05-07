package gin

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	ginhttp "github.com/gin-gonic/gin"

	http2 "origadmin/application/origcms/internal/helpers/http"
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

// ==================== RouterAdapter ====================
// RouterAdapter implements the http2.Router interface by adapting
// a *gin.RouterGroup. It is the central piece of the migration:
// all handlers register routes via http2.Router, and only this
// adapter knows about gin.

type RouterAdapter struct {
	group   *ginhttp.RouterGroup
	filters []http2.FilterFunc
}

var _ http2.Router = (*RouterAdapter)(nil)

// NewRouterAdapter creates a new RouterAdapter wrapping the given gin.RouterGroup.
func NewRouterAdapter(group *ginhttp.RouterGroup) *RouterAdapter {
	return &RouterAdapter{group: group}
}

// Group returns a new RouterAdapter for the given prefix, inheriting
// any filters from the parent and appending the new ones.
func (a *RouterAdapter) Group(prefix string, filters ...http2.FilterFunc) http2.Router {
	return &RouterAdapter{
		group:   a.group.Group(prefix),
		filters: append(a.filters, filters...),
	}
}

// Use adds gin middleware to the underlying RouterGroup.
// This is a convenience method for gin middleware that cannot be
// expressed as http2.FilterFunc (e.g., middleware using c.Next()).
// Handlers should type-assert the http2.Router to *RouterAdapter
// when they need to add gin middleware.
func (a *RouterAdapter) Use(mw ...ginhttp.HandlerFunc) {
	a.group.Use(mw...)
}

func (a *RouterAdapter) GET(path string, h http2.HandlerFunc, filters ...http2.FilterFunc) {
	a.register(http.MethodGet, path, h, filters)
}

func (a *RouterAdapter) POST(path string, h http2.HandlerFunc, filters ...http2.FilterFunc) {
	a.register(http.MethodPost, path, h, filters)
}

func (a *RouterAdapter) PUT(path string, h http2.HandlerFunc, filters ...http2.FilterFunc) {
	a.register(http.MethodPut, path, h, filters)
}

func (a *RouterAdapter) DELETE(path string, h http2.HandlerFunc, filters ...http2.FilterFunc) {
	a.register(http.MethodDelete, path, h, filters)
}

func (a *RouterAdapter) PATCH(path string, h http2.HandlerFunc, filters ...http2.FilterFunc) {
	a.register(http.MethodPatch, path, h, filters)
}

func (a *RouterAdapter) register(method, relativePath string, h http2.HandlerFunc, filters []http2.FilterFunc) {
	handler := a.wrapHandler(h)
	allFilters := append(a.filters, filters...)
	if len(allFilters) > 0 {
		handler = a.applyFilters(handler, allFilters)
	}
	switch method {
	case http.MethodGet:
		a.group.GET(relativePath, handler)
	case http.MethodPost:
		a.group.POST(relativePath, handler)
	case http.MethodPut:
		a.group.PUT(relativePath, handler)
	case http.MethodDelete:
		a.group.DELETE(relativePath, handler)
	case http.MethodPatch:
		a.group.PATCH(relativePath, handler)
	}
}

func (a *RouterAdapter) wrapHandler(h http2.HandlerFunc) ginhttp.HandlerFunc {
	return func(c *ginhttp.Context) {
		ctx := &contextWrapper{ginCtx: c}
		if err := h(ctx); err != nil {
			handleError(c, err)
		}
	}
}

func (a *RouterAdapter) applyFilters(handler ginhttp.HandlerFunc, filters []http2.FilterFunc) ginhttp.HandlerFunc {
	// Build the http.Handler chain from filters.
	// The innermost handler converts the gin.HandlerFunc into an http.Handler
	// that retrieves the gin.Context from request context.
	h := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gc := GetGinContext(r)
		if gc != nil {
			handler(gc)
		}
	}))
	for i := len(filters) - 1; i >= 0; i-- {
		h = filters[i](h)
	}
	return func(c *ginhttp.Context) {
		// Store gin context in request before passing to filter chain
		r := SetGinContext(c.Request, c)
		h.ServeHTTP(c.Writer, r)
	}
}

func handleError(c *ginhttp.Context, err error) {
	c.JSON(http.StatusInternalServerError, http2.Response{
		Code:    http2.ErrInternal,
		Message: err.Error(),
	})
}

// ==================== contextWrapper ====================
// contextWrapper adapts a *gin.Context to implement the http2.Context interface.
// This is the true adapter layer that decouples handler logic from gin.

type contextWrapper struct {
	ginCtx *ginhttp.Context
}

var _ http2.Context = (*contextWrapper)(nil)

func (c *contextWrapper) Deadline() (time.Time, bool)       { return c.ginCtx.Request.Context().Deadline() }
func (c *contextWrapper) Done() <-chan struct{}             { return c.ginCtx.Request.Context().Done() }
func (c *contextWrapper) Err() error                        { return c.ginCtx.Request.Context().Err() }
func (c *contextWrapper) Value(key interface{}) interface{} { return c.ginCtx.Request.Context().Value(key) }

func (c *contextWrapper) Vars() url.Values {
	params := c.ginCtx.Params
	vars := make(url.Values, len(params))
	for _, p := range params {
		vars[p.Key] = []string{p.Value}
	}
	return vars
}

func (c *contextWrapper) Query() url.Values             { return c.ginCtx.Request.URL.Query() }
func (c *contextWrapper) Request() *http.Request        { return c.ginCtx.Request }
func (c *contextWrapper) Response() http.ResponseWriter { return c.ginCtx.Writer }
func (c *contextWrapper) Header() http.Header           { return c.ginCtx.Request.Header }

func (c *contextWrapper) Form() url.Values {
	if err := c.ginCtx.Request.ParseForm(); err != nil {
		return url.Values{}
	}
	return c.ginCtx.Request.Form
}

func (c *contextWrapper) Bind(v interface{}) error      { return c.ginCtx.ShouldBindJSON(v) }
func (c *contextWrapper) BindVars(v interface{}) error  { return c.ginCtx.ShouldBindUri(v) }
func (c *contextWrapper) BindQuery(v interface{}) error { return c.ginCtx.ShouldBindQuery(v) }
func (c *contextWrapper) BindForm(v interface{}) error  { return c.ginCtx.ShouldBind(v) }

func (c *contextWrapper) Returns(v interface{}, err error) error {
	if err != nil {
		return err
	}
	c.ginCtx.JSON(http.StatusOK, v)
	return nil
}

func (c *contextWrapper) Result(code int, v interface{}) error {
	c.ginCtx.JSON(code, v)
	return nil
}

func (c *contextWrapper) JSON(code int, v interface{}) error {
	c.ginCtx.JSON(code, v)
	return nil
}

func (c *contextWrapper) String(code int, text string) error {
	c.ginCtx.String(code, text)
	return nil
}

func (c *contextWrapper) Blob(code int, contentType string, data []byte) error {
	c.ginCtx.Data(code, contentType, data)
	return nil
}

func (c *contextWrapper) Stream(code int, contentType string, rd io.Reader) error {
	c.ginCtx.DataFromReader(code, -1, contentType, rd, nil)
	return nil
}

func (c *contextWrapper) Reset(res http.ResponseWriter, req *http.Request) {
	c.ginCtx.Request = req
	c.ginCtx.Writer = &ginResponseWriterAdapter{w: res}
}

// GinContext returns the underlying *gin.Context.
// This is useful for handlers that need gin-specific features
// (e.g., c.Stream, c.FormFile) during the migration period.
func (c *contextWrapper) GinContext() *ginhttp.Context {
	return c.ginCtx
}

// GinContextFromHTTP extracts the underlying *gin.Context from an http2.Context.
// Returns nil if the context is not a contextWrapper.
func GinContextFromHTTP(ctx http2.Context) *ginhttp.Context {
	if wrapper, ok := ctx.(*contextWrapper); ok {
		return wrapper.GinContext()
	}
	return nil
}

// ==================== ginResponseWriterAdapter ====================
// ginResponseWriterAdapter adapts an http.ResponseWriter to gin.ResponseWriter.
// Used by contextWrapper.Reset to swap the writer.

type ginResponseWriterAdapter struct {
	w http.ResponseWriter
}

func (g *ginResponseWriterAdapter) Header() http.Header            { return g.w.Header() }
func (g *ginResponseWriterAdapter) Write(data []byte) (int, error) { return g.w.Write(data) }
func (g *ginResponseWriterAdapter) WriteHeader(code int)           { g.w.WriteHeader(code) }

func (g *ginResponseWriterAdapter) Status() int  { return 0 }
func (g *ginResponseWriterAdapter) Size() int     { return 0 }
func (g *ginResponseWriterAdapter) Written() bool { return false }
func (g *ginResponseWriterAdapter) WriteString(s string) (int, error) {
	return g.w.Write([]byte(s))
}
func (g *ginResponseWriterAdapter) WriteHeaderNow() {}

func (g *ginResponseWriterAdapter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := g.w.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}
func (g *ginResponseWriterAdapter) CloseNotify() <-chan bool { return nil }
func (g *ginResponseWriterAdapter) Flush() {
	if f, ok := g.w.(http.Flusher); ok {
		f.Flush()
	}
}
func (g *ginResponseWriterAdapter) Pusher() http.Pusher {
	if p, ok := g.w.(http.Pusher); ok {
		return p
	}
	return nil
}
