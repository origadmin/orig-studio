/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"strings"
	"testing"
	"time"

	origadmin_conf "origadmin/application/origstudio/internal/conf"
)

func TestAssetStorage_Avatar(t *testing.T) {
	tmpDir := t.TempDir()
	sp := origadmin_conf.NewStoragePaths(tmpDir)
	asset := NewAssetStorage(sp, "v1")

	refTime := time.Date(2026, 7, 31, 10, 0, 0, 0, time.UTC)
	got := asset.Avatar("user-123", refTime, "abc123")

	// Expected: assets/avatars/user-123/202607/abc123.jpg
	expected := "assets/avatars/user-123/202607/abc123.jpg"
	if got != expected {
		t.Errorf("Avatar: expected %q, got %q", expected, got)
	}
}

func TestAssetStorage_AllMethods(t *testing.T) {
	tmpDir := t.TempDir()
	sp := origadmin_conf.NewStoragePaths(tmpDir)
	asset := NewAssetStorage(sp, "v1")

	refTime := time.Date(2026, 7, 31, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Ad", asset.Ad(refTime, "hash1"), "assets/ads/202607/hash1.jpg"},
		{"Banner", asset.Banner(refTime, "hash2"), "assets/banners/202607/hash2.jpg"},
		{"MediaCover", asset.MediaCover("uuid-1"), "assets/covers/uuid-1.jpg"},
		{"Category", asset.Category("cat-1", "hash3"), "assets/categories/cat-1/hash3.jpg"},
		{"Channel", asset.Channel("ch-1", "avatar", "hash4"), "assets/channels/ch-1/avatar/hash4.jpg"},
		{"Article", asset.Article(refTime, "hash5"), "assets/articles/202607/hash5.jpg"},
		{"Misc", asset.Misc(refTime, "hash6"), "assets/misc/202607/hash6.jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: expected %q, got %q", tt.name, tt.expected, tt.got)
			}
		})
	}
}

func TestGalleryStorage_ImagePath(t *testing.T) {
	tmpDir := t.TempDir()
	sp := origadmin_conf.NewStoragePaths(tmpDir)
	gallery := NewGalleryStorage(sp, "v1")

	tests := []struct {
		name     string
		uuid     string
		idx      int
		isThumb  bool
		expected string
	}{
		{"image-0", "gallery-uuid", 0, false, "gallery/gallery-uuid/00.jpg"},
		{"image-1", "gallery-uuid", 1, false, "gallery/gallery-uuid/01.jpg"},
		{"thumb-0", "gallery-uuid", 0, true, "gallery/gallery-uuid/00_thumb.jpg"},
		{"thumb-5", "gallery-uuid", 5, true, "gallery/gallery-uuid/05_thumb.jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gallery.ImagePath(tt.uuid, tt.idx, tt.isThumb)
			if got != tt.expected {
				t.Errorf("ImagePath(%q, %d, %v): expected %q, got %q",
					tt.uuid, tt.idx, tt.isThumb, tt.expected, got)
			}
		})
	}
}

func TestGalleryStorage_CoverPath(t *testing.T) {
	tmpDir := t.TempDir()
	sp := origadmin_conf.NewStoragePaths(tmpDir)
	gallery := NewGalleryStorage(sp, "v1")

	got := gallery.CoverPath("gallery-uuid")
	expected := "gallery/gallery-uuid/cover.jpg"
	if got != expected {
		t.Errorf("CoverPath: expected %q, got %q", expected, got)
	}
}

// TestGalleryStorage_NoPanic verifies that GalleryStorage methods don't panic
// after NewStoragePaths (P0 fix: gallery must be auto-registered).
func TestGalleryStorage_NoPanic(t *testing.T) {
	tmpDir := t.TempDir()
	sp := origadmin_conf.NewStoragePaths(tmpDir)
	gallery := NewGalleryStorage(sp, "v1")

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GalleryStorage panicked: %v", r)
		}
	}()

	_ = gallery.ImagePath("uuid", 0, false)
	_ = gallery.CoverPath("uuid")
}

// TestRegisterGalleryPaths_Idempotent verifies that calling RegisterGalleryPaths
// after NewStoragePaths doesn't fail (gallery already auto-registered).
func TestRegisterGalleryPaths_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	sp := origadmin_conf.NewStoragePaths(tmpDir)

	// NewStoragePaths already registered gallery; calling again should succeed
	// (Register overwrites the existing entry).
	err := RegisterGalleryPaths(sp)
	if err != nil {
		t.Errorf("RegisterGalleryPaths after auto-register failed: %v", err)
	}

	// Verify gallery still works
	gallery := NewGalleryStorage(sp, "v1")
	rel := gallery.CoverPath("uuid")
	if !strings.HasPrefix(rel, "gallery/") {
		t.Errorf("gallery path after re-register: expected prefix 'gallery/', got %q", rel)
	}
}
