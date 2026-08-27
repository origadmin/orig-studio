/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package conf

import (
	"context"
	"fmt"

	systembiz "origadmin/application/origstudio/internal/features/system/biz"
)

// NewStorageConfigFromSettings creates storage config from defaults, overridden by settings.
func NewStorageConfigFromSettings(settingUC *systembiz.SettingUseCase) *StorageConfig {
	cfg := DefaultStorageConfig()
	if basePath := settingUC.Get(context.Background(), "storage_base_path"); basePath != "" {
		cfg.BasePath = basePath
	}
	if storageType := settingUC.Get(context.Background(), "storage_type"); storageType != "" {
		cfg.Type = StorageType(storageType)
	}
	if endpoint := settingUC.Get(context.Background(), "s3_endpoint"); endpoint != "" {
		cfg.S3.Endpoint = endpoint
	}
	if region := settingUC.Get(context.Background(), "s3_region"); region != "" {
		cfg.S3.Region = region
	}
	if bucket := settingUC.Get(context.Background(), "s3_bucket"); bucket != "" {
		cfg.S3.Bucket = bucket
	}
	if accessKey := settingUC.Get(context.Background(), "s3_access_key"); accessKey != "" {
		cfg.S3.AccessKey = accessKey
	}
	if secretKey := settingUC.Get(context.Background(), "s3_secret_key"); secretKey != "" {
		cfg.S3.SecretKey = secretKey
	}
	if usePathStyle := settingUC.GetBool(context.Background(), "s3_use_path_style"); usePathStyle {
		cfg.S3.UsePathStyle = true
	}
	return cfg
}

// NewUploadConfigFromDefaults creates upload config from defaults.
func NewUploadConfigFromDefaults() *UploadConfig {
	return DefaultUploadConfig()
}

// NewStoragePathsFromConfig creates StoragePaths from StorageConfig.
// Passes overrides map for per-dir path customization.
func NewStoragePathsFromConfig(cfg *StorageConfig) *StoragePaths {
	return NewStoragePathsWithOverrides(cfg.BasePath, cfg.Overrides)
}

// NewTranscodeConfigFromDefaults creates transcode config from defaults.
func NewTranscodeConfigFromDefaults() *TranscodeConfig {
	return DefaultTranscodeConfig()
}

// NewRateLimiterConfig creates a rate limiter with rpm from settings.
// Static asset paths and health endpoints are excluded from rate limiting
// to prevent chunk load failures (ChunkLoadError) and video streaming interruptions.
func NewRateLimiterConfig(settingUC *systembiz.SettingUseCase) (rpm int, excludePrefixes []string) {
	defaultRPM := 600
	if settingUC != nil {
		if val := settingUC.Get(context.Background(), "api_rate_limit"); val != "" {
			var rpmVal int
			if _, err := fmt.Sscanf(val, "%d", &rpmVal); err == nil && rpmVal > 0 {
				defaultRPM = rpmVal
			}
		}
	}
	return defaultRPM, []string{
		"/api/v1/uploads",
		"/api/v1/admin/medias/transcoding/events",
		"/static",
		"/assets",
		"/locales",
		"/themes",
		"/files",
		"/health",
		"/favicon.ico",
		"/robots.txt",
		"/manifest.json",
		"/logo",
		"/banner",
	}
}
