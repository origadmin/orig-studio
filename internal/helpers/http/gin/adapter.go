package gin

import (
	"bufio"
	"context"
	"io"
	"mime/multipart"
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
	group *ginhttp.RouterGroup
	mws   []http2.MiddlewareFunc
}

var _ http2.Router = (*RouterAdapter)(nil)

// NewRouterAdapter creates a new RouterAdapter wrapping the given gin.RouterGroup.
func NewRouterAdapter(group *ginhttp.RouterGroup) *RouterAdapter {
	return &RouterAdapter{group: group}
}

// Group returns a new RouterAdapter for the given prefix, inheriting
// any middleware from the parent and appending the new ones.
func (a *RouterAdapter) Group(prefix string, mws ...http2.MiddlewareFunc) http2.Router {
	ginGroup := a.group.Group(prefix)
	for _, mw := range mws {
		ginGroup.Use(middlewareToGin(mw))
	}
	return &RouterAdapter{
		group: ginGroup,
		mws:   append(a.mws, mws...),
	}
}

// Use adds framework-agnostic middleware to the underlying RouterGroup.
func (a *RouterAdapter) Use(mws ...http2.MiddlewareFunc) {
	for _, mw := range mws {
		a.group.Use(middlewareToGin(mw))
	}
	a.mws = append(a.mws, mws...)
}

// UseGin adds gin middleware to the underlying RouterGroup.
// This is a convenience method for gin middleware that cannot be
// expressed as http2.MiddlewareFunc (e.g., middleware using c.Next()).
// Handlers should type-assert the http2.Router to *RouterAdapter
// when they need to add gin middleware.
func (a *RouterAdapter) UseGin(mw ...ginhttp.HandlerFunc) {
	a.group.Use(mw...)
}

func (a *RouterAdapter) GET(path string, h http2.HandlerFunc, mws ...http2.MiddlewareFunc) {
	a.register(http.MethodGet, path, h, mws)
}

func (a *RouterAdapter) POST(path string, h http2.HandlerFunc, mws ...http2.MiddlewareFunc) {
	a.register(http.MethodPost, path, h, mws)
}

func (a *RouterAdapter) PUT(path string, h http2.HandlerFunc, mws ...http2.MiddlewareFunc) {
	a.register(http.MethodPut, path, h, mws)
}

func (a *RouterAdapter) DELETE(path string, h http2.HandlerFunc, mws ...http2.MiddlewareFunc) {
	a.register(http.MethodDelete, path, h, mws)
}

func (a *RouterAdapter) PATCH(path string, h http2.HandlerFunc, mws ...http2.MiddlewareFunc) {
	a.register(http.MethodPatch, path, h, mws)
}

// Static serves static files from the given root directory.
func (a *RouterAdapter) Static(relativePath, root string) {
	a.group.Static(relativePath, root)
}

func (a *RouterAdapter) register(method, relativePath string, h http2.HandlerFunc, mws []http2.MiddlewareFunc) {
	handler := a.wrapHandler(h)
	allMws := append(a.mws, mws...)
	if len(allMws) > 0 {
		handler = a.applyMiddleware(handler, allMws)
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

// middlewareToGin converts a framework-agnostic MiddlewareFunc to a gin.HandlerFunc.
func middlewareToGin(mw http2.MiddlewareFunc) ginhttp.HandlerFunc {
	return func(c *ginhttp.Context) {
		ctx := &contextWrapper{ginCtx: c}
		next := func(ctx http2.Context) error {
			c.Next()
			return nil
		}
		wrapped := mw(next)
		if err := wrapped(ctx); err != nil {
			handleError(c, err)
		}
	}
}

// applyMiddleware wraps a gin.HandlerFunc with the given middleware chain.
func (a *RouterAdapter) applyMiddleware(handler ginhttp.HandlerFunc, mws []http2.MiddlewareFunc) ginhttp.HandlerFunc {
	// Convert the gin handler to a framework-agnostic HandlerFunc,
	// apply the middleware chain, then convert back to a gin handler.
	agnosticHandler := func(ctx http2.Context) error {
		// The ctx is a contextWrapper; extract the gin context to call the original handler.
		if wrapper, ok := ctx.(*contextWrapper); ok {
			handler(wrapper.ginCtx)
			return nil
		}
		return nil
	}

	// Apply middleware chain: outermost first
	wrapped := agnosticHandler
	for i := len(mws) - 1; i >= 0; i-- {
		wrapped = mws[i](wrapped)
	}

	return func(c *ginhttp.Context) {
		ctx := &contextWrapper{ginCtx: c}
		if err := wrapped(ctx); err != nil {
			handleError(c, err)
		}
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

func (c *contextWrapper) Var(name string) string {
	val, _ := c.ginCtx.Params.Get(name)
	return val
}

func (c *contextWrapper) Query() url.Values             { return c.ginCtx.Request.URL.Query() }
func (c *contextWrapper) QueryVar(name string) string   { return c.ginCtx.Query(name) }
func (c *contextWrapper) QueryVarDefault(name, defaultValue string) string {
	return c.ginCtx.DefaultQuery(name, defaultValue)
}

func (c *contextWrapper) Form() url.Values {
	if err := c.ginCtx.Request.ParseForm(); err != nil {
		return url.Values{}
	}
	return c.ginCtx.Request.Form
}
func (c *contextWrapper) FormVar(name string) string { return c.ginCtx.PostForm(name) }

func (c *contextWrapper) Request() *http.Request        { return c.ginCtx.Request }
func (c *contextWrapper) Response() http.ResponseWriter { return c.ginCtx.Writer }
func (c *contextWrapper) Header() http.Header           { return c.ginCtx.Request.Header }
func (c *contextWrapper) GetHeader(name string) string  { return c.ginCtx.GetHeader(name) }

func (c *contextWrapper) Bind(v interface{}) error      { return c.ginCtx.ShouldBindJSON(v) }
func (c *contextWrapper) BindJSON(v interface{}) error  { return c.ginCtx.ShouldBindJSON(v) }
func (c *contextWrapper) BindVars(v interface{}) error  { return c.ginCtx.ShouldBindUri(v) }
func (c *contextWrapper) BindQuery(v interface{}) error { return c.ginCtx.ShouldBindQuery(v) }
func (c *contextWrapper) BindForm(v interface{}) error  { return c.ginCtx.ShouldBind(v) }

func (c *contextWrapper) FormFile(name string) (multipart.File, *multipart.FileHeader, error) {
	fileHeader, err := c.ginCtx.FormFile(name)
	if err != nil {
		return nil, nil, err
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, nil, err
	}
	return file, fileHeader, nil
}

func (c *contextWrapper) MultipartForm() (*multipart.Form, error) {
	return c.ginCtx.MultipartForm()
}

func (c *contextWrapper) GetRawData() ([]byte, error) {
	return c.ginCtx.GetRawData()
}

func (c *contextWrapper) Set(key string, value interface{}) {
	c.ginCtx.Set(key, value)
}

func (c *contextWrapper) Get(key string) (interface{}, bool) {
	return c.ginCtx.Get(key)
}

func (c *contextWrapper) GetString(key string) string {
	val, _ := c.ginCtx.Get(key)
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}

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

func (c *contextWrapper) File(path string) error {
	c.ginCtx.File(path)
	return nil
}

func (c *contextWrapper) Reset(res http.ResponseWriter, req *http.Request) {
	c.ginCtx.Request = req
	c.ginCtx.Writer = &ginResponseWriterAdapter{w: res}
}

// GinContext returns the underlying *gin.Context.
// Deprecated: Use the extended Context interface methods instead.
// This is retained for the migration period only.
func (c *contextWrapper) GinContext() *ginhttp.Context {
	return c.ginCtx
}

// GinContextFromHTTP extracts the underlying *gin.Context from an http2.Context.
// Deprecated: Use the extended Context interface methods instead.
// This is retained for the migration period only.
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
