// Package kratos provides a Kratos HTTP server adapter that implements
// the framework-agnostic http.Router and http.Context interfaces.
//
// This adapter allows registering routes directly on a Kratos
// *transhttp.Server (which uses gorilla/mux internally) without
// going through a Gin engine. It translates Gin-style path parameters
// (:param) to gorilla/mux-style ({param}) at registration time.
package kratos

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	stdhttp "net/http"
	"net/url"
	"path"
	"strings"
	"time"

	transhttp "github.com/go-kratos/kratos/v2/transport/http"

	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/pkg/http/validate"
)

// ==================== RouterAdapter ====================

// RouterAdapter implements the http2.Router interface by adapting
// a Kratos *transhttp.Router. It is the Kratos counterpart of the
// Gin adapter: all handlers register routes via http2.Router, and
// only this adapter knows about Kratos.
type RouterAdapter struct {
	srv    *transhttp.Server // underlying Kratos server (for HandlePrefix if needed)
	router *transhttp.Router // Kratos router for method-based registration
	prefix string            // accumulated prefix (for Static path computation)
	mws    []http2.MiddlewareFunc
}

var _ http2.Router = (*RouterAdapter)(nil)

// NewRouterAdapter creates a new RouterAdapter wrapping a Kratos server.
// The prefix is passed to srv.Route(prefix) to obtain a Kratos *Router.
func NewRouterAdapter(srv *transhttp.Server, prefix string) *RouterAdapter {
	return &RouterAdapter{
		srv:    srv,
		router: srv.Route(prefix),
		prefix: prefix,
	}
}

// NewRouterAdapterFromRouter creates a RouterAdapter from an existing
// Kratos *Router (e.g., obtained from srv.Route). The srv is needed
// for Static support; pass nil if Static is not used.
func NewRouterAdapterFromRouter(srv *transhttp.Server, router *transhttp.Router, prefix string) *RouterAdapter {
	return &RouterAdapter{
		srv:    srv,
		router: router,
		prefix: prefix,
	}
}

// Group returns a new RouterAdapter for the given prefix, inheriting
// any middleware from the parent and appending the new ones.
func (a *RouterAdapter) Group(prefix string, mws ...http2.MiddlewareFunc) http2.Router {
	newPrefix := path.Join(a.prefix, prefix)
	return &RouterAdapter{
		srv:    a.srv,
		router: a.router.Group(prefix),
		prefix: newPrefix,
		mws:    append(append([]http2.MiddlewareFunc{}, a.mws...), mws...),
	}
}

// Use adds framework-agnostic middleware to this router.
// Middleware applies to all subsequent route registrations.
func (a *RouterAdapter) Use(mws ...http2.MiddlewareFunc) {
	a.mws = append(a.mws, mws...)
}

func (a *RouterAdapter) GET(path string, h http2.HandlerFunc, mws ...http2.MiddlewareFunc) {
	a.register(stdhttp.MethodGet, path, h, mws)
}

func (a *RouterAdapter) POST(path string, h http2.HandlerFunc, mws ...http2.MiddlewareFunc) {
	a.register(stdhttp.MethodPost, path, h, mws)
}

func (a *RouterAdapter) PUT(path string, h http2.HandlerFunc, mws ...http2.MiddlewareFunc) {
	a.register(stdhttp.MethodPut, path, h, mws)
}

func (a *RouterAdapter) DELETE(path string, h http2.HandlerFunc, mws ...http2.MiddlewareFunc) {
	a.register(stdhttp.MethodDelete, path, h, mws)
}

func (a *RouterAdapter) PATCH(path string, h http2.HandlerFunc, mws ...http2.MiddlewareFunc) {
	a.register(stdhttp.MethodPatch, path, h, mws)
}

// Static serves static files from the given root directory.
// It registers a catch-all route under relativePath.
func (a *RouterAdapter) Static(relativePath, root string) {
	absolutePath := path.Join(a.prefix, relativePath)
	fileServer := stdhttp.StripPrefix(absolutePath, stdhttp.FileServer(stdhttp.Dir(root)))
	catchAllPath := translatePath(relativePath + "/*filepath")
	a.router.Handle(stdhttp.MethodGet, catchAllPath, func(ctx transhttp.Context) error {
		fileServer.ServeHTTP(ctx.Response(), ctx.Request())
		return nil
	})
}

// register wraps an http2.HandlerFunc as a Kratos HandlerFunc and
// registers it on the underlying Kratos router.
func (a *RouterAdapter) register(method, relativePath string, h http2.HandlerFunc, mws []http2.MiddlewareFunc) {
	allMws := append(append([]http2.MiddlewareFunc{}, a.mws...), mws...)

	kratosHandler := func(ctx transhttp.Context) error {
		wrapper := &contextWrapper{kratosCtx: ctx}
		handler := h
		if len(allMws) > 0 {
			chain := http2.Chain(allMws...)
			handler = chain(h)
		}
		return handler(wrapper)
	}

	a.router.Handle(method, translatePath(relativePath), kratosHandler)
}

// ==================== contextWrapper ====================

// contextWrapper adapts a Kratos transhttp.Context to implement
// the http2.Context interface. It delegates 21 methods to the
// underlying Kratos context and implements 13 methods itself.
type contextWrapper struct {
	kratosCtx transhttp.Context
	keys      map[string]interface{}
}

var _ http2.Context = (*contextWrapper)(nil)

// --- context.Context methods (delegated) ---

func (c *contextWrapper) Deadline() (time.Time, bool)       { return c.kratosCtx.Deadline() }
func (c *contextWrapper) Done() <-chan struct{}             { return c.kratosCtx.Done() }
func (c *contextWrapper) Err() error                        { return c.kratosCtx.Err() }
func (c *contextWrapper) Value(key interface{}) interface{} { return c.kratosCtx.Value(key) }

// --- Request data access (delegated) ---

func (c *contextWrapper) Request() *stdhttp.Request { return c.kratosCtx.Request() }
func (c *contextWrapper) Vars() url.Values       { return c.kratosCtx.Vars() }
func (c *contextWrapper) Query() url.Values      { return c.kratosCtx.Query() }
func (c *contextWrapper) Form() url.Values       { return c.kratosCtx.Form() }
func (c *contextWrapper) Header() stdhttp.Header    { return c.kratosCtx.Header() }

// --- Request data access (implemented) ---

func (c *contextWrapper) Var(name string) string {
	return c.kratosCtx.Vars().Get(name)
}

func (c *contextWrapper) QueryVar(name string) string {
	return c.kratosCtx.Query().Get(name)
}

func (c *contextWrapper) QueryVarDefault(name, defaultValue string) string {
	if v := c.kratosCtx.Query().Get(name); v != "" {
		return v
	}
	return defaultValue
}

func (c *contextWrapper) FormVar(name string) string {
	return c.kratosCtx.Form().Get(name)
}

func (c *contextWrapper) GetHeader(name string) string {
	return c.kratosCtx.Header().Get(name)
}

// --- Request binding (delegated) ---

func (c *contextWrapper) Bind(v interface{}) error      { return c.kratosCtx.Bind(v) }
func (c *contextWrapper) BindVars(v interface{}) error  { return c.kratosCtx.BindVars(v) }
func (c *contextWrapper) BindQuery(v interface{}) error { return c.kratosCtx.BindQuery(v) }
func (c *contextWrapper) BindForm(v interface{}) error  { return c.kratosCtx.BindForm(v) }

// --- Request binding (implemented) ---

func (c *contextWrapper) BindJSON(v interface{}) error {
	// Kratos transhttp.Context.Bind uses its own codec registry and enforces
	// strict content-type matching (which may fail when content-type is empty
	// or contains charset suffixes). To keep binding behaviour aligned with
	// the Gin adapter we decode JSON from the raw body directly. To guard
	// against the case where the Kratos transport has already drained the
	// body (e.g. during RequestID / recovery middleware), we first slurp the
	// body into a buffer and re-wrap a fresh ReadCloser so future reads also
	// succeed.
	req := c.kratosCtx.Request()
	raw, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	req.Body = io.NopCloser(bytes.NewReader(raw))
	if len(raw) > 0 {
		if err := json.NewDecoder(bytes.NewReader(raw)).Decode(v); err != nil {
			return err
		}
	}
	return validate.Validate(v)
}

// --- Multipart / file upload (implemented) ---

func (c *contextWrapper) FormFile(name string) (multipart.File, *multipart.FileHeader, error) {
	return c.kratosCtx.Request().FormFile(name)
}

func (c *contextWrapper) MultipartForm() (*multipart.Form, error) {
	if c.kratosCtx.Request().MultipartForm == nil {
		if err := c.kratosCtx.Request().ParseMultipartForm(32 << 20); err != nil {
			return nil, err
		}
	}
	return c.kratosCtx.Request().MultipartForm, nil
}

func (c *contextWrapper) GetRawData() ([]byte, error) {
	return io.ReadAll(c.kratosCtx.Request().Body)
}

// --- Context value storage (implemented) ---

func (c *contextWrapper) Set(key string, value interface{}) {
	if c.keys == nil {
		c.keys = make(map[string]interface{})
	}
	c.keys[key] = value
}

func (c *contextWrapper) Get(key string) (interface{}, bool) {
	v, ok := c.keys[key]
	return v, ok
}

func (c *contextWrapper) GetString(key string) string {
	if v, ok := c.Get(key); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// --- Client information (implemented) ---

// ClientIP returns the real client IP, checking X-Forwarded-For and X-Real-IP headers.
func (c *contextWrapper) ClientIP() string {
	req := c.kratosCtx.Request()
	if ip := req.Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For may contain multiple IPs, take the first (original client)
		if idx := strings.Index(ip, ","); idx >= 0 {
			return strings.TrimSpace(ip[:idx])
		}
		return strings.TrimSpace(ip)
	}
	if ip := req.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	// Fall back to RemoteAddr (strip port)
	if idx := strings.LastIndex(req.RemoteAddr, ":"); idx >= 0 {
		return req.RemoteAddr[:idx]
	}
	return req.RemoteAddr
}

// --- Response writing (delegated) ---

func (c *contextWrapper) Response() stdhttp.ResponseWriter       { return c.kratosCtx.Response() }
func (c *contextWrapper) JSON(code int, v interface{}) error  { return c.kratosCtx.JSON(code, v) }
func (c *contextWrapper) String(code int, text string) error  { return c.kratosCtx.String(code, text) }
func (c *contextWrapper) Blob(code int, contentType string, data []byte) error {
	return c.kratosCtx.Blob(code, contentType, data)
}
func (c *contextWrapper) Stream(code int, contentType string, rd io.Reader) error {
	return c.kratosCtx.Stream(code, contentType, rd)
}
func (c *contextWrapper) Result(code int, v interface{}) error        { return c.kratosCtx.Result(code, v) }
func (c *contextWrapper) Returns(v interface{}, err error) error     { return c.kratosCtx.Returns(v, err) }
func (c *contextWrapper) Reset(res stdhttp.ResponseWriter, req *stdhttp.Request) {
	c.kratosCtx.Reset(res, req)
}

// --- Response writing (implemented) ---

func (c *contextWrapper) File(filePath string) error {
	stdhttp.ServeFile(c.kratosCtx.Response(), c.kratosCtx.Request(), filePath)
	return nil
}

// ==================== Path Translation ====================

// translatePath converts Gin-style path parameters to gorilla/mux-style.
//   - :param  → {param}
//   - *param  → {param:.*}  (catch-all)
//   - *       → {path:.*}   (unnamed catch-all)
//
// Literal segments (no : or * prefix) are left unchanged.
func translatePath(p string) string {
	segments := strings.Split(p, "/")
	for i, seg := range segments {
		if len(seg) == 0 {
			continue
		}
		switch seg[0] {
		case ':':
			segments[i] = "{" + seg[1:] + "}"
		case '*':
			name := seg[1:]
			if name == "" {
				name = "path"
			}
			segments[i] = "{" + name + ":.*}"
		}
	}
	return strings.Join(segments, "/")
}

// ==================== Kratos Filter Bridge ====================
// The Filter bridge allows http2.MiddlewareFunc to be applied to
// gRPC-Gateway generated stdhttp.Handler routes via Kratos Filter option.
// This is the key enabler for securing proto-generated routes without
// wrapping each handler individually.

// MiddlewareToFilter converts a framework-agnostic MiddlewareFunc to a
// Kratos FilterFunc, so it can protect gRPC-Gateway generated stdhttp.Handler
// routes (registered via RegisterXxxHTTPServer).
// Usage: kratoshttp.Filter(kratosadapter.MiddlewareToFilter(server.JWTMiddlewareCtx(jwt)))
func MiddlewareToFilter(mw http2.MiddlewareFunc) transhttp.FilterFunc {
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			// Create a minimal http2.Context backed by stdlib types.
			ctx := &filterContext{
				req: r,
				w:   w,
			}
			handler := func(c http2.Context) error {
				next.ServeHTTP(c.Response(), c.Request())
				return nil
			}
			mw(handler)(ctx)
		})
	}
}

// PathPrefixMatcher returns a Kratos MatchFunc that matches if the request
// path starts with any of the given prefixes. Used with Filter or MatchFilter
// to apply middleware to a group of routes by path prefix.
func PathPrefixMatcher(prefixes ...string) func(*stdhttp.Request) bool {
	return func(r *stdhttp.Request) bool {
		for _, p := range prefixes {
			if strings.HasPrefix(r.URL.Path, p) {
				return true
			}
		}
		return false
	}
}

// filterContext is a minimal http2.Context implementation used by the
// Filter bridge. It only needs to support the methods that middleware
// typically use: Request, Response, Header, GetHeader, QueryVar, Set, Get,
// ClientIP, Result, JSON.
type filterContext struct {
	req  *stdhttp.Request
	w    stdhttp.ResponseWriter
	keys map[string]interface{}
}

var _ http2.Context = (*filterContext)(nil)

func (c *filterContext) Deadline() (time.Time, bool)       { return c.req.Context().Deadline() }
func (c *filterContext) Done() <-chan struct{}             { return c.req.Context().Done() }
func (c *filterContext) Err() error                        { return c.req.Context().Err() }
func (c *filterContext) Value(key interface{}) interface{} { return c.req.Context().Value(key) }

func (c *filterContext) Request() *stdhttp.Request { return c.req }
func (c *filterContext) Response() stdhttp.ResponseWriter {
	return &statusWriter{w: c.w}
}

func (c *filterContext) Vars() url.Values  { return url.Values{} }
func (c *filterContext) Var(name string) string { return "" }
func (c *filterContext) Query() url.Values { return c.req.URL.Query() }
func (c *filterContext) QueryVar(name string) string { return c.req.URL.Query().Get(name) }
func (c *filterContext) QueryVarDefault(name, defaultValue string) string {
	if v := c.req.URL.Query().Get(name); v != "" {
		return v
	}
	return defaultValue
}
func (c *filterContext) Form() url.Values {
	_ = c.req.ParseForm()
	return c.req.Form
}
func (c *filterContext) FormVar(name string) string { return c.req.FormValue(name) }
func (c *filterContext) Header() stdhttp.Header    { return c.req.Header }
func (c *filterContext) GetHeader(name string) string { return c.req.Header.Get(name) }

func (c *filterContext) Bind(v interface{}) error      { return c.BindJSON(v) }
func (c *filterContext) BindJSON(v interface{}) error {
	if err := json.NewDecoder(c.req.Body).Decode(v); err != nil {
		return err
	}
	return validate.Validate(v)
}
func (c *filterContext) BindVars(v interface{}) error  { return nil }
func (c *filterContext) BindQuery(v interface{}) error { return nil }
func (c *filterContext) BindForm(v interface{}) error  { return nil }

func (c *filterContext) FormFile(name string) (multipart.File, *multipart.FileHeader, error) {
	return c.req.FormFile(name)
}
func (c *filterContext) MultipartForm() (*multipart.Form, error) {
	return c.req.MultipartForm, nil
}
func (c *filterContext) GetRawData() ([]byte, error) { return io.ReadAll(c.req.Body) }

func (c *filterContext) Set(key string, value interface{}) {
	if c.keys == nil {
		c.keys = make(map[string]interface{})
	}
	c.keys[key] = value
	// Also inject into request context so downstream handlers can access it
	c.req = c.req.WithContext(context.WithValue(c.req.Context(), contextKey(key), value))
}
func (c *filterContext) Get(key string) (interface{}, bool) {
	if v, ok := c.keys[key]; ok {
		return v, true
	}
	v := c.req.Context().Value(contextKey(key))
	return v, v != nil
}
func (c *filterContext) GetString(key string) string {
	if v, ok := c.Get(key); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (c *filterContext) ClientIP() string {
	if ip := c.req.Header.Get("X-Forwarded-For"); ip != "" {
		if idx := strings.Index(ip, ","); idx >= 0 {
			return strings.TrimSpace(ip[:idx])
		}
		return strings.TrimSpace(ip)
	}
	if ip := c.req.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	if idx := strings.LastIndex(c.req.RemoteAddr, ":"); idx >= 0 {
		return c.req.RemoteAddr[:idx]
	}
	return c.req.RemoteAddr
}

func (c *filterContext) JSON(code int, v interface{}) error {
	return c.Result(code, v)
}
func (c *filterContext) String(code int, text string) error {
	c.w.WriteHeader(code)
	_, err := io.WriteString(c.w, text)
	return err
}
func (c *filterContext) Blob(code int, contentType string, data []byte) error {
	c.w.Header().Set("Content-Type", contentType)
	c.w.WriteHeader(code)
	_, err := c.w.Write(data)
	return err
}
func (c *filterContext) Stream(code int, contentType string, rd io.Reader) error {
	c.w.Header().Set("Content-Type", contentType)
	c.w.WriteHeader(code)
	_, err := io.Copy(c.w, rd)
	return err
}
func (c *filterContext) File(filePath string) error {
	stdhttp.ServeFile(c.w, c.req, filePath)
	return nil
}
func (c *filterContext) Result(code int, v interface{}) error {
	c.w.WriteHeader(code)
	if v == nil {
		return nil
	}
	c.w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(c.w).Encode(v)
}
func (c *filterContext) Returns(v interface{}, err error) error {
	if err != nil {
		return err
	}
	return c.Result(stdhttp.StatusOK, v)
}
func (c *filterContext) Reset(res stdhttp.ResponseWriter, req *stdhttp.Request) {
	c.w = res
	c.req = req
	c.keys = nil
}

type contextKey string

// statusWriter wraps stdhttp.ResponseWriter to track status code.
type statusWriter struct {
	w      stdhttp.ResponseWriter
	status int
}

func (s *statusWriter) Header() stdhttp.Header { return s.w.Header() }
func (s *statusWriter) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = stdhttp.StatusOK
	}
	return s.w.Write(b)
}
func (s *statusWriter) WriteHeader(code int) {
	s.status = code
	s.w.WriteHeader(code)
}
