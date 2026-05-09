/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package conf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewStoragePaths(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	if sp.BasePath() == "" {
		t.Fatal("BasePath should not be empty")
	}

	// Verify absolute path resolution
	if !filepath.IsAbs(sp.BasePath()) {
		t.Fatalf("BasePath should be absolute, got: %s", sp.BasePath())
	}

	// Verify all directories are created
	for _, dir := range []string{
		sp.OriginalsDir,
		sp.TempDir,
		sp.ThumbnailsDir,
		sp.HLSDir,
		sp.PreviewsDir,
		sp.SpritesDir,
	} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("directory should exist: %s", dir)
		}
	}
}

func TestOriginalPath(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	path := sp.OriginalPath("user123", "abc-def.mp4")

	// Should contain originals/user123/yyyy/MM/abc-def.mp4
	if !strings.Contains(path, "originals") {
		t.Errorf("path should contain 'originals': %s", path)
	}
	if !strings.Contains(path, "user123") {
		t.Errorf("path should contain user ID: %s", path)
	}
	if !strings.Contains(path, "abc-def.mp4") {
		t.Errorf("path should contain filename: %s", path)
	}
	if !filepath.IsAbs(path) {
		t.Errorf("path should be absolute: %s", path)
	}
}

func TestOriginalPathAt(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	tm := time.Date(2026, 5, 7, 10, 0, 0, 0, time.UTC)
	path := sp.OriginalPathAt("user123", "test.mp4", tm)

	expected := filepath.Join(sp.OriginalsDir, "user123", "2026", "05", "test.mp4")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestTempUploadDir(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	path := sp.TempUploadDir("user123", "upload-abc")

	if !strings.Contains(path, "temp") {
		t.Errorf("path should contain 'temp': %s", path)
	}
	if !strings.Contains(path, "user123") {
		t.Errorf("path should contain user ID: %s", path)
	}
	if !strings.Contains(path, "upload-abc") {
		t.Errorf("path should contain upload ID: %s", path)
	}
	if !filepath.IsAbs(path) {
		t.Errorf("path should be absolute: %s", path)
	}
}

func TestTempPartPath(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	path := sp.TempPartPath("user123", "upload-abc", 1)

	if !strings.Contains(path, "part_00001") {
		t.Errorf("path should contain 'part_00001': %s", path)
	}
}

func TestTempMergedPath(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	path := sp.TempMergedPath("user123", "upload-abc.mp4")

	if !strings.Contains(path, "temp") {
		t.Errorf("path should contain 'temp': %s", path)
	}
	if !strings.Contains(path, "upload-abc.mp4") {
		t.Errorf("path should contain filename: %s", path)
	}

	// Merged file should be a sibling of the parts directory, not inside it
	partsDir := sp.TempUploadDir("user123", "upload-abc")
	if strings.HasPrefix(path, partsDir+string(filepath.Separator)) {
		t.Errorf("merged file should NOT be inside parts directory: %s", path)
	}
}

func TestRelativeOriginal(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	rel := sp.RelativeOriginal("user123", "test.mp4")

	if !strings.HasPrefix(rel, "originals/user123/") {
		t.Errorf("relative path should start with 'originals/user123/': %s", rel)
	}
	if !strings.HasSuffix(rel, "/test.mp4") {
		t.Errorf("relative path should end with '/test.mp4': %s", rel)
	}
}

func TestRelativeOriginalAt(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	tm := time.Date(2026, 5, 7, 10, 0, 0, 0, time.UTC)
	rel := sp.RelativeOriginalAt("user123", "test.mp4", tm)

	expected := "originals/user123/2026/05/test.mp4"
	if rel != expected {
		t.Errorf("expected %s, got %s", expected, rel)
	}
}

func TestRelativeTemp(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	rel := sp.RelativeTemp("user123", "test.mp4")

	if !strings.HasPrefix(rel, "temp/user123/") {
		t.Errorf("relative path should start with 'temp/user123/': %s", rel)
	}
	if !strings.HasSuffix(rel, "/test.mp4") {
		t.Errorf("relative path should end with '/test.mp4': %s", rel)
	}
}

func TestFullPath(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	absPath := sp.FullPath("originals/user123/2026/05/test.mp4")

	expected := filepath.Join(sp.BasePath(), "originals/user123/2026/05/test.mp4")
	if absPath != expected {
		t.Errorf("expected %s, got %s", expected, absPath)
	}
}

func TestFullPathBackwardCompatibility(t *testing.T) {
	// Old data stored as bare filenames like "abc123.mp4" should still resolve
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	absPath := sp.FullPath("abc123.mp4")
	expected := filepath.Join(sp.BasePath(), "abc123.mp4")
	if absPath != expected {
		t.Errorf("backward compat: expected %s, got %s", expected, absPath)
	}
}

func TestPromoteToOriginal(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	// Create a temp file to promote
	filename := "test-upload.mp4"
	userID := "user123"

	tempPath := sp.TempMergedPath(userID, filename)
	if err := os.MkdirAll(filepath.Dir(tempPath), 0755); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	if err := os.WriteFile(tempPath, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	// Promote
	relPath, err := sp.PromoteToOriginal(userID, filename)
	if err != nil {
		t.Fatalf("PromoteToOriginal failed: %v", err)
	}

	// Verify relative path format
	if !strings.HasPrefix(relPath, "originals/user123/") {
		t.Errorf("relative path should start with 'originals/user123/': %s", relPath)
	}

	// Verify file moved to originals
	originalAbsPath := sp.FullPath(relPath)
	data, err := os.ReadFile(originalAbsPath)
	if err != nil {
		t.Fatalf("failed to read promoted file: %v", err)
	}
	if string(data) != "test content" {
		t.Errorf("file content mismatch: expected 'test content', got '%s'", string(data))
	}

	// Verify temp file no longer exists
	if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
		t.Error("temp file should have been removed after promotion")
	}
}

func TestCleanupTempParts(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	userID := "user123"
	uploadID := "upload-abc"

	// Create parts directory with a part file
	partsDir := sp.TempUploadDir(userID, uploadID)
	if err := os.MkdirAll(partsDir, 0755); err != nil {
		t.Fatalf("failed to create parts dir: %v", err)
	}
	partPath := filepath.Join(partsDir, "part_00001")
	if err := os.WriteFile(partPath, []byte("part data"), 0644); err != nil {
		t.Fatalf("failed to write part file: %v", err)
	}

	// Cleanup
	if err := sp.CleanupTempParts(userID, uploadID); err != nil {
		t.Fatalf("CleanupTempParts failed: %v", err)
	}

	// Verify parts directory removed
	if _, err := os.Stat(partsDir); !os.IsNotExist(err) {
		t.Error("parts directory should have been removed")
	}
}

func TestThumbnailAbsPath(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	path := sp.ThumbnailAbsPath("media-uuid-123")
	expected := filepath.Join(sp.ThumbnailsDir, "media-uuid-123.jpg")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestHLSDirForMedia(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	path := sp.HLSDirForMedia("media-uuid-123")
	expected := filepath.Join(sp.HLSDir, "media-uuid-123")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestPreviewAbsPath(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	path := sp.PreviewAbsPath("media-uuid-123")
	expected := filepath.Join(sp.PreviewsDir, "media-uuid-123.gif")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestSpriteImageAbsPath(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	path := sp.SpriteImageAbsPath("media-uuid-123")
	expected := filepath.Join(sp.SpritesDir, "media-uuid-123", "sprite.jpg")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestSpriteVTTAbsPath(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	path := sp.SpriteVTTAbsPath("media-uuid-123")
	expected := filepath.Join(sp.SpritesDir, "media-uuid-123", "sprite.vtt")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestStaticRouteMap(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	routes := sp.StaticRouteMap()

	expectedRoutes := map[string]string{
		"/uploads":    sp.OriginalsDir,
		"/thumbnails": sp.ThumbnailsDir,
		"/hls":        sp.HLSDir,
		"/sprites":    sp.SpritesDir,
		"/previews":   sp.PreviewsDir,
	}

	for prefix, expectedDir := range expectedRoutes {
		if routes[prefix] != expectedDir {
			t.Errorf("route %s: expected %s, got %s", prefix, expectedDir, routes[prefix])
		}
	}
}

func TestRelativePaths(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"RelativeThumbnail", sp.RelativeThumbnail("uuid1"), "thumbnails/uuid1.jpg"},
		{"RelativeHLSMaster", sp.RelativeHLSMaster("uuid1"), "hls/uuid1/master.m3u8"},
		{"RelativeHLSProfile", sp.RelativeHLSProfile("uuid1", "720p"), "hls/uuid1/720p/index.m3u8"},
		{"RelativePreview", sp.RelativePreview("uuid1"), "previews/uuid1.gif"},
		{"RelativeSpriteImage", sp.RelativeSpriteImage("uuid1"), "sprites/uuid1/sprite.jpg"},
		{"RelativeSpriteVTT", sp.RelativeSpriteVTT("uuid1"), "sprites/uuid1/sprite.vtt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.got)
			}
		})
	}
}

func TestFullPathRoundTrip(t *testing.T) {
	// Verify that FullPath(RelativeXxx(...)) == XxxAbsPath(...)
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	// Thumbnail
	rel := sp.RelativeThumbnail("uuid1")
	abs := sp.FullPath(rel)
	expected := sp.ThumbnailAbsPath("uuid1")
	if abs != expected {
		t.Errorf("thumbnail round-trip: expected %s, got %s", expected, abs)
	}

	// Preview
	rel = sp.RelativePreview("uuid1")
	abs = sp.FullPath(rel)
	expected = sp.PreviewAbsPath("uuid1")
	if abs != expected {
		t.Errorf("preview round-trip: expected %s, got %s", expected, abs)
	}

	// Sprite image
	rel = sp.RelativeSpriteImage("uuid1")
	abs = sp.FullPath(rel)
	expected = sp.SpriteImageAbsPath("uuid1")
	if abs != expected {
		t.Errorf("sprite image round-trip: expected %s, got %s", expected, abs)
	}
}

func TestUserIDIsolation(t *testing.T) {
	tmpDir := t.TempDir()
	sp := NewStoragePaths(tmpDir)

	path1 := sp.OriginalPath("user_a", "file.mp4")
	path2 := sp.OriginalPath("user_b", "file.mp4")

	// Different users should have different paths
	if path1 == path2 {
		t.Error("different users should have different paths")
	}

	// Paths should contain respective user IDs
	if !strings.Contains(path1, "user_a") {
		t.Errorf("path1 should contain 'user_a': %s", path1)
	}
	if !strings.Contains(path2, "user_b") {
		t.Errorf("path2 should contain 'user_b': %s", path2)
	}
}
