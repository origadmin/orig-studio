package frontend

import (
	"github.com/gin-gonic/gin"
)

func RegisterGinRoutes(r *gin.Engine) {
	handler := NewSPAHandler()
	if handler.IsEmpty() {
		return
	}

	for _, route := range DefaultStaticRoutes {
		prefix := route.Prefix
		cc := route.CacheControl
		r.GET(prefix+"/*filepath", func(c *gin.Context) {
			c.Header("Cache-Control", cc)
			c.Request.URL.Path = prefix + c.Param("filepath")
			handler.fileServer.ServeHTTP(c.Writer, c.Request)
		})
	}

	for _, name := range DefaultRootFiles {
		n := name
		r.GET(n, func(c *gin.Context) {
			handler.ServeRootFile(c.Writer, c.Request, n)
		})
	}

	r.NoRoute(func(c *gin.Context) {
		handler.ServeSPAFallback(c.Writer, c.Request)
	})
}

func RegisterGinRoutesWithHandler(r *gin.Engine, handler *SPAHandler) {
	if handler.IsEmpty() {
		return
	}

	for _, route := range DefaultStaticRoutes {
		prefix := route.Prefix
		cc := route.CacheControl
		r.GET(prefix+"/*filepath", func(c *gin.Context) {
			c.Header("Cache-Control", cc)
			c.Request.URL.Path = prefix + c.Param("filepath")
			handler.fileServer.ServeHTTP(c.Writer, c.Request)
		})
	}

	for _, name := range DefaultRootFiles {
		n := name
		r.GET(n, func(c *gin.Context) {
			handler.ServeRootFile(c.Writer, c.Request, n)
		})
	}

	r.NoRoute(func(c *gin.Context) {
		handler.ServeSPAFallback(c.Writer, c.Request)
	})
}
