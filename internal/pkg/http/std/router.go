// Package std provides a framework-agnostic implementation of http2.Router and
// http2.Context backed only by the standard library (net/http).
//
// It exists so that Enterprise Edition (EE) modules can serve HTTP without
// depending on gin. EE handler code must stay framework-agnostic (http2.Context)
// and is wired to this adapter (or the Kratos adapter) at the runtime boundary.
// This keeps EE compliant with the highest rule: "EE 不用 gin, CE 使用 gin".
package std

import (
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/pkg/http/validate"
)

// Router is a minimal http2.Router implementation over net/http.
// It also implements http.Handler so it can be mounted directly.
type Router struct {
	prefix string
	mws    []http2.MiddlewareFunc
	routes *[]route
}

type route struct {
	method  string
	path    string
	handler http2.HandlerFunc
	mws     []http2.MiddlewareFunc
}

// NewRouter creates an empty std Router.
func NewRouter() *Router {
	return &Router{routes: &[]route{}}
}

// Group returns a sub-router that inherits the parent prefix and middleware.
func (rt *Router) Group(prefix string, mws ...http2.MiddlewareFunc) http2.Router {
	return &Router{
		prefix: rt.prefix + prefix,
		mws:    append(append([]http2.MiddlewareFunc{}, rt.mws...), mws...),
		routes: rt.routes,
	}
}

// Use registers framework-agnostic middleware on the router.
func (rt *Router) Use(mws ...http2.MiddlewareFunc) {
	rt.mws = append(rt.mws, mws...)
}

func (rt *Router) add(method, path string, h http2.HandlerFunc, mws []http2.MiddlewareFunc) {
	full := rt.prefix + path
	combined := append(append([]http2.MiddlewareFunc{}, rt.mws...), mws...)
	*rt.routes = append(*rt.routes, route{method: method, path: full, handler: h, mws: combined})
}

// Static registers a static file server (best-effort, optional for EE).
func (rt *Router) Static(relativePath, root string) {
	fileServer := http.FileServer(http.Dir(root))
	*rt.routes = append(*rt.routes, route{
		method:  http.MethodGet,
		path:    rt.prefix + relativePath + "/*filepath",
		handler: func(c http2.Context) error { fileServer.ServeHTTP(c.Response(), c.Request()); return nil },
	})
}

func (rt *Router) GET(path string, h http2.HandlerFunc, mws ...http2.MiddlewareFunc) {
	rt.add(http.MethodGet, path, h, mws)
}
func (rt *Router) POST(path string, h http2.HandlerFunc, mws ...http2.MiddlewareFunc) {
	rt.add(http.MethodPost, path, h, mws)
}
func (rt *Router) PUT(path string, h http2.HandlerFunc, mws ...http2.MiddlewareFunc) {
	rt.add(http.MethodPut, path, h, mws)
}
func (rt *Router) DELETE(path string, h http2.HandlerFunc, mws ...http2.MiddlewareFunc) {
	rt.add(http.MethodDelete, path, h, mws)
}
func (rt *Router) PATCH(path string, h http2.HandlerFunc, mws ...http2.MiddlewareFunc) {
	rt.add(http.MethodPatch, path, h, mws)
}

// ServeHTTP dispatches an http.Request to the matching route.
func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range *rt.routes {
		if route.method != r.Method {
			continue
		}
		params, ok := matchPath(route.path, r.URL.Path)
		if !ok {
			continue
		}
		ctx := &contextWrapper{req: r, w: w, vars: params}
		h := route.handler
		for i := len(route.mws) - 1; i >= 0; i-- {
			h = route.mws[i](h)
		}
		if err := h(ctx); err != nil {
			writeError(w, err)
		}
		return
	}
	http.NotFound(w, r)
}

// matchPath converts a ":param" route pattern to a regex and extracts values.
func matchPath(pattern, path string) (url.Values, bool) {
	if !strings.Contains(pattern, ":") {
		if pattern == path {
			return url.Values{}, true
		}
		return nil, false
	}
	parts := strings.Split(pattern, "/")
	reStr := ""
	names := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		if strings.HasPrefix(p, ":") {
			names = append(names, p[1:])
			reStr += "/([^/]+)"
		} else if strings.HasPrefix(p, "*") {
			names = append(names, p[1:])
			reStr += "/(.*)"
		} else {
			reStr += "/" + regexp.QuoteMeta(p)
		}
	}
	re := regexp.MustCompile("^" + reStr + "$")
	m := re.FindStringSubmatch(path)
	if m == nil {
		return nil, false
	}
	vals := url.Values{}
	for i, n := range names {
		vals.Set(n, m[i+1])
	}
	return vals, true
}

func writeError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    500,
		"reason":  "INTERNAL_ERROR",
		"message": err.Error(),
	})
}

// contextWrapper adapts *http.Request / http.ResponseWriter to http2.Context.
type contextWrapper struct {
	req     *http.Request
	w       http.ResponseWriter
	vars    url.Values
	formParsed bool
}

func (c *contextWrapper) Deadline() (deadline time.Time, ok bool) { return c.req.Context().Deadline() }
func (c *contextWrapper) Done() <-chan struct{}                         { return c.req.Context().Done() }
func (c *contextWrapper) Err() error                                    { return c.req.Context().Err() }
func (c *contextWrapper) Value(key interface{}) interface{}             { return c.req.Context().Value(key) }

func (c *contextWrapper) Request() *http.Request { return c.req }
func (c *contextWrapper) Response() http.ResponseWriter {
	return &statusWriter{w: c.w, status: 0}
}

func (c *contextWrapper) Vars() url.Values { return c.vars }
func (c *contextWrapper) Var(name string) string {
	return c.vars.Get(name)
}
func (c *contextWrapper) Query() url.Values { return c.req.URL.Query() }
func (c *contextWrapper) QueryVar(name string) string {
	return c.req.URL.Query().Get(name)
}
func (c *contextWrapper) QueryVarDefault(name, defaultValue string) string {
	if v := c.req.URL.Query().Get(name); v != "" {
		return v
	}
	return defaultValue
}
func (c *contextWrapper) Header() http.Header { return c.req.Header }
func (c *contextWrapper) GetHeader(name string) string {
	return c.req.Header.Get(name)
}
func (c *contextWrapper) Form() url.Values {
	if !c.formParsed {
		_ = c.req.ParseForm()
		c.formParsed = true
	}
	return c.req.Form
}
func (c *contextWrapper) FormVar(name string) string { return c.req.FormValue(name) }

// ClientIP returns the real client IP, checking X-Forwarded-For and X-Real-IP headers.
func (c *contextWrapper) ClientIP() string {
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

func (c *contextWrapper) Bind(v interface{}) error { return c.BindJSON(v) }
func (c *contextWrapper) BindJSON(v interface{}) error {
	body, err := io.ReadAll(c.req.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, v); err != nil {
		return err
	}
	return validate.Validate(v)
}
func (c *contextWrapper) BindVars(v interface{}) error { return bindVars(c.vars, v) }
func (c *contextWrapper) BindQuery(v interface{}) error {
	return bindQuery(c.req.URL.Query(), v)
}
func (c *contextWrapper) BindForm(v interface{}) error {
	return bindQuery(c.Form(), v)
}

func (c *contextWrapper) GetRawData() ([]byte, error) { return io.ReadAll(c.req.Body) }

func (c *contextWrapper) FormFile(name string) (multipart.File, *multipart.FileHeader, error) {
	f, h, err := c.req.FormFile(name)
	if err != nil {
		return nil, nil, err
	}
	return f, h, nil
}
func (c *contextWrapper) MultipartForm() (*multipart.Form, error) {
	return c.req.MultipartForm, nil
}

func (c *contextWrapper) Set(key string, value interface{}) {
	c.req = c.req.WithContext(context.WithValue(c.req.Context(), key, value))
}
func (c *contextWrapper) Get(key string) (interface{}, bool) {
	return c.req.Context().Value(key), true
}
func (c *contextWrapper) GetString(key string) string {
	if v, ok := c.req.Context().Value(key).(string); ok {
		return v
	}
	return ""
}

func (c *contextWrapper) Result(code int, v interface{}) error {
	c.w.WriteHeader(code)
	if v == nil {
		return nil
	}
	c.w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(c.w).Encode(v)
}
func (c *contextWrapper) JSON(code int, v interface{}) error { return c.Result(code, v) }
func (c *contextWrapper) String(code int, text string) error {
	c.w.WriteHeader(code)
	_, err := io.WriteString(c.w, text)
	return err
}
func (c *contextWrapper) Blob(code int, contentType string, data []byte) error {
	c.w.Header().Set("Content-Type", contentType)
	c.w.WriteHeader(code)
	_, err := c.w.Write(data)
	return err
}
func (c *contextWrapper) Stream(code int, contentType string, rd io.Reader) error {
	c.w.Header().Set("Content-Type", contentType)
	c.w.WriteHeader(code)
	_, err := io.Copy(c.w, rd)
	return err
}
func (c *contextWrapper) File(path string) error {
	http.ServeFile(c.w, c.req, path)
	return nil
}
func (c *contextWrapper) Returns(v interface{}, err error) error {
	if err != nil {
		return err
	}
	return c.Result(http.StatusOK, v)
}
func (c *contextWrapper) Reset(res http.ResponseWriter, req *http.Request) {
	c.w = res
	c.req = req
	c.vars = nil
	c.formParsed = false
}

// statusWriter wraps http.ResponseWriter to satisfy the interface used by Result.
type statusWriter struct {
	w     http.ResponseWriter
	status int
}

func (s *statusWriter) Header() http.Header { return s.w.Header() }
func (s *statusWriter) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	return s.w.Write(b)
}
func (s *statusWriter) WriteHeader(code int) { s.status = code; s.w.WriteHeader(code) }

// bindVars maps url.Values (path params) into a struct using `uri` tags.
func bindVars(vars url.Values, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return nil
	}
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		tag := field.Tag.Get("uri")
		if tag == "" || tag == "-" {
			continue
		}
		val := vars.Get(tag)
		if val == "" {
			continue
		}
		if err := setField(rv.Field(i), val); err != nil {
			return err
		}
	}
	return nil
}

// bindQuery maps url.Values (query/form) into a struct using `json`/`form` tags.
func bindQuery(values url.Values, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return nil
	}
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		tag := field.Tag.Get("json")
		if tag == "" {
			tag = field.Tag.Get("form")
		}
		if tag == "" || tag == "-" {
			continue
		}
		name := strings.Split(tag, ",")[0]
		val := values.Get(name)
		if val == "" {
			continue
		}
		if err := setField(rv.Field(i), val); err != nil {
			return err
		}
	}
	return nil
}

func setField(fv reflect.Value, val string) error {
	if !fv.CanSet() {
		return nil
	}
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		fv.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		fv.SetUint(n)
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		fv.SetBool(b)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		fv.SetFloat(n)
	}
	return nil
}
