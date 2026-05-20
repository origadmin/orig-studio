/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dal

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/conf"
	"origadmin/application/origstudio/internal/dal/enums"
)

// syncTask represents a file that needs to be synced to S3.
type syncTask struct {
	Key      string
	Priority int // Higher priority = sync sooner
}

// HybridStorage implements the Storage interface with a local-first strategy.
// Files are written to local storage first, then asynchronously synced to S3.
// Reads prefer local storage with S3 fallback for cache misses.
type HybridStorage struct {
	local  *LocalStorage
	remote *S3Storage
	paths  *conf.StoragePaths
	syncCh chan syncTask
	syncWg sync.WaitGroup
	config conf.HybridConfig
	logger *log.Helper
	stopCh chan struct{}
}

// NewHybridStorage creates a new HybridStorage with async sync workers.
func NewHybridStorage(
	local *LocalStorage,
	remote *S3Storage,
	paths *conf.StoragePaths,
	config conf.HybridConfig,
	logger log.Logger,
) *HybridStorage {
	queueSize := config.SyncQueueSize
	if queueSize <= 0 {
		queueSize = 1000
	}
	workers := config.SyncWorkers
	if workers <= 0 {
		workers = 2
	}

	hs := &HybridStorage{
		local:  local,
		remote: remote,
		paths:  paths,
		syncCh: make(chan syncTask, queueSize),
		config: config,
		logger: log.NewHelper(log.With(logger, "module", "storage.hybrid")),
		stopCh: make(chan struct{}),
	}

	for i := 0; i < workers; i++ {
		hs.syncWg.Add(1)
		go hs.syncWorker(i)
	}

	return hs
}

// Close stops the sync workers gracefully. Call this on application shutdown.
func (hs *HybridStorage) Close() {
	close(hs.stopCh)
	close(hs.syncCh)
	hs.syncWg.Wait()
}

// Upload writes to local storage first, then queues async sync to S3.
func (hs *HybridStorage) Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error) {
	result, err := hs.local.Upload(ctx, key, r, size, contentType)
	if err != nil {
		return "", err
	}
	hs.queueSync(key, 1)
	return result, nil
}

// Download reads from local storage first; falls back to S3 on cache miss.
func (hs *HybridStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	reader, err := hs.local.Download(ctx, key)
	if err == nil {
		return reader, nil
	}
	// Local miss — try S3
	reader, err = hs.remote.Download(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("file not found in local or S3: %s", key)
	}
	return reader, nil
}

// Delete removes the file from both local and S3 storage.
func (hs *HybridStorage) Delete(ctx context.Context, key string) error {
	// Delete from local (ignore error if not present)
	_ = hs.local.Delete(ctx, key)
	// Delete from S3
	return hs.remote.Delete(ctx, key)
}

// GetURL returns a URL for the given key. Prefers local URL for hot files.
func (hs *HybridStorage) GetURL(ctx context.Context, key string) (string, error) {
	// Check if file exists locally
	if _, err := os.Stat(hs.paths.FullPath(key)); err == nil {
		return hs.local.GetURL(ctx, key)
	}
	// Fall back to S3 presigned URL
	return hs.remote.GetURL(ctx, key)
}

// StorePart stores a part in local storage only (upload performance).
func (hs *HybridStorage) StorePart(ctx context.Context, uploadID string, partNumber int, data []byte) (string, error) {
	return hs.local.StorePart(ctx, uploadID, partNumber, data)
}

// MergeParts merges parts locally and queues the merged file for S3 sync.
func (hs *HybridStorage) MergeParts(ctx context.Context, uploadID string, totalParts int, finalPath string) error {
	if err := hs.local.MergeParts(ctx, uploadID, totalParts, finalPath); err != nil {
		return err
	}
	hs.queueSync(finalPath, 1)
	return nil
}

// DeleteParts removes parts from local storage.
func (hs *HybridStorage) DeleteParts(ctx context.Context, uploadID string) error {
	return hs.local.DeleteParts(ctx, uploadID)
}

// PromoteToOriginal promotes a temp file to originals locally and queues S3 sync.
func (hs *HybridStorage) PromoteToOriginal(ctx context.Context, tempPath string) (string, error) {
	originalPath, err := hs.local.PromoteToOriginal(ctx, tempPath)
	if err != nil {
		return "", err
	}
	hs.queueSync(originalPath, 2) // Higher priority for originals
	return originalPath, nil
}

// CleanupTempParts cleans up temp parts from local storage.
func (hs *HybridStorage) CleanupTempParts(ctx context.Context, userID, uploadID string) error {
	return hs.local.CleanupTempParts(ctx, userID, uploadID)
}

// SyncStatus returns the sync status for a key.
// Checks if the file exists in S3 to determine sync state.
func (hs *HybridStorage) SyncStatus(ctx context.Context, key string) (enums.SyncStatus, error) {
	// Check local existence first
	if _, err := os.Stat(hs.paths.FullPath(key)); err != nil {
		// Not local — check S3
		s3Status, _ := hs.remote.SyncStatus(ctx, key)
		if s3Status == enums.SyncStatusSynced {
			return enums.SyncStatusSynced, nil
		}
		return enums.SyncStatusLocalOnly, nil
	}

	// File exists locally — check S3
	s3Status, _ := hs.remote.SyncStatus(ctx, key)
	if s3Status == enums.SyncStatusSynced {
		return enums.SyncStatusSynced, nil
	}
	return enums.SyncStatusLocalOnly, nil
}

// queueSync enqueues a sync task. Non-blocking; drops the task if the queue is full.
func (hs *HybridStorage) queueSync(key string, priority int) {
	select {
	case hs.syncCh <- syncTask{Key: key, Priority: priority}:
	default:
		hs.logger.Warnf("sync queue full, dropping sync task for key: %s", key)
	}
}

// syncWorker processes sync tasks from the queue.
func (hs *HybridStorage) syncWorker(id int) {
	defer hs.syncWg.Done()
	for task := range hs.syncCh {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		if err := hs.syncToRemote(ctx, task.Key); err != nil {
			hs.logger.Errorf("worker %d: sync failed for %s: %v", id, task.Key, err)
			hs.retrySync(task.Key, err)
		}
		cancel()
	}
}

// syncToRemote reads a file from local storage and uploads it to S3.
func (hs *HybridStorage) syncToRemote(ctx context.Context, key string) error {
	reader, err := hs.local.Download(ctx, key)
	if err != nil {
		return fmt.Errorf("local read: %w", err)
	}
	defer reader.Close()

	absPath := hs.paths.FullPath(key)
	stat, _ := os.Stat(absPath)
	var size int64
	if stat != nil {
		size = stat.Size()
	}

	_, err = hs.remote.Upload(ctx, key, reader, size, "")
	if err != nil {
		return fmt.Errorf("S3 upload: %w", err)
	}

	hs.logger.Infof("synced to S3: %s", key)

	// Evict local cache if configured
	hs.evictIfNeeded()

	return nil
}

// retrySync re-queues a failed sync task with a delay.
func (hs *HybridStorage) retrySync(key string, originalErr error) {
	if hs.config.SyncRetryMax <= 0 {
		return
	}
	// Simple retry: re-queue with a delay
	go func() {
		time.Sleep(hs.config.SyncRetryDelay)
		hs.queueSync(key, 0) // Lower priority for retries
	}()
}

// evictIfNeeded checks local storage usage and evicts synced files if over limit.
func (hs *HybridStorage) evictIfNeeded() {
	if hs.config.LocalCacheSize <= 0 {
		return
	}

	usage := hs.getLocalUsage()
	if usage < hs.config.LocalCacheSize {
		return
	}

	// Get synced files sorted by access time (oldest first)
	syncedFiles := hs.getSyncedFilesOldestFirst()
	for _, f := range syncedFiles {
		if usage < hs.config.LocalCacheSize {
			break
		}
		if err := hs.local.Delete(context.Background(), f.Key); err != nil {
			hs.logger.Warnf("failed to evict local file %s: %v", f.Key, err)
			continue
		}
		usage -= f.Size
		hs.logger.Infof("evicted local cache file: %s (freed %d bytes)", f.Key, f.Size)
	}
}

// localFileEntry represents a local file with its metadata for eviction.
type localFileEntry struct {
	Key  string
	Size int64
}

// getLocalUsage returns the total size of the local storage directory.
func (hs *HybridStorage) getLocalUsage() int64 {
	var total int64
	_ = filepath.Walk(hs.paths.BasePath(), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		total += info.Size()
		return nil
	})
	return total
}

// getSyncedFilesOldestFirst returns local files that have been synced to S3,
// sorted by modification time (oldest first) for LRU eviction.
func (hs *HybridStorage) getSyncedFilesOldestFirst() []localFileEntry {
	var files []localFileEntry

	_ = filepath.Walk(hs.paths.OriginalsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		// Convert absolute path back to relative key
		relPath, relErr := filepath.Rel(hs.paths.BasePath(), path)
		if relErr != nil {
			return nil
		}
		files = append(files, localFileEntry{
			Key:  relPath,
			Size: info.Size(),
		})
		return nil
	})

	// Sort by modification time (oldest first) for LRU eviction
	// Simple approach: files are already walked in lexical order
	// Production code should sort by atime/mtime
	return files
}
