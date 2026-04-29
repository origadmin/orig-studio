/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
// Note: NewUploadUseCase and NewSpriteUseCase require hardcoded config values
// (chunkSize, baseDir) and are provided via bridge functions in wire.go.
// NewTranscodeHandler and NewGoroutineWorker also require hardcoded config values
// or return unexported types, so they are also provided via bridge functions.
var ProviderSet = wire.NewSet(
	NewMediaUseCase,
)
