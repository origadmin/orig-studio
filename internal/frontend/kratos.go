package frontend

import (
	"net/http"

	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

func RegisterKratosRoutes(srv *kratoshttp.Server) {
	handler := NewSPAHandler()
	if handler.IsEmpty() {
		return
	}

	for _, route := range DefaultStaticRoutes {
		prefix := route.Prefix
		cc := route.CacheControl
		srv.HandlePrefix(prefix+"/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", cc)
			handler.fileServer.ServeHTTP(w, r)
		}))
	}

	for _, name := range DefaultRootFiles {
		n := name
		srv.HandleFunc(n, func(w http.ResponseWriter, r *http.Request) {
			handler.ServeRootFile(w, r, n)
		})
	}

	srv.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || (!IsAPIPath(r.URL.Path) && !IsStaticAssetPath(r.URL.Path) && !IsRootFile(r.URL.Path)) {
			handler.ServeSPAFallback(w, r)
			return
		}
		handler.fileServer.ServeHTTP(w, r)
	})
}
