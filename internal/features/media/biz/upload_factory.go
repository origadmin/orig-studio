/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/conf"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
	"origadmin/application/origstudio/internal/features/media/dto"
)

// NewUploadUseCaseWithConfig creates a new upload use case with config from UploadConfig.
// This is the wire-compatible constructor that extracts ChunkSize from UploadConfig.
func NewUploadUseCaseWithConfig(
	uploadRepo UploadRepo,
	mediaRepo MediaRepo,
	profileRepo dto.EncodeProfileRepo,
	taskRepo dto.EncodingTaskRepo,
	mediaUC *MediaUseCase,
	storage Storage,
	sp *conf.StoragePaths,
	cfg *conf.UploadConfig,
	logger log.Logger,
	settingUC *systembiz.SettingUseCase,
) *UploadUseCase {
	return NewUploadUseCase(
		uploadRepo,
		mediaRepo,
		profileRepo,
		taskRepo,
		mediaUC,
		storage,
		sp,
		cfg.ChunkSize,
		logger,
		settingUC,
	)
}
