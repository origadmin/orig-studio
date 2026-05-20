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
func NewStoragePathsFromConfig(cfg *StorageConfig) *StoragePaths {
	return NewStoragePaths(cfg.BasePath)
}

// NewTranscodeConfigFromDefaults creates transcode config from defaults.
func NewTranscodeConfigFromDefaults() *TranscodeConfig {
	return DefaultTranscodeConfig()
}

// NewRateLimiterConfig creates a rate limiter with rpm from settings.
// Upload endpoints are excluded from rate limiting to prevent upload failures.
func NewRateLimiterConfig(settingUC *systembiz.SettingUseCase) (rpm int, excludePrefix string) {
	defaultRPM := 60
	if settingUC != nil {
		if val := settingUC.Get(context.Background(), "api_rate_limit"); val != "" {
			var rpmVal int
			if _, err := fmt.Sscanf(val, "%d", &rpmVal); err == nil && rpmVal > 0 {
				defaultRPM = rpmVal
			}
		}
	}
	return defaultRPM, "/api/v1/uploads"
}
