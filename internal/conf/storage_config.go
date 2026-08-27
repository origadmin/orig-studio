package conf

import (
	"os"
	"strconv"
	"time"
)

type StorageType string

const (
	StorageTypeLocal  StorageType = "local"
	StorageTypeS3     StorageType = "s3"
	StorageTypeHybrid StorageType = "hybrid"
)

type StorageConfig struct {
	Type       StorageType        `json:"type" yaml:"type"`
	BasePath   string             `json:"base_path" yaml:"base_path"`
	CDNBaseURL string             `json:"cdn_base_url" yaml:"cdn_base_url"`
	Overrides  map[string]string  `json:"overrides" yaml:"overrides"`     // per-dir path override: name → absolute path
	Strategies map[string]string  `json:"strategies" yaml:"strategies"`   // impl version per module: name → "v1"|"v2"
	S3         S3Config           `json:"s3" yaml:"s3"`
	Hybrid     HybridConfig       `json:"hybrid" yaml:"hybrid"`
	Temp       TempConfig         `json:"temp" yaml:"temp"`
}

type S3Config struct {
	Endpoint      string        `json:"endpoint" yaml:"endpoint"`
	Region        string        `json:"region" yaml:"region"`
	Bucket        string        `json:"bucket" yaml:"bucket"`
	AccessKey     string        `json:"access_key" yaml:"access_key"`
	SecretKey     string        `json:"secret_key" yaml:"secret_key"`
	UsePathStyle  bool          `json:"use_path_style" yaml:"use_path_style"`
	PresignExpiry time.Duration `json:"presign_expiry" yaml:"presign_expiry"`
}

type HybridConfig struct {
	SyncWorkers    int           `json:"sync_workers" yaml:"sync_workers"`
	SyncQueueSize  int           `json:"sync_queue_size" yaml:"sync_queue_size"`
	LocalCacheSize int64         `json:"local_cache_size" yaml:"local_cache_size"`
	SyncRetryMax   int           `json:"sync_retry_max" yaml:"sync_retry_max"`
	SyncRetryDelay time.Duration `json:"sync_retry_delay" yaml:"sync_retry_delay"`
}

type TempConfig struct {
	TTL             time.Duration `json:"ttl" yaml:"ttl"`
	CleanupInterval time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`
}

func DefaultStorageConfig() *StorageConfig {
	cfg := &StorageConfig{
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

	if v := os.Getenv("STORAGE_BASE_PATH"); v != "" {
		cfg.BasePath = v
	}
	if v := os.Getenv("CDN_BASE_URL"); v != "" {
		cfg.CDNBaseURL = v
	}
	if v := os.Getenv("S3_ENDPOINT"); v != "" {
		cfg.S3.Endpoint = v
		if os.Getenv("STORAGE_TYPE") == "" {
			cfg.Type = StorageTypeS3
		}
	}
	if v := os.Getenv("S3_REGION"); v != "" {
		cfg.S3.Region = v
	}
	if v := os.Getenv("S3_BUCKET"); v != "" {
		cfg.S3.Bucket = v
	}
	if v := os.Getenv("S3_ACCESS_KEY"); v != "" {
		cfg.S3.AccessKey = v
	}
	if v := os.Getenv("S3_SECRET_KEY"); v != "" {
		cfg.S3.SecretKey = v
	}
	if v := os.Getenv("S3_USE_PATH_STYLE"); v != "" {
		cfg.S3.UsePathStyle, _ = strconv.ParseBool(v)
	}
	if v := os.Getenv("STORAGE_TYPE"); v != "" {
		cfg.Type = StorageType(v)
	}

	return cfg
}

func (c *StorageConfig) GetURLPrefix() string {
	if c.CDNBaseURL != "" {
		return c.CDNBaseURL
	}
	return "/files"
}
