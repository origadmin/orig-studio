package service

// This file re-exports server.HTTPToHandlerFunc as httpToHandlerFunc
// for use within the content/service package.
import (
	"net/http"

	http2 "origadmin/application/origstudio/internal/helpers/http"
	"origadmin/application/origstudio/internal/server"
)

// httpToHandlerFunc wraps a standard http.HandlerFunc into an http2.HandlerFunc.
// It delegates to server.HTTPToHandlerFunc.
func httpToHandlerFunc(h http.HandlerFunc) http2.HandlerFunc {
	return server.HTTPToHandlerFunc(h)
}
