/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/conf"
	"origadmin/application/origstudio/internal/features/media/dto"
)

// NewTranscodeHandlerWithConfig creates a new transcode handler with paths from StoragePaths.
// This is the wire-compatible constructor that extracts TaskTimeout from TranscodeConfig.
func NewTranscodeHandlerWithConfig(
	mediaUC *MediaUseCase,
	profileRepo dto.EncodeProfileRepo,
	taskRepo dto.EncodingTaskRepo,
	mediaRepo MediaRepo,
	worker TranscodeWorker,
	publisher message.Publisher,
	logger log.Logger,
	sp *conf.StoragePaths,
	cfg *conf.TranscodeConfig,
	spriteUC *SpriteUseCase,
) *TranscodeHandler {
	return NewTranscodeHandler(
		mediaUC,
		profileRepo,
		taskRepo,
		mediaRepo,
		worker,
		publisher,
		logger,
		sp,
		cfg.TaskTimeout,
		spriteUC,
	)
}
