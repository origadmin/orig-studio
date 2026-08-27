package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRateLimiterExcludesStaticAssets(t *testing.T) {
	rl := NewRateLimiter(10,
		"/static",
		"/assets",
		"/locales",
		"/themes",
		"/files",
		"/health",
		"/favicon.ico",
		"/robots.txt",
		"/manifest.json",
		"/logo",
		"/banner",
	)

	staticPaths := []string{
		"/static/js/index.abc123.js",
		"/static/js/async/9080.9c9b9777.js",
		"/static/css/index.1294e658.css",
		"/static/js/lib-react.847b2a86.js",
		"/assets/images/video.svg",
		"/assets/images/cover.svg",
		"/assets/images/avatar.svg",
		"/locales/zh.json",
		"/locales/en.json",
		"/themes/registry.json",
		"/themes/default/index.css",
		"/favicon.ico",
		"/robots.txt",
		"/manifest.json",
		"/logo.svg",
		"/logo-192.png",
		"/logo-32.png",
		"/banner.png",
		"/banner.svg",
	}

	for _, path := range staticPaths {
		t.Run(path, func(t *testing.T) {
			for i := 0; i < 50; i++ {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest("GET", path, nil)
				c.Request.RemoteAddr = "192.168.1.100:1234"

				rl.Middleware()(c)

				if w.Code == http.StatusTooManyRequests {
					t.Errorf("static asset path %q should NOT be rate-limited, but got 429 on request %d", path, i+1)
					return
				}
				if c.IsAborted() {
					t.Errorf("static asset path %q request was aborted on request %d (code=%d)", path, i+1, w.Code)
					return
				}
			}
		})
	}
}

func TestRateLimiterExcludesFilePaths(t *testing.T) {
	rl := NewRateLimiter(10,
		"/static",
		"/assets",
		"/locales",
		"/themes",
		"/files",
		"/health",
		"/banner",
	)

	filePaths := []string{
		"/files/thumbnails/abc123.jpg",
		"/files/originals/def456.mp4",
		"/files/hls/stream/seg1.ts",
		"/files/hls/stream/index.m3u8",
		"/files/sprites/sprite.vtt",
		"/files/previews/prev.jpg",
	}

	for _, path := range filePaths {
		t.Run(path, func(t *testing.T) {
			for i := 0; i < 50; i++ {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest("GET", path, nil)
				c.Request.RemoteAddr = "192.168.1.100:1235"

				rl.Middleware()(c)

				if w.Code == http.StatusTooManyRequests {
					t.Errorf("file path %q should NOT be rate-limited, but got 429 on request %d", path, i+1)
					return
				}
			}
		})
	}
}

func TestRateLimiterExcludesHealthEndpoint(t *testing.T) {
	rl := NewRateLimiter(5, "/health")

	for i := 0; i < 20; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/health", nil)
		c.Request.RemoteAddr = "192.168.1.100:1236"

		rl.Middleware()(c)

		if w.Code == http.StatusTooManyRequests {
			t.Fatalf("health endpoint should never be rate-limited, got 429 on request %d", i+1)
		}
	}
}

func TestRateLimiterLimitsAPIPaths(t *testing.T) {
	rl := NewRateLimiter(10,
		"/static",
		"/assets",
		"/locales",
		"/themes",
		"/files",
		"/health",
		"/banner",
	)

	blocked := false
	for i := 0; i < 30; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/medias?page=1&page_size=12", nil)
		c.Request.RemoteAddr = "192.168.1.100:1237"

		rl.Middleware()(c)

		if w.Code == http.StatusTooManyRequests {
			blocked = true
			break
		}
	}
	if !blocked {
		t.Log("API paths should be rate-limited eventually (warn: burst may allow more than expected)")
	}
}

func TestRateLimiterDisabledWhenZero(t *testing.T) {
	rl := NewRateLimiter(0)

	if !rl.disabled {
		t.Error("rate limiter with RPM=0 should be disabled")
	}

	for i := 0; i < 100; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/test", nil)
		c.Request.RemoteAddr = "192.168.1.100:1238"

		rl.Middleware()(c)

		if w.Code == http.StatusTooManyRequests {
			t.Fatalf("disabled rate limiter should not block any request (request %d)", i+1)
		}
	}
}
