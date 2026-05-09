/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package conf

import (
	"time"
)

// StorageType defines the type of storage backend.
type StorageType string

const (
	StorageTypeLocal  StorageType = "local"
	StorageTypeS3     StorageType = "s3"
	StorageTypeHybrid StorageType = "hybrid"
)

// StorageConfig holds the configuration for the storage subsystem.
// It determines which storage backend to use and provides
// backend-specific configuration blocks.
type StorageConfig struct {
	Type     StorageType  `json:"type" yaml:"type"`
	BasePath string       `json:"base_path" yaml:"base_path"`
	S3       S3Config     `json:"s3" yaml:"s3"`
	Hybrid   HybridConfig `json:"hybrid" yaml:"hybrid"`
	Temp     TempConfig   `json:"temp" yaml:"temp"`
}

// S3Config holds the configuration for S3/MinIO object storage.
type S3Config struct {
	Endpoint      string        `json:"endpoint" yaml:"endpoint"`
	Region        string        `json:"region" yaml:"region"`
	Bucket        string        `json:"bucket" yaml:"bucket"`
	AccessKey     string        `json:"access_key" yaml:"access_key"`
	SecretKey     string        `json:"secret_key" yaml:"secret_key"`
	UsePathStyle  bool          `json:"use_path_style" yaml:"use_path_style"`
	PresignExpiry time.Duration `json:"presign_expiry" yaml:"presign_expiry"`
}

// HybridConfig holds the configuration for hybrid (local + S3) storage.
type HybridConfig struct {
	SyncWorkers    int           `json:"sync_workers" yaml:"sync_workers"`
	SyncQueueSize  int           `json:"sync_queue_size" yaml:"sync_queue_size"`
	LocalCacheSize int64         `json:"local_cache_size" yaml:"local_cache_size"`
	SyncRetryMax   int           `json:"sync_retry_max" yaml:"sync_retry_max"`
	SyncRetryDelay time.Duration `json:"sync_retry_delay" yaml:"sync_retry_delay"`
}

// TempConfig holds the configuration for temp file lifecycle management.
type TempConfig struct {
	TTL             time.Duration `json:"ttl" yaml:"ttl"`
	CleanupInterval time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`
}

// DefaultStorageConfig returns a StorageConfig with sensible defaults.
// The default type is "local" for backward compatibility.
func DefaultStorageConfig() *StorageConfig {
	return &StorageConfig{
		Type:     StorageTypeLocal,
		BasePath: "./data/uploads",
		Temp: TempConfig{
			TTL:             48 * time.Hour,
			CleanupInterval: 1 * time.Hour,
		},
		Hybrid: HybridConfig{
			SyncWorkers:    2,
			SyncQueueSize:  1000,
			LocalCacheSize: 0,
			SyncRetryMax:   3,
			SyncRetryDelay: 30 * time.Second,
		},
	}
}
