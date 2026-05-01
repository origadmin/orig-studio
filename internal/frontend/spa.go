package frontend

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	if IsDistEmpty() {
		return
	}

	sub, err := fs.Sub(DistFS, "dist")
	if err != nil {
		return
	}
	fileServer := http.FileServer(http.FS(sub))

	indexHTML, indexErr := fs.ReadFile(sub, "index.html")
	_ = indexErr

	staticPaths := []string{"/assets", "/static", "/locales"}
	for _, prefix := range staticPaths {
		p := prefix
		r.GET(p+"/*filepath", func(c *gin.Context) {
			switch p {
			case "/assets":
				c.Header("Cache-Control", "public, max-age=31536000, immutable")
			case "/locales":
				c.Header("Cache-Control", "public, max-age=3600")
			default:
				c.Header("Cache-Control", "public, max-age=86400")
			}
			c.Request.URL.Path = p + c.Param("filepath")
			fileServer.ServeHTTP(c.Writer, c.Request)
		})
	}

	rootFiles := []string{"/favicon.ico", "/robots.txt", "/manifest.json", "/logo-192.png", "/logo-512.png"}
	for _, name := range rootFiles {
		n := name
		r.GET(n, func(c *gin.Context) {
			c.Request.URL.Path = n
			fileServer.ServeHTTP(c.Writer, c.Request)
		})
	}

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		if strings.HasPrefix(path, "/api/") ||
			strings.HasPrefix(path, "/uploads/") ||
			strings.HasPrefix(path, "/thumbnails/") ||
			strings.HasPrefix(path, "/hls/") {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}

		c.Header("Cache-Control", "no-cache")
		if indexErr != nil {
			c.String(500, "index.html not found in embedded frontend")
			return
		}
		c.Data(200, "text/html; charset=utf-8", indexHTML)
	})
}

func IsDistEmpty() bool {
	if DistFS == nil {
		return true
	}
	entries, err := fs.ReadDir(DistFS, "dist")
	return err != nil || len(entries) == 0
}
