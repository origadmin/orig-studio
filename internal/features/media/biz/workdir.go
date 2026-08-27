package biz

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-kratos/kratos/v2/log"
)

// LocalWorkDir represents a temporary local working directory for remote files.
// It provides cleanup functionality to remove the directory after processing.
type LocalWorkDir struct {
	LocalPath string
	dir       string
	logger    *log.Helper
}

// Cleanup removes the temporary working directory and all its contents.
func (w *LocalWorkDir) Cleanup() {
	if w == nil || w.dir == "" {
		return
	}
	if err := os.RemoveAll(w.dir); err != nil {
		w.logger.Warnf("failed to cleanup work dir %s: %v", w.dir, err)
	}
}

// DownloadToLocalWorkDir downloads a remote file to a temporary local directory.
// Returns a LocalWorkDir struct with the local path and cleanup functionality.
func DownloadToLocalWorkDir(ctx context.Context, storage Storage, remoteUrl string, fullPathFunc func(string) string) (*LocalWorkDir, error) {
	logger := log.NewHelper(log.With(log.DefaultLogger, "module", "media.workdir"))

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "media-work-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	// Determine local file path
	localPath := filepath.Join(tmpDir, filepath.Base(remoteUrl))

	// Download from storage
	reader, err := storage.Download(ctx, remoteUrl)
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("download from storage: %w", err)
	}
	defer reader.Close()

	// Write to local file
	file, err := os.Create(localPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("create local file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("write local file: %w", err)
	}

	return &LocalWorkDir{
		LocalPath: localPath,
		dir:       tmpDir,
		logger:    logger,
	}, nil
}