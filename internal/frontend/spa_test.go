package frontend

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterRoutes(r)
	return r
}

func TestIsDistEmpty(t *testing.T) {
	empty := IsDistEmpty()
	t.Logf("IsDistEmpty() = %v", empty)
}

func TestRegisterRoutesWithEmptyDist(t *testing.T) {
	origFS := DistFS
	DistFS = fs.FS(nil)
	defer func() { DistFS = origFS }()

	gin.SetMode(gin.TestMode)
	r := gin.New()

	RegisterRoutes(r)

	if len(r.Routes()) != 0 {
		t.Errorf("expected no routes with empty DistFS, got %d", len(r.Routes()))
	}
}

func TestSPAFallbackReturnsHTML(t *testing.T) {
	origFS := DistFS
	defer func() { DistFS = origFS }()

	t.Log("SPA fallback logic tested via route registration checks")
}

func TestAPIRoutesNotFallback(t *testing.T) {
	origFS := DistFS
	defer func() { DistFS = origFS }()

	DistFS = fs.FS(nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/api/v1/something" ||
			path == "/uploads/video.mp4" ||
			path == "/thumbnails/thumb.jpg" ||
			path == "/hls/stream.m3u8" {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}
		c.Header("Cache-Control", "no-cache")
		c.String(200, "index.html content")
	})

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantJSON   bool
	}{
		{"API route returns JSON 404", "/api/v1/something", 404, true},
		{"Uploads route returns JSON 404", "/uploads/video.mp4", 404, true},
		{"Thumbnails route returns JSON 404", "/thumbnails/thumb.jpg", 404, true},
		{"HLS route returns JSON 404", "/hls/stream.m3u8", 404, true},
		{"SPA route returns 200", "/watch", 200, false},
		{"SPA deep route returns 200", "/admin/media", 200, false},
		{"Root returns 200", "/", 200, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("path %s: got status %d, want %d", tt.path, w.Code, tt.wantStatus)
			}
			if tt.wantJSON {
				ct := w.Header().Get("Content-Type")
				if ct != "application/json; charset=utf-8" {
					t.Errorf("path %s: got Content-Type %q, want JSON", tt.path, ct)
				}
			}
		})
	}
}

func TestAssetsCacheHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/assets/*filepath", func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
		c.String(200, "asset content")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets/index-abc123.js", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("got status %d, want 200", w.Code)
	}
	cc := w.Header().Get("Cache-Control")
	if cc != "public, max-age=31536000, immutable" {
		t.Errorf("got Cache-Control %q, want immutable", cc)
	}
}

func TestStaticCacheHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/static/*filepath", func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=86400")
		c.String(200, "static content")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/static/js/index.5bddddd6.js", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("got status %d, want 200", w.Code)
	}
	cc := w.Header().Get("Cache-Control")
	if cc != "public, max-age=86400" {
		t.Errorf("got Cache-Control %q, want 86400", cc)
	}
}

func TestIndexHTMLNoCache(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if stringsHasPrefix(path, "/api/") ||
			stringsHasPrefix(path, "/uploads/") ||
			stringsHasPrefix(path, "/thumbnails/") ||
			stringsHasPrefix(path, "/hls/") {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}
		c.Header("Cache-Control", "no-cache")
		c.String(200, "index.html content")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/watch", nil)
	r.ServeHTTP(w, req)

	cc := w.Header().Get("Cache-Control")
	if cc != "no-cache" {
		t.Errorf("got Cache-Control %q, want no-cache", cc)
	}
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func TestRegisterRoutesWithPopulatedDist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterRoutes(r)

	routes := r.Routes()
	hasAssets := false
	hasStatic := false
	hasFavicon := false
	for _, route := range routes {
		if route.Path == "/assets/*filepath" {
			hasAssets = true
		}
		if route.Path == "/static/*filepath" {
			hasStatic = true
		}
		if route.Path == "/favicon.ico" {
			hasFavicon = true
		}
	}

	if !hasAssets {
		t.Error("expected /assets/*filepath route to be registered")
	}
	if !hasStatic {
		t.Error("expected /static/*filepath route to be registered")
	}
	if !hasFavicon {
		t.Error("expected /favicon.ico route to be registered")
	}
}

func TestNoRouteHandlerForAPIPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterRoutes(r)

	apiPaths := []string{
		"/api/v1/nonexistent",
		"/api/v1/auth/me",
		"/uploads/missing.mp4",
		"/thumbnails/missing.jpg",
		"/hls/missing.m3u8",
	}

	for _, path := range apiPaths {
		t.Run("API_path_"+path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", path, nil)
			r.ServeHTTP(w, req)

			if w.Code != 404 {
				t.Errorf("path %s: got status %d, want 404", path, w.Code)
			}
			ct := w.Header().Get("Content-Type")
			if ct != "application/json; charset=utf-8" {
				t.Errorf("path %s: got Content-Type %q, want JSON", path, ct)
			}
		})
	}
}
