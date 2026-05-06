package http

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

type Context interface {
	context.Context
	Vars() url.Values
	Query() url.Values
	Form() url.Values
	Header() http.Header
	Request() *http.Request
	Response() http.ResponseWriter
	Bind(v interface{}) error
	BindVars(v interface{}) error
	BindQuery(v interface{}) error
	BindForm(v interface{}) error
	Returns(v interface{}, err error) error
	Result(code int, v interface{}) error
	JSON(code int, v interface{}) error
	String(code int, text string) error
	Blob(code int, contentType string, data []byte) error
	Stream(code int, contentType string, rd io.Reader) error
	Reset(res http.ResponseWriter, req *http.Request)
}

type HandlerFunc func(Context) error

type FilterFunc func(http.Handler) http.Handler

type Router interface {
	Group(prefix string, filters ...FilterFunc) Router
	GET(path string, h HandlerFunc, filters ...FilterFunc)
	POST(path string, h HandlerFunc, filters ...FilterFunc)
	PUT(path string, h HandlerFunc, filters ...FilterFunc)
	DELETE(path string, h HandlerFunc, filters ...FilterFunc)
	PATCH(path string, h HandlerFunc, filters ...FilterFunc)
}

type Module interface {
	RegisterRoutes(r Router)
}
