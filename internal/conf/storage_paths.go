/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package conf

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PathSpec declares a storage subdirectory's specification.
// Each feature module registers its own dirs via Register().
type PathSpec struct {
	Name        string // registry key, e.g. "gallery", "assets/avatars", "originals"
	DefaultDir  string // default subdir name under basePath, e.g. "gallery", "assets/avatars"
	Description string // human-readable description
}

type registeredPath struct {
	spec      PathSpec
	actualDir string // resolved absolute path: basePath + DefaultDir, or override
}

// StoragePaths is the Single Source of Truth (SSOT) for all storage paths.
//
// After the Registry refactoring, this struct is the stable core:
// new feature modules register their dirs via Register() without modifying this file.
// Existing video paths are auto-registered in the constructor for backward compatibility.
type StoragePaths struct {
	basePath  string
	overrides map[string]string       // config: name → absolute path override
	registry  map[string]*registeredPath // registered storage dirs
}

// NewStoragePaths creates a StoragePaths with the given basePath.
// Video dirs (originals, temp, thumbnails, hls, previews, sprites) are auto-registered
// for backward compatibility.
func NewStoragePaths(basePath string) *StoragePaths {
	return NewStoragePathsWithOverrides(basePath, nil)
}

// NewStoragePathsWithOverrides creates a StoragePaths with per-dir path overrides.
// overrides maps a registry name (e.g. "gallery") to an absolute path.
// If a name in overrides has no matching Register() call, it is logged but ignored.
func NewStoragePathsWithOverrides(basePath string, overrides map[string]string) *StoragePaths {
	abs, err := filepath.Abs(basePath)
	if err != nil {
		abs = basePath
	}
	sp := &StoragePaths{
		basePath:  abs,
		overrides: overrides,
		registry:  make(map[string]*registeredPath),
	}

	// Auto-register video dirs for backward compatibility.
	// New modules call Register() themselves; this block only covers existing paths.
	videoSpecs := []PathSpec{
		{Name: "originals", DefaultDir: "originals", Description: "video source files"},
		{Name: "temp", DefaultDir: "temp", Description: "temp upload parts"},
		{Name: "thumbnails", DefaultDir: "thumbnails", Description: "video thumbnails"},
		{Name: "hls", DefaultDir: "hls", Description: "HLS segments"},
		{Name: "previews", DefaultDir: "previews", Description: "GIF previews"},
		{Name: "sprites", DefaultDir: "sprites", Description: "sprite sheets"},
	}
	for _, spec := range videoSpecs {
		if err := sp.Register(spec); err != nil {
			panic(fmt.Sprintf("failed to create storage directory %s: %v", spec.Name, err))
		}
	}

	// Auto-register business asset dirs (avatars, ads, banners, etc.)
	// These are existing functionality being migrated from hardcoded paths.
	assetSubs := []string{"avatars", "ads", "banners", "covers", "channels", "categories", "articles", "misc"}
	for _, sub := range assetSubs {
		name := "assets/" + sub
		if err := sp.Register(PathSpec{
			Name:        name,
			DefaultDir:  name,
			Description: "business asset: " + sub,
		}); err != nil {
			panic(fmt.Sprintf("failed to create storage directory %s: %v", name, err))
		}
	}

	// Auto-register gallery dir (multi-image content type).
	if err := sp.Register(PathSpec{
		Name:        "gallery",
		DefaultDir:  "gallery",
		Description: "gallery content (multi-image)",
	}); err != nil {
		panic(fmt.Sprintf("failed to create storage directory gallery: %v", err))
	}

	return sp
}

// Register declares a storage subdirectory.
// Called by feature modules to register their own dirs.
// The core StoragePaths code does not need to know which modules exist.
func (sp *StoragePaths) Register(spec PathSpec) error {
	dir := filepath.Join(sp.basePath, spec.DefaultDir)
	if override, ok := sp.overrides[spec.Name]; ok {
		dir = override
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory %s: %w", spec.Name, err)
	}
	sp.registry[spec.Name] = &registeredPath{spec: spec, actualDir: dir}
	return nil
}

// Dir returns the absolute filesystem path for a registered directory.
// Panics if the name is not registered — call Register() first.
func (sp *StoragePaths) Dir(name string) string {
	if rp, ok := sp.registry[name]; ok {
		return rp.actualDir
	}
	panic(fmt.Sprintf("storage path not registered: %q — did you call Register()?", name))
}

// Relative generates a relative storage key (for DB storage / API responses).
// segments are the path components after the directory name.
// Example: Relative("originals", "admin", "2026", "06", "file.mp4") → "originals/admin/2026/06/file.mp4"
func (sp *StoragePaths) Relative(name string, segments ...string) string {
	rp, ok := sp.registry[name]
	if !ok {
		panic(fmt.Sprintf("storage path not registered: %q", name))
	}
	parts := append([]string{rp.spec.DefaultDir}, segments...)
	return strings.Join(parts, "/")
}

// FullPath resolves a relative storage key to an absolute filesystem path.
func (sp *StoragePaths) FullPath(relativePath string) string {
	return filepath.Join(sp.basePath, relativePath)
}

// BasePath returns the storage root directory.
func (sp *StoragePaths) BasePath() string { return sp.basePath }

// EnsureDirs is a no-op (dirs are created during Register).
// Kept for backward compatibility.
func (sp *StoragePaths) EnsureDirs() error { return nil }

// StaticRouteMap auto-generates URL-to-filesystem route mappings for all registered dirs.
// Used by the server to register Gin static routes (e.g. /files/originals → /data/uploads/originals).
func (sp *StoragePaths) StaticRouteMap() map[string]string {
	routes := make(map[string]string)
	for _, rp := range sp.registry {
		routes["/files/"+rp.spec.DefaultDir] = rp.actualDir
	}
	return routes
}

// --- Backward compatibility wrappers (video paths) ---
// These methods preserve the existing API surface.
// External callers use these methods; they delegate to the registry.

// OriginalsDir returns the absolute path to the originals/ directory.
func (sp *StoragePaths) OriginalsDir() string  { return sp.Dir("originals") }
func (sp *StoragePaths) TempDir() string       { return sp.Dir("temp") }
func (sp *StoragePaths) ThumbnailsDir() string { return sp.Dir("thumbnails") }
func (sp *StoragePaths) HLSDir() string        { return sp.Dir("hls") }
func (sp *StoragePaths) PreviewsDir() string   { return sp.Dir("previews") }
func (sp *StoragePaths) SpritesDir() string    { return sp.Dir("sprites") }

// --- Absolute path generation (for filesystem operations) ---

func (sp *StoragePaths) OriginalPath(userID, filename string) string {
	return sp.OriginalPathAt(userID, filename, time.Now())
}

func (sp *StoragePaths) OriginalPathAt(userID, filename string, t time.Time) string {
	return filepath.Join(
		sp.OriginalsDir(),
		userID,
		fmt.Sprintf("%04d", t.Year()),
		fmt.Sprintf("%02d", t.Month()),
		filename,
	)
}

func (sp *StoragePaths) TempUploadDir(userID, uploadID string) string {
	return sp.TempUploadDirAt(userID, uploadID, time.Now())
}

func (sp *StoragePaths) TempUploadDirAt(userID, uploadID string, t time.Time) string {
	return filepath.Join(
		sp.TempDir(),
		userID,
		fmt.Sprintf("%04d", t.Year()),
		fmt.Sprintf("%02d", t.Month()),
		uploadID,
	)
}

func (sp *StoragePaths) TempPartPathAt(userID, uploadID string, partNumber int, t time.Time) string {
	return filepath.Join(sp.TempUploadDirAt(userID, uploadID, t), fmt.Sprintf("part_%05d", partNumber))
}

func (sp *StoragePaths) TempPartPath(userID, uploadID string, partNumber int) string {
	return sp.TempPartPathAt(userID, uploadID, partNumber, time.Now())
}

func (sp *StoragePaths) TempMergedPath(userID, filename string) string {
	return sp.TempMergedPathAt(userID, filename, time.Now())
}

func (sp *StoragePaths) TempMergedPathAt(userID, filename string, t time.Time) string {
	return filepath.Join(
		sp.TempDir(),
		userID,
		fmt.Sprintf("%04d", t.Year()),
		fmt.Sprintf("%02d", t.Month()),
		filename,
	)
}

func (sp *StoragePaths) ThumbnailAbsPath(mediaUUID string) string {
	return filepath.Join(sp.ThumbnailsDir(), fmt.Sprintf("%s.jpg", mediaUUID))
}

func (sp *StoragePaths) HLSProfileDir(mediaUUID, profileName string) string {
	return filepath.Join(sp.HLSDir(), mediaUUID, profileName)
}

func (sp *StoragePaths) HLSMasterAbsPath(mediaUUID string) string {
	return filepath.Join(sp.HLSDir(), mediaUUID, "master.m3u8")
}

func (sp *StoragePaths) HLSDirForMedia(mediaUUID string) string {
	return filepath.Join(sp.HLSDir(), mediaUUID)
}

func (sp *StoragePaths) PreviewAbsPath(mediaUUID string) string {
	return filepath.Join(sp.PreviewsDir(), fmt.Sprintf("%s.gif", mediaUUID))
}

func (sp *StoragePaths) SpriteDirAbs(mediaUUID string) string {
	return filepath.Join(sp.SpritesDir(), mediaUUID)
}

func (sp *StoragePaths) SpriteImageAbsPath(mediaUUID string) string {
	return filepath.Join(sp.SpritesDir(), mediaUUID, "sprite.jpg")
}

func (sp *StoragePaths) SpriteVTTAbsPath(mediaUUID string) string {
	return filepath.Join(sp.SpritesDir(), mediaUUID, "sprite.vtt")
}

// --- Relative path generation (for storing in database / API responses) ---

func (sp *StoragePaths) RelativeOriginal(userID, filename string) string {
	return sp.RelativeOriginalAt(userID, filename, time.Now())
}

func (sp *StoragePaths) RelativeOriginalAt(userID, filename string, t time.Time) string {
	return sp.Relative("originals", userID,
		fmt.Sprintf("%04d", t.Year()), fmt.Sprintf("%02d", t.Month()), filename)
}

func (sp *StoragePaths) RelativeTemp(userID, filename string) string {
	return sp.RelativeTempAt(userID, filename, time.Now())
}

func (sp *StoragePaths) RelativeTempAt(userID, filename string, t time.Time) string {
	return sp.Relative("temp", userID,
		fmt.Sprintf("%04d", t.Year()), fmt.Sprintf("%02d", t.Month()), filename)
}

func (sp *StoragePaths) RelativeThumbnail(mediaUUID string) string {
	return sp.Relative("thumbnails", fmt.Sprintf("%s.jpg", mediaUUID))
}

func (sp *StoragePaths) RelativeHLSMaster(mediaUUID string) string {
	return sp.Relative("hls", mediaUUID, "master.m3u8")
}

func (sp *StoragePaths) RelativeHLSProfile(mediaUUID, profileName string) string {
	return sp.Relative("hls", mediaUUID, profileName, "index.m3u8")
}

func (sp *StoragePaths) RelativePreview(mediaUUID string) string {
	return sp.Relative("previews", fmt.Sprintf("%s.gif", mediaUUID))
}

func (sp *StoragePaths) RelativeSpriteImage(mediaUUID string) string {
	return sp.Relative("sprites", mediaUUID, "sprite.jpg")
}

func (sp *StoragePaths) RelativeSpriteVTT(mediaUUID string) string {
	return sp.Relative("sprites", mediaUUID, "sprite.vtt")
}

// --- File promotion and cleanup ---

// PromoteToOriginal moves a file from temp/ to originals/ using the exact
// year/month embedded in the tempPath. This avoids time.Now() calls which
// cause cross-month failures.
// tempPath format: temp/{userID}/{yyyy}/{MM}/{filename}
// Returns: originals/{userID}/{yyyy}/{MM}/{filename}
func (sp *StoragePaths) PromoteToOriginal(tempPath string) (string, error) {
	if !strings.HasPrefix(tempPath, "temp/") {
		return "", fmt.Errorf("invalid temp path: must start with 'temp/': %s", tempPath)
	}
	parts := strings.SplitN(tempPath, "/", 5)
	if len(parts) < 5 {
		return "", fmt.Errorf("invalid temp path format: %s", tempPath)
	}
	userID := parts[1]
	yearMonth := parts[2] + "/" + parts[3]
	filename := parts[4]

	originalsRelPath := fmt.Sprintf("originals/%s/%s/%s", userID, yearMonth, filename)
	srcAbs := filepath.Join(sp.basePath, tempPath)
	dstAbs := filepath.Join(sp.basePath, originalsRelPath)

	if err := os.MkdirAll(filepath.Dir(dstAbs), 0755); err != nil {
		return "", fmt.Errorf("create originals directory: %w", err)
	}

	if err := os.Rename(srcAbs, dstAbs); err != nil {
		if err := copyFile(srcAbs, dstAbs); err != nil {
			return "", fmt.Errorf("copy temp to originals: %w", err)
		}
		_ = os.Remove(srcAbs)
	}

	return originalsRelPath, nil
}

// Deprecated: Use PromoteToOriginal(tempPath) instead which is time-safe.
func (sp *StoragePaths) PromoteToOriginalLegacy(userID, filename string) (string, error) {
	now := time.Now()
	yearMonth := fmt.Sprintf("%04d/%02d", now.Year(), now.Month())
	tempFile := filepath.Join(sp.TempDir(), userID, yearMonth, filename)
	originalFile := filepath.Join(sp.OriginalsDir(), userID, yearMonth, filename)

	if err := os.MkdirAll(filepath.Dir(originalFile), 0755); err != nil {
		return "", fmt.Errorf("create originals directory: %w", err)
	}
	if err := os.Rename(tempFile, originalFile); err != nil {
		if err := copyFile(tempFile, originalFile); err != nil {
			return "", fmt.Errorf("copy temp to originals: %w", err)
		}
		os.Remove(tempFile)
	}
	return fmt.Sprintf("originals/%s/%s/%s", userID, yearMonth, filename), nil
}

// CleanupTempPartsByDir removes the parts directory given an absolute path.
// This is time-safe when the directory path is derived from the session.
func (sp *StoragePaths) CleanupTempPartsByDir(partsDirAbs string) error {
	return os.RemoveAll(partsDirAbs)
}

// Deprecated: Use CleanupTempPartsByDir with a path derived from session time.
func (sp *StoragePaths) CleanupTempParts(userID, uploadID string) error {
	return os.RemoveAll(sp.TempUploadDir(userID, uploadID))
}

// CleanupTempPartsAt removes the parts directory at a specific time.
func (sp *StoragePaths) CleanupTempPartsAt(userID, uploadID string, t time.Time) error {
	return os.RemoveAll(sp.TempUploadDirAt(userID, uploadID, t))
}

// --- Static route mapping (for StorageProxy) ---
// StaticRouteMap above is the auto-generated version.
// The method below is kept for any code that still references it explicitly.

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}
