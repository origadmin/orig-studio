/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"fmt"
	"time"

	origadmin_conf "origadmin/application/origstudio/internal/conf"
)

// MediaStorage is the stable contract for media (video) storage path generation.
// Callers depend on this interface, never on concrete implementations.
// Adding new methods is backward-compatible; changing signatures is not.
type MediaStorage interface {
	// Original returns the relative key for a video source file.
	Original(userID, filename string, t time.Time) string
	// Thumbnail returns the relative key for a video thumbnail.
	Thumbnail(uuid string) string
	// HLSMaster returns the relative key for the HLS master playlist.
	HLSMaster(uuid string) string
	// HLSProfile returns the relative key for a specific HLS profile playlist.
	HLSProfile(uuid, profile string) string
	// Preview returns the relative key for a GIF preview.
	Preview(uuid string) string
	// SpriteImage returns the relative key for the sprite sheet image.
	SpriteImage(uuid string) string
	// SpriteVTT returns the relative key for the sprite VTT file.
	SpriteVTT(uuid string) string
	// Temp returns the relative key for a temp upload file.
	Temp(userID, filename string, t time.Time) string
	// OriginalsDir returns the absolute path to the originals directory (for filesystem ops).
	OriginalsDir() string
	// ThumbnailsDir returns the absolute path to the thumbnails directory.
	ThumbnailsDir() string
	// HLSDir returns the absolute path to the HLS directory.
	HLSDir() string
	// PreviewsDir returns the absolute path to the previews directory.
	PreviewsDir() string
	// SpritesDir returns the absolute path to the sprites directory.
	SpritesDir() string
}

// mediaStorageV1 is the default implementation that delegates to StoragePaths.
// It exists alongside the interface so that future V2 implementations
// (e.g. hash-based sharding) can be added without changing callers.
type mediaStorageV1 struct {
	sp *origadmin_conf.StoragePaths
}

// NewMediaStorage creates a MediaStorage implementation.
// strategy selects the implementation version ("v1" default, "v2" future).
func NewMediaStorage(sp *origadmin_conf.StoragePaths, strategy string) MediaStorage {
	switch strategy {
	// case "v2": return &mediaStorageV2{sp: sp}  // future
	default:
		return &mediaStorageV1{sp: sp}
	}
}

func (m *mediaStorageV1) Original(userID, filename string, t time.Time) string {
	return m.sp.RelativeOriginalAt(userID, filename, t)
}

func (m *mediaStorageV1) Thumbnail(uuid string) string {
	return m.sp.RelativeThumbnail(uuid)
}

func (m *mediaStorageV1) HLSMaster(uuid string) string {
	return m.sp.RelativeHLSMaster(uuid)
}

func (m *mediaStorageV1) HLSProfile(uuid, profile string) string {
	return m.sp.RelativeHLSProfile(uuid, profile)
}

func (m *mediaStorageV1) Preview(uuid string) string {
	return m.sp.RelativePreview(uuid)
}

func (m *mediaStorageV1) SpriteImage(uuid string) string {
	return m.sp.RelativeSpriteImage(uuid)
}

func (m *mediaStorageV1) SpriteVTT(uuid string) string {
	return m.sp.RelativeSpriteVTT(uuid)
}

func (m *mediaStorageV1) Temp(userID, filename string, t time.Time) string {
	return m.sp.RelativeTempAt(userID, filename, t)
}

func (m *mediaStorageV1) OriginalsDir() string  { return m.sp.OriginalsDir() }
func (m *mediaStorageV1) ThumbnailsDir() string { return m.sp.ThumbnailsDir() }
func (m *mediaStorageV1) HLSDir() string        { return m.sp.HLSDir() }
func (m *mediaStorageV1) PreviewsDir() string   { return m.sp.PreviewsDir() }
func (m *mediaStorageV1) SpritesDir() string    { return m.sp.SpritesDir() }

// RegisterVideoPaths registers the 6 video storage directories.
// Note: NewStoragePaths already auto-registers these for backward compatibility,
// so this function is a no-op when called after NewStoragePaths.
// It exists for explicit registration clarity and future modules that may
// use a fresh StoragePaths without auto-registration.
func RegisterVideoPaths(sp *origadmin_conf.StoragePaths) error {
	// Video dirs are already auto-registered by NewStoragePaths.
	// This is a placeholder for explicit registration when auto-registration is removed.
	return nil
}

// formatYearMonth returns "yyyy/MM" for path construction.
func formatYearMonth(t time.Time) string {
	return fmt.Sprintf("%04d/%02d", t.Year(), t.Month())
}
