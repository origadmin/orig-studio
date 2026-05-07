package http

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

// Context represents the framework-agnostic request context.
type Context interface {
	context.Context

	// Request data access
	Request() *http.Request
	Vars() url.Values
	Var(name string) string                         // convenience for Vars().Get(name)
	Query() url.Values
	QueryVar(name string) string                    // convenience for Query().Get(name)
	QueryVarDefault(name, defaultValue string) string // with default
	Form() url.Values
	FormVar(name string) string                     // convenience for Form().Get(name)
	Header() http.Header
	GetHeader(name string) string                   // convenience for Header().Get(name)

	// Request binding
	Bind(v interface{}) error
	BindJSON(v interface{}) error                   // explicit JSON binding
	BindVars(v interface{}) error
	BindQuery(v interface{}) error
	BindForm(v interface{}) error

	// Multipart / file upload
	FormFile(name string) (multipart.File, *multipart.FileHeader, error)
	MultipartForm() (*multipart.Form, error)
	GetRawData() ([]byte, error)

	// Context value storage
	Set(key string, value interface{})
	Get(key string) (interface{}, bool)
	GetString(key string) string

	// Response writing
	Response() http.ResponseWriter
	JSON(code int, v interface{}) error
	String(code int, text string) error
	Blob(code int, contentType string, data []byte) error
	Stream(code int, contentType string, rd io.Reader) error
	File(path string) error                         // serve static file
	Result(code int, v interface{}) error
	Returns(v interface{}, err error) error
	Reset(res http.ResponseWriter, req *http.Request)
}

// HandlerFunc is the framework-agnostic handler signature.
type HandlerFunc func(Context) error

// MiddlewareFunc is the framework-agnostic middleware signature.
// It wraps a HandlerFunc and returns a new HandlerFunc (decorator pattern).
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// Chain composes multiple middleware into a single MiddlewareFunc.
// The first middleware in the list is the outermost wrapper.
func Chain(mws ...MiddlewareFunc) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		for i := len(mws) - 1; i >= 0; i-- {
			next = mws[i](next)
		}
		return next
	}
}

// Router represents the framework-agnostic route registration interface.
type Router interface {
	Group(prefix string, mws ...MiddlewareFunc) Router
	Use(mws ...MiddlewareFunc)
	GET(path string, h HandlerFunc, mws ...MiddlewareFunc)
	POST(path string, h HandlerFunc, mws ...MiddlewareFunc)
	PUT(path string, h HandlerFunc, mws ...MiddlewareFunc)
	DELETE(path string, h HandlerFunc, mws ...MiddlewareFunc)
	PATCH(path string, h HandlerFunc, mws ...MiddlewareFunc)
	Static(relativePath, root string)
}

// Module represents a feature module that registers its routes.
type Module interface {
	RegisterRoutes(r Router)
}
