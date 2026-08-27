/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewMediaUseCase,
	NewUploadUseCaseWithConfig,
	NewSpriteUseCaseWithConfig,
	NewTranscodeHandlerWithConfig,
)

var TranscodeProviderSet = wire.NewSet(
	NewAsynqWorker,
	NewAsynqMonitor,
)
