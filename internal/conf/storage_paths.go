/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package conf

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// StoragePaths is a value object that derives all subdirectory paths from a
// single StorageBasePath. It provides user-isolation and date-sharded path
// generation for original and temporary files, and flat path generation for
// thumbnails, HLS, previews, and sprites.
type StoragePaths struct {
	basePath      string
	OriginalsDir  string
	TempDir       string
	ThumbnailsDir string
	HLSDir        string
	PreviewsDir   string
	SpritesDir    string
}

// NewStoragePaths creates a StoragePaths value object from a base path.
// It resolves the base path to an absolute path and ensures all
// subdirectories exist (fail-fast at startup).
func NewStoragePaths(basePath string) *StoragePaths {
	abs, err := filepath.Abs(basePath)
	if err != nil {
		abs = basePath
	}
	sp := &StoragePaths{
		basePath:      abs,
		OriginalsDir:  filepath.Join(abs, "originals"),
		TempDir:       filepath.Join(abs, "temp"),
		ThumbnailsDir: filepath.Join(abs, "thumbnails"),
		HLSDir:        filepath.Join(abs, "hls"),
		PreviewsDir:   filepath.Join(abs, "previews"),
		SpritesDir:    filepath.Join(abs, "sprites"),
	}
	if err := sp.EnsureDirs(); err != nil {
		panic(fmt.Sprintf("failed to create storage directories: %v", err))
	}
	return sp
}

// BasePath returns the resolved absolute base path.
func (sp *StoragePaths) BasePath() string { return sp.basePath }

// EnsureDirs creates all top-level subdirectories if they do not exist.
func (sp *StoragePaths) EnsureDirs() error {
	for _, dir := range []string{
		sp.OriginalsDir,
		sp.TempDir,
		sp.ThumbnailsDir,
		sp.HLSDir,
		sp.PreviewsDir,
		sp.SpritesDir,
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}
	return nil
}

// --- Absolute path generation (for filesystem operations) ---

// OriginalPath returns the absolute filesystem path for an original file.
// userID: the uploader's user ID
// filename: the file name including extension (e.g., "{uploadID}.mp4")
// The file will be stored at: originals/{userID}/{yyyy}/{MM}/{filename}
func (sp *StoragePaths) OriginalPath(userID, filename string) string {
	now := time.Now()
	return filepath.Join(
		sp.OriginalsDir,
		userID,
		fmt.Sprintf("%04d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		filename,
	)
}

// OriginalPathAt returns the absolute path for an original file at a specific time.
// Used when the upload time is known (e.g., from session creation time).
func (sp *StoragePaths) OriginalPathAt(userID, filename string, t time.Time) string {
	return filepath.Join(
		sp.OriginalsDir,
		userID,
		fmt.Sprintf("%04d", t.Year()),
		fmt.Sprintf("%02d", t.Month()),
		filename,
	)
}

// TempUploadDir returns the absolute path for an upload session's parts directory.
// Structure: temp/{userID}/{yyyy}/{MM}/{uploadID}/
func (sp *StoragePaths) TempUploadDir(userID, uploadID string) string {
	now := time.Now()
	return filepath.Join(
		sp.TempDir,
		userID,
		fmt.Sprintf("%04d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		uploadID,
	)
}

// TempUploadDirAt returns the parts directory at a specific time.
func (sp *StoragePaths) TempUploadDirAt(userID, uploadID string, t time.Time) string {
	return filepath.Join(
		sp.TempDir,
		userID,
		fmt.Sprintf("%04d", t.Year()),
		fmt.Sprintf("%02d", t.Month()),
		uploadID,
	)
}

// TempPartPath returns the absolute path for a specific part file.
func (sp *StoragePaths) TempPartPath(userID, uploadID string, partNumber int) string {
	return filepath.Join(sp.TempUploadDir(userID, uploadID), fmt.Sprintf("part_%05d", partNumber))
}

// TempMergedPath returns the absolute path for the merged file in temp.
// The merged file is a sibling of the parts directory, not inside it.
// Structure: temp/{userID}/{yyyy}/{MM}/{filename}
func (sp *StoragePaths) TempMergedPath(userID, filename string) string {
	now := time.Now()
	return filepath.Join(
		sp.TempDir,
		userID,
		fmt.Sprintf("%04d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		filename,
	)
}

// TempMergedPathAt returns the merged file path at a specific time.
func (sp *StoragePaths) TempMergedPathAt(userID, filename string, t time.Time) string {
	return filepath.Join(
		sp.TempDir,
		userID,
		fmt.Sprintf("%04d", t.Year()),
		fmt.Sprintf("%02d", t.Month()),
		filename,
	)
}

// ThumbnailAbsPath returns the absolute path for a thumbnail image.
func (sp *StoragePaths) ThumbnailAbsPath(mediaUUID string) string {
	return filepath.Join(sp.ThumbnailsDir, fmt.Sprintf("%s.jpg", mediaUUID))
}

// HLSProfileDir returns the absolute path for an HLS profile directory.
func (sp *StoragePaths) HLSProfileDir(mediaUUID, profileName string) string {
	return filepath.Join(sp.HLSDir, mediaUUID, profileName)
}

// HLSMasterAbsPath returns the absolute path for the HLS master playlist.
func (sp *StoragePaths) HLSMasterAbsPath(mediaUUID string) string {
	return filepath.Join(sp.HLSDir, mediaUUID, "master.m3u8")
}

// HLSDirForMedia returns the absolute HLS directory for a specific media UUID.
func (sp *StoragePaths) HLSDirForMedia(mediaUUID string) string {
	return filepath.Join(sp.HLSDir, mediaUUID)
}

// PreviewAbsPath returns the absolute path for a GIF preview.
func (sp *StoragePaths) PreviewAbsPath(mediaUUID string) string {
	return filepath.Join(sp.PreviewsDir, fmt.Sprintf("%s.gif", mediaUUID))
}

// SpriteDirAbs returns the absolute path for a sprite directory.
func (sp *StoragePaths) SpriteDirAbs(mediaUUID string) string {
	return filepath.Join(sp.SpritesDir, mediaUUID)
}

// SpriteImageAbsPath returns the absolute path for a sprite image.
func (sp *StoragePaths) SpriteImageAbsPath(mediaUUID string) string {
	return filepath.Join(sp.SpritesDir, mediaUUID, "sprite.jpg")
}

// SpriteVTTAbsPath returns the absolute path for a sprite VTT file.
func (sp *StoragePaths) SpriteVTTAbsPath(mediaUUID string) string {
	return filepath.Join(sp.SpritesDir, mediaUUID, "sprite.vtt")
}

// FullPath resolves a relative path (stored in DB) to an absolute filesystem path.
// Primary method for converting DB-stored paths to filesystem paths.
func (sp *StoragePaths) FullPath(relativePath string) string {
	return filepath.Join(sp.basePath, relativePath)
}

// --- Relative path generation (for storing in database / API responses) ---

// RelativeOriginal generates the relative path for an original file.
// Returns: originals/{userID}/{yyyy}/{MM}/{filename}
func (sp *StoragePaths) RelativeOriginal(userID, filename string) string {
	now := time.Now()
	return fmt.Sprintf("originals/%s/%04d/%02d/%s",
		userID, now.Year(), now.Month(), filename)
}

// RelativeOriginalAt generates the relative path for an original file at a specific time.
func (sp *StoragePaths) RelativeOriginalAt(userID, filename string, t time.Time) string {
	return fmt.Sprintf("originals/%s/%04d/%02d/%s",
		userID, t.Year(), t.Month(), filename)
}

// RelativeTemp generates the relative path for a merged file in temp.
func (sp *StoragePaths) RelativeTemp(userID, filename string) string {
	now := time.Now()
	return fmt.Sprintf("temp/%s/%04d/%02d/%s",
		userID, now.Year(), now.Month(), filename)
}

// RelativeTempAt generates the relative path for a merged file in temp at a specific time.
func (sp *StoragePaths) RelativeTempAt(userID, filename string, t time.Time) string {
	return fmt.Sprintf("temp/%s/%04d/%02d/%s",
		userID, t.Year(), t.Month(), filename)
}

// RelativeThumbnail generates the relative path for a thumbnail image.
func (sp *StoragePaths) RelativeThumbnail(mediaUUID string) string {
	return fmt.Sprintf("thumbnails/%s.jpg", mediaUUID)
}

// RelativeHLSMaster generates the relative path for the HLS master playlist.
func (sp *StoragePaths) RelativeHLSMaster(mediaUUID string) string {
	return fmt.Sprintf("hls/%s/master.m3u8", mediaUUID)
}

// RelativeHLSProfile generates the relative path for an HLS profile playlist.
func (sp *StoragePaths) RelativeHLSProfile(mediaUUID, profileName string) string {
	return fmt.Sprintf("hls/%s/%s/index.m3u8", mediaUUID, profileName)
}

// RelativePreview generates the relative path for a GIF preview.
func (sp *StoragePaths) RelativePreview(mediaUUID string) string {
	return fmt.Sprintf("previews/%s.gif", mediaUUID)
}

// RelativeSpriteImage generates the relative path for a sprite image.
func (sp *StoragePaths) RelativeSpriteImage(mediaUUID string) string {
	return fmt.Sprintf("sprites/%s/sprite.jpg", mediaUUID)
}

// RelativeSpriteVTT generates the relative path for a sprite VTT file.
func (sp *StoragePaths) RelativeSpriteVTT(mediaUUID string) string {
	return fmt.Sprintf("sprites/%s/sprite.vtt", mediaUUID)
}

// --- File promotion and cleanup ---

// PromoteToOriginal moves a file from temp/ to originals/ (same path structure).
// temp/{uid}/{yyyy}/{MM}/{filename} -> originals/{uid}/{yyyy}/{MM}/{filename}
// Returns the relative path of the promoted file.
func (sp *StoragePaths) PromoteToOriginal(userID, filename string) (string, error) {
	now := time.Now()
	yearMonth := fmt.Sprintf("%04d/%02d", now.Year(), now.Month())

	tempFile := filepath.Join(sp.TempDir, userID, yearMonth, filename)
	originalFile := filepath.Join(sp.OriginalsDir, userID, yearMonth, filename)

	if err := os.MkdirAll(filepath.Dir(originalFile), 0755); err != nil {
		return "", fmt.Errorf("create originals directory: %w", err)
	}

	if err := os.Rename(tempFile, originalFile); err != nil {
		// Cross-device link: fall back to copy+delete
		if err := copyFile(tempFile, originalFile); err != nil {
			return "", fmt.Errorf("copy temp to originals: %w", err)
		}
		os.Remove(tempFile)
	}

	return fmt.Sprintf("originals/%s/%s/%s", userID, yearMonth, filename), nil
}

// PromoteToOriginalAt moves a file from temp/ to originals/ at a specific time.
func (sp *StoragePaths) PromoteToOriginalAt(userID, filename string, t time.Time) (string, error) {
	yearMonth := fmt.Sprintf("%04d/%02d", t.Year(), t.Month())

	tempFile := filepath.Join(sp.TempDir, userID, yearMonth, filename)
	originalFile := filepath.Join(sp.OriginalsDir, userID, yearMonth, filename)

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

// CleanupTempParts removes the parts directory after promotion.
// Deletes: temp/{uid}/{yyyy}/{MM}/{uploadID}/ (parts only, merged file already moved)
func (sp *StoragePaths) CleanupTempParts(userID, uploadID string) error {
	now := time.Now()
	partsDir := filepath.Join(sp.TempDir, userID,
		fmt.Sprintf("%04d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		uploadID)
	return os.RemoveAll(partsDir)
}

// CleanupTempPartsAt removes the parts directory at a specific time.
func (sp *StoragePaths) CleanupTempPartsAt(userID, uploadID string, t time.Time) error {
	partsDir := filepath.Join(sp.TempDir, userID,
		fmt.Sprintf("%04d", t.Year()),
		fmt.Sprintf("%02d", t.Month()),
		uploadID)
	return os.RemoveAll(partsDir)
}

// --- Static route mapping (for server.go) ---

// StaticRouteMap returns a mapping of URL prefixes to filesystem directories
// for static file serving.
func (sp *StoragePaths) StaticRouteMap() map[string]string {
	return map[string]string{
		"/uploads":    sp.OriginalsDir,
		"/thumbnails": sp.ThumbnailsDir,
		"/hls":        sp.HLSDir,
		"/sprites":    sp.SpritesDir,
		"/previews":   sp.PreviewsDir,
	}
}

// copyFile copies a file from src to dst using a streaming copy.
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

	// Preserve permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}
