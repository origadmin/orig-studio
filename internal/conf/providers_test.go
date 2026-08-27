package conf

import (
	"strings"
	"testing"
)

func TestNewRateLimiterConfig_DefaultRPM(t *testing.T) {
	rpm, excludes := NewRateLimiterConfig(nil)

	if rpm < 300 {
		t.Errorf("default RPM should be at least 300 to prevent ChunkLoadError in SPA, got %d", rpm)
	}

	requiredPrefixes := []string{
		"/static",
		"/assets",
		"/locales",
		"/themes",
		"/files",
		"/health",
		"/banner",
		"/api/v1/uploads",
	}

	for _, required := range requiredPrefixes {
		found := false
		for _, ex := range excludes {
			if ex == required || strings.HasPrefix(required, ex) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("rate limiter excludes must contain %q to prevent static asset 429/ChunkLoadError; got %v", required, excludes)
		}
	}
}

func TestNewRateLimiterConfig_StaticAssetPrefixesCoverJSChunks(t *testing.T) {
	_, excludes := NewRateLimiterConfig(nil)

	jsChunkPaths := []string{
		"/static/js/async/9080.9c9b9777.js",
		"/static/js/lib-react.847b2a86.js",
		"/static/js/index.5964e4e8.js",
		"/static/css/index.1294e658.css",
		"/themes/registry.json",
		"/themes/default/index.css",
		"/banner.png",
		"/banner.svg",
	}

	for _, path := range jsChunkPaths {
		excluded := false
		for _, prefix := range excludes {
			if strings.HasPrefix(path, prefix) {
				excluded = true
				break
			}
		}
		if !excluded {
			t.Errorf("JS chunk path %q would be rate-limited! This causes ChunkLoadError. Check exclude prefixes: %v", path, excludes)
		}
	}
}

func TestNewRateLimiterConfig_MediaFilePaths(t *testing.T) {
	_, excludes := NewRateLimiterConfig(nil)

	mediaPaths := []string{
		"/files/thumbnails/abc.jpg",
		"/files/hls/stream/seg1.ts",
		"/files/hls/stream/index.m3u8",
		"/files/originals/video.mp4",
		"/files/sprites/sprite.vtt",
	}

	for _, path := range mediaPaths {
		excluded := false
		for _, prefix := range excludes {
			if strings.HasPrefix(path, prefix) {
				excluded = true
				break
			}
		}
		if !excluded {
			t.Errorf("media file path %q would be rate-limited! This causes video playback failures. Check exclude prefixes: %v", path, excludes)
		}
	}
}
