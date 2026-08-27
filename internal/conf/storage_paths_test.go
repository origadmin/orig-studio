package conf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStoragePaths_RelativePaths(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	refTime := time.Date(2026, 6, 25, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{
			name:     "RelativeOriginalAt",
			got:      sp.RelativeOriginalAt("admin", "abc123.mp4", refTime),
			expected: "originals/admin/2026/06/abc123.mp4",
		},
		{
			name:     "RelativeTempAt",
			got:      sp.RelativeTempAt("admin", "abc123.mp4", refTime),
			expected: "temp/admin/2026/06/abc123.mp4",
		},
		{
			name:     "RelativeThumbnail",
			got:      sp.RelativeThumbnail("media-uuid"),
			expected: "thumbnails/media-uuid.jpg",
		},
		{
			name:     "RelativeHLSMaster",
			got:      sp.RelativeHLSMaster("media-uuid"),
			expected: "hls/media-uuid/master.m3u8",
		},
		{
			name:     "RelativeSpriteVTT",
			got:      sp.RelativeSpriteVTT("media-uuid"),
			expected: "sprites/media-uuid/sprite.vtt",
		},
		{
			name:     "RelativeSpriteImage",
			got:      sp.RelativeSpriteImage("media-uuid"),
			expected: "sprites/media-uuid/sprite.jpg",
		},
		{
			name:     "RelativePreview",
			got:      sp.RelativePreview("media-uuid"),
			expected: "previews/media-uuid.gif",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.got)
			}
		})
	}
}

func TestStoragePaths_TempUploadDirAt(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	refTime := time.Date(2026, 6, 30, 23, 59, 0, 0, time.UTC)
	got := sp.TempUploadDirAt("admin", "upload-xyz", refTime)
	gotSlash := filepath.ToSlash(got)

	// Must contain year/month from refTime, not current time
	expectedSuffix := "temp/admin/2026/06/upload-xyz"
	if !strings.HasSuffix(gotSlash, expectedSuffix) {
		t.Errorf("TempUploadDirAt expected suffix %q, got %q", expectedSuffix, gotSlash)
	}
}

func TestStoragePaths_TempPartPathAt(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	refTime := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	got := sp.TempPartPathAt("user1", "up-1", 1, refTime)
	gotSlash := filepath.ToSlash(got)

	expectedSuffix := "temp/user1/2026/01/up-1/part_00001"
	if !strings.HasSuffix(gotSlash, expectedSuffix) {
		t.Errorf("TempPartPathAt expected suffix %q, got %q", expectedSuffix, gotSlash)
	}
}

func TestPromoteToOriginal_CrossMonth(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	// Simulate a file created in June, promoted in July (cross-month scenario)
	juneTime := time.Date(2026, 6, 30, 23, 59, 0, 0, time.UTC)
	tempPath := sp.RelativeTempAt("admin", "cross-month.mp4", juneTime)

	// Write a test file to the temp location
	srcFull := sp.FullPath(tempPath)
	if err := os.MkdirAll(filepath.Dir(srcFull), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(srcFull, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	// PromoteToOriginal must preserve the original June path, not use July (current time)
	promotedPath, err := sp.PromoteToOriginal(tempPath)
	if err != nil {
		t.Fatalf("PromoteToOriginal failed: %v", err)
	}

	// The promoted path must be originals/admin/2026/06/... (June), not 07 (July)
	expected := "originals/admin/2026/06/cross-month.mp4"
	if promotedPath != expected {
		t.Errorf("PromoteToOriginal cross-month: expected %q, got %q", expected, promotedPath)
	}

	// Verify temp file no longer exists
	if _, err := os.Stat(srcFull); !os.IsNotExist(err) {
		t.Error("temp file should be removed after promotion")
	}

	// Verify original file exists
	dstFull := sp.FullPath(promotedPath)
	data, err := os.ReadFile(dstFull)
	if err != nil {
		t.Fatalf("promoted file not found: %v", err)
	}
	if string(data) != "test content" {
		t.Errorf("promoted file content mismatch")
	}
}

func TestPromoteToOriginal_InvalidPath(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{name: "not temp prefix", path: "originals/admin/2026/06/x.mp4", wantErr: true},
		{name: "too few parts", path: "temp/admin", wantErr: true},
		{name: "empty", path: "", wantErr: true},
		{name: "valid temp path", path: "temp/admin/2026/06/test.mp4", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				// Create the temp file first
				srcFull := sp.FullPath(tt.path)
				if err := os.MkdirAll(filepath.Dir(srcFull), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(srcFull, []byte("x"), 0644); err != nil {
					t.Fatal(err)
				}
			}
			_, err := sp.PromoteToOriginal(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("PromoteToOriginal(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestStoragePaths_FullPath(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	relKey := "originals/admin/2026/06/video.mp4"
	full := sp.FullPath(relKey)
	expected := filepath.Join(tmpDir, filepath.FromSlash(relKey))
	if full != expected {
		t.Errorf("FullPath: expected %q, got %q", expected, full)
	}
}

// TestStoragePaths_RegistryAutoRegister verifies that NewStoragePaths auto-registers
// all expected directories (video, assets, gallery).
func TestStoragePaths_RegistryAutoRegister(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	// All auto-registered dirs must be accessible via Dir() without panic.
	expectedDirs := []string{
		"originals", "temp", "thumbnails", "hls", "previews", "sprites",
		"assets/avatars", "assets/ads", "assets/banners", "assets/covers",
		"assets/channels", "assets/categories", "assets/articles", "assets/misc",
		"gallery",
	}

	for _, name := range expectedDirs {
		t.Run("Dir/"+name, func(t *testing.T) {
			dir := sp.Dir(name)
			if dir == "" {
				t.Errorf("Dir(%q) returned empty", name)
			}
			// Verify the directory exists on filesystem
			if _, err := os.Stat(dir); err != nil {
				t.Errorf("Dir(%q) path %q does not exist: %v", name, dir, err)
			}
		})
	}
}

// TestStoragePaths_GalleryRegistered verifies gallery dir is registered
// (P0 fix: previously gallery was never registered, causing panic).
func TestStoragePaths_GalleryRegistered(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	// Relative("gallery", ...) must not panic
	rel := sp.Relative("gallery", "uuid-123", "cover.jpg")
	expected := "gallery/uuid-123/cover.jpg"
	if rel != expected {
		t.Errorf("gallery Relative: expected %q, got %q", expected, rel)
	}

	// Dir("gallery") must return existing path
	dir := sp.Dir("gallery")
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("gallery dir does not exist: %v", err)
	}
}

// TestStoragePaths_RegisterCustom verifies custom dir registration via Register().
func TestStoragePaths_RegisterCustom(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	// Register a custom dir (e.g. future "document" type)
	err := sp.Register(PathSpec{
		Name:        "document",
		DefaultDir:  "document",
		Description: "document content",
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Dir must work
	dir := sp.Dir("document")
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("document dir does not exist: %v", err)
	}

	// Relative must work
	rel := sp.Relative("document", "pdf", "file.pdf")
	expected := "document/pdf/file.pdf"
	if rel != expected {
		t.Errorf("document Relative: expected %q, got %q", expected, rel)
	}
}

// TestStoragePaths_StaticRouteMap verifies that StaticRouteMap generates
// /files/{dir} routes for all registered dirs.
func TestStoragePaths_StaticRouteMap(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	routes := sp.StaticRouteMap()

	// Must include video dirs
	expectedRoutes := []string{
		"/files/originals", "/files/temp", "/files/thumbnails",
		"/files/hls", "/files/previews", "/files/sprites",
	}
	for _, route := range expectedRoutes {
		if _, ok := routes[route]; !ok {
			t.Errorf("StaticRouteMap missing route %q", route)
		}
	}

	// Must include gallery (P0 fix)
	if _, ok := routes["/files/gallery"]; !ok {
		t.Errorf("StaticRouteMap missing /files/gallery (gallery should be auto-registered)")
	}

	// Must include at least one asset route
	if _, ok := routes["/files/assets/avatars"]; !ok {
		t.Errorf("StaticRouteMap missing /files/assets/avatars")
	}
}

// TestStoragePaths_Overrides verifies per-dir path override.
func TestStoragePaths_Overrides(t *testing.T) {
	tmpDir := t.TempDir()
	overrideDir := t.TempDir()

	overrides := map[string]string{
		"gallery": overrideDir, // put gallery on a different path
	}
	sp := NewStoragePathsWithOverrides(tmpDir, overrides)

	// Dir("gallery") must return the override path, not basePath/gallery
	galleryDir := sp.Dir("gallery")
	if galleryDir != overrideDir {
		t.Errorf("gallery Dir with override: expected %q, got %q", overrideDir, galleryDir)
	}

	// Relative("gallery", ...) still uses DefaultDir for the relative key
	// (relative keys are for DB storage, not filesystem paths)
	rel := sp.Relative("gallery", "img.jpg")
	if rel != "gallery/img.jpg" {
		t.Errorf("gallery Relative with override: expected %q, got %q", "gallery/img.jpg", rel)
	}
}

// TestStoragePaths_DirPanicsOnUnregistered verifies Dir() panics for unregistered names.
func TestStoragePaths_DirPanicsOnUnregistered(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Dir() with unregistered name should panic")
		}
	}()
	_ = sp.Dir("nonexistent-dir")
}
