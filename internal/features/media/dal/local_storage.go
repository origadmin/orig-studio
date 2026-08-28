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
	"time"

	"origadmin/application/origstudio/internal/conf"
	"origadmin/application/origstudio/internal/data/enums"
	"origadmin/application/origstudio/internal/features/media/biz"
)

var publicBaseURL = func() string {
	if v := os.Getenv("PUBLIC_BASE_URL"); v != "" {
		return v
	}
	if v := os.Getenv("CDN_BASE_URL"); v != "" {
		return v
	}
	return "/files"
}()

type LocalStorage struct {
	paths     *conf.StoragePaths
	urlPrefix string
}

func NewLocalStorage(paths *conf.StoragePaths) *LocalStorage {
	return &LocalStorage{
		paths:     paths,
		urlPrefix: publicBaseURL,
	}
}

func (s *LocalStorage) Paths() *conf.StoragePaths {
	return s.paths
}

func sessionTime(ctx context.Context) time.Time {
	if t, ok := biz.SessionCreateTimeFromContext(ctx); ok && !t.IsZero() {
		return t
	}
	return time.Now()
}

func (s *LocalStorage) tempUploadDir(ctx context.Context, userID, uploadID string) string {
	return s.paths.TempUploadDirAt(userID, uploadID, sessionTime(ctx))
}

func (s *LocalStorage) tempPartPath(ctx context.Context, userID, uploadID string, partNumber int) string {
	return s.paths.TempPartPathAt(userID, uploadID, partNumber, sessionTime(ctx))
}

// StorePart stores a single upload part using streaming I/O.
func (s *LocalStorage) StorePart(ctx context.Context, uploadID string, partNumber int, r io.Reader, size int64) (string, error) {
	userID := biz.UserIDFromContext(ctx)
	tempDir := s.tempUploadDir(ctx, userID, uploadID)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}

	partPath := s.tempPartPath(ctx, userID, uploadID, partNumber)
	f, err := os.Create(partPath)
	if err != nil {
		return "", fmt.Errorf("failed to create part file: %v", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return "", fmt.Errorf("failed to write part: %v", err)
	}

	etag, err := computeEtag(partPath)
	if err != nil {
		return "", fmt.Errorf("failed to compute etag: %v", err)
	}
	return etag, nil
}

func computeEtag(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	buf := make([]byte, 65536)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return "", err
	}
	if n > 0 {
		_, _ = h.Write(buf[:n])
	}
	fi, err := f.Stat()
	if err == nil && fi.Size() > 131072 {
		offset := fi.Size() - 65536
		_, err := f.Seek(offset, io.SeekStart)
		if err == nil {
			n2, err := io.ReadFull(f, buf)
			if err == nil || err == io.ErrUnexpectedEOF {
				if n2 > 0 {
					_, _ = h.Write(buf[:n2])
				}
			}
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// MergeParts merges all parts into a single file at the specified relative path.
func (s *LocalStorage) MergeParts(ctx context.Context, uploadID string, totalParts int, finalPath string) error {
	userID := biz.UserIDFromContext(ctx)
	tempDir := s.tempUploadDir(ctx, userID, uploadID)
	finalFilePath := s.paths.FullPath(finalPath)

	if err := os.MkdirAll(filepath.Dir(finalFilePath), 0755); err != nil {
		return fmt.Errorf("failed to create final directory: %v", err)
	}

	dst, err := os.Create(finalFilePath)
	if err != nil {
		return fmt.Errorf("failed to create final file: %v", err)
	}
	defer dst.Close()

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
	return os.RemoveAll(s.tempUploadDir(ctx, userID, uploadID))
}

func (s *LocalStorage) GetFile(ctx context.Context, path string) ([]byte, error) {
	return os.ReadFile(s.paths.FullPath(path))
}

func (s *LocalStorage) PutFile(ctx context.Context, path string, data []byte) error {
	filePath := s.paths.FullPath(path)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	return os.WriteFile(filePath, data, 0644)
}

func (s *LocalStorage) DeleteFile(ctx context.Context, path string) error {
	return os.Remove(s.paths.FullPath(path))
}

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

// Upload uploads a file by key using streaming I/O (no full buffering in memory).
func (s *LocalStorage) Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error) {
	filePath := s.paths.FullPath(key)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	// Skip if reader is a file pointing to the same destination path (already in place)
	// This prevents os.Create from truncating the source file before io.Copy reads it
	if f, ok := r.(*os.File); ok {
		if absPath, absErr := filepath.Abs(f.Name()); absErr == nil {
			if absDest, destErr := filepath.Abs(filePath); destErr == nil && absPath == absDest {
				return key, nil
			}
		}
	}

	f, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}

	return key, nil
}

func (s *LocalStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return os.Open(s.paths.FullPath(key))
}

func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	return os.Remove(s.paths.FullPath(key))
}

func (s *LocalStorage) GetURL(ctx context.Context, key string) (string, error) {
	return s.urlPrefix + "/" + key, nil
}

// PromoteToOriginal moves a file from temp/ to originals/ by replacing the path prefix.
// tempPath format: temp/{userID}/{yyyy}/{MM}/{filename}
// Returns: originals/{userID}/{yyyy}/{MM}/{filename}
func (s *LocalStorage) PromoteToOriginal(ctx context.Context, tempPath string) (string, error) {
	if !strings.HasPrefix(tempPath, "temp/") {
		return "", fmt.Errorf("invalid temp path: must start with 'temp/': %s", tempPath)
	}
	return s.paths.PromoteToOriginal(tempPath)
}

func (s *LocalStorage) CleanupTempParts(ctx context.Context, userID, uploadID string) error {
	return os.RemoveAll(s.tempUploadDir(ctx, userID, uploadID))
}

func (s *LocalStorage) SyncStatus(ctx context.Context, key string) (enums.SyncStatus, error) {
	fullPath := s.paths.FullPath(key)
	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return enums.SyncStatusLocalOnly, nil
		}
		return enums.SyncStatusLocalOnly, fmt.Errorf("stat file: %w", err)
	}
	return enums.SyncStatusLocalOnly, nil
}

func (s *LocalStorage) DownloadToFile(ctx context.Context, key string, localPath string) error {
	srcPath := s.paths.FullPath(key)
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return fmt.Errorf("mkdir for download key=%s: %w", key, err)
	}
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open source key=%s: %w", key, err)
	}
	defer src.Close()
	dst, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("create local file key=%s: %w", key, err)
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("copy key=%s to %s: %w", key, localPath, err)
	}
	return nil
}

func (s *LocalStorage) UploadDir(ctx context.Context, localDir string, keyPrefix string) error {
	return filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(localDir, path)
		if err != nil {
			return err
		}
		destKey := keyPrefix + "/" + filepath.ToSlash(rel)
		destPath := s.paths.FullPath(destKey)
		// Skip if source and destination are the same file (already in place)
		if absPath, absErr := filepath.Abs(path); absErr == nil {
			if absDest, destErr := filepath.Abs(destPath); destErr == nil && absPath == absDest {
				return nil
			}
		}
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("mkdir for upload %s: %w", destKey, err)
		}
		src, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open %s: %w", path, err)
		}
		defer src.Close()
		dst, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("create %s: %w", destKey, err)
		}
		defer dst.Close()
		if _, err := io.Copy(dst, src); err != nil {
			return fmt.Errorf("copy %s → %s: %w", path, destKey, err)
		}
		return nil
	})
}

func (s *LocalStorage) DeletePrefix(ctx context.Context, keyPrefix string) error {
	return os.RemoveAll(s.paths.FullPath(keyPrefix))
}
