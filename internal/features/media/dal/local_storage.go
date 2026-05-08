/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"origadmin/application/origcms/internal/conf"
	"origadmin/application/origcms/internal/data/enums"
	"origadmin/application/origcms/internal/features/media/biz"
)

// LocalStorage implements local filesystem storage.
type LocalStorage struct {
	paths *conf.StoragePaths
}

// NewLocalStorage creates a new LocalStorage instance.
func NewLocalStorage(paths *conf.StoragePaths) *LocalStorage {
	return &LocalStorage{paths: paths}
}

// Paths returns the underlying StoragePaths (used by transcode handler for promotion).
func (s *LocalStorage) Paths() *conf.StoragePaths {
	return s.paths
}

// StorePart stores a single upload part.
func (s *LocalStorage) StorePart(ctx context.Context, uploadID string, partNumber int, data []byte) (string, error) {
	userID := biz.UserIDFromContext(ctx)
	tempDir := s.paths.TempUploadDir(userID, uploadID)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}

	partPath := s.paths.TempPartPath(userID, uploadID, partNumber)
	if err := os.WriteFile(partPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write part: %v", err)
	}

	// Calculate ETag
	hash := sha256.Sum256(data)
	etag := hex.EncodeToString(hash[:])

	return etag, nil
}

// MergeParts merges all parts into a single file at the specified relative path.
func (s *LocalStorage) MergeParts(ctx context.Context, uploadID string, totalParts int, finalPath string) error {
	userID := biz.UserIDFromContext(ctx)
	tempDir := s.paths.TempUploadDir(userID, uploadID)
	finalFilePath := s.paths.FullPath(finalPath)

	// Ensure target directory exists
	if err := os.MkdirAll(filepath.Dir(finalFilePath), 0755); err != nil {
		return fmt.Errorf("failed to create final directory: %v", err)
	}

	// Create the destination file
	dst, err := os.Create(finalFilePath)
	if err != nil {
		return fmt.Errorf("failed to create final file: %v", err)
	}
	defer dst.Close()

	// Read and write parts in order
	for i := 1; i <= totalParts; i++ {
		partPath := filepath.Join(tempDir, fmt.Sprintf("part_%05d", i))
		src, err := os.Open(partPath)
		if err != nil {
			return fmt.Errorf("failed to open part %d: %v", i, err)
		}

		if _, err := io.Copy(dst, src); err != nil {
			src.Close()
			return fmt.Errorf("failed to copy part %d: %v", i, err)
		}

		src.Close()
	}

	return nil
}

// DeleteParts removes the parts directory for an upload session.
func (s *LocalStorage) DeleteParts(ctx context.Context, uploadID string) error {
	userID := biz.UserIDFromContext(ctx)
	return os.RemoveAll(s.paths.TempUploadDir(userID, uploadID))
}

// GetFile reads a file by relative path.
func (s *LocalStorage) GetFile(ctx context.Context, path string) ([]byte, error) {
	return os.ReadFile(s.paths.FullPath(path))
}

// PutFile writes a file by relative path.
func (s *LocalStorage) PutFile(ctx context.Context, path string, data []byte) error {
	filePath := s.paths.FullPath(path)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

// DeleteFile removes a file by relative path.
func (s *LocalStorage) DeleteFile(ctx context.Context, path string) error {
	return os.Remove(s.paths.FullPath(path))
}

// Exists checks if a file exists by relative path.
func (s *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	_, err := os.Stat(s.paths.FullPath(path))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Upload uploads a file by key.
func (s *LocalStorage) Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	filePath := s.paths.FullPath(key)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", err
	}

	return key, nil
}

// Download downloads a file by key.
func (s *LocalStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return os.Open(s.paths.FullPath(key))
}

// Delete removes a file by key.
func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	return os.Remove(s.paths.FullPath(key))
}

// GetURL returns a URL for the given key.
func (s *LocalStorage) GetURL(ctx context.Context, key string) (string, error) {
	return "http://localhost:8080/" + key, nil
}

// PromoteToOriginal moves a file from temp/ to originals/ using StoragePaths.
// tempPath is a relative path like "temp/{userID}/{yyyy}/{MM}/{filename}".
// Returns the relative path of the promoted file in originals/.
func (s *LocalStorage) PromoteToOriginal(ctx context.Context, tempPath string) (string, error) {
	// Parse the relative path to extract userID and filename.
	// Expected format: temp/{userID}/{yyyy}/{MM}/{filename}
	parts := strings.SplitN(tempPath, "/", 5)
	if len(parts) < 5 || parts[0] != "temp" {
		return "", fmt.Errorf("invalid temp path format: %s", tempPath)
	}
	userID := parts[1]
	filename := parts[4]

	promotedPath, err := s.paths.PromoteToOriginal(userID, filename)
	if err != nil {
		return "", fmt.Errorf("promote to original: %w", err)
	}
	return promotedPath, nil
}

// CleanupTempParts removes the parts directory for an upload session.
// It delegates to StoragePaths.CleanupTempParts for the actual filesystem removal.
func (s *LocalStorage) CleanupTempParts(ctx context.Context, userID, uploadID string) error {
	return s.paths.CleanupTempParts(userID, uploadID)
}

// SyncStatus returns the sync status for a key.
// For local-only storage, files are always local_only (no remote sync).
func (s *LocalStorage) SyncStatus(ctx context.Context, key string) (enums.SyncStatus, error) {
	// Check if the file exists locally
	fullPath := s.paths.FullPath(key)
	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return enums.SyncStatusLocalOnly, nil
		}
		return enums.SyncStatusLocalOnly, fmt.Errorf("stat file: %w", err)
	}
	return enums.SyncStatusLocalOnly, nil
}
